package stores

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/twpayne/go-geom"
)

type DBGisStore struct {
	db database.Dbx
}
type GisStore interface {
	CreateCountry(ctx context.Context, country *models.Country) (*models.Country, error)
	CreateManyCountries(ctx context.Context, countries []*models.Country) error
	FindCountryByPoint(ctx context.Context, point *geom.Point) (*models.Country, error)
}

var _ GisStore = &DBGisStore{}

func NewGisStore(db database.Dbx) *DBGisStore {
	return &DBGisStore{
		db: db,
	}
}

func (s *DBGisStore) CreateManyCountries(ctx context.Context, countries []*models.Country) error {
	res, err := repository.Country.PostExec(ctx, s.db, mapper.Map(countries, func(t *models.Country) models.Country {
		return *t
	}))
	if err != nil {
		return err
	}
	if int64(len(countries)) != res {
		return fmt.Errorf("failed to insert all countries")
	}
	return nil
}

func (s *DBGisStore) CreateCountry(ctx context.Context, country *models.Country) (*models.Country, error) {
	return repository.Country.PostOne(ctx, s.db, country)
}

func (s *DBGisStore) FindCountryByPoint(ctx context.Context, point *geom.Point) (*models.Country, error) {
	query := `
		SELECT gid, name, iso_a2_eh, iso_a3_eh, geom
		FROM gis.countries
		WHERE ST_Contains(geom, $1) 
		LIMIT 1
	`

	rows, err := s.db.Query(ctx, query, point)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	countries, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[models.Country])
	if err != nil {
		return nil, err
	}
	if len(countries) > 0 {
		return countries[0], nil
	}

	return nil, nil
}
