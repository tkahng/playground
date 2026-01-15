package geocoder

import "github.com/twpayne/go-geom"

func PointFromLonLat(lon, lat float64) *geom.Point {
	return geom.NewPoint(geom.XY).MustSetCoords(
		// lon,lat
		geom.Coord{lon, lat},
	).SetSRID(4326)
}
