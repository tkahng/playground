package geocoder

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGetLocationFromIp(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		ip       string
		wantJson []byte
		wantErr  bool
	}{
		{
			name: "San Jose",
			ip:   "185.81.124.193",
			wantJson: []byte(`
			{
				"query": "185.81.124.193",
				"status": "success",
				"country": "United States",
				"countryCode": "US",
				"region": "CA",
				"regionName": "California",
				"city": "San Jose",
				"zip": "95141",
				"lat": 37.3388,
				"lon": -121.8916,
				"timezone": "America/Los_Angeles",
				"isp": "Datacamp Limited",
				"org": "Packethub S.A",
				"as": "AS212238 Datacamp Limited"
			}
				`),
			wantErr: false,
		},
		{
			name: "Seoul",
			ip:   "160.238.37.79",
			wantJson: []byte(`
						{
						"status": "success",
						"country": "South Korea",
						"countryCode": "KR",
						"region": "11",
						"regionName": "Seoul",
						"city": "Seoul",
						"zip": "04524",
						"lat": 37.5658,
						"lon": 126.978,
						"timezone": "Asia/Seoul",
						"isp": "PacketHub S.A.",
						"org": "Packethub S.A",
						"as": "AS147049 PacketHub S.A.",
						"query": "160.238.37.79"
						}
				`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := GetLocationFromIp(t.Context(), tt.ip)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetLocationFromIp() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetLocationFromIp() succeeded unexpectedly")
			}
			var want IpLocationResponse
			err := json.Unmarshal(tt.wantJson, &want)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetLocationFromIp() failed: %v", err)
				}
			}
			if !reflect.DeepEqual(*got, want) {
				t.Errorf("GetLocationFromIp() = %v, want %v", got, want)
			}
		})
	}
}
