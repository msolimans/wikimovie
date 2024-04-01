package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/require"
)

func Test_CreateDownloader(t *testing.T) {
	client := CreateS3Client(&aws.Config{Endpoint: aws.String("http://localhost.doesnotexist:8080")})
	downloader := CreateDownloader(client)
	require.Implements(t, (*s3manageriface.DownloaderAPI)(nil), downloader)
}
