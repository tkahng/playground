package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type (
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
		PlaceID     int64   `json:"place_id"`
		Licence     string  `json:"licence"`
		OsmType     string  `json:"osm_type"`
		OsmID       int64   `json:"osm_id"`
		Lat         string  `json:"lat"`
		Lon         string  `json:"lon"`
		Category    string  `json:"category"`
		Type        string  `json:"type"`
		PlaceRank   int64   `json:"place_rank"`
		Importance  float64 `json:"importance"`
		Addresstype string  `json:"addresstype"`
		Name        string  `json:"name"`
		DisplayName string  `json:"display_name"`
		Address     Address `json:"address"`
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
		return nil, fmt.Errorf("error getting response: %w", err)
	}
	if res.StatusCode > 300 {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		slog.ErrorContext(ctx, "error getting response", slog.String("body", string(bodyBytes)))
		return nil, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	var place ReversePlace
	if err := json.NewDecoder(res.Body).Decode(&place); err != nil {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		slog.ErrorContext(ctx, "error decoding response", slog.String("body", string(bodyBytes)))
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &place, nil
}
