package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/errors"

	"golang-boilerplate/internal/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Adapter struct {
	config *config.Config
	client *s3.Client
	bucket string
}

// NewS3Adapter creates a new S3 adapter instance implementing storage.StorageAdapter
func NewS3Adapter(config *config.Config) (*S3Adapter, error) {
	ctx := context.Background()

	// Load AWS config
	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(config.S3Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(config.S3AccessKey, config.S3SecretKey, ""),
		),
	)
	if err != nil {
		return nil, errors.ExternalServiceError("failed to load AWS config", err).
			WithOperation("load_aws_config").
			WithResource("storage")
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &S3Adapter{
		config: config,
		client: client,
		bucket: config.S3Bucket,
	}, nil
}

func (a *S3Adapter) UploadFile(ctx context.Context, file *multipart.FileHeader, key string) (*UploadResult, error) {
	f, err := file.Open()
	if err != nil {
		logger.Sugar.Errorf("failed to open file: %v", err)
		return nil, errors.ExternalServiceError("failed to open file", err).
			WithOperation("open_file").
			WithResource("storage")
	}
	defer f.Close()

	// Use S3 uploader for efficient uploads
	uploader := manager.NewUploader(a.client)

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		logger.Sugar.Errorf("failed to upload to S3: %v", err)
		return nil, errors.ExternalServiceError("failed to upload to S3", err).
			WithOperation("upload_to_s3").
			WithResource("storage")
	}

	url := a.GetObjectURL(key)

	return &UploadResult{
		URL:      url,
		Key:      key,
		Bucket:   a.bucket,
		Location: url,
	}, nil
}

func (a *S3Adapter) GetObjectURL(key string) string {
	// Public URL pattern for S3
	// Format: https://<bucket>.s3.<region>.amazonaws.com/<key>
	// Or: https://s3.<region>.amazonaws.com/<bucket>/<key>
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", a.bucket, a.config.S3Region, key)
}

func (a *S3Adapter) GetPresignedURL(ctx context.Context, key string, duration ...time.Duration) (string, error) {
	expiry := a.config.S3PresignedURLDuration
	if len(duration) > 0 && duration[0] > 0 {
		expiry = duration[0]
	}

	// Create presigner
	presigner := s3.NewPresignClient(a.client)

	// Create the request
	request, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		logger.Sugar.Errorf("failed to generate presigned URL: %v", err)
		return "", errors.ExternalServiceError("failed to generate presigned URL", err).
			WithOperation("generate_presigned_url").
			WithResource("storage")
	}

	return request.URL, nil
}

func (a *S3Adapter) UploadFiles(ctx context.Context, files []*multipart.FileHeader) (*BatchUploadResult, error) {
	result := &BatchUploadResult{Files: make([]UploadResult, 0, len(files))}
	type uploadResult struct {
		result *UploadResult
		err    error
	}

	ch := make(chan uploadResult, len(files))

	for _, fh := range files {
		go func(fh *multipart.FileHeader) {
			res, err := a.UploadFile(ctx, fh, fh.Filename)
			ch <- uploadResult{res, err}
		}(fh)
	}

	for i := 0; i < len(files); i++ {
		r := <-ch
		if r.err != nil {
			return nil, errors.ExternalServiceError("failed to upload file", r.err).
				WithOperation("upload_file").
				WithResource("storage")
		}
		result.Files = append(result.Files, *r.result)
	}

	return result, nil
}
