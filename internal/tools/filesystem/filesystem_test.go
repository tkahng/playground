package filesystem_test

import (
	"bytes"
	"context"
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
		dto, err := fs.PutFileFromBytes(ctx, []byte("public url test"), "pub.txt")
		require.NoError(t, err)

		key := dto.Directory + "/" + dto.Filename
		url := fs.PublicURL(key)
		assert.Contains(t, url, dto.Filename)
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

// makePNGBytes returns a minimal valid PNG header followed by padding.
func makePNGBytes() []byte {
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return append(header, make([]byte, 100)...)
}
