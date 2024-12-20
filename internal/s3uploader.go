package internal

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Uploader struct {
	bucket   string
	uploader *manager.Uploader
}

func LoadConfig() (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
	)
}

func NewS3Uploader(bucket string, cfg aws.Config) (*S3Uploader, error) {
	awsS3APIEndpoint := os.Getenv("AWS_S3_API_ENDPOINT")
	var client *s3.Client
	if awsS3APIEndpoint == "" {
		client = s3.NewFromConfig(cfg)
	} else {
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(awsS3APIEndpoint)
		})
	}
	// it may require 2GB of memory to run
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 512 * 1024 * 1024 // 512 MiB
		u.Concurrency = 4
	})
	return &S3Uploader{
		bucket:   bucket,
		uploader: uploader,
	}, nil
}

func (u *S3Uploader) Upload(reader io.Reader, filename string) error {
	ctx := context.TODO()

	progressReader := &ProgressReader{
		reader: reader,
	}

	result, err := u.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(filename),
		Body:   progressReader,
	})
	if err != nil {
		return err
	}
	log.Println("Uploaded", filename, result.Location)
	return nil
}
