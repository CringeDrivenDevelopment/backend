package service

import (
	"backend/cmd/app"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type S3Service struct {
	client *minio.Client

	bucketName string
}

func NewS3Service(app *app.App) (*S3Service, error) {
	client, err := minio.New(app.Settings.MinioHost, &minio.Options{
		Creds:  credentials.NewStaticV4(app.Settings.MinioAccessKey, app.Settings.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), app.Settings.MinioBucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(context.Background(), app.Settings.MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &S3Service{
		client:     client,
		bucketName: app.Settings.MinioBucket,
	}, nil
}

func (s *S3Service) DeletePhoto(ctx context.Context, filename string) error {
	opts := minio.RemoveObjectOptions{
		ForceDelete: true,
	}

	err := s.client.RemoveObject(ctx, s.bucketName, filename, opts)
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Service) UploadPhoto(ctx context.Context, filename string, data io.Reader) (minio.UploadInfo, error) {
	uploadInfo, err := s.client.PutObject(
		ctx,
		s.bucketName,
		filename,
		data,
		-1,
		minio.PutObjectOptions{
			ContentType:  "application/octet-stream",
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		},
	)

	return uploadInfo, err
}
