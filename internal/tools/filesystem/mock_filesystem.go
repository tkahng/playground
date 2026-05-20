package filesystem

import (
	"context"
	"io"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tkahng/playground/internal/conf"
)

type StorageClientDecorator struct {
	StorageClientFunc func() StorageClient
	PutObjectFunc     func(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
}

func (s *StorageClientDecorator) PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
	if s.PutObjectFunc != nil {
		return s.PutObjectFunc(ctx, params, optFns...)
	}
	return s.StorageClientFunc().PutObject(ctx, params, optFns...)
}

func NewMockFileSystem(cfg conf.StorageConfig) FileSystem {
	return &S3FileSystemDecorator{Delegate: &S3FileSystem{
		cfg: cfg,
	}}
}

type S3FileSystemDecorator struct {
	Delegate              *S3FileSystem
	PutFileFunc           func(ctx context.Context, authority string, key string, file io.Reader) error
	PutFileFromBytesFunc  func(ctx context.Context, b []byte, name string) (*FileDto, error)
	PutNewFileFromURLFunc func(ctx context.Context, url string) (*FileDto, error)
	PublicURLFunc         func(key string) string
	StorageClientFunc     func() StorageClient
	HttpClientFunc        func() HttpRequestDoer
}

// PutFile implements FileSystem.
func (s *S3FileSystemDecorator) PutFile(ctx context.Context, authority string, key string, file io.Reader) error {
	if s.PutFileFunc != nil {
		return s.PutFileFunc(ctx, authority, key, file)
	}
	return s.Delegate.PutFile(ctx, authority, key, file)
}

// PutFileFromBytes implements FileSystem.
func (s *S3FileSystemDecorator) PutFileFromBytes(ctx context.Context, b []byte, name string) (*FileDto, error) {
	if s.PutFileFromBytesFunc != nil {
		return s.PutFileFromBytesFunc(ctx, b, name)
	}
	return s.Delegate.PutFileFromBytes(ctx, b, name)
}

// PutNewFileFromURL implements FileSystem.
func (s *S3FileSystemDecorator) PutNewFileFromURL(ctx context.Context, url string) (*FileDto, error) {
	if s.PutNewFileFromURLFunc != nil {
		return s.PutNewFileFromURLFunc(ctx, url)
	}
	return s.Delegate.PutNewFileFromURL(ctx, url)
}

// PublicURL implements FileSystem.
func (s *S3FileSystemDecorator) PublicURL(key string) string {
	if s.PublicURLFunc != nil {
		return s.PublicURLFunc(key)
	}
	return s.Delegate.PublicURL(key)
}
