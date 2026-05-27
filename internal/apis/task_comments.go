package apis

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/workers"
)

type TaskComment struct {
	ID                uuid.UUID   `json:"id"`
	TaskID            uuid.UUID   `json:"task_id"`
	CreatedByMemberID uuid.UUID   `json:"created_by_member_id"`
	Content           string      `json:"content"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	CreatedByMember   *TeamMember `json:"created_by_member,omitempty"`
}

func fromModelTaskComment(c *models.TaskComment) *TaskComment {
	if c == nil {
		return nil
	}
	return &TaskComment{
		ID:                c.ID,
		TaskID:            c.TaskID,
		CreatedByMemberID: c.CreatedByMemberID,
		Content:           c.Content,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
		CreatedByMember:   fromTeamMemberModel(c.CreatedByMember),
	}
}

type TaskCommentListInput struct {
	TaskID string `path:"task-id" format:"uuid" required:"true"`
}

type TaskCommentListOutput struct {
	Body []*TaskComment
}

func (api *Api) TaskCommentList(ctx context.Context, input *TaskCommentListInput) (*TaskCommentListOutput, error) {
	taskID, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid task ID")
	}
	comments, err := api.App().Adapter().TaskComment().ListTaskComments(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return &TaskCommentListOutput{
		Body: mapper.Map(comments, fromModelTaskComment),
	}, nil
}

type CreateTaskCommentBody struct {
	Content string `json:"content" required:"true" minLength:"1" maxLength:"10000"`
}

type TaskCommentCreateInput struct {
	TaskID string                `path:"task-id" format:"uuid" required:"true"`
	Body   CreateTaskCommentBody
}

type TaskCommentCreateOutput struct {
	Body *TaskComment
}

func (api *Api) TaskCommentCreate(ctx context.Context, input *TaskCommentCreateInput) (*TaskCommentCreateOutput, error) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("team info not found")
	}
	taskID, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid task ID")
	}

	comment := &models.TaskComment{
		TaskID:            taskID,
		CreatedByMemberID: teamInfo.Member.ID,
		Content:           input.Body.Content,
	}
	created, err := api.App().Adapter().TaskComment().CreateTaskComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Reload with relations
	created, err = api.App().Adapter().TaskComment().FindTaskCommentByID(ctx, created.ID)
	if err != nil {
		return nil, err
	}

	mentionedIDs := models.ParseMentionedMemberIDs(input.Body.Content)

	if err := api.App().JobService().EnqueueTaskCommentCreatedJob(ctx, &workers.TaskCommentCreatedJobArgs{
		TaskID:    taskID,
		CommentID: created.ID,
		AuthorID:  teamInfo.Member.ID,
	}); err != nil {
		return nil, err
	}

	for _, mentionedID := range mentionedIDs {
		if err := api.App().JobService().EnqueueTaskCommentMentionJob(ctx, &workers.TaskCommentMentionJobArgs{
			TaskID:      taskID,
			CommentID:   created.ID,
			AuthorID:    teamInfo.Member.ID,
			MentionedID: mentionedID,
		}); err != nil {
			return nil, err
		}
	}

	return &TaskCommentCreateOutput{Body: fromModelTaskComment(created)}, nil
}

type UpdateTaskCommentBody struct {
	Content string `json:"content" required:"true" minLength:"1" maxLength:"10000"`
}

type TaskCommentUpdateInput struct {
	TaskID    string                `path:"task-id" format:"uuid" required:"true"`
	CommentID string                `path:"comment-id" format:"uuid" required:"true"`
	Body      UpdateTaskCommentBody
}

type TaskCommentUpdateOutput struct {
	Body *TaskComment
}

func (api *Api) TaskCommentUpdate(ctx context.Context, input *TaskCommentUpdateInput) (*TaskCommentUpdateOutput, error) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("team info not found")
	}
	commentID, err := uuid.Parse(input.CommentID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid comment ID")
	}

	comment, err := api.App().Adapter().TaskComment().FindTaskCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, huma.Error404NotFound("comment not found")
	}
	if comment.CreatedByMemberID != teamInfo.Member.ID {
		return nil, huma.Error403Forbidden("only the author can edit this comment")
	}

	comment.Content = input.Body.Content
	updated, err := api.App().Adapter().TaskComment().UpdateTaskComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Reload with relations
	updated, err = api.App().Adapter().TaskComment().FindTaskCommentByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}
	return &TaskCommentUpdateOutput{Body: fromModelTaskComment(updated)}, nil
}

type TaskCommentDeleteInput struct {
	TaskID    string `path:"task-id" format:"uuid" required:"true"`
	CommentID string `path:"comment-id" format:"uuid" required:"true"`
}

func (api *Api) TaskCommentDelete(ctx context.Context, input *TaskCommentDeleteInput) (*struct{}, error) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("team info not found")
	}
	commentID, err := uuid.Parse(input.CommentID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid comment ID")
	}

	comment, err := api.App().Adapter().TaskComment().FindTaskCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, huma.Error404NotFound("comment not found")
	}

	isAuthor := comment.CreatedByMemberID == teamInfo.Member.ID
	isOwner := teamInfo.Member.Role == models.TeamMemberRoleOwner
	if !isAuthor && !isOwner {
		return nil, huma.Error403Forbidden("not authorized to delete this comment")
	}

	if err := api.App().Adapter().TaskComment().DeleteTaskComment(ctx, commentID); err != nil {
		return nil, err
	}
	return nil, nil
}
