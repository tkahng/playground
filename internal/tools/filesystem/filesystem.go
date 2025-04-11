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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/c2fo/vfs/v7/backend/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
)

type FileSystem struct {
	client *awss3.Client
	fs     *s3.FileSystem
	cfg    conf.StorageConfig
}

func NewFileSystem(cfg conf.StorageConfig) (*FileSystem, error) {
	config, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.ClientId, cfg.ClientSecret, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := awss3.NewFromConfig(config, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(cfg.EndpointUrl)
		o.UsePathStyle = true
	})
	bucketAuth := s3.NewFileSystem(s3.WithOptions(s3.Options{
		AccessKeyID:     cfg.ClientId,
		SecretAccessKey: cfg.ClientSecret,
		Region:          cfg.Region,
		Endpoint:        cfg.EndpointUrl,
		ForcePathStyle:  true,
	}), s3.WithClient(client))

	return &FileSystem{
		client: client,
		fs:     bucketAuth,
		cfg:    cfg,
	}, nil
}

func (fs *FileSystem) GeneratePresignedURL(ctx context.Context, bucket, key string) (string, error) {
	client := fs.client

	presignClient := awss3.NewPresignClient(client)

	presignResult, err := presignClient.PresignGetObject(ctx, &awss3.GetObjectInput{
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

func (fs *FileSystem) NewFileFromURL(ctx context.Context, url string) (*FileDto, error) {
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

	return fs.NewFileFromBytes(buf.Bytes(), path.Base(url))
}

type FileDto struct {
	ID           uuid.UUID `json:"id"`
	Disk         string    `db:"disk" json:"disk"`
	Directory    string    `db:"directory" json:"directory"`
	Filename     string    `db:"filename" json:"filename"`
	OriginalName string    `db:"original_name" json:"original_name"`
	Extension    string    `db:"extension" json:"extension"`
	MimeType     string    `db:"mime_type" json:"mime_type"`
	Size         int64     `db:"size" json:"size"`
}

func (fs *FileSystem) NewFileFromBytes(b []byte, name string) (*FileDto, error) {
	id := uuid.New()
	size := len(b)
	if size == 0 {
		return nil, errors.New("cannot create an empty file")
	}
	mime := http.DetectContentType(b)
	ext := path.Ext(name)
	if ext == "" {
		ext = mimetype.Detect(b).Extension()
	}
	key := "media/" + id.String() + ext

	f, err := fs.fs.NewFile(fs.cfg.BucketName, "/"+key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ok, err := f.Exists()
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errors.New("file already exists")
	}

	if _, err := f.Write(b); err != nil {
		return nil, err
	}
	dto := &FileDto{
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
