package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/workers"
)

type Notifier interface {
	NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error
	NotifyAssignedToTask(ctx context.Context, taskID uuid.UUID, assignedByMemberID uuid.UUID, assigneeMemberID uuid.UUID) error

	NotifyTaskDueToday(ctx context.Context, taskID uuid.UUID) error
	NotifyTaskCompleted(ctx context.Context, taskID uuid.UUID, completedByMemberID uuid.UUID, completedAt time.Time) error
	NotifyTaskOverdue(ctx context.Context, taskID uuid.UUID) error
	NotifyTaskStatusChanged(ctx context.Context, taskID uuid.UUID, oldStatus string, newStatus string, changedByMemberID uuid.UUID) error
	NotifyProjectStatusChanged(ctx context.Context, projectID uuid.UUID, oldStatus string, newStatus string, changedByMemberID uuid.UUID) error
}

var _ Notifier = (*DbNotifier)(nil)

func NewDbNotificationPublisher(sseManager sse.Manager, teamService TeamService, adapter stores.StorageAdapterInterface) *DbNotifier {
	return &DbNotifier{
		sseManager:  sseManager,
		teamService: teamService,
		adapter:     adapter,
	}
}

type DbNotifier struct {
	sseManager  sse.Manager
	teamService TeamService
	adapter     stores.StorageAdapterInterface
}

// collectMemberIDs returns unique non-nil UUIDs from nullable pointers.
func collectMemberIDs(ids ...*uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == nil {
			continue
		}
		if _, ok := seen[*id]; !ok {
			seen[*id] = struct{}{}
			result = append(result, *id)
		}
	}
	return result
}

// sendToMembers persists notifications for each member and broadcasts via SSE.
func (d *DbNotifier) sendToMembers(ctx context.Context, memberIDs []uuid.UUID, notifType string, payloadBytes []byte, ssePayload any) error {
	if len(memberIDs) == 0 {
		return nil
	}
	members, err := d.adapter.TeamMember().FindTeamMembers(ctx, &stores.TeamMemberFilter{Ids: memberIDs})
	if err != nil {
		return err
	}
	notifications := make([]models.Notification, 0, len(members))
	for _, member := range members {
		notifications = append(notifications, models.Notification{
			TeamMemberID: &member.ID,
			Channel:      "team_member_id:" + member.ID.String(),
			Type:         notifType,
			Payload:      payloadBytes,
			Metadata:     map[string]any{},
		})
	}
	if _, err = d.adapter.Notification().InsertManyNotifications(ctx, notifications); err != nil {
		return err
	}
	for _, n := range notifications {
		memberID := *n.TeamMemberID
		if sendErr := d.sseManager.Send("team_member_id:"+memberID.String(), ssePayload); sendErr != nil {
			slog.ErrorContext(ctx, "error sending notification", slog.Any("error", sendErr))
		}
	}
	return nil
}

// NotifyAssignedToTask notifies the assignee that they have been assigned to a task.
func (d *DbNotifier) NotifyAssignedToTask(ctx context.Context, taskID uuid.UUID, assignedByMemberID uuid.UUID, assigneeMemberID uuid.UUID) error {
	assigner, err := d.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{Ids: []uuid.UUID{assignedByMemberID}})
	if err != nil {
		return err
	}
	if assigner == nil {
		return errors.New("assigner not found")
	}
	if assigner.UserID == nil {
		return errors.New("assigner has no user")
	}
	assignerUser, err := d.adapter.User().FindUser(ctx, &stores.UserFilter{Ids: []uuid.UUID{*assigner.UserID}})
	if err != nil {
		return err
	}
	if assignerUser == nil {
		return errors.New("assigner user not found")
	}
	task, err := d.adapter.Task().FindTask(ctx, &stores.TaskFilter{Ids: []uuid.UUID{taskID}})
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	payload := notification.AssignedToTaskNotificationData{
		AssignedByMemberID: assigner.ID,
		AssigneeMemberID:   assigneeMemberID,
		TaskID:             task.ID,
	}
	notificationPayload := notification.NewNotificationPayload(
		"You have been assigned to a task.",
		assignerUser.Email+" has assigned you to a task.",
		payload,
	)
	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		return err
	}
	return d.sendToMembers(ctx, []uuid.UUID{assigneeMemberID}, payload.Kind(), payloadBytes, notificationPayload)
}

// NotifyMembersOfNewMember notifies all existing team members of a new member joining.
func (d *DbNotifier) NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error {
	newMember, err := d.teamService.FindTeamMemberWithUserAndTeam(ctx, teamMemberID)
	if err != nil {
		return err
	}
	if newMember == nil {
		return nil
	}
	members, err := d.adapter.TeamMember().FindTeamMembers(ctx, &stores.TeamMemberFilter{TeamIds: []uuid.UUID{newMember.Team.ID}})
	if err != nil {
		return err
	}
	payload := notification.NewTeamMemberNotificationData{
		TeamMemberID: teamMemberID,
		TeamID:       newMember.Team.ID,
		Email:        newMember.User.Email,
	}
	notificationPayload := notification.NewNotificationPayload(
		"New member joined your team.",
		payload.Email+" has joined your team.",
		payload,
	)
	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		return err
	}
	recipientIDs := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		if m.ID != teamMemberID {
			recipientIDs = append(recipientIDs, m.ID)
		}
	}
	return d.sendToMembers(ctx, recipientIDs, payload.Kind(), payloadBytes, notificationPayload)
}

// NotifyTaskDueToday notifies task stakeholders that the task is due today.
func (d *DbNotifier) NotifyTaskDueToday(ctx context.Context, taskID uuid.UUID) error {
	task, err := d.adapter.Task().FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	if !isWithinPastHours(task.EndAt, 24*time.Hour) {
		slog.DebugContext(ctx, "task is not due today", slog.String("task_id", taskID.String()))
		return nil
	}
	payload := notification.TaskDueTodayNotificationData{
		TaskID:  task.ID,
		DueDate: *task.EndAt,
	}
	notificationPayload := notification.NewNotificationPayload(
		"There is a task due today.",
		task.Name+" is due today.",
		payload,
	)
	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		return err
	}
	memberIDs := collectMemberIDs(task.AssigneeID, task.ReporterID, task.CreatedByMemberID)
	if len(memberIDs) == 0 {
		slog.DebugContext(ctx, "no members to notify", slog.String("task_id", taskID.String()))
		return nil
	}
	return d.sendToMembers(ctx, memberIDs, payload.Kind(), payloadBytes, notificationPayload)
}

// NotifyTaskCompleted notifies task stakeholders that the task has been completed.
func (d *DbNotifier) NotifyTaskCompleted(ctx context.Context, taskID uuid.UUID, completedByMemberID uuid.UUID, completedAt time.Time) error {
	task, err := d.adapter.Task().FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	if task.Status != models.TaskStatusDone {
		return errors.New("task is not completed")
	}
	payload := notification.TaskCompletedNotificationData{
		TaskID:              taskID,
		CompletedByMemberID: completedByMemberID,
		CompletedAt:         completedAt,
	}
	notificationPayload := notification.NewNotificationPayload(
		"Task completed.",
		task.Name+" was completed today.",
		payload,
	)
	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		return err
	}
	memberIDs := collectMemberIDs(task.AssigneeID, task.ReporterID, task.CreatedByMemberID)
	if len(memberIDs) == 0 {
		slog.DebugContext(ctx, "no members to notify", slog.String("task_id", taskID.String()))
		return nil
	}
	return d.sendToMembers(ctx, memberIDs, payload.Kind(), payloadBytes, notificationPayload)
}

// NotifyTaskOverdue notifies task stakeholders that the task is overdue.
func (d *DbNotifier) NotifyTaskOverdue(ctx context.Context, taskID uuid.UUID) error {
	task, err := d.adapter.Task().FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	if task.Status == models.TaskStatusDone {
		slog.DebugContext(ctx, "task already done, skipping overdue notification", slog.String("task_id", taskID.String()))
		return nil
	}
	if task.EndAt == nil || !time.Now().After(*task.EndAt) {
		slog.DebugContext(ctx, "task not overdue", slog.String("task_id", taskID.String()))
		return nil
	}
	payload := notification.TaskOverdueNotificationData{
		TaskID:  task.ID,
		DueDate: *task.EndAt,
	}
	notificationPayload := notification.NewNotificationPayload(
		"Task is overdue.",
		task.Name+" is overdue.",
		payload,
	)
	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		return err
	}
	memberIDs := collectMemberIDs(task.AssigneeID, task.ReporterID, task.CreatedByMemberID)
	if len(memberIDs) == 0 {
		slog.DebugContext(ctx, "no members to notify for overdue task", slog.String("task_id", taskID.String()))
		return nil
	}
	return d.sendToMembers(ctx, memberIDs, payload.Kind(), payloadBytes, notificationPayload)
}

// NotifyTaskStatusChanged notifies task stakeholders of a status transition (non-done).
func (d *DbNotifier) NotifyTaskStatusChanged(ctx context.Context, taskID uuid.UUID, oldStatus string, newStatus string, changedByMemberID uuid.UUID) error {
	task, err := d.adapter.Task().FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	payload := notification.TaskStatusChangedNotificationData{
		TaskID:            task.ID,
		OldStatus:         oldStatus,
		NewStatus:         newStatus,
		ChangedByMemberID: changedByMemberID,
	}
	notificationPayload := notification.NewNotificationPayload(
		"Task status changed.",
		task.Name+" status changed from "+oldStatus+" to "+newStatus+".",
		payload,
	)
	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		return err
	}
	memberIDs := collectMemberIDs(task.AssigneeID, task.ReporterID, task.CreatedByMemberID)
	if len(memberIDs) == 0 {
		slog.DebugContext(ctx, "no members to notify for status change", slog.String("task_id", taskID.String()))
		return nil
	}
	return d.sendToMembers(ctx, memberIDs, payload.Kind(), payloadBytes, notificationPayload)
}

// NotifyProjectStatusChanged notifies project stakeholders of a status transition.
func (d *DbNotifier) NotifyProjectStatusChanged(ctx context.Context, projectID uuid.UUID, oldStatus string, newStatus string, changedByMemberID uuid.UUID) error {
	project, err := d.adapter.Task().FindTaskProjectByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.New("project not found")
	}
	payload := notification.ProjectStatusChangedNotificationData{
		ProjectID:         project.ID,
		OldStatus:         oldStatus,
		NewStatus:         newStatus,
		ChangedByMemberID: changedByMemberID,
	}
	notificationPayload := notification.NewNotificationPayload(
		"Project status changed.",
		project.Name+" status changed from "+oldStatus+" to "+newStatus+".",
		payload,
	)
	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		return err
	}
	memberIDs := collectMemberIDs(project.AssigneeID, project.ReporterID, project.CreatedByMemberID)
	if len(memberIDs) == 0 {
		slog.DebugContext(ctx, "no members to notify for project status change", slog.String("project_id", projectID.String()))
		return nil
	}
	return d.sendToMembers(ctx, memberIDs, payload.Kind(), payloadBytes, notificationPayload)
}

// workers — bridge job args to Notifier methods.

type AssignedToTaskWorker struct {
	notifier Notifier
}

func (a *AssignedToTaskWorker) Work(ctx context.Context, job *jobs.Job[workers.AssignedToTaskJobArgs]) error {
	return a.notifier.NotifyAssignedToTask(ctx, job.Args.TaskID, job.Args.AssignedByMemberID, job.Args.AssigneeMemberID)
}

func NewAssignedToTaskWorker(notifier Notifier) *AssignedToTaskWorker {
	return &AssignedToTaskWorker{notifier: notifier}
}

var _ jobs.Worker[workers.AssignedToTaskJobArgs] = (*AssignedToTaskWorker)(nil)

type TaskDueTodayWorker struct {
	notifier Notifier
}

func (a *TaskDueTodayWorker) Work(ctx context.Context, job *jobs.Job[workers.TaskDueTodayJobArgs]) error {
	return a.notifier.NotifyTaskDueToday(ctx, job.Args.TaskID)
}

func NewTaskDueTodayWorker(notifier Notifier) *TaskDueTodayWorker {
	return &TaskDueTodayWorker{notifier: notifier}
}

var _ jobs.Worker[workers.TaskDueTodayJobArgs] = (*TaskDueTodayWorker)(nil)

type TaskCompletedWorker struct {
	notifier Notifier
}

func (a *TaskCompletedWorker) Work(ctx context.Context, job *jobs.Job[workers.TaskCompletedJobArgs]) error {
	return a.notifier.NotifyTaskCompleted(ctx, job.Args.TaskID, job.Args.CompletedByMemberID, job.Args.CompletedAt)
}

func NewTaskCompletedWorker(notifier Notifier) *TaskCompletedWorker {
	return &TaskCompletedWorker{notifier: notifier}
}

var _ jobs.Worker[workers.TaskCompletedJobArgs] = (*TaskCompletedWorker)(nil)

type TaskOverdueWorker struct {
	notifier Notifier
}

func (a *TaskOverdueWorker) Work(ctx context.Context, job *jobs.Job[workers.TaskOverdueJobArgs]) error {
	return a.notifier.NotifyTaskOverdue(ctx, job.Args.TaskID)
}

func NewTaskOverdueWorker(notifier Notifier) *TaskOverdueWorker {
	return &TaskOverdueWorker{notifier: notifier}
}

var _ jobs.Worker[workers.TaskOverdueJobArgs] = (*TaskOverdueWorker)(nil)

type TaskStatusChangedWorker struct {
	notifier Notifier
}

func (a *TaskStatusChangedWorker) Work(ctx context.Context, job *jobs.Job[workers.TaskStatusChangedJobArgs]) error {
	return a.notifier.NotifyTaskStatusChanged(ctx, job.Args.TaskID, job.Args.OldStatus, job.Args.NewStatus, job.Args.ChangedByMemberID)
}

func NewTaskStatusChangedWorker(notifier Notifier) *TaskStatusChangedWorker {
	return &TaskStatusChangedWorker{notifier: notifier}
}

var _ jobs.Worker[workers.TaskStatusChangedJobArgs] = (*TaskStatusChangedWorker)(nil)

type ProjectStatusChangedWorker struct {
	notifier Notifier
}

func (a *ProjectStatusChangedWorker) Work(ctx context.Context, job *jobs.Job[workers.ProjectStatusChangedJobArgs]) error {
	return a.notifier.NotifyProjectStatusChanged(ctx, job.Args.ProjectID, job.Args.OldStatus, job.Args.NewStatus, job.Args.ChangedByMemberID)
}

func NewProjectStatusChangedWorker(notifier Notifier) *ProjectStatusChangedWorker {
	return &ProjectStatusChangedWorker{notifier: notifier}
}

var _ jobs.Worker[workers.ProjectStatusChangedJobArgs] = (*ProjectStatusChangedWorker)(nil)

func isWithinPastHours(t *time.Time, dur time.Duration) bool {
	if t == nil {
		return false
	}
	diff := time.Since(*t)
	return diff > 0 && diff <= dur
}
