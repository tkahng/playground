package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/conf"
)

type StorageClient interface {
	PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
}

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type S3FileSystem struct {
	httpClient    HttpRequestDoer
	storageClient StorageClient
	cfg           conf.StorageConfig
}

func (fs *S3FileSystem) PutFile(ctx context.Context, authority string, key string, file io.Reader) error {
	_, err := fs.storageClient.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(authority),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}

func NewFileSystem(ctx context.Context, cfg conf.StorageConfig) (FileSystem, error) {
	newConfig, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.ClientId, cfg.ClientSecret, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}

	client := awss3.NewFromConfig(newConfig, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(cfg.EndpointUrl)
		o.UsePathStyle = true
	})

	return &S3FileSystem{
		storageClient: client,
		cfg:           cfg,
		httpClient:    http.DefaultClient,
	}, nil
}

func (fs *S3FileSystem) DeleteObject(ctx context.Context, key string) error {
	_, err := fs.storageClient.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(fs.cfg.BucketName),
		Key:    aws.String(key),
	})
	return err
}

func (fs *S3FileSystem) publicBase() string {
	if fs.cfg.PublicBaseURL != "" {
		return strings.TrimRight(fs.cfg.PublicBaseURL, "/")
	}
	// Derive from endpoint + bucket when STORAGE_PUBLIC_BASE_URL is not set.
	return strings.TrimRight(fs.cfg.EndpointUrl, "/") + "/" + fs.cfg.BucketName
}

func (fs *S3FileSystem) PublicURL(key string) string {
	return fs.publicBase() + "/" + key
}

var snakecaseSplitRegex = regexp.MustCompile(`[\W_]+`)

func Snakecase(str string) string {
	var result strings.Builder

	// split at any non word character and underscore
	words := snakecaseSplitRegex.Split(str, -1)

	for _, word := range words {
		if word == "" {
			continue
		}

		if result.Len() > 0 {
			result.WriteString("_")
		}

		var prev rune
		for i, c := range word {
			if unicode.IsUpper(c) && i > 0 && !unicode.IsUpper(prev) {
				result.WriteString("_")
			}
			result.WriteRune(c)
			prev = c
		}
	}

	return strings.ToLower(result.String())
}

func (fs *S3FileSystem) PutNewFileFromURL(ctx context.Context, url string) (*FileDto, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := fs.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 399 {
		return nil, fmt.Errorf("failed to download url %s (%d)", url, res.StatusCode)
	}

	const maxDownloadBytes = 50 * 1024 * 1024 // 50 MB

	lr := &io.LimitedReader{R: res.Body, N: maxDownloadBytes + 1}
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, lr); err != nil {
		return nil, err
	}
	if int64(buf.Len()) > maxDownloadBytes {
		return nil, fmt.Errorf("remote file at %s exceeds the %d MB size limit",
			url, maxDownloadBytes/(1024*1024))
	}

	return fs.PutFileFromBytes(ctx, buf.Bytes(), path.Base(url))
}

func (fs *S3FileSystem) PutFileFromBytes(ctx context.Context, b []byte, name string) (*FileDto, error) {
	id := uuid.New()
	size := len(b)
	if size == 0 {
		return nil, errors.New("cannot create an empty file")
	}
	mime := http.DetectContentType(b)
	ext := path.Ext(name)
	if ext == "" {
		mt := mimetype.Detect(b)
		ext = mt.Extension()
		mime = mt.String()
	}
	key := "media/" + id.String() + ext

	err := fs.PutFile(ctx, fs.cfg.BucketName, key, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	dto := &FileDto{
		ID:           id,
		StorageKey:   key,
		PublicURL:    fs.PublicURL(key),
		MimeType:     mime,
		Size:         int64(size),
		OriginalName: name,
		Extension:    ext,
		Disk:         fs.cfg.BucketName,
		Directory:    path.Dir(key),
		Filename:     path.Base(key),
	}
	return dto, nil
}
