package notification

import "github.com/google/uuid"

type NewTeamMemberNotificationData struct {
	TeamMemberID uuid.UUID `json:"team_member_id" required:"true"`
	TeamID       uuid.UUID `json:"team_id" required:"true"`
	Email        string    `json:"email" required:"true"`
}

func (n NewTeamMemberNotificationData) Kind() string {
	return "new_team_member"
}

type AssignedToTaskNotificationData struct {
	AssignedByMemeberID uuid.UUID `json:"assigned_by_member_id" required:"true"`
	AssigneeMemberID    uuid.UUID `json:"assignee_member_id" required:"true"`
	TaskID              uuid.UUID `json:"task_id" required:"true"`
}

func (n AssignedToTaskNotificationData) Kind() string {
	return "assigned_to_task"
}
