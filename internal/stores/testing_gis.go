package stores

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/twpayne/go-geom"
)

// City represents a city with its name and geographic location
type City struct {
	Name     string
	Location *geom.Point
}

// TestCities contains sample city data for testing
var TestCities = []City{
	{
		Name:     "Seoul",
		Location: geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{126.9780, 37.5665}).SetSRID(4326),
	},
	{
		Name:     "Jeju",
		Location: geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{126.5219, 33.4996}).SetSRID(4326),
	},
	{
		Name:     "Pyongyang",
		Location: geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{125.7625, 39.0392}).SetSRID(4326),
	},
}

// TestNearCities contains sample points near cities for testing
var TestNearCities = []City{
	{
		Name:     "Seoul",
		Location: geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{127.2442, 37.6562}).SetSRID(4326), // 37.65623258608295, 127.2441705774566 namyang
	},
	{
		Name:     "Jeju",
		Location: geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{126.9575, 33.5096}).SetSRID(4326), // 33.509571100302324, 126.95748624280017 udo island
	},
	{
		Name:     "Pyongyang",
		Location: geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{126.6769, 38.7394}).SetSRID(4326), // 38.73936017702917, 126.67691505774272 koksan
	},
}

// GetCityByName returns a test city by name
func GetCityByName(t *testing.T, name string) *City {
	for _, city := range TestCities {
		if city.Name == name {
			return &city
		}
	}
	t.Fatal("city not found")
	return nil
}

// GetCityByName returns a test city by name
func GetNearCityByName(t *testing.T, name string) *City {
	for _, city := range TestNearCities {
		if city.Name == name {
			return &city
		}
	}
	t.Fatal("city not found")
	return nil
}

func MarshalCountries(t *testing.T, b []byte) []*models.Country {
	var countries []*models.Country
	err := json.Unmarshal(b, &countries)
	require.NoError(t, err)
	return countries
}

func LoadCountriesToDB(t *testing.T, db database.Dbx) {
	res := test.ReadFileFromDataFs(t, "data/populated_places.json")
	countries := MarshalCountries(t, res)
	gisStore := NewGisStore(db)
	err := gisStore.CreateManyCountries(t.Context(), countries)
	require.NoError(t, err)
}

func MarshalPopulatedPlaces(t *testing.T, b []byte) []*models.PopulatedPlace {
	var places []*models.PopulatedPlace
	err := json.Unmarshal(b, &places)
	require.NoError(t, err)
	return places
}

func LoadPopulatedPlacesToDB(t *testing.T, db database.Dbx) {
	res := test.ReadFileFromDataFs(t, "data/populated_places.json")
	countries := MarshalPopulatedPlaces(t, res)
	gisStore := NewGisStore(db)
	err := gisStore.CreateManyPlaces(t.Context(), countries)
	require.NoError(t, err)
}
