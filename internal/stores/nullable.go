package stores

import (
	"encoding/json"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// NullableUUID distinguishes between a JSON field being absent (Set=false)
// and explicitly null or a UUID value (Set=true).  Use it on PATCH DTOs where
// omitting a field means "leave unchanged" and sending null means "clear".
type NullableUUID struct {
	Set   bool
	Value *uuid.UUID
}

func (n *NullableUUID) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var id uuid.UUID
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	n.Value = &id
	return nil
}

func (NullableUUID) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		OneOf: []*huma.Schema{
			{Type: "string", Format: "uuid"},
			{Type: "null"},
		},
	}
}

var _ huma.SchemaProvider = NullableUUID{}
var _ json.Unmarshaler = (*NullableUUID)(nil)

// NullableString distinguishes between a JSON field being absent (Set=false)
// and explicitly null or a string value (Set=true).
type NullableString struct {
	Set   bool
	Value *string
}

func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s
	return nil
}

func (NullableString) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		OneOf: []*huma.Schema{
			{Type: "string"},
			{Type: "null"},
		},
	}
}

var _ huma.SchemaProvider = NullableString{}
var _ json.Unmarshaler = (*NullableString)(nil)
