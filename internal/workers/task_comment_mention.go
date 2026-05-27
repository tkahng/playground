package workers

import "github.com/google/uuid"

type TaskCommentMentionJobArgs struct {
	TaskID      uuid.UUID `json:"task_id"`
	CommentID   uuid.UUID `json:"comment_id"`
	AuthorID    uuid.UUID `json:"author_id"`
	MentionedID uuid.UUID `json:"mentioned_id"`
}

func (a TaskCommentMentionJobArgs) Kind() string {
	return "task_comment_mention"
}
