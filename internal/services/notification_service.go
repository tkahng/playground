package services

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/notification"
	"github.com/tkahng/authgo/internal/stores"
)

type NotificationService interface {
	NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error
}

type DbNotificationService struct {
	teamService TeamService
	adapter     stores.StorageAdapterInterface
}

// NotifyMembersOfNewMember implements NotificationService.
// 1. find team member with team and user.
// 2. find all team members of the team.
// 3. send notification to all team members except the team member.
func (d *DbNotificationService) NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error {
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
		}
		notifications = append(notifications, notification)
	}

	_, err = d.adapter.Notification().InsertManyNotifications(ctx, notifications)
	if err != nil {
		return err
	}
	return nil
}

var _ NotificationService = (*DbNotificationService)(nil)
