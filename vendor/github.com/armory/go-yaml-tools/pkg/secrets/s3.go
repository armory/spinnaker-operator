package secrets

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"strings"
)

const (
	MaxApiRetry = 10
)

type S3Decrypter struct {
	region   string
	bucket   string
	filepath string
	key      string
	isFile   bool
}

func NewS3Decrypter(ctx context.Context, isFile bool, params string) (Decrypter, error) {
	s3 := &S3Decrypter{isFile: isFile}
	if err := s3.parse(params); err != nil {
		return nil, err
	}
	return s3, nil
}

func (s3 *S3Decrypter) Decrypt() (string, error) {
	sec, err := s3.fetchSecret()
	if err != nil || !s3.isFile {
		return sec, err
	}
	return ToTempFile([]byte(sec))
}

func (s3 *S3Decrypter) IsFile() bool {
	return s3.isFile
}

func (s3 *S3Decrypter) parse(params string) error {
	tokens := strings.Split(params, "!")
	for _, element := range tokens {
		kv := strings.Split(element, ":")
		if len(kv) == 2 {
			switch kv[0] {
			case "r":
				s3.region = kv[1]
			case "b":
				s3.bucket = kv[1]
			case "f":
				s3.filepath = kv[1]
			case "k":
				s3.key = kv[1]
			}
		}
	}

	if s3.region == "" {
		return fmt.Errorf("secret format error - 'r' for region is required")
	}
	if s3.bucket == "" {
		return fmt.Errorf("secret format error - 'b' for bucket is required")
	}
	if s3.filepath == "" {
		return fmt.Errorf("secret format error - 'f' for file is required")
	}
	return nil
}

func (s3 *S3Decrypter) fetchSecret() (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String(s3.region),
		MaxRetries: aws.Int(MaxApiRetry),
	})
	if err != nil {
		return "", err
	}

	downloader := s3manager.NewDownloader(sess)

	contents := aws.NewWriteAtBuffer([]byte{})
	size, err := downloader.Download(contents,
		&awss3.GetObjectInput{
			Bucket: aws.String(s3.bucket),
			Key:    aws.String(s3.filepath),
		})
	if err != nil {
		return "", fmt.Errorf("unable to download item %q: %v", s3.filepath, err)
	}
	if size == 0 {
		return "", fmt.Errorf("file %q empty", s3.filepath)
	}

	if len(s3.key) > 0 {
		bytes := contents.Bytes()
		return parseSecretFile(bytes, s3.key)
	}

	return string(contents.Bytes()), nil
}
