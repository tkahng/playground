package models

import (
	"encoding/json"
	"errors"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

type PopulatedPlace struct {
	_          struct{}    `db:"populated_places" schema:"gis" json:"-"`
	GID        int         `json:"gid" db:"gid,pk"`
	Geom       *geom.Point `json:"-" db:"geom"` // ST_Point(lon, lat)
	Scalerank  int8        `json:"scalerank" db:"scalerank"`
	Labelrank  int8        `json:"labelrank" db:"labelrank"`
	Featurecla string      `json:"featurecla" db:"featurecla"`
	Name       string      `json:"name" db:"name"`
	Nameascii  string      `json:"nameascii" db:"nameascii"`
	Sov0name   string      `json:"sov0name" db:"sov0name"`
	Adm0name   string      `json:"adm0name" db:"adm0name"`
	Adm0A3     string      `json:"adm0_a3" db:"adm0_a3"`
	Adm1name   string      `json:"adm1name" db:"adm1name"`
	IsoA2      string      `json:"iso_a2" db:"iso_a2"`
	PopMax     int64       `json:"pop_max" db:"pop_max"`
	MinZoom    float32     `json:"min_zoom" db:"min_zoom"`
}

func (w *PopulatedPlace) MarshalJSON() ([]byte, error) {
	geometry, err := geojson.Marshal(w.Geom)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&struct {
		GID        int             `json:"gid" db:"gid"`
		Geometry   json.RawMessage `json:"geometry"`
		Scalerank  int8            `json:"scalerank" db:"scalerank"`
		Labelrank  int8            `json:"labelrank" db:"labelrank"`
		Featurecla string          `json:"featurecla" db:"featurecla"`
		Name       string          `json:"name" db:"name"`
		Nameascii  string          `json:"nameascii" db:"nameascii"`
		Sov0name   string          `json:"sov0name" db:"sov0name"`
		Adm0name   string          `json:"adm0name" db:"adm0name"`
		Adm0A3     string          `json:"adm0_a3" db:"adm0_a3"`
		Adm1name   string          `json:"adm1name" db:"adm1name"`
		IsoA2      string          `json:"iso_a2" db:"iso_a2"`
		PopMax     int64           `json:"pop_max" db:"pop_max"`
		MinZoom    float32         `json:"min_zoom" db:"min_zoom"`
	}{
		GID:        w.GID,
		Geometry:   geometry,
		Scalerank:  w.Scalerank,
		Labelrank:  w.Labelrank,
		Featurecla: w.Featurecla,
		Name:       w.Name,
		Nameascii:  w.Nameascii,
		Sov0name:   w.Sov0name,
		Adm0name:   w.Adm0name,
		Adm0A3:     w.Adm0A3,
		Adm1name:   w.Adm1name,
		IsoA2:      w.IsoA2,
		PopMax:     w.PopMax,
		MinZoom:    w.MinZoom,
	})
}

func (w *PopulatedPlace) UnmarshalJSON(data []byte) error {
	if w == nil {
		return errors.New("models.Country: UnmarshalJSON on nil pointer")
	}
	var country struct {
		GID        int             `json:"gid" db:"gid"`
		Geometry   json.RawMessage `json:"geometry"`
		Scalerank  int8            `json:"scalerank" db:"scalerank"`
		Labelrank  int8            `json:"labelrank" db:"labelrank"`
		Featurecla string          `json:"featurecla" db:"featurecla"`
		Name       string          `json:"name" db:"name"`
		Nameascii  string          `json:"nameascii" db:"nameascii"`
		Sov0name   string          `json:"sov0name" db:"sov0name"`
		Adm0name   string          `json:"adm0name" db:"adm0name"`
		Adm0A3     string          `json:"adm0_a3" db:"adm0_a3"`
		Adm1name   string          `json:"adm1name" db:"adm1name"`
		IsoA2      string          `json:"iso_a2" db:"iso_a2"`
		PopMax     int64           `json:"pop_max" db:"pop_max"`
		MinZoom    float32         `json:"min_zoom" db:"min_zoom"`
	}
	if err := json.Unmarshal(data, &country); err != nil {
		return err
	}
	var geo geom.T
	if err := geojson.Unmarshal(country.Geometry, &geo); err != nil {
		return err
	}
	point, ok := geo.(*geom.Point)
	if !ok {
		return errors.New("geometry is not a point")
	}
	w.GID = country.GID
	w.Scalerank = country.Scalerank
	w.Labelrank = country.Labelrank
	w.Featurecla = country.Featurecla
	w.Name = country.Name
	w.Nameascii = country.Nameascii
	w.Sov0name = country.Sov0name
	w.Adm0name = country.Adm0name
	w.Adm0A3 = country.Adm0A3
	w.Adm1name = country.Adm1name
	w.IsoA2 = country.IsoA2
	w.PopMax = country.PopMax
	w.MinZoom = country.MinZoom
	w.Geom = point
	return nil
}
