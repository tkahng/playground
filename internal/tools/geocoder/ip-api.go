package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type IpLocationResponse struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
}

// GetLocationFromIp returns location from ip
// it uses ip-api with free plan.
// rate limit is 1 request per second
func GetLocationFromIp(ctx context.Context, ip string) (*IpLocationResponse, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	if res.StatusCode > 300 {
		return nil, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	var response IpLocationResponse

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}
