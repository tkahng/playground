package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StripeMetadata map[string]string

func (h *StripeMetadata) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte")
	}

	return json.Unmarshal(bytes, &h)
}

func (h StripeMetadata) Value() (driver.Value, error) {
	return json.Marshal(h)
}
