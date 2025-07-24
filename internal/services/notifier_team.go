package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type AssignedToTaskWorker struct {
	notifier Notifier
}

// Work implements workers.AssignedToTaskWorker.
func (a *AssignedToTaskWorker) Work(ctx context.Context, job *jobs.Job[workers.AssignedToTasJobArgs]) error {
	return a.notifier.NotifyAssignedToTask(ctx, job.Args.TaskID, job.Args.AssignedByMemeberID, job.Args.AssigneeMemberID)
}

func NewAssignedToTaskWorker(notifier Notifier) *AssignedToTaskWorker {
	return &AssignedToTaskWorker{
		notifier: notifier,
	}
}

var _ workers.AssignedToTaskWorker = (*AssignedToTaskWorker)(nil)

// NotifyAssignedToTask implements Notifier.
// 1. find assignee
// 2. find task assigned
// 3. create notification
// 4. send notification
func (d *DbNotifier) NotifyAssignedToTask(ctx context.Context, taskID uuid.UUID, assignedByMemberID uuid.UUID, assigneeMemberID uuid.UUID) error {
	assigneeMember, err := d.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		Ids: []uuid.UUID{
			assigneeMemberID,
		},
	})
	if err != nil {
		return err
	}
	if assigneeMember == nil {
		return errors.New("assignee member not found")
	}
	// 1. find assigner
	assigner, err := d.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		Ids: []uuid.UUID{
			assignedByMemberID,
		},
	})
	if err != nil {
		return err
	}
	if assigner == nil {
		return errors.New("assignee not found")
	}
	if assigner.UserID == nil {
		return errors.New("user id not found")
	}
	assignerUser, err := d.adapter.User().FindUser(ctx, &stores.UserFilter{
		Ids: []uuid.UUID{
			*assigner.UserID,
		},
	})
	if err != nil {
		return err
	}
	if assignerUser == nil {
		return errors.New("assigned user not found")
	}
	// 2. find task assigned
	task, err := d.adapter.Task().FindTask(ctx, &stores.TaskFilter{
		Ids: []uuid.UUID{
			taskID,
		},
	})
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	// 3. create notification
	payload := notification.AssignedToTaskNotificationData{
		AssignedByMemeberID: assigner.ID,
		TaskID:              task.ID,
	}
	// 3. send notification to all team members
	notifcationPaylod := notification.NewNotificationPayload(
		"You have been assigned to a task.",
		assignerUser.Email+" has assigned you to a task.",
		payload,
	)
	notificationPayloadBytes, err := json.Marshal(notifcationPaylod)
	if err != nil {
		return err
	}
	_, err = d.adapter.Notification().CreateNotification(ctx, &models.Notification{
		TeamMemberID: &assigneeMember.ID,
		Channel:      "team_member_id:" + assigneeMember.ID.String(),
		Type:         payload.Kind(),
		Payload:      notificationPayloadBytes,
		Metadata:     map[string]any{},
	})
	if err != nil {
		return err
	}
	err = d.sseManager.Send(
		"team_member_id:"+assigneeMember.ID.String(),
		notifcationPaylod,
	)
	if err != nil {
		return err
	}
	return nil
}

type NewTeamMemberWorker struct {
	notifier *DbNotifier
}

// Work implements workers.NewTeamMemberWorker.
func (a *NewTeamMemberWorker) Work(ctx context.Context, job *jobs.Job[workers.NewMemberNotificationJobArgs]) error {
	return a.notifier.NotifyMembersOfNewMember(ctx, job.Args.TeamMemberID)
}

var _ jobs.Worker[workers.NewMemberNotificationJobArgs] = (*NewTeamMemberWorker)(nil)

func NewNewTeamMemberWorker(notifier *DbNotifier) *NewTeamMemberWorker {
	return &NewTeamMemberWorker{
		notifier: notifier,
	}
}

// NotifyMembersOfNewMember implements NotificationService.
// 1. find team member with team and user.
// 2. find all team members of the team.
// 3. send notification to all team members except the team member.
func (d *DbNotifier) NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error {
	// 1. find team member with team and user
	newMember, err := d.teamService.FindTeamInfoByMemberID(ctx, teamMemberID)
	if err != nil {
		return err
	}
	if newMember == nil {
		return nil
	}
	// 2. find all team members of the team
	members, err := d.adapter.TeamMember().FindTeamMembers(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{
			newMember.Team.ID,
		},
	})
	if err != nil {
		return err
	}
	payload := notification.NewTeamMemberNotificationData{
		TeamMemberID: teamMemberID,
		TeamID:       newMember.Team.ID,
		Email:        newMember.User.Email,
	}
	// 3. send notification to all team members
	notifcationPaylod := notification.NewNotificationPayload(
		"New member joined your team.",
		payload.Email+" has joined your team.",
		payload,
	)
	notificationPayloadBytes, err := json.Marshal(notifcationPaylod)
	if err != nil {
		return err
	}
	var notifications []models.Notification
	for _, member := range members {
		if member.ID == teamMemberID {
			continue
		}
		notification := models.Notification{
			TeamMemberID: &member.ID,
			Channel:      "team_member_id:" + member.ID.String(),
			Type:         payload.Kind(),
			Payload:      notificationPayloadBytes,
			Metadata:     map[string]any{},
		}
		notifications = append(notifications, notification)
	}

	_, err = d.adapter.Notification().InsertManyNotifications(ctx, notifications)
	if err != nil {
		return err
	}
	for _, notification := range notifications {
		if notification.TeamMemberID == nil {
			continue
		}
		teamMemberID := *notification.TeamMemberID
		err = d.sseManager.Send("team_member_id:"+teamMemberID.String(), notifcationPaylod)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"error sending notification",
				slog.Any("error", err),
			)
		}
	}
	return nil
}

// func isWithinLast24Hours(t *time.Time) bool {
// 	if t == nil {
// 		return false
// 	}
// 	return t.After(time.Now().Add(-24 * time.Hour))
// }

func isWithinPastHours(t *time.Time, dur time.Duration) bool {
	if t == nil {
		return false
	}
	now := time.Now()
	// Calculate the duration between 'now' and 't'
	diff := now.Sub(*t)

	// Define the 24-hour duration

	// Check if the difference is positive (t is in the past)
	// and if the difference is less than or equal to 24 hours
	return diff > 0 && diff <= dur
}

type TaskDueTodayWorker struct {
	notifier Notifier
}

// Work implements workers.TaskDueTodayWorker.
func (a *TaskDueTodayWorker) Work(ctx context.Context, job *jobs.Job[workers.TaskDueTodayJobArgs]) error {
	return a.notifier.NotifyTaskDueToday(ctx, job.Args.TaskID)
}

func NewTaskDueTodayWorker(notifier Notifier) *TaskDueTodayWorker {
	return &TaskDueTodayWorker{
		notifier: notifier,
	}
}

var _ jobs.Worker[workers.TaskDueTodayJobArgs] = (*TaskDueTodayWorker)(nil)

// NotifyTaskDueToday implements Notifier.
//  1. find task
//  2. check task end at is now
//  3. if so, create notification
//  4. else, do nothing
func (d *DbNotifier) NotifyTaskDueToday(ctx context.Context, taskID uuid.UUID) error {
	// 1. find task
	task, err := d.adapter.Task().FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	// 2. check task end at is now
	taskEndAtIsNow := isWithinPastHours(task.EndAt, 24*time.Hour)

	if taskEndAtIsNow {
		fmt.Println("task due today")
		payload := notification.TaskDueTodayNotificationData{
			TaskID:  task.ID,
			DueDate: *task.EndAt,
		}
		// 3. send notification to all team members
		notifcationPaylod := notification.NewNotificationPayload(
			"There is a task due today.",
			task.Name+" is due today.",
			payload,
		)
		notificationPayloadBytes, err := json.Marshal(notifcationPaylod)
		if err != nil {
			return err
		}
		var notifyMemberIds []uuid.UUID
		if task.AssigneeID != nil {
			notifyMemberIds = append(notifyMemberIds, *task.AssigneeID)
		}
		if task.ReporterID != nil {
			notifyMemberIds = append(notifyMemberIds, *task.ReporterID)
		}
		if task.CreatedByMemberID != nil {
			notifyMemberIds = append(notifyMemberIds, *task.CreatedByMemberID)
		}
		if len(notifyMemberIds) == 0 {
			fmt.Println("no members to notify")
			return nil
		}
		notifyMembers, err := d.adapter.TeamMember().FindTeamMembers(ctx, &stores.TeamMemberFilter{
			Ids: notifyMemberIds,
		})
		if err != nil {
			return err
		}
		var notifications []models.Notification
		for _, member := range notifyMembers {
			notification := models.Notification{
				TeamMemberID: &member.ID,
				Channel:      "team_member_id:" + member.ID.String(),
				Type:         payload.Kind(),
				Payload:      notificationPayloadBytes,
				Metadata:     map[string]any{},
			}
			notifications = append(notifications, notification)
		}

		_, err = d.adapter.Notification().InsertManyNotifications(ctx, notifications)
		if err != nil {
			return err
		}
		for _, notification := range notifications {
			teamMemberID := *notification.TeamMemberID
			err = d.sseManager.Send("team_member_id:"+teamMemberID.String(), notifcationPaylod)
			if err != nil {
				slog.ErrorContext(
					ctx,
					"error sending notification",
					slog.Any("error", err),
				)
			}
		}
		if err != nil {
			return err
		}
	} else {
		fmt.Println("task is not due today")
		return nil
	}
	return nil
}
