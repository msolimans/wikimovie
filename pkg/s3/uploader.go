package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// UploadManager interface is the minimum interface required from a s3 upload manager
type UploadManager interface {
	UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

func CreateUploader(client s3iface.S3API) s3manageriface.UploaderAPI {
	return s3manager.NewUploaderWithClient(client)
}
