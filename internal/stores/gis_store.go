package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/twpayne/go-geom"
)

type DBGisStore struct {
	db database.Dbx
}
type GisStore interface {
	CreateCountry(ctx context.Context, country *models.Country) (*models.Country, error)
	CreateManyCountries(ctx context.Context, countries []*models.Country) error
	FindCountryByPoint(ctx context.Context, point *geom.Point) (*models.Country, error)
	CreatePopulatedPlace(ctx context.Context, place *models.PopulatedPlace) (*models.PopulatedPlace, error)
	CreateManyPopulatedPlaces(ctx context.Context, places []*models.PopulatedPlace) error
	FindPopulatedPlaceByPoint(ctx context.Context, point *geom.Point) (*models.PopulatedPlace, error)
}

var _ GisStore = &DBGisStore{}

func NewGisStore(db database.Dbx) *DBGisStore {
	return &DBGisStore{
		db: db,
	}
}
