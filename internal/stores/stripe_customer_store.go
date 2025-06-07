package stores

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type DbCustomerStore struct {
	db database.Dbx
}

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

func (s *DbCustomerStore) ListCustomers(ctx context.Context, input *shared.StripeCustomerListParams) ([]*models.StripeCustomer, error) {

	filter := input.StripeCustomerListFilter
	pageInput := &input.PaginatedInput

	limit, offset := database.PaginateRepo(pageInput)
	where := listCustomerFilterFunc(&filter)
	order := stripeCustomerOrderByFunc(input)
	data, err := repository.StripeCustomer.Get(
		ctx,
		s.db,
		where,
		order,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (s *DbCustomerStore) CountCustomers(ctx context.Context, filter *shared.StripeCustomerListFilter) (int64, error) {
	where := listCustomerFilterFunc(filter)
	data, err := repository.StripeCustomer.Count(ctx, s.db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func stripeCustomerOrderByFunc(input *shared.StripeCustomerListParams) *map[string]string {
	if input == nil {
		return nil
	}
	order := make(map[string]string)
	if slices.Contains(models.StripeCustomerTable.Columns, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}
	return &order
}

func listCustomerFilterFunc(filter *shared.StripeCustomerListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
	if len(filter.Ids) > 0 {
		where[models.StripeCustomerTable.ID] = map[string]any{
			"_in": filter.Ids,
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
func (s *DbCustomerStore) FindCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	if customer == nil {
		return nil, nil
	}
	where := map[string]any{}
	if customer.ID != "" {
		where[models.StripeCustomerTable.ID] = map[string]any{
			"_eq": customer.ID,
		}
	}
	if customer.TeamID != nil {
		where[models.StripeCustomerTable.TeamID] = map[string]any{
			"_eq": customer.TeamID,
		}
	}
	if customer.UserID != nil {
		where[models.StripeCustomerTable.UserID] = map[string]any{
			"_eq": customer.UserID,
		}
	}
	data, err := repository.StripeCustomer.GetOne(
		ctx,
		s.db,
		&where,
	)
	return database.OptionalRow(data, err)
}
