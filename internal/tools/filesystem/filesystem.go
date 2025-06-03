package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/shared"
)

type StorageClient interface {
	PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
}

type PresignClient interface {
	PresignGetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type S3FileSystem struct {
	httpClient    HttpRequestDoer
	storageClient StorageClient
	presignClient PresignClient
	cfg           conf.StorageConfig
}

func (fs *S3FileSystem) PutFile(ctx context.Context, authority string, key string, file io.Reader) error {

	_, err := fs.storageClient.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(fs.cfg.BucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}

func NewFileSystem(cfg conf.StorageConfig) (FileSystem, error) {
	newConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.ClientId, cfg.ClientSecret, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := awss3.NewFromConfig(newConfig, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(cfg.EndpointUrl)
		o.UsePathStyle = true
	})

	presignClient := awss3.NewPresignClient(client)
	httpClient := http.DefaultClient
	return &S3FileSystem{
		storageClient: client,
		cfg:           cfg,
		presignClient: presignClient,
		httpClient:    httpClient,
	}, nil
}

func (fs *S3FileSystem) GeneratePresignedURL(ctx context.Context, bucket, key string) (string, error) {

	presignResult, err := fs.presignClient.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, awss3.WithPresignExpires(10*time.Minute))

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
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

		for i, c := range word {
			if unicode.IsUpper(c) && i > 0 &&
				// is not a following uppercase character
				!unicode.IsUpper(rune(word[i-1])) {
				result.WriteString("_")
			}

			result.WriteRune(c)
		}
	}

	return strings.ToLower(result.String())
}

func (fs *S3FileSystem) PutNewFileFromURL(ctx context.Context, url string) (*shared.FileDto, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 399 {
		return nil, fmt.Errorf("failed to download url %s (%d)", url, res.StatusCode)
	}

	var buf bytes.Buffer

	if _, err = io.Copy(&buf, res.Body); err != nil {
		return nil, err
	}

	return fs.PutFileFromBytes(ctx, buf.Bytes(), path.Base(url))
}

func (fs *S3FileSystem) PutFileFromBytes(ctx context.Context, b []byte, name string) (*shared.FileDto, error) {
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

	dto := &shared.FileDto{
		ID:           id,
		Disk:         fs.cfg.BucketName,
		Directory:    path.Dir(key),
		Filename:     path.Base(key),
		OriginalName: name,
		Extension:    ext,
		MimeType:     mime,
		Size:         int64(size),
	}
	return dto, nil
}
