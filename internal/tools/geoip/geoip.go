package geoip

import (
	"embed"
	"fmt"
	"net/netip"

	"github.com/oschwald/geoip2-golang/v2"
)

func openFs(f embed.FS, path string) (*geoip2.Reader, error) {
	data, err := f.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return geoip2.FromBytes(data)
}

func City(ip string) (*geoip2.City, error) {
	parsedIp, err := netip.ParseAddr(ip)
	if err != nil {
		return nil, fmt.Errorf("error parsing ip: %w", err)
	}
	db, err := openFs(DataFs, "data/GeoLite2-City.mmdb")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	city, err := db.City(parsedIp)
	if err != nil {
		return nil, err
	}
	return city, nil
}
