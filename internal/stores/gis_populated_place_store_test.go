package stores

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestFindPopulatedPlaceByPoint(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gisStore := NewGisStore(db)

		res := test.ReadFileFromDataFs(t, "data/populated_places.json")
		countries := MarshalPopulatedPlaces(t, res)
		err := gisStore.CreateManyPlaces(ctx, countries)
		require.NoError(t, err)

		var seoul, pyongyang, jeju *models.PopulatedPlace
		for _, c := range countries {
			switch c.Name {
			case "Seoul":
				seoul = c
			case "Pyongyang":
				pyongyang = c
			case "Jeju":
				jeju = c
			}
		}
		if seoul == nil || pyongyang == nil || jeju == nil {
			t.Fatal("failed to find populated places")
		}
		tests := []struct {
			name string // description of this test case
			// Named input parameters for target function.
			city    *City
			want    *models.PopulatedPlace
			wantErr bool
		}{
			{
				name:    "Seoul",
				city:    GetNearCityByName(t, "Seoul"),
				want:    seoul,
				wantErr: false,
			},

			{
				name:    "Jeju",
				city:    GetNearCityByName(t, "Jeju"),
				want:    jeju,
				wantErr: false,
			},

			{
				name:    "Pyongyang",
				city:    GetNearCityByName(t, "Pyongyang"),
				want:    pyongyang,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, gotErr := gisStore.FindPopulatedPlaceByPoint(ctx, tt.city.Location)
				if gotErr != nil {
					if !tt.wantErr {
						t.Errorf("FindPopulatedPlaceByPoint() failed: %v", gotErr)
					}
					return
				}
				if tt.wantErr {
					t.Fatal("FindPopulatedPlaceByPoint() succeeded unexpectedly")
				}
				if got == nil {
					t.Errorf("FindPopulatedPlaceByPoint() = %v, want %v", got, nil)
				}
				if got.IsoA2 != tt.want.IsoA2 {
					t.Errorf("FindPopulatedPlaceByPoint() = %v, want %v", got.IsoA2, tt.want.IsoA2)
				}
				if got.Adm0name != tt.want.Adm0name {
					t.Errorf("FindPopulatedPlaceByPoint() = %v, want %v", got.Adm0name, tt.want.Adm0name)
				}
				if got.Name != tt.want.Name {
					t.Errorf("FindPopulatedPlaceByPoint() = %v, want %v", got.Name, tt.want.Name)
				}
			})
		}
	})
}
