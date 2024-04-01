package s3

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// DownloadManager interface is the minimum interface required from a s3 download manager
type DownloadManager interface {
	DownloadWithContext(aws.Context, io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)
}

func CreateDownloader(client s3iface.S3API) s3manageriface.DownloaderAPI {
	return s3manager.NewDownloaderWithClient(client)
}
