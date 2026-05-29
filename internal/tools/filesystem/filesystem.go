package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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

// AllowedMIMETypes is the set of MIME types accepted for upload.
// SVG is intentionally excluded: it supports embedded scripts and is unsafe
// when served from the same origin as the app.
var AllowedMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"image/avif":      true,
	"image/heic":      true,
	"image/heif":      true,
	"video/mp4":       true,
	"video/webm":      true,
	"video/ogg":       true,
	"audio/mpeg":      true,
	"audio/ogg":       true,
	"audio/wav":       true,
	"audio/webm":      true,
	"application/pdf": true,
	"text/plain":      true,
	"text/csv":        true,
}

// privateIPBlocks holds all IP ranges that are non-routable / internal.
var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // loopback
		"::1/128",        // IPv6 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // link-local (AWS metadata, etc.)
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local
		"100.64.0.0/10",  // CGNAT
		"::ffff:0:0/96",  // IPv4-mapped IPv6
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// validateRemoteURL rejects non-http/https schemes and URLs that resolve to
// private or loopback addresses (SSRF protection).
func validateRemoteURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL scheme %q not allowed; only http and https are accepted", u.Scheme)
	}
	host := u.Hostname()
	ips, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("cannot resolve host %q: %w", host, err)
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if isPrivateIP(ip) {
			return fmt.Errorf("URL resolves to a private or reserved address: %s", ipStr)
		}
	}
	return nil
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

func (fs *S3FileSystem) PutNewFileFromURL(ctx context.Context, rawURL string) (*FileDto, error) {
	if err := validateRemoteURL(rawURL); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := fs.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 399 {
		return nil, fmt.Errorf("failed to download url %s (%d)", rawURL, res.StatusCode)
	}

	const maxDownloadBytes = 50 * 1024 * 1024 // 50 MB

	lr := &io.LimitedReader{R: res.Body, N: maxDownloadBytes + 1}
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, lr); err != nil {
		return nil, err
	}
	if int64(buf.Len()) > maxDownloadBytes {
		return nil, fmt.Errorf("remote file at %s exceeds the %d MB size limit",
			rawURL, maxDownloadBytes/(1024*1024))
	}

	return fs.PutFileFromBytes(ctx, buf.Bytes(), path.Base(rawURL))
}

func (fs *S3FileSystem) PutFileFromBytes(ctx context.Context, b []byte, name string) (*FileDto, error) {
	id := uuid.New()
	size := len(b)
	if size == 0 {
		return nil, errors.New("cannot create an empty file")
	}

	// Use mimetype (magic-byte detection) for accurate MIME type identification.
	mt := mimetype.Detect(b)
	mime := mt.String()
	// Strip parameters (e.g. "text/plain; charset=utf-8" → "text/plain")
	if idx := strings.Index(mime, ";"); idx != -1 {
		mime = strings.TrimSpace(mime[:idx])
	}

	if !AllowedMIMETypes[mime] {
		return nil, fmt.Errorf("file type %q is not allowed", mime)
	}

	ext := path.Ext(name)
	if ext == "" {
		ext = mt.Extension()
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
