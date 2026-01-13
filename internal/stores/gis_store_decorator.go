package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	geom "github.com/twpayne/go-geom"
)

type DBGisStoreDecorator struct {
	Delegate GisStore
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

var _ GisStore = (*DBGisStoreDecorator)(nil)
