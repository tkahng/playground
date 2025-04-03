package utils

import (
	"encoding/json"
	"fmt"
)

func MarshalJSONByte[T any](v T) []byte {
	jsonData, _ := json.Marshal(v)
	return jsonData
}

// func MarshalJSON[T any](v T) string {
// 	jsonData, _ := json.MarshalIndent(v, "", "  ")
// 	return string(jsonData)
// }

func UnmarshalJSON[T any](r []byte) (T, error) {
	var v T
	if err := json.Unmarshal(r, &v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func PrettyPrintJSON[T any](v T) {
	jsonData, _ := json.MarshalIndent(v, "", "  ")
	println(string(jsonData))
}
