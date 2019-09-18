package secrets

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	yamlParse "gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

const (
	MaxApiRetry = 10
)

type S3Secret struct {
	region   string
	bucket   string
	filepath string
	key      string
}

type S3Decrypter struct {
	params map[string]string
}

func NewS3Decrypter(params map[string]string) *S3Decrypter {
	return &S3Decrypter{params}
}

func (s3 *S3Decrypter) Decrypt() (string, error) {
	s3Secret, err := ParseS3Secret(s3.params)
	if err != nil {
		return "", err
	}
	return s3Secret.fetchSecret()
}

func ParseS3Secret(params map[string]string) (S3Secret, error) {
	var s3Secret S3Secret

	region, ok := params["r"]
	if !ok {
		return S3Secret{}, fmt.Errorf("secret format error - 'r' for region is required")
	}
	s3Secret.region = region

	bucket, ok := params["b"]
	if !ok {
		return S3Secret{}, fmt.Errorf("secret format error - 'b' for bucket is required")
	}
	s3Secret.bucket = bucket

	filepath, ok := params["f"]
	if !ok {
		return S3Secret{}, fmt.Errorf("secret format error - 'f' for file is required")
	}
	s3Secret.filepath = filepath

	key, ok := params["k"]
	if ok {
		s3Secret.key = key
	}

	return s3Secret, nil
}

func (secret *S3Secret) fetchSecret() (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String(secret.region),
		MaxRetries: aws.Int(MaxApiRetry),
	})
	if err != nil {
		return "", err
	}

	downloader := s3manager.NewDownloader(sess)

	contents := aws.NewWriteAtBuffer([]byte{})
	size, err := downloader.Download(contents,
		&s3.GetObjectInput{
			Bucket: aws.String(secret.bucket),
			Key:    aws.String(secret.filepath),
		})
	if err != nil {
		return "", fmt.Errorf("unable to download item %q: %v", secret.filepath, err)
	}
	if size == 0 {
		return "", fmt.Errorf("file %q empty", secret.filepath)
	}

	if len(secret.key) > 0 {
		bytes := contents.Bytes()
		return parseSecretFile(bytes, secret.key)
	}

	return string(contents.Bytes()), nil
}

func parseSecretFile(fileContents []byte, key string) (string, error) {
	m := make(map[interface{}]interface{})
	if err := yamlParse.Unmarshal(fileContents, &m); err != nil {
		return "", err
	}

	for _, yamlKey := range strings.Split(key, ".") {
		switch s := m[yamlKey].(type) {
		case map[interface{}]interface{}:
			m = s
		case string:
			return s, nil
		case nil:
			return "", fmt.Errorf("error parsing secret file: couldn't find key %q in yaml", key)
		default:
			return "", fmt.Errorf("error parsing secret file: unknown type %q with value %q",
				reflect.TypeOf(s), s)
		}
	}

	return "", fmt.Errorf("error parsing secret file for key %q", key)
}
