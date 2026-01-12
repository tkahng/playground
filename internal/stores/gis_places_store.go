package stores

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/twpayne/go-geom"
)

func (s *DBGisStore) CreateManyPlaces(ctx context.Context, places []*models.PopulatedPlace) error {
	res, err := repository.PopulatedPlace.PostExec(ctx, s.db, mapper.Map(places, func(t *models.PopulatedPlace) models.PopulatedPlace {
		return *t
	}))
	if err != nil {
		return err
	}
	if int64(len(places)) != res {
		return fmt.Errorf("failed to insert all places")
	}
	return nil
}

func (s *DBGisStore) CreatePopulatedPlace(ctx context.Context, country *models.PopulatedPlace) (*models.PopulatedPlace, error) {
	return repository.PopulatedPlace.PostOne(ctx, s.db, country)
}

func (s *DBGisStore) FindPopulatedPlaceByPoint(ctx context.Context, point *geom.Point) (*models.PopulatedPlace, error) {
	query := `
SELECT
    *,
    gis.ST_Distance(geom::gis.geometry, $1) AS distance_meters
FROM
    gis.populated_places
ORDER BY
    distance_meters ASC
LIMIT 1;
	`

	rows, err := s.db.Query(ctx, query, point)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	places, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[models.PopulatedPlace])
	if err != nil {
		return nil, err
	}
	if len(places) > 0 {
		return places[0], nil
	}

	return nil, nil
}
