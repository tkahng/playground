package filesystem

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/tkahng/playground/internal/conf"
)

const (
	minioImage       = "minio/minio:RELEASE.2024-01-16T16-07-38Z"
	testBucketName   = "test-bucket"
	testBucketRegion = "us-east-1"
)

// WithMinioContainer starts a MinIO container, creates a test bucket, builds a
// real FileSystem wired to it, and calls fn. The container is terminated when
// the test ends. Use test.SkipIfShort to skip in -short mode.
func WithMinioContainer(t testing.TB, fn func(ctx context.Context, fs FileSystem, cfg conf.StorageConfig)) {
	t.Helper()
	ctx := context.Background()

	ctr, err := tcminio.Run(ctx, minioImage)
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("start MinIO container: %v", err)
	}

	host, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("MinIO connection string: %v", err)
	}
	endpoint := "http://" + host

	cfg := conf.StorageConfig{
		ClientId:      ctr.Username,
		ClientSecret:  ctr.Password,
		BucketName:    testBucketName,
		EndpointUrl:   endpoint,
		Region:        testBucketRegion,
		PublicBaseURL: endpoint + "/" + testBucketName,
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.ClientId, cfg.ClientSecret, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		t.Fatalf("build AWS config for MinIO: %v", err)
	}

	s3c := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(cfg.EndpointUrl)
		o.UsePathStyle = true
	})
	if _, err = s3c.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(cfg.BucketName),
	}); err != nil {
		t.Fatalf("create test bucket: %v", err)
	}

	// Allow anonymous reads so PublicURL-based fetches work in tests.
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"s3:GetObject","Resource":"arn:aws:s3:::` + testBucketName + `/*"}]}`
	if _, err = s3c.PutBucketPolicy(ctx, &awss3.PutBucketPolicyInput{
		Bucket: aws.String(cfg.BucketName),
		Policy: aws.String(policy),
	}); err != nil {
		t.Fatalf("set bucket public-read policy: %v", err)
	}

	fs, err := NewFileSystem(ctx, cfg)
	if err != nil {
		t.Fatalf("build filesystem: %v", err)
	}

	fn(ctx, fs, cfg)
}
