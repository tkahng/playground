package stores

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/tools/utils"
)

type DbCustomerStoreInterface interface {
	ListCustomers(ctx context.Context, input *StripeCustomerFilter) ([]*models.StripeCustomer, error)
	CountCustomers(ctx context.Context, filter *StripeCustomerFilter) (int64, error)
	CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error)
	FindCustomer(ctx context.Context, customer *StripeCustomerFilter) (*models.StripeCustomer, error)
	LoadCustomersByIds(ctx context.Context, ids ...string) ([]*models.StripeCustomer, error)
}

type DbCustomerStore struct {
	db database.Dbx
}

// LoadCustomersByIds implements DbCustomerStoreInterface.
func (s *DbCustomerStore) LoadCustomersByIds(ctx context.Context, ids ...string) ([]*models.StripeCustomer, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	prices, err := repository.StripeCustomer.Get(
		ctx,
		s.db,
		&map[string]any{
			models.StripeCustomerTable.ID: map[string]any{
				"_in": ids,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(prices, ids, func(t *models.StripeCustomer) string {
		if t == nil {
			return ""
		}
		return t.ID
	}), nil
}

var _ DbCustomerStoreInterface = (*DbCustomerStore)(nil)

func NewDbCustomerStore(db database.Dbx) *DbCustomerStore {
	return &DbCustomerStore{
		db: db,
	}
}

func (s *DbCustomerStore) WithTx(tx database.Dbx) *DbCustomerStore {
	return &DbCustomerStore{
		db: tx,
	}
}

func (s *DbCustomerStore) ListCustomers(ctx context.Context, input *StripeCustomerFilter) ([]*models.StripeCustomer, error) {

	limit, offset := pagination(input)
	where := s.filter(input)
	order := s.stripeCustomerOrderByFunc(input)
	data, err := repository.StripeCustomer.Get(
		ctx,
		s.db,
		where,
		order,
		&limit,
		&offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (s *DbCustomerStore) CountCustomers(ctx context.Context, filter *StripeCustomerFilter) (int64, error) {
	where := s.filter(filter)
	data, err := repository.StripeCustomer.Count(ctx, s.db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func (s *DbCustomerStore) stripeCustomerOrderByFunc(input *StripeCustomerFilter) *map[string]string {
	if input == nil {
		return nil
	}
	order := make(map[string]string)
	if slices.Contains(models.StripeCustomerTable.Columns, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}
	return &order
}

type StripeCustomerFilter struct {
	PaginatedInput
	SortParams
	Q            string                                         `query:"q,omitempty" required:"false"`
	Ids          []string                                       `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Emails       []string                                       `db:"emails" json:"emails,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Names        []string                                       `db:"names" json:"names,omitempty" required:"false"`
	UserIds      []uuid.UUID                                    `db:"user_ids" json:"user_ids,omitempty" required:"false"`
	TeamIds      []uuid.UUID                                    `db:"team_ids" json:"team_ids,omitempty" required:"false"`
	CustomerType types.OptionalParam[models.StripeCustomerType] `query:"customer_type,omitempty" json:"customer_type,omitempty" enum:"user,team" required:"false"`
}

func (s *DbCustomerStore) filter(filter *StripeCustomerFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
	if len(filter.Ids) > 0 {
		where[models.StripeCustomerTable.ID] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Emails) > 0 {
		where[models.StripeCustomerTable.Email] = map[string]any{
			"_in": filter.Emails,
		}
	}
	if len(filter.Names) > 0 {
		where[models.StripeCustomerTable.Name] = map[string]any{
			"_in": filter.Names,
		}
	}
	if len(filter.UserIds) > 0 {
		where[models.StripeCustomerTable.UserID] = map[string]any{
			"_in": filter.UserIds,
		}
	}
	if len(filter.TeamIds) > 0 {
		where[models.StripeCustomerTable.TeamID] = map[string]any{
			"_in": filter.TeamIds,
		}
	}
	if filter.CustomerType.IsSet {
		where[models.StripeCustomerTable.CustomerType] = map[string]any{
			"_eq": filter.CustomerType.Value,
		}
	}
	return &where
}

func SelectStripeCustomerColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.Column(models.StripeCustomerTablePrefix.ID + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.ID))).
		Column(models.StripeCustomerTablePrefix.Email + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.Email))).
		Column(models.StripeCustomerTablePrefix.Name + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.Name))).
		Column(models.StripeCustomerTablePrefix.UserID + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.UserID))).
		Column(models.StripeCustomerTablePrefix.TeamID + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.TeamID))).
		Column(models.StripeCustomerTablePrefix.CustomerType + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.CustomerType))).
		Column(models.StripeCustomerTablePrefix.BillingAddress + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.BillingAddress))).
		Column(models.StripeCustomerTablePrefix.PaymentMethod + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.PaymentMethod))).
		Column(models.StripeCustomerTablePrefix.CreatedAt + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.CreatedAt))).
		Column(models.StripeCustomerTablePrefix.UpdatedAt + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeCustomerTable.UpdatedAt)))
	return qs
}

func (s *DbCustomerStore) CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	if customer == nil {
		return nil, errors.New("customer is nil")
	}
	if customer.UserID != nil {
		customer.CustomerType = models.StripeCustomerTypeUser
	} else if customer.TeamID != nil {
		customer.CustomerType = models.StripeCustomerTypeTeam
	} else {
		return nil, errors.New("customer type is not set")
	}
	return repository.StripeCustomer.PostOne(
		ctx,
		s.db,
		customer,
	)
}

// FindCustomer implements PaymentStore.
func (s *DbCustomerStore) FindCustomer(ctx context.Context, filter *StripeCustomerFilter) (*models.StripeCustomer, error) {
	if filter == nil {
		return nil, nil
	}
	where := s.filter(filter)
	data, err := repository.StripeCustomer.GetOne(
		ctx,
		s.db,
		where,
	)
	return database.OptionalRow(data, err)
}
