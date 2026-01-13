package stores

import (
	"context"
	"fmt"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	geom "github.com/twpayne/go-geom"
)

type DBGisStoreDecorator struct {
	Delegate                GisStore
	CreateCountryFunc       func(ctx context.Context, country *models.Country) (*models.Country, error)
	CreateManyCountriesFunc func(ctx context.Context, countries []*models.Country) error
	FindCountryByPointFunc  func(ctx context.Context, point *geom.Point) (*models.Country, error)
}

// CreateManyPopulatedPlaces implements [GisStore].
func (d *DBGisStoreDecorator) CreateManyPopulatedPlaces(ctx context.Context, places []*models.PopulatedPlace) error {
	panic("unimplemented")
}

// FindPopulatedPlaceByPoint implements [GisStore].
func (d *DBGisStoreDecorator) FindPopulatedPlaceByPoint(ctx context.Context, point *geom.Point) (*models.PopulatedPlace, error) {
	panic("unimplemented")
}

// CreatePopulatedPlace implements [GisStore].
func (d *DBGisStoreDecorator) CreatePopulatedPlace(ctx context.Context, place *models.PopulatedPlace) (*models.PopulatedPlace, error) {
	panic("unimplemented")
}

func NewDBGisStoreDecorator(db database.Dbx) *DBGisStoreDecorator {
	return &DBGisStoreDecorator{
		Delegate: NewGisStore(db),
	}
}

// CreateCountry implements [GisStore].
func (d *DBGisStoreDecorator) CreateCountry(ctx context.Context, country *models.Country) (*models.Country, error) {
	if d.CreateCountryFunc != nil {
		return d.CreateCountryFunc(ctx, country)
	}
	if d.Delegate != nil {
		return d.Delegate.CreateCountry(ctx, country)
	}
	return nil, fmt.Errorf("DBGisStoreDecorator CreateCountry %w", ErrDelegateNil)
}

// CreateManyCountries implements [GisStore].
func (d *DBGisStoreDecorator) CreateManyCountries(ctx context.Context, countries []*models.Country) error {
	if d.CreateManyCountriesFunc != nil {
		return d.CreateManyCountriesFunc(ctx, countries)
	}
	if d.Delegate != nil {
		return d.Delegate.CreateManyCountries(ctx, countries)
	}
	return fmt.Errorf("DBGisStoreDecorator CreateManyCountries %w", ErrDelegateNil)
}

// FindCountryByPoint implements [GisStore].
func (d *DBGisStoreDecorator) FindCountryByPoint(ctx context.Context, point *geom.Point) (*models.Country, error) {
	if d.FindCountryByPointFunc != nil {
		return d.FindCountryByPointFunc(ctx, point)
	}
	if d.Delegate != nil {
		return d.Delegate.FindCountryByPoint(ctx, point)
	}
	return nil, fmt.Errorf("DBGisStoreDecorator FindCountryByPoint %w", ErrDelegateNil)
}

var _ GisStore = (*DBGisStoreDecorator)(nil)
