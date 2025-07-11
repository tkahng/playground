package stores

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type JobFilter struct {
	PaginatedInput
	SortParams
	Ids        []uuid.UUID                    `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Kinds      []string                       `db:"kinds" json:"kinds" query:"kinds" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	UniqueKeys []string                       `db:"unique_keys" json:"unique_keys" query:"unique_keys" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Statuses   []models.JobStatus             `db:"statuses" json:"statuses" query:"statuses" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	RunAfter   types.OptionalParam[time.Time] `db:"run_after" json:"run_after" query:"run_after" required:"false"`
	Attempt    types.OptionalParam[int64]     `db:"attempt" json:"attempt" query:"attempt" required:"false"`
	LastErrors []string                       `db:"last_errors" json:"last_errors" query:"last_errors" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
}

type JobStore interface {
	FindJob(ctx context.Context, filter *JobFilter) (*models.JobRow, error)
	FindJobs(ctx context.Context, filter *JobFilter) ([]*models.JobRow, error)
	CountJobs(ctx context.Context, filter *JobFilter) (int64, error)
	CreateJob(ctx context.Context, job *models.JobRow) (*models.JobRow, error)
	UpdateJob(ctx context.Context, job *models.JobRow) (*models.JobRow, error)
	DeleteJob(ctx context.Context, filter *JobFilter) (int64, error)
}

type DbJobStore struct {
	db database.Dbx
}

func NewDbJobStore(db database.Dbx) *DbJobStore {
	return &DbJobStore{
		db: db,
	}
}

func (d *DbJobStore) filter(filter *JobFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}

	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Kinds) > 0 {
		where["kind"] = map[string]any{
			"_in": filter.Kinds,
		}
	}
	if len(filter.UniqueKeys) > 0 {
		where["unique_key"] = map[string]any{
			"_in": filter.UniqueKeys,
		}
	}

	if len(filter.Statuses) > 0 {
		where["status"] = map[string]any{
			"_in": filter.Statuses,
		}
	}
	if filter.RunAfter.IsSet {
		where["run_after"] = map[string]any{
			"_gte": filter.RunAfter.Value,
		}
	}
	if filter.Attempt.IsSet {
		where["attempt"] = map[string]any{
			"_gte": filter.Attempt.Value,
		}
	}
	if len(filter.LastErrors) > 0 {
		where["last_error"] = map[string]any{
			"_in": filter.LastErrors,
		}
	}
	return &where
}

func (d *DbJobStore) sort(filter *JobFilter) *map[string]string {
	sortBy, sortOrder := filter.Sort()
	if slices.Contains(repository.JobBuilder.ColumnNames(), utils.Quote(sortBy)) {
		return &map[string]string{
			sortBy: strings.ToUpper(sortOrder),
		}
	}
	return nil
}

// CountJobs implements JobStore.
func (d *DbJobStore) CountJobs(ctx context.Context, filter *JobFilter) (int64, error) {
	where := d.filter(filter)
	return repository.Job.Count(ctx, d.db, where)
}

// CreateJob implements JobStore.
func (d *DbJobStore) CreateJob(ctx context.Context, job *models.JobRow) (*models.JobRow, error) {
	return repository.Job.PostOne(ctx, d.db, job)
}

// DeleteJob implements JobStore.
func (d *DbJobStore) DeleteJob(ctx context.Context, filter *JobFilter) (int64, error) {
	where := d.filter(filter)
	return repository.Job.Delete(ctx, d.db, where)
}

// FindJob implements JobStore.
func (d *DbJobStore) FindJob(ctx context.Context, filter *JobFilter) (*models.JobRow, error) {
	where := d.filter(filter)
	return repository.Job.GetOne(ctx, d.db, where)
}

// FindJobs implements JobStore.
func (d *DbJobStore) FindJobs(ctx context.Context, filter *JobFilter) ([]*models.JobRow, error) {
	where := d.filter(filter)
	sort := d.sort(filter)
	limit, offset := filter.LimitOffset()
	return repository.Job.Get(
		ctx,
		d.db,
		where,
		sort,
		&limit,
		&offset,
	)
}

// UpdateJob implements JobStore.
func (d *DbJobStore) UpdateJob(ctx context.Context, job *models.JobRow) (*models.JobRow, error) {
	return repository.Job.PutOne(ctx, d.db, job)
}

func (d *DbJobStore) WithTx(db database.Dbx) JobStore {
	return &DbJobStore{
		db: db,
	}
}

type JobStoreDecorator struct {
	Delegate      JobStore
	CountJobsFunc func(ctx context.Context, filter *JobFilter) (int64, error)
	FindJobsFunc  func(ctx context.Context, filter *JobFilter) ([]*models.JobRow, error)
	FindJobFunc   func(ctx context.Context, filter *JobFilter) (*models.JobRow, error)
	CreateJobFunc func(ctx context.Context, job *models.JobRow) (*models.JobRow, error)
	UpdateJobFunc func(ctx context.Context, job *models.JobRow) (*models.JobRow, error)
	DeleteJobFunc func(ctx context.Context, filter *JobFilter) (int64, error)
	RunInTxFunc   func(ctx context.Context, fn func(JobStore) error) error
}

// CountJobs implements JobStore.
func (j *JobStoreDecorator) CountJobs(ctx context.Context, filter *JobFilter) (int64, error) {
	if j.CountJobsFunc != nil {
		return j.CountJobsFunc(ctx, filter)
	}
	if j.Delegate == nil {
		return 0, errors.New("delegate for CountJobs in JobStore is nil")
	}
	return j.Delegate.CountJobs(ctx, filter)
}

// CreateJob implements JobStore.
func (j *JobStoreDecorator) CreateJob(ctx context.Context, job *models.JobRow) (*models.JobRow, error) {
	if j.CreateJobFunc != nil {
		return j.CreateJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return nil, errors.New("delegate for CreateJob in JobStore is nil")
	}
	return j.Delegate.CreateJob(ctx, job)
}

// DeleteJob implements JobStore.
func (j *JobStoreDecorator) DeleteJob(ctx context.Context, filter *JobFilter) (int64, error) {
	if j.DeleteJobFunc != nil {
		return j.DeleteJobFunc(ctx, filter)
	}
	if j.Delegate == nil {
		return 0, errors.New("delegate for DeleteJob in JobStore is nil")
	}
	return j.Delegate.DeleteJob(ctx, filter)
}

// FindJob implements JobStore.
func (j *JobStoreDecorator) FindJob(ctx context.Context, filter *JobFilter) (*models.JobRow, error) {
	if j.FindJobFunc != nil {
		return j.FindJobFunc(ctx, filter)
	}
	if j.Delegate == nil {
		return nil, errors.New("delegate for FindJob in JobStore is nil")
	}
	return j.Delegate.FindJob(ctx, filter)
}

// FindJobs implements JobStore.
func (j *JobStoreDecorator) FindJobs(ctx context.Context, filter *JobFilter) ([]*models.JobRow, error) {
	if j.FindJobsFunc != nil {
		return j.FindJobsFunc(ctx, filter)
	}
	if j.Delegate == nil {
		return nil, errors.New("delegate for FindJobs in JobStore is nil")
	}
	return j.Delegate.FindJobs(ctx, filter)
}

// UpdateJob implements JobStore.
func (j *JobStoreDecorator) UpdateJob(ctx context.Context, job *models.JobRow) (*models.JobRow, error) {
	if j.UpdateJobFunc != nil {
		return j.UpdateJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return nil, errors.New("delegate for UpdateJob in JobStore is nil")
	}
	return j.Delegate.UpdateJob(ctx, job)
}

var _ JobStore = &JobStoreDecorator{}
