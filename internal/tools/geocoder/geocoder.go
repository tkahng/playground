package geocoder

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/tkahng/playground/internal/tools/ttlmap"
	"golang.org/x/time/rate"
)

type Location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}
type Coordinates struct {
	Latitude  float64 `json:"latitude" required:"true"`
	Longitude float64 `json:"longitude" required:"true"`
}
type GeocodingInput struct {
	IP          string       `json:"ip" required:"false"`
	Coordinates *Coordinates `json:"coordinates" required:"false"`
}
type Geocoder interface {
	GetLocation(ctx context.Context, input GeocodingInput) (*Location, error)
}

type GeocoderImpl struct {
	reverseLimiter *rate.Limiter
	ipLimiter      *rate.Limiter
	cache          *ttlmap.TTLMap
}

var _ Geocoder = (*GeocoderImpl)(nil)

func NewGeocoder() *GeocoderImpl {
	m := ttlmap.New(100, 10)
	return &GeocoderImpl{
		cache: m,
		// reverse limit 45 requests per minute.
		reverseLimiter: rate.NewLimiter(rate.Every(1*time.Minute), 45),
		ipLimiter:      rate.NewLimiter(rate.Every(1*time.Second), 1),
	}
}

func quantizeCoord(lat, lon float64, precision int) (float64, float64) {
	factor := math.Pow10(precision)
	// truncate; use math.Round(...) if you prefer rounding
	qLat := math.Trunc(lat*factor) / factor
	qLon := math.Trunc(lon*factor) / factor
	return qLat, qLon
}

func cacheKeyCoordinates(lat, lon float64, precision int) string {
	qLat, qLon := quantizeCoord(lat, lon, precision)
	return fmt.Sprintf("%.6f,%.6f", qLat, qLon)
}

func cacheKeyIP(ip string) string {
	return ip
}

// GetLocation implements [Geocoder].
func (g *GeocoderImpl) GetLocation(ctx context.Context, input GeocodingInput) (*Location, error) {
	if input.Coordinates != nil {
		key := cacheKeyCoordinates(input.Coordinates.Latitude, input.Coordinates.Longitude, 3)
		if v, ok := g.cache.Get(key); ok {
			return v.(*Location), nil
		}
		if g.reverseLimiter.Allow() {
			place, err := Reverse(ctx, input.Coordinates.Longitude, input.Coordinates.Latitude)
			if err != nil {
				return nil, err
			}
			address := place.Address
			if address.CountryCode == "" {
				return nil, errors.New("country code is empty")
			}
			location := &Location{
				Country: strings.ToUpper(address.CountryCode),
			}

			switch {
			// municipality, city, town, village
			case address.City != "":
				location.City = address.City
			case address.Town != "":
				location.City = address.Town
			case address.Village != "":
				location.City = address.Village
			}
			if location.City == "" {
				return nil, errors.New("city is empty")
			}
			g.cache.Put(key, location)
			return location, nil
		}
	}

	if g.ipLimiter.Allow() {
		key := cacheKeyIP(input.IP)
		if v, ok := g.cache.Get(key); ok {
			return v.(*Location), nil
		}
		place, err := GetLocationFromIp(ctx, input.IP)
		if err != nil {
			return nil, err
		}
		location := &Location{
			Country: place.CountryCode,
			City:    place.City,
		}
		g.cache.Put(key, location)
		return location, nil
	}

	if input.IP != "" {
		city, err := City(input.IP)
		if err != nil {
			return nil, err
		}
		location := &Location{}
		location.City = city.City.Names.English
		location.Country = city.Country.ISOCode
		return location, nil
	}

	return nil, errors.New("unable to get location")
}
