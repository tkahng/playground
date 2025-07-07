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

type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error)
	CreateManyNotifications(ctx context.Context, notifications []models.Notification) ([]*models.Notification, error)
	FindNotification(ctx context.Context, args *NotificationFilter) (*models.Notification, error)
	FindNotifications(ctx context.Context, args *NotificationFilter) ([]*models.Notification, error)
	CountNotification(ctx context.Context, args *NotificationFilter) (int64, error)
	UpdateNotification(ctx context.Context, notification *models.Notification) error
}

type DbNotificationStore struct {
	db database.Dbx
}

// UpdateNotification implements NotificationStore.
func (s *DbNotificationStore) UpdateNotification(ctx context.Context, notification *models.Notification) error {
	_, err := repository.Notification.PutOne(
		ctx,
		s.db,
		notification,
	)
	return err
}

var _ NotificationStore = (*DbNotificationStore)(nil)

func NewDbNotificationStore(db database.Dbx) *DbNotificationStore {
	return &DbNotificationStore{
		db: db,
	}
}

func (s *DbNotificationStore) CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	return repository.Notification.PostOne(
		ctx,
		s.db,
		notification,
	)
}

func (s *DbNotificationStore) CreateManyNotifications(ctx context.Context, notifications []models.Notification) ([]*models.Notification, error) {
	return repository.Notification.Post(
		ctx,
		s.db,
		notifications,
	)
}

type NotificationFilter struct {
	PaginatedInput
	SortParams
	Ids           []uuid.UUID                    `query:"ids" json:"ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	UserIds       []uuid.UUID                    `query:"user_ids" json:"user_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	TeamIds       []uuid.UUID                    `query:"team_ids" json:"team_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	TeamMemberIds []uuid.UUID                    `query:"team_member_ids" json:"team_member_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Channels      []string                       `query:"channels" json:"channels,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Types         []string                       `query:"types" json:"types,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	ReadAt        types.OptionalParam[time.Time] `query:"read_at" json:"read_at" required:"false"`
}

func (s *DbNotificationStore) FindNotification(ctx context.Context, args *NotificationFilter) (*models.Notification, error) {
	where := s.filter(args)
	return repository.Notification.GetOne(
		ctx,
		s.db,
		where,
	)
}

func (s *DbNotificationStore) FindNotifications(ctx context.Context, args *NotificationFilter) ([]*models.Notification, error) {
	where := s.filter(args)
	sort := s.sort(args)
	limit, offset := args.LimitOffset()
	return repository.Notification.Get(
		ctx,
		s.db,
		where,
		sort,
		&limit,
		&offset,
	)
}

func (s *DbNotificationStore) CountNotification(ctx context.Context, args *NotificationFilter) (int64, error) {
	where := s.filter(args)
	return repository.Notification.Count(
		ctx,
		s.db,
		where,
	)
}

func (s *DbNotificationStore) filter(args *NotificationFilter) *map[string]any {
	if args == nil {
		return nil
	}
	where := map[string]any{}
	if len(args.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": args.Ids,
		}
	}
	if len(args.UserIds) > 0 {
		where["user_id"] = map[string]any{
			"_in": args.UserIds,
		}
	}
	if len(args.TeamIds) > 0 {
		where["team_id"] = map[string]any{
			"_in": args.TeamIds,
		}
	}
	if len(args.TeamMemberIds) > 0 {
		where["team_member_id"] = map[string]any{
			"_in": args.TeamMemberIds,
		}
	}
	if len(args.Channels) > 0 {
		where["channel"] = map[string]any{
			"_in": args.Channels,
		}
	}
	if len(args.Types) > 0 {
		where["type"] = map[string]any{
			"_in": args.Types,
		}
	}
	if args.ReadAt.IsSet {
		where["read_at"] = map[string]any{
			"_eq": args.ReadAt.Value,
		}
	}
	return &where
}

func (d *DbNotificationStore) sort(filter *NotificationFilter) *map[string]string {
	sortBy, sortOrder := filter.Sort()
	if slices.Contains(repository.NotificationBuilder.ColumnNames(), utils.Quote(sortBy)) {
		return &map[string]string{
			sortBy: strings.ToUpper(sortOrder),
		}
	}
	return nil
}

type NotificationStoreDecorator struct {
	Delegate              *DbNotificationStore
	CountFunc             func(ctx context.Context, filter *NotificationFilter) (int64, error)
	CreateFunc            func(ctx context.Context, notification *models.Notification) (*models.Notification, error)
	CreateManyFunc        func(ctx context.Context, notifications []models.Notification) ([]*models.Notification, error)
	FindNotificationFunc  func(ctx context.Context, args *NotificationFilter) (*models.Notification, error)
	FindNotificationsFunc func(ctx context.Context, args *NotificationFilter) ([]*models.Notification, error)
	UpdateFunc            func(ctx context.Context, notification *models.Notification) error
}

// CountNotification implements NotificationStore.
func (n *NotificationStoreDecorator) CountNotification(ctx context.Context, args *NotificationFilter) (int64, error) {
	if n.CountFunc != nil {
		return n.CountFunc(ctx, args)
	}
	if n.Delegate == nil {
		return 0, errors.New("delegate is nil in CountNotification")
	}
	return n.Delegate.CountNotification(ctx, args)
}

// CreateManyNotifications implements NotificationStore.
func (n *NotificationStoreDecorator) CreateManyNotifications(ctx context.Context, notifications []models.Notification) ([]*models.Notification, error) {
	if n.CreateFunc != nil {
		return n.CreateManyFunc(ctx, notifications)
	}
	if n.Delegate == nil {
		return nil, errors.New("delegate is nil in CreateManyNotifications")
	}
	return n.Delegate.CreateManyNotifications(ctx, notifications)
}

// CreateNotification implements NotificationStore.
func (n *NotificationStoreDecorator) CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	if n.CreateFunc != nil {
		return n.CreateFunc(ctx, notification)
	}
	if n.Delegate == nil {
		return nil, errors.New("delegate is nil in CreateNotification")
	}
	return n.Delegate.CreateNotification(ctx, notification)
}

// FindNotification implements NotificationStore.
func (n *NotificationStoreDecorator) FindNotification(ctx context.Context, args *NotificationFilter) (*models.Notification, error) {
	if n.FindNotificationFunc != nil {
		return n.FindNotificationFunc(ctx, args)
	}
	if n.Delegate == nil {
		return nil, errors.New("delegate is nil in FindNotificationFuncNotification")
	}
	return n.Delegate.FindNotification(ctx, args)
}

// FindNotifications implements NotificationStore.
func (n *NotificationStoreDecorator) FindNotifications(ctx context.Context, args *NotificationFilter) ([]*models.Notification, error) {
	if n.FindNotificationsFunc != nil {
		return n.FindNotificationsFunc(ctx, args)
	}
	if n.Delegate == nil {
		return nil, errors.New("delegate is nil in FindNotifications")
	}
	return n.Delegate.FindNotifications(ctx, args)
}

// UpdateNotification implements NotificationStore.
func (n *NotificationStoreDecorator) UpdateNotification(ctx context.Context, notification *models.Notification) error {
	if n.UpdateFunc != nil {
		return n.UpdateFunc(ctx, notification)
	}
	if n.Delegate == nil {
		return errors.New("delegate is nil in UpdateNotification")
	}
	return n.Delegate.UpdateNotification(ctx, notification)
}

var _ NotificationStore = (*NotificationStoreDecorator)(nil)
