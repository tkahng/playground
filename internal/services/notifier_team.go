package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/notification"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/sse"
)

type Notifier interface {
	NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error
	NotifyAssignedToTask(ctx context.Context, taskID uuid.UUID, assignedByMemberID uuid.UUID, assigneeMemberID uuid.UUID) error
}

type DbNotifier struct {
	sseManager  sse.Manager
	teamService TeamService
	adapter     stores.StorageAdapterInterface
}

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

var _ Notifier = (*DbNotifier)(nil)

func NewDbNotificationPublisher(sseManager sse.Manager, teamService TeamService, adapter stores.StorageAdapterInterface) *DbNotifier {
	return &DbNotifier{
		sseManager:  sseManager,
		teamService: teamService,
		adapter:     adapter,
	}
}
