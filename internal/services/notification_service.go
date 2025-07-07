package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/stores"
)

type NotificationService interface {
	NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error
}

type DbNotificationService struct {
	adapter stores.StorageAdapterInterface
}

// NotifyMembersOfNewMember implements NotificationService.
// 1. find team member with team and user.
// 2. find all team members of the team.
// 3. send notification to all team members except the team member.
func (d *DbNotificationService) NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID) error {
	panic("unimplemented")
}

var _ NotificationService = (*DbNotificationService)(nil)
