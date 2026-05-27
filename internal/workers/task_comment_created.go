package workers

import "github.com/google/uuid"

type TaskCommentCreatedJobArgs struct {
	TaskID    uuid.UUID `json:"task_id"`
	CommentID uuid.UUID `json:"comment_id"`
	AuthorID  uuid.UUID `json:"author_id"`
}

func (a TaskCommentCreatedJobArgs) Kind() string {
	return "task_comment_created"
}
