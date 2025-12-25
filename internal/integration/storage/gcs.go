package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/errors"

	"golang-boilerplate/internal/logger"

	gcstorage "cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type GCSAdapter struct {
	config *config.Config
	client *gcstorage.Client
	bucket *gcstorage.BucketHandle
}

// NewGCSAdapter creates a new GCS adapter instance implementing storage.Adapter
func NewGCSAdapter(config *config.Config) (*GCSAdapter, error) {
	ctx := context.Background()
	var client *gcstorage.Client
	var err error

	// Prefer credentials JSON path when provided
	if config.GCSCredentialsJSONPath != "" {
		if err := config.PopulateFromJSON(config.GCSCredentialsJSONPath); err != nil {
			return nil, errors.ExternalServiceError("failed to load GCS credentials", err).
				WithOperation("load_gcs_credentials").
				WithResource("storage")
		}
		client, err = gcstorage.NewClient(ctx, option.WithCredentialsFile(config.GCSCredentialsJSONPath))
		if err != nil {
			return nil, errors.ExternalServiceError("failed to create GCS client", err).
				WithOperation("create_gcs_client").
				WithResource("storage")
		}
	} else {
		// Fallback to ADC or environment
		creds, err := google.FindDefaultCredentials(ctx, gcstorage.ScopeReadWrite)
		if err != nil {
			return nil, errors.ExternalServiceError("failed to get default credentials", err).
				WithOperation("get_default_credentials").
				WithResource("storage")
		}
		client, err = gcstorage.NewClient(ctx, option.WithCredentials(creds))
		if err != nil {
			return nil, errors.ExternalServiceError("failed to create GCS client", err).
				WithOperation("create_gcs_client").
				WithResource("storage")
		}
	}

	bucket := client.Bucket(config.GCSBucket)

	return &GCSAdapter{
		config: config,
		client: client,
		bucket: bucket,
	}, nil
}

func (a *GCSAdapter) UploadFile(ctx context.Context, file *multipart.FileHeader, key string) (*UploadResult, error) {
	f, err := file.Open()
	if err != nil {
		logger.Sugar.Errorf("failed to open file: %v", err)
		return nil, errors.ExternalServiceError("failed to open file", err).
			WithOperation("open_file").
			WithResource("storage")
	}
	defer f.Close()

	obj := a.bucket.Object(key)
	wc := obj.NewWriter(ctx)

	if _, err := io.Copy(wc, f); err != nil {
		_ = wc.Close()
		logger.Sugar.Errorf("failed to write to GCS: %v", err)
		return nil, errors.ExternalServiceError("failed to write to GCS", err).
			WithOperation("write_to_gcs").
			WithResource("storage")
	}

	if err := wc.Close(); err != nil {
		logger.Sugar.Errorf("failed to close writer: %v", err)
		return nil, errors.ExternalServiceError("failed to close writer", err).
			WithOperation("close_writer").
			WithResource("storage")
	}

	url := a.GetObjectURL(key)

	return &UploadResult{URL: url, Key: key, Bucket: a.config.GCSBucket, Location: url}, nil
}

func (a *GCSAdapter) GetObjectURL(key string) string {
	// Public URL pattern (object must be public or via signed URL for access)
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", a.config.GCSBucket, key)
}

func (a *GCSAdapter) GetPresignedURL(ctx context.Context, key string, duration ...time.Duration) (string, error) {
	expiry := a.config.GCSPresignedURLDuration
	if len(duration) > 0 && duration[0] > 0 {
		expiry = duration[0]
	}

	// Read service account JSON key
	keyJSON, err := os.ReadFile(a.config.GCSCredentialsJSONPath)
	if err != nil {
		logger.Sugar.Errorf("failed to read service account JSON: %v", err)
		return "", errors.ExternalServiceError("failed to read service account JSON", err).
			WithOperation("read_service_account_json").
			WithResource("storage")
	}

	// Parse service account JSON
	conf, err := google.JWTConfigFromJSON(keyJSON)
	if err != nil {
		logger.Sugar.Errorf("failed to parse service account JSON: %v", err)
		return "", errors.ExternalServiceError("failed to parse service account JSON", err).
			WithOperation("parse_service_account_json").
			WithResource("storage")
	}

	opts := &gcstorage.SignedURLOptions{
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Method:         "GET",
		Expires:        time.Now().Add(expiry),
	}

	url, err := gcstorage.SignedURL(a.config.GCSBucket, key, opts)
	if err != nil {
		logger.Sugar.Errorf("failed to sign GCS URL: %v", err)
		return "", errors.ExternalServiceError("failed to sign GCS URL", err).
			WithOperation("sign_gcs_url").
			WithResource("storage")
	}

	return url, nil
}

func (a *GCSAdapter) UploadFiles(ctx context.Context, files []*multipart.FileHeader) (*BatchUploadResult, error) {
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
