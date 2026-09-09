package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioResumeStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioResumeStorage(ctx context.Context, endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioResumeStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinioResumeStorage{client: client, bucket: bucket}, nil
}

func (s *MinioResumeStorage) Save(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, filename, content, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return filename, nil
}

func (s *MinioResumeStorage) Get(ctx context.Context, objectPath string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, nil
}
