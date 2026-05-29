package notification

const (
	TypeAssignedToTask       = "assigned_to_task"
	TypeTaskCompleted        = "task_completed"
	TypeTaskDueToday         = "task_due_today"
	TypeTaskOverdue          = "task_overdue"
	TypeTaskStatusChanged    = "task_status_changed"
	TypeProjectStatusChanged = "project_status_changed"
	TypeNewTeamMember        = "new_team_member"
	TypeTaskCommentCreated   = "task_comment_created"
	TypeTaskCommentMention   = "task_comment_mention"
)

// teamNotificationTypes is the authoritative set of valid team notification type strings.
var teamNotificationTypes = map[string]struct{}{
	TypeAssignedToTask:       {},
	TypeTaskCompleted:        {},
	TypeTaskDueToday:         {},
	TypeTaskOverdue:          {},
	TypeTaskStatusChanged:    {},
	TypeProjectStatusChanged: {},
	TypeNewTeamMember:        {},
	TypeTaskCommentCreated:   {},
	TypeTaskCommentMention:   {},
}

// IsValidTeamNotificationType reports whether t is a known team notification type.
func IsValidTeamNotificationType(t string) bool {
	_, ok := teamNotificationTypes[t]
	return ok
}

// TeamNotificationTypeList returns all valid team notification types in sorted order.
func TeamNotificationTypeList() []string {
	list := make([]string, 0, len(teamNotificationTypes))
	for k := range teamNotificationTypes {
		list = append(list, k)
	}
	return list
}
