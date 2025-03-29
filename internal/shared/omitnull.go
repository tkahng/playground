package shared

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

var JSONNull = []byte("null")

// type Omit[T any] = omit.Val[T]

// Schema returns a schema representing this value on the wire.
// It returns the schema of the contained type.
// func (o Omit[T]) Schema(r huma.Registry) *huma.Schema {
// 	return r.Schema(reflect.TypeOf(o.Value), true, "")
// }

// OmitNull is a field which can be omitted from the input,
// set to `null`, or set to a value. Each state is tracked and can
// be checked for in handling code.
type OmitNull[T any] struct {
	Sent  bool
	Null  bool
	Value T
}

func (o *OmitNull[T]) IsUnset() bool { return !o.Sent }

func (o *OmitNull[T]) IsSet() bool { return o.Sent }

func From[T any](val T) OmitNull[T] {
	return OmitNull[T]{
		Value: val,
		Sent:  true,
	}
}

func (o OmitNull[T]) MustGet() T { return o.Value } //MustGet()

// UnmarshalJSON unmarshals this value from JSON input.
func (o *OmitNull[T]) UnmarshalJSON(b []byte) error {
	if len(b) > 0 {
		o.Sent = true
		if bytes.Equal(b, []byte("null")) {
			o.Null = true
			return nil
		}
		return json.Unmarshal(b, &o.Value)
	}
	return nil
}
func (v *OmitNull[T]) MarshalJSON() ([]byte, error) {

	if v.Sent && !v.Null {
		return json.Marshal(v.Value)
	}

	return JSONNull, nil
}

// Schema returns a schema representing this value on the wire.
// It returns the schema of the contained type.
func (o OmitNull[T]) Schema(r huma.Registry) *huma.Schema {
	return r.Schema(reflect.TypeOf(o.Value), true, "")
}
