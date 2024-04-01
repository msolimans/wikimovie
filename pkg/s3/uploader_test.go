package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/stretchr/testify/require"
)

func Test_CreateUploader(t *testing.T) {
	//since we don't want to require real environment - use random endpoint
	client := CreateS3Client(&aws.Config{Endpoint: aws.String("http://localhost.doesnotexist:8080")})
	uploader := CreateUploader(client)
	//check that it satisfies interface
	require.Implements(t, (*s3manageriface.UploaderAPI)(nil), uploader)
}
