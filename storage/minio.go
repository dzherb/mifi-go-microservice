package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type MiniIO[T any] struct {
	log        *slog.Logger
	client     *minio.Client
	bucketName string
}

type MiniIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

func NewMiniIO[T any](
	ctx context.Context,
	log *slog.Logger,
	cfg MiniIOConfig,
) (*MiniIO[T], error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MiniIO[T]{
		log:        log,
		client:     minioClient,
		bucketName: cfg.BucketName,
	}

	// Создаем bucket если его нет
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(
			ctx,
			cfg.BucketName,
			minio.MakeBucketOptions{},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		log.Info("bucket created", slog.String("bucket_name", cfg.BucketName))
	}

	return storage, nil
}

func (s *MiniIO[T]) Set(ctx context.Context, key string, data T) error {
	dataSerialized, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling data for key %s: %w", key, err)
	}

	reader := bytes.NewReader(dataSerialized)

	_, err = s.client.PutObject(
		ctx,
		s.bucketName,
		key,
		reader,
		int64(len(dataSerialized)),
		minio.PutObjectOptions{
			ContentType: "application/json",
		},
	)
	if err != nil {
		return fmt.Errorf("saving data to s3 with key %s: %w", key, err)
	}

	s.log.Info(
		"data saved to s3",
		slog.String("bucket_name", s.bucketName),
		slog.String("key", key),
	)

	return nil
}

func (s *MiniIO[T]) Get(ctx context.Context, key string) (T, error) {
	var res T

	object, err := s.client.GetObject(
		ctx,
		s.bucketName,
		key,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return res, fmt.Errorf("failed to get object from s3: %w", err)
	}

	defer func() {
		err := object.Close()
		if err != nil {
			s.log.Error("failed to close s3 object", slog.String("key", key))
		}
	}()

	err = json.NewDecoder(object).Decode(&res)
	if err != nil {
		var minioErr minio.ErrorResponse

		if errors.As(err, &minioErr) && minioErr.Code == minio.NoSuchKey {
			return res, ErrKeyNotFound
		}

		return res, fmt.Errorf("failed to unmarshal s3 object: %w", err)
	}

	return res, nil
}

func (s *MiniIO[T]) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(
		ctx,
		s.bucketName,
		key,
		minio.RemoveObjectOptions{},
	)

	if err != nil {
		return fmt.Errorf("failed to remove s3 object: %w", err)
	}

	return nil
}
