package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const NotFoundErrorCode = "NotFound"

type UploadRequest struct {
	Bucket      string
	Key         string
	ContentType string
	Body        io.Reader
}

type ManagerAPI interface {
	DownloadObject(ctx context.Context, bucket, key string) ([]byte, error)
	ListObjects(ctx context.Context, bucket, prefix string) ([]*s3.Object, error)
	MoveObject(ctx context.Context, bucket, sourceKey, destinationKey string) error
	ObjectExists(ctx context.Context, bucket, key string) (bool, error)
	UploadObject(ctx context.Context, request UploadRequest) (string, error)
}

type manager struct {
	client     Client
	downloader DownloadManager
	uploader   UploadManager
}

func NewManager(client Client, downloader DownloadManager, uploader UploadManager) ManagerAPI {
	return &manager{
		client:     client,
		downloader: downloader,
		uploader:   uploader,
	}
}

func NewS3Manager(cfg *aws.Config) ManagerAPI {
	client := CreateS3Client(cfg)
	downloader := CreateDownloader(client)
	uploader := CreateUploader(client)
	return NewManager(client, downloader, uploader)
}

// DownloadObject will download a specified s3 file and return a []byte
func (m *manager) DownloadObject(ctx context.Context, bucket, key string) ([]byte, error) {
	buff := &aws.WriteAtBuffer{}
	if _, err := m.downloader.DownloadWithContext(ctx, buff, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

// ListObjects will return a []*s3.Object found inside a bucket
func (m *manager) ListObjects(ctx context.Context, bucket, prefix string) ([]*s3.Object, error) {
	output, err := m.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}
	return output.Contents, nil
}

// MoveObject moves object within the same bucket from sourceKey to destinationKey
func (m *manager) MoveObject(ctx context.Context, bucket, sourceKey, destinationKey string) error {
	copySource := bucket + "/" + sourceKey
	if _, err := m.client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     &bucket,
		CopySource: &copySource,
		Key:        &destinationKey,
	}); err != nil {
		return err
	}
	_, err := m.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &sourceKey,
	})
	return err
}

func (m *manager) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	if _, err := m.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == NotFoundErrorCode {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UploadObject uploads file to S3
func (m *manager) UploadObject(ctx context.Context, request UploadRequest) (string, error) {
	var location string
	result, err := m.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Body:        request.Body,
		Bucket:      &request.Bucket,
		ContentType: &request.ContentType,
		Key:         &request.Key,
	})
	if result != nil {
		location = result.Location
	}
	return location, err
}
