package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/msolimans/wikimovie/cmd"
	"github.com/msolimans/wikimovie/pkg/appconf"
	"github.com/msolimans/wikimovie/pkg/s3"

	"github.com/sirupsen/logrus"
)

func loadConfig() *appconf.Configuration {
	conf := &appconf.Configuration{}
	err := appconf.LoadConfig(".", conf)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	return conf
}

func uploadFile(cfg *appconf.Configuration, fixtureFileName, bucket, key string) error {
	file, err := os.ReadFile(filepath.Join("./cmd/seed/fixtures", fixtureFileName))
	if err != nil {
		return err
	}

	client := s3.CreateS3Client(&aws.Config{
		Endpoint: cfg.Aws.Endpoint,
		Region:   cfg.Aws.Region,
	})
	uploader := s3manager.NewUploaderWithClient(client)
	_, err = uploader.UploadWithContext(context.Background(), &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(file),
	})
	return err
}

func main() {
	logger := logrus.New()

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		logger.Fatal("what type of seed you want to do here? pass one of (s3 - sqs)")
	}
	cfg := loadConfig()

	switch args[0] {
	case "s3":
		filename := "1900s.json"
		if len(args) > 1 && args[1] != "" {
			filename = args[1]
		}
		key := strings.Join([]string{cmd.S3KeyPrefix, cmd.S3KeyNewMoviesprefix, filename}, "/")
		if err := uploadFile(cfg, filename, cfg.Bucket, key); err != nil {
			logger.WithError(err).Fatal("execution failed")
		}
	default:
		logger.Fatalf("unsupported option %s", args[0])
	}
	os.Exit(0)
}
