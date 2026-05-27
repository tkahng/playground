package stores

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error)
	InsertManyNotifications(ctx context.Context, notifications []models.Notification) (int64, error)
	FindNotification(ctx context.Context, args *NotificationFilter) (*models.Notification, error)
	FindNotifications(ctx context.Context, args *NotificationFilter) ([]*models.Notification, error)
	CountNotification(ctx context.Context, args *NotificationFilter) (int64, error)
	UpdateNotification(ctx context.Context, notification *models.Notification) error
	DeleteNotifications(ctx context.Context, args *NotificationFilter) (int64, error)
	MarkAllNotificationsRead(ctx context.Context, teamMemberID uuid.UUID) error
	// FindDisabledMemberIDs returns the subset of memberIDs that have explicitly
	// disabled the given notification type in their preferences.
	FindDisabledMemberIDs(ctx context.Context, memberIDs []uuid.UUID, notifType string) ([]uuid.UUID, error)
	// UpsertNotificationPreference creates or updates a preference row.
	UpsertNotificationPreference(ctx context.Context, teamMemberID uuid.UUID, notifType string, enabled bool) error
}

type DbNotificationStore struct {
	db database.Dbx
}

// DeleteNotifications implements NotificationStore.
func (s *DbNotificationStore) DeleteNotifications(ctx context.Context, args *NotificationFilter) (int64, error) {
	where := s.filter(args)
	return repository.Notification.Delete(
		ctx,
		s.db,
		where,
	)
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

func (s *DbNotificationStore) WithTx(db database.Dbx) *DbNotificationStore {
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

func (s *DbNotificationStore) InsertManyNotifications(ctx context.Context, notifications []models.Notification) (int64, error) {
	return repository.Notification.PostExec(
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
	Unread        bool                           `query:"unread" json:"unread,omitempty" required:"false"`
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
	if args.Unread {
		where["read_at"] = map[string]any{
			"_isnull": nil,
		}
	}
	return &where
}

func (s *DbNotificationStore) MarkAllNotificationsRead(ctx context.Context, teamMemberID uuid.UUID) error {
	db := database.GetContextOrDefaultDbx(ctx, s.db)
	_, err := db.Exec(ctx, `
		UPDATE messaging.notifications
		SET read_at = clock_timestamp(), updated_at = clock_timestamp()
		WHERE team_member_id = $1 AND read_at IS NULL
	`, teamMemberID)
	return err
}

func (s *DbNotificationStore) FindDisabledMemberIDs(ctx context.Context, memberIDs []uuid.UUID, notifType string) ([]uuid.UUID, error) {
	if len(memberIDs) == 0 {
		return nil, nil
	}
	db := database.GetContextOrDefaultDbx(ctx, s.db)
	rows, err := db.Query(ctx, `
		SELECT team_member_id
		FROM messaging.team_notification_preferences
		WHERE team_member_id = ANY($1) AND type = $2 AND enabled = false
	`, memberIDs, notifType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var disabled []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		disabled = append(disabled, id)
	}
	return disabled, rows.Err()
}

func (s *DbNotificationStore) UpsertNotificationPreference(ctx context.Context, teamMemberID uuid.UUID, notifType string, enabled bool) error {
	db := database.GetContextOrDefaultDbx(ctx, s.db)
	_, err := db.Exec(ctx, `
		INSERT INTO messaging.team_notification_preferences (team_member_id, type, enabled)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_member_id, type) DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = clock_timestamp()
	`, teamMemberID, notifType, enabled)
	return err
}

func (d *DbNotificationStore) sort(filter *NotificationFilter) *map[string]string {
	sortBy, sortOrder := filter.Sort()
	if slices.Contains(repository.NotificationBuilder.FieldNames(), sortBy) {
		return &map[string]string{
			sortBy: strings.ToUpper(sortOrder),
		}
	}
	return nil
}

type NotificationStoreDecorator struct {
	Delegate                  *DbNotificationStore
	CountFunc                 func(ctx context.Context, filter *NotificationFilter) (int64, error)
	CreateFunc                func(ctx context.Context, notification *models.Notification) (*models.Notification, error)
	CreateManyFunc            func(ctx context.Context, notifications []models.Notification) (int64, error)
	FindNotificationFunc      func(ctx context.Context, args *NotificationFilter) (*models.Notification, error)
	FindNotificationsFunc     func(ctx context.Context, args *NotificationFilter) ([]*models.Notification, error)
	UpdateFunc                func(ctx context.Context, notification *models.Notification) error
	DeleteFunc                func(ctx context.Context, args *NotificationFilter) (int64, error)
	MarkAllReadFunc           func(ctx context.Context, teamMemberID uuid.UUID) error
}

// DeleteNotifications implements NotificationStore.
func (n *NotificationStoreDecorator) DeleteNotifications(ctx context.Context, args *NotificationFilter) (int64, error) {
	if n.DeleteFunc != nil {
		return n.DeleteFunc(ctx, args)
	}
	if n.Delegate == nil {
		return 0, errors.New("delegate is nil in DeleteNotifications")
	}
	return n.Delegate.DeleteNotifications(ctx, args)
}

func NewNotificationStoreDecorator(db database.Dbx) *NotificationStoreDecorator {
	return &NotificationStoreDecorator{
		Delegate: &DbNotificationStore{db: db},
	}
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

// InsertManyNotifications implements NotificationStore.
func (n *NotificationStoreDecorator) InsertManyNotifications(ctx context.Context, notifications []models.Notification) (int64, error) {
	if n.CreateManyFunc != nil {
		return n.CreateManyFunc(ctx, notifications)
	}
	if n.Delegate == nil {
		return 0, errors.New("delegate is nil in CreateManyNotifications")
	}
	return n.Delegate.InsertManyNotifications(ctx, notifications)
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

// MarkAllNotificationsRead implements NotificationStore.
func (n *NotificationStoreDecorator) MarkAllNotificationsRead(ctx context.Context, teamMemberID uuid.UUID) error {
	if n.MarkAllReadFunc != nil {
		return n.MarkAllReadFunc(ctx, teamMemberID)
	}
	if n.Delegate == nil {
		return errors.New("delegate is nil in MarkAllNotificationsRead")
	}
	return n.Delegate.MarkAllNotificationsRead(ctx, teamMemberID)
}

// FindDisabledMemberIDs implements NotificationStore.
func (n *NotificationStoreDecorator) FindDisabledMemberIDs(ctx context.Context, memberIDs []uuid.UUID, notifType string) ([]uuid.UUID, error) {
	if n.Delegate == nil {
		return nil, errors.New("delegate is nil in FindDisabledMemberIDs")
	}
	return n.Delegate.FindDisabledMemberIDs(ctx, memberIDs, notifType)
}

// UpsertNotificationPreference implements NotificationStore.
func (n *NotificationStoreDecorator) UpsertNotificationPreference(ctx context.Context, teamMemberID uuid.UUID, notifType string, enabled bool) error {
	if n.Delegate == nil {
		return errors.New("delegate is nil in UpsertNotificationPreference")
	}
	return n.Delegate.UpsertNotificationPreference(ctx, teamMemberID, notifType, enabled)
}

var _ NotificationStore = (*NotificationStoreDecorator)(nil)
