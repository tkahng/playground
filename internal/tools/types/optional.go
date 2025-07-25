package types

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

type OptionalParam[T any] struct {
	Value T
	IsSet bool
}

// Define schema to use wrapped type
func (o OptionalParam[T]) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(o.Value))
}

// Expose wrapped value to receive parsed value from Huma
// MUST have pointer receiver
func (o *OptionalParam[T]) Receiver() reflect.Value {
	return reflect.ValueOf(o).Elem().Field(0)
}

// React to request param being parsed to update internal state
// MUST have pointer receiver
func (o *OptionalParam[T]) OnParamSet(isSet bool, parsed any) {
	o.IsSet = isSet
}

// Define request input with the wrapper type
type MyRequestInput struct {
	MaybeText OptionalParam[string] `query:"text"`
}
