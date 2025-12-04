package userreaction

import (
	"context"
	"log/slog"
	"strings"

	"github.com/tkahng/playground/internal/tools/geocoder"
	"github.com/tkahng/playground/internal/tools/geoip"
)

type Location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

func GetLocationFromBody(ctx context.Context, lon, lat float64) *Location {
	place, err := geocoder.Reverse(ctx, lon, lat)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"there was an error during reverse geocoding.",
			slog.Any("error", err),
			slog.Float64("latitude", lat),
			slog.Float64("longitude", lon),
		)
		return nil
	}
	address := place.Address
	if address.CountryCode == "" {
		return nil
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
		return nil
	}
	return location
}

func GetLocationFromIp(ctx context.Context, ip string) *Location {
	address, err := geoip.GetLocationFromIp(ctx, ip)

	if err != nil {
		slog.ErrorContext(
			ctx,
			"error getting city",
			slog.String("ip", ip),
			slog.Any("error", err),
		)
		return nil
	}

	return &Location{
		Country: address.CountryCode,
		City:    address.City,
	}
}
