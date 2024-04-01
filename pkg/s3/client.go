package s3

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Client interface is the minimum interface required from a s3 implementation
type Client interface {
	CopyObjectWithContext(aws.Context, *s3.CopyObjectInput, ...request.Option) (*s3.CopyObjectOutput, error)
	DeleteObjectWithContext(aws.Context, *s3.DeleteObjectInput, ...request.Option) (*s3.DeleteObjectOutput, error)
	HeadObjectWithContext(aws.Context, *s3.HeadObjectInput, ...request.Option) (*s3.HeadObjectOutput, error)
	ListObjectsV2WithContext(aws.Context, *s3.ListObjectsV2Input, ...request.Option) (*s3.ListObjectsV2Output, error)
}

func CreateS3Client(awsConfigs ...*aws.Config) s3iface.S3API {
	awsSession := session.Must(session.NewSession(&aws.Config{
		HTTPClient: &http.Client{
			Timeout: time.Second * 15,
		},
	}))

	return s3.New(awsSession, awsConfigs...)

	// sess := session.Must(session.NewSession(awsConfigs))

	// return s3.New(sess)

}
