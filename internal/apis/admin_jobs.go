package apis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

// 'pending', 'processing', 'done', 'failed'
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	_           struct{}  `db:"jobs" json:"-"`
	ID          uuid.UUID `db:"id" json:"id"`
	Kind        string    `db:"kind" json:"kind"`
	UniqueKey   *string   `db:"unique_key" json:"unique_key"`
	Payload     []byte    `db:"payload" json:"payload"`
	Status      JobStatus `db:"status" json:"status"`
	RunAfter    time.Time `db:"run_after" json:"run_after"`
	Attempts    int64     `db:"attempts" json:"attempts"`
	MaxAttempts int64     `db:"max_attempts" json:"max_attempts"`
	LastError   *string   `db:"last_error" json:"last_error"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

func ToJob(j *models.JobRow) *Job {
	if j == nil {
		return nil
	}
	return &Job{
		ID:          j.ID,
		Kind:        j.Kind,
		UniqueKey:   j.UniqueKey,
		Payload:     j.Payload,
		Status:      JobStatus(j.Status),
		RunAfter:    j.RunAfter,
		Attempts:    j.Attempts,
		MaxAttempts: j.MaxAttempts,
		LastError:   j.LastError,
		CreatedAt:   j.CreatedAt,
		UpdatedAt:   j.UpdatedAt,
	}
}

type JobFilter struct {
	PaginatedInput
	SortParams
	Ids        []string                       `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Kinds      []string                       `db:"kinds" json:"kinds" query:"kinds" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	UniqueKeys []string                       `db:"unique_keys" json:"unique_keys" query:"unique_keys" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Statuses   []JobStatus                    `db:"statuses" json:"statuses" query:"statuses" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	RunAfter   types.OptionalParam[time.Time] `db:"run_after" json:"run_after" query:"run_after" required:"false"`
	Attempt    types.OptionalParam[int64]     `db:"attempt" json:"attempt" query:"attempt" required:"false"`
	LastErrors []string                       `db:"last_errors" json:"last_errors" query:"last_errors" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
}

func (api *Api) AdminGetJobs(
	ctx context.Context,
	input *JobFilter,
) (*ApiPaginatedResponse[*Job], error) {
	filter := &stores.JobFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy, filter.SortOrder = input.Sort()
	filter.Ids = utils.ParseValidUUIDs(input.Ids...)
	filter.Kinds = input.Kinds
	filter.UniqueKeys = input.UniqueKeys
	filter.Statuses = mapper.Map(input.Statuses, func(s JobStatus) models.JobStatus {
		return models.JobStatus(s)
	})
	filter.RunAfter = input.RunAfter
	filter.Attempt = input.Attempt
	filter.LastErrors = input.LastErrors

	jobs, err := api.app.Adapter().Job().FindJobs(ctx, filter)
	if err != nil {
		return nil, err
	}
	count, err := api.app.Adapter().Job().CountJobs(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedResponse[*Job]{
		Data: mapper.Map(jobs, ToJob),
		Meta: ApiGenerateMeta(&input.PaginatedInput, count),
	}, nil
}

type FindJobInput struct {
	ID string `path:"job-id" required:"true" format:"uuid"`
}

func (api *Api) AdminGetJob(
	ctx context.Context,
	input *FindJobInput,
) (*Job, error) {
	id := uuid.MustParse(input.ID)
	j, err := api.app.Adapter().Job().FindJob(ctx, &stores.JobFilter{
		Ids: []uuid.UUID{id},
	})
	if err != nil {
		return nil, err
	}
	return ToJob(j), nil
}

type JobUpdateDto struct {
	Kind        string    `db:"kind" json:"kind"`
	UniqueKey   *string   `db:"unique_key" json:"unique_key"`
	Payload     []byte    `db:"payload" json:"payload"`
	Status      JobStatus `db:"status" json:"status"`
	RunAfter    time.Time `db:"run_after" json:"run_after"`
	Attempts    int64     `db:"attempts" json:"attempts"`
	MaxAttempts int64     `db:"max_attempts" json:"max_attempts"`
	LastError   *string   `db:"last_error" json:"last_error"`
}

func (api *Api) AdminUpdateJob(
	ctx context.Context,
	input *struct {
		FindJobInput
		Body JobUpdateDto `json:"body" required:"true"`
	},
) (*struct{}, error) {
	id := uuid.MustParse(input.ID)
	j, err := api.app.Adapter().Job().FindJob(ctx, &stores.JobFilter{
		Ids: []uuid.UUID{id},
	})
	if err != nil {
		return nil, err
	}
	j.Kind = input.Body.Kind
	j.UniqueKey = input.Body.UniqueKey
	j.Payload = input.Body.Payload
	j.Status = models.JobStatus(input.Body.Status)
	j.RunAfter = input.Body.RunAfter
	j.Attempts = input.Body.Attempts
	j.MaxAttempts = input.Body.MaxAttempts
	j.LastError = input.Body.LastError
	_, err = api.app.Adapter().Job().UpdateJob(ctx, j)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
