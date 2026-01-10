package models

import (
	"encoding/json"
	"errors"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

type Country struct {
	_     struct{} `db:"countries" schema:"gis" json:"-"`
	GID   int64    `db:"gid" json:"gid"`
	Name  string   `db:"name" json:"name"`
	ISOA2 string   `db:"iso_a2_eh" json:"iso_a2_eh"`
	ISOA3 string   `db:"iso_a3_eh" json:"iso_a3_eh"`
	Geom  geom.T   `db:"geom" json:"-"`
}

func (w *Country) MarshalJSON() ([]byte, error) {
	geometry, err := geojson.Marshal(w.Geom)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&struct {
		GID      int64           `json:"gid"`
		Name     string          `json:"name"`
		ISOA2    string          `json:"iso_a2_eh"`
		ISOA3    string          `json:"iso_a3_eh"`
		Geometry json.RawMessage `json:"geometry"`
	}{
		GID:      w.GID,
		Name:     w.Name,
		ISOA2:    w.ISOA2,
		ISOA3:    w.ISOA3,
		Geometry: geometry,
	})
}

func (w *Country) UnmarshalJSON(data []byte) error {
	if w == nil {
		return errors.New("models.Country: UnmarshalJSON on nil pointer")
	}
	var country struct {
		GID      int64           `json:"gid"`
		Name     string          `json:"name"`
		ISOA2    string          `json:"iso_a2_eh"`
		ISOA3    string          `json:"iso_a3_eh"`
		Geometry json.RawMessage `json:"geometry"`
	}
	if err := json.Unmarshal(data, &country); err != nil {
		return err
	}
	var geom geom.T
	if err := geojson.Unmarshal(country.Geometry, &geom); err != nil {
		return err
	}
	w.GID = country.GID
	w.Name = country.Name
	w.ISOA2 = country.ISOA2
	w.ISOA3 = country.ISOA3
	w.Geom = geom
	return nil
}
