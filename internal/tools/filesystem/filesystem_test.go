//go:build integration

package filesystem_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/tools/filesystem"
)

// skipIfShort skips the test when -short is passed; MinIO container tests are slow.
func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping container test in short mode")
	}
}

func TestFilesystem_PutFileFromBytes(t *testing.T) {
	skipIfShort(t)
	filesystem.WithMinioContainer(t, func(ctx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		content := []byte("hello minio")
		dto, err := fs.PutFileFromBytes(ctx, content, "hello.txt")
		require.NoError(t, err)

		assert.Equal(t, cfg.BucketName, dto.Disk)
		assert.Equal(t, ".txt", dto.Extension)
		assert.Equal(t, int64(len(content)), dto.Size)
		assert.Equal(t, "hello.txt", dto.OriginalName)
		assert.NotEmpty(t, dto.Filename)
		assert.NotEmpty(t, dto.Directory)
		assert.NotEqual(t, dto.ID, "")
	})
}

func TestFilesystem_PutFileFromBytes_Empty(t *testing.T) {
	skipIfShort(t)
	filesystem.WithMinioContainer(t, func(ctx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		_, err := fs.PutFileFromBytes(ctx, []byte{}, "empty.txt")
		assert.Error(t, err, "empty file should be rejected before upload")
	})
}

func TestFilesystem_PutFile(t *testing.T) {
	skipIfShort(t)
	filesystem.WithMinioContainer(t, func(ctx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		content := []byte("raw put content")
		key := "media/raw-test.txt"
		err := fs.PutFile(ctx, cfg.BucketName, key, bytes.NewReader(content))
		require.NoError(t, err)
	})
}

func TestFilesystem_PublicURL(t *testing.T) {
	skipIfShort(t)
	filesystem.WithMinioContainer(t, func(ctx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		content := []byte("public url test content")
		dto, err := fs.PutFileFromBytes(ctx, content, "pub.txt")
		require.NoError(t, err)

		key := dto.Directory + "/" + dto.Filename
		url := fs.PublicURL(key)

		// URL must be exactly PublicBaseURL/key — no signing params.
		assert.Equal(t, cfg.PublicBaseURL+"/"+key, url)

		// File must be reachable without credentials.
		resp, err := http.Get(url) //nolint:noctx
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		got, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, content, got)
	})
}

func TestFilesystem_PutFileFromBytes_NoExtension(t *testing.T) {
	skipIfShort(t)
	filesystem.WithMinioContainer(t, func(ctx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		// PNG magic bytes — mime detection should kick in when no ext given.
		png := makePNGBytes()
		dto, err := fs.PutFileFromBytes(ctx, png, "image-no-ext")
		require.NoError(t, err)
		assert.Equal(t, ".png", dto.Extension)
		assert.Contains(t, dto.MimeType, "image/png")
	})
}

func TestFilesystem_MultipleUploads_UniqueKeys(t *testing.T) {
	skipIfShort(t)
	filesystem.WithMinioContainer(t, func(ctx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		content := []byte("same content")
		dto1, err := fs.PutFileFromBytes(ctx, content, "file.txt")
		require.NoError(t, err)

		dto2, err := fs.PutFileFromBytes(ctx, content, "file.txt")
		require.NoError(t, err)

		// Each upload gets a unique UUID-based key even with the same filename.
		assert.NotEqual(t, dto1.Filename, dto2.Filename)
	})
}

func TestFilesystem_DeleteObject(t *testing.T) {
	skipIfShort(t)
	filesystem.WithMinioContainer(t, func(ctx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		content := []byte("delete me")
		dto, err := fs.PutFileFromBytes(ctx, content, "delete.txt")
		require.NoError(t, err)

		// File should be reachable before deletion.
		resp, err := http.Get(fs.PublicURL(dto.StorageKey)) //nolint:noctx
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Delete the object.
		require.NoError(t, fs.DeleteObject(ctx, dto.StorageKey))

		// File should no longer be reachable (MinIO returns 403 for public buckets
		// when the object is absent, rather than 404, so we just check != 200).
		resp2, err := http.Get(fs.PublicURL(dto.StorageKey)) //nolint:noctx
		require.NoError(t, err)
		resp2.Body.Close()
		assert.NotEqual(t, http.StatusOK, resp2.StatusCode, "deleted object should not be accessible")
	})
}

// makePNGBytes returns a minimal valid PNG header followed by padding.
func makePNGBytes() []byte {
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return append(header, make([]byte, 100)...)
}
