package geocoder

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestReverse(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		lon      float64
		lat      float64
		wantJson []byte
		wantErr  bool
	}{
		{
			name: "Seoul Korea",
			lon:  126.99000,
			lat:  37.56000,
			wantJson: []byte(`{
  "place_id": 207412994,
  "licence": "Data © OpenStreetMap contributors, ODbL 1.0. http://osm.org/copyright",
  "osm_type": "way",
  "osm_id": 783518856,
  "lat": "37.5599540",
  "lon": "126.9901069",
  "category": "man_made",
  "type": "bridge",
  "place_rank": 30,
  "importance": 0.0000729800869763793,
  "addresstype": "man_made",
  "name": "",
  "display_name": "Samil-daero, Namhak-dong, Pil-dong, Jung-gu, Seoul, 04553, South Korea",
  "address": {
    "road": "Samil-daero",
    "quarter": "Namhak-dong",
    "suburb": "Pil-dong",
    "borough": "Jung-gu",
    "city": "Seoul",
    "ISO3166-2-lvl4": "KR-11",
    "postcode": "04553",
    "country": "South Korea",
    "country_code": "kr"
  },
  "boundingbox": [
    "37.5598225",
    "37.5600855",
    "126.9899486",
    "126.9902653"
  ]
}`),
			wantErr: false,
		},
		{
			name: "Brea, Ca",
			lon:  -117.86181,
			lat:  33.90950,
			wantJson: []byte(`{
  "place_id": 297956286,
  "licence": "Data © OpenStreetMap contributors, ODbL 1.0. http://osm.org/copyright",
  "osm_type": "way",
  "osm_id": 1028144387,
  "lat": "33.9097948",
  "lon": "-117.8605066",
  "category": "building",
  "type": "yes",
  "place_rank": 30,
  "importance": 0.00005346227206831491,
  "addresstype": "building",
  "name": "",
  "display_name": "2771, Saturn Street, Brea, Orange County, California, 92821, United States",
  "address": {
    "house_number": "2771",
    "road": "Saturn Street",
    "town": "Brea",
    "county": "Orange County",
    "state": "California",
    "ISO3166-2-lvl4": "US-CA",
    "postcode": "92821",
    "country": "United States",
    "country_code": "us"
  },
  "boundingbox": [
    "33.9097511",
    "33.9098387",
    "-117.8608763",
    "-117.8601370"
  ]
}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := Reverse(context.Background(), tt.lon, tt.lat)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Reverse() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Reverse() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			var want ReversePlace
			err := json.Unmarshal(tt.wantJson, &want)
			if err != nil {
				t.Errorf("Reverse() failed: %v", err)
			}
			if !reflect.DeepEqual(got.Address, want.Address) {
				t.Errorf("Reverse().Address = %v, want %v", got.Address, want.Address)
			}
		})
	}
}
