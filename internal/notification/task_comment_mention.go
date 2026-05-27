package notification

import "github.com/google/uuid"

type TaskCommentMentionNotificationData struct {
	TaskID      uuid.UUID `json:"task_id"`
	CommentID   uuid.UUID `json:"comment_id"`
	AuthorID    uuid.UUID `json:"author_id"`
	MentionedID uuid.UUID `json:"mentioned_id"`
	TaskName    string    `json:"task_name"`
	Excerpt     string    `json:"excerpt"`
}

func (n TaskCommentMentionNotificationData) Kind() string {
	return "task_comment_mention"
}
