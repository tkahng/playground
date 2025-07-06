package filesystem

import (
	"context"
	"io"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tkahng/authgo/internal/conf"
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

// type Mock

func NewMockFileSystem(cfg conf.StorageConfig) FileSystem {
	return &S3FileSystemDecorator{Delegate: &S3FileSystem{
		cfg: cfg,
	}}
}

type S3FileSystemDecorator struct {
	Delegate                 *S3FileSystem
	GeneratePresignedURLFunc func(ctx context.Context, bucket string, key string) (string, error)
	PutFileFunc              func(ctx context.Context, authority string, key string, file io.Reader) error
	PutFileFromBytesFunc     func(ctx context.Context, b []byte, name string) (*FileDto, error)
	PutNewFileFromURLFunc    func(ctx context.Context, url string) (*FileDto, error)
	StorageClientFunc        func() StorageClient
	PresignClientFunc        func() PresignClient
	HttpClientFunc           func() HttpRequestDoer
}

// GeneratePresignedURL implements FileSystem.
func (s *S3FileSystemDecorator) GeneratePresignedURL(ctx context.Context, bucket string, key string) (string, error) {
	if s.GeneratePresignedURLFunc != nil {
		return s.GeneratePresignedURLFunc(ctx, bucket, key)
	}
	return s.Delegate.GeneratePresignedURL(ctx, bucket, key)
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
