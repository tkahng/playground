package stores

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestFindCountryByPoint(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gisStore := NewGisStore(db)

		res := test.ReadFileFromDataFs(t, "data/countries.json")
		countries := MarshalCountries(t, res)
		err := gisStore.CreateManyCountries(ctx, countries)
		require.NoError(t, err)

		var sk, nk *models.Country
		for _, c := range countries {
			switch c.ISOA2 {
			case "KR":
				sk = c
			case "KP":
				nk = c
			}
		}
		if sk == nil || nk == nil {
			t.Fatal("failed to find SK or NK")
		}
		tests := []struct {
			name string // description of this test case
			// Named input parameters for target function.
			city    *City
			want    *models.Country
			wantErr bool
		}{
			{
				name:    "Seoul",
				city:    GetCityByName(t, "Seoul"),
				want:    sk,
				wantErr: false,
			},

			{
				name:    "Jeju",
				city:    GetCityByName(t, "Jeju"),
				want:    sk,
				wantErr: false,
			},

			{
				name:    "Pyongyang",
				city:    GetCityByName(t, "Pyongyang"),
				want:    nk,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, gotErr := gisStore.FindCountryByPoint(ctx, tt.city.Location)
				if gotErr != nil {
					if !tt.wantErr {
						t.Errorf("FindCountryByPoint() failed: %v", gotErr)
					}
					return
				}
				if tt.wantErr {
					t.Fatal("FindCountryByPoint() succeeded unexpectedly")
				}
				if got == nil {
					t.Errorf("FindCountryByPoint() = %v, want %v", got, nil)
				}
				if got.ISOA2 != tt.want.ISOA2 {
					t.Errorf("FindCountryByPoint() = %v, want %v", got.ISOA2, tt.want.ISOA2)
				}
				if got.ISOA3 != tt.want.ISOA3 {
					t.Errorf("FindCountryByPoint() = %v, want %v", got.ISOA3, tt.want.ISOA3)
				}
				if got.Name != tt.want.Name {
					t.Errorf("FindCountryByPoint() = %v, want %v", got.Name, tt.want.Name)
				}
			})
		}
	})
}
