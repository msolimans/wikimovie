package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
)

func Test_CreateS3Client(t *testing.T) {
	//since we don't want to require real environment - use random endpoint
	client := CreateS3Client(&aws.Config{Endpoint: aws.String("http://localhost.doesnotexist:8080")})
	//interface check
	require.Implements(t, (*s3iface.S3API)(nil), client)
}
