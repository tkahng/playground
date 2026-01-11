package test

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed data
var DataFs embed.FS

func ReadFileFromDataFs(t testing.TB, path string) []byte {
	res, err := DataFs.ReadFile(path)
	require.NoError(t, err, "There was an error reading the file at path: "+path)
	return res
}
