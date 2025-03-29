package utils

import "encoding/json"

func MarshalJSONByte[T any](v T) []byte {
	jsonData, _ := json.Marshal(v)
	return jsonData
}

func MarshalJSON[T any](v T) string {
	jsonData, _ := json.MarshalIndent(v, "", "  ")
	return string(jsonData)
}

func PrettyPrintJSON[T any](v T) {
	jsonData, _ := json.MarshalIndent(v, "", "  ")
	println(string(jsonData))
}
