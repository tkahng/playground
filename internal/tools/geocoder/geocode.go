package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	Place struct {
		PlaceId     int `json:"place_id"`
		Licence     string
		OsmType     string `json:"osm_type"`
		OsmId       int    `json:"osm_id"`
		Lat         string
		Lon         string
		DisplayName string `json:"display_name"`
		PlaceRank   int    `json:"place_rank"`
		Category    string
		Type        string
		Importance  float64
		Icon        string
	}
	Address struct {
		Road         string `json:"road"`
		County       string `json:"county"`
		Town         string `json:"town"`
		State        string `json:"state"`
		City         string `json:"city"`
		Municipality string `json:"municipality"`
		Village      string `json:"village"`
		ISO31662Lvl4 string `json:"ISO3166-2-lvl4"`
		Postcode     string `json:"postcode"`
		Country      string `json:"country"`
		CountryCode  string `json:"country_code"`
	}
	ReversePlace struct {
		Place
		AddressType string `json:"addresstype"`
		Name        string
		Address     Address
	}
)

func Reverse(ctx context.Context, lon, lat float64) (*ReversePlace, error) {
	uri := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=jsonv2", lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept-language", "en")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var place ReversePlace
	if err := json.NewDecoder(res.Body).Decode(&place); err != nil {
		return nil, err
	}
	return &place, nil
}
