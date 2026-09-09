package config

import (
	"errors"
	"os"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type NatsConfig struct {
	URL           string
	StreamName    string
	StreamSubject string
}

type DBConfig struct {
	DSN string
}

func LoadMinioConfig() (*MinioConfig, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, errors.New("MINIO_ENDPOINT environment variable not set")
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		return nil, errors.New("MINIO_ACCESS_KEY environment variable not set")
	}
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("MINIO_SECRET_KEY environment variable not set")
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		return nil, errors.New("MINIO_BUCKET environment variable not set")
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	return &MinioConfig{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		UseSSL:    useSSL,
	}, nil
}

func LoadNatsConfig() (*NatsConfig, error) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		return nil, errors.New("NATS_URL environment variable not set")
	}
	streamName := os.Getenv("NATS_STREAM_NAME")
	if streamName == "" {
		return nil, errors.New("NATS_STREAM_NAME environment variable not set")
	}
	streamSubject := os.Getenv("NATS_STREAM_SUBJECT")
	if streamSubject == "" {
		return nil, errors.New("NATS_STREAM_SUBJECT environment variable not set")
	}
	return &NatsConfig{
		URL:           url,
		StreamName:    streamName,
		StreamSubject: streamSubject,
	}, nil
}

func LoadDBConfig() (*DBConfig, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}
	return &DBConfig{DSN: dsn}, nil
}
