package secrets

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
)

type GcsSecret struct {
}

type GcsDecrypter struct {
	bucket   string
	filepath string
	key      string
	ctx      context.Context
	isFile   bool
}

func NewGcsDecrypter(ctx context.Context, isFile bool, params string) (Decrypter, error) {
	gcs := &GcsDecrypter{isFile: isFile, ctx: ctx}
	if err := gcs.parse(params); err != nil {
		return nil, err
	}
	return gcs, nil
}

func (gcs *GcsDecrypter) Decrypt() (string, error) {
	sec, err := gcs.fetchSecret(gcs.ctx)
	if err != nil || !gcs.isFile {
		return sec, err
	}
	return ToTempFile([]byte(sec))
}

func (gcs *GcsDecrypter) IsFile() bool {
	return gcs.isFile
}

func (gcs *GcsDecrypter) parse(params string) error {
	tokens := strings.Split(params, "!")
	for _, element := range tokens {
		kv := strings.Split(element, ":")
		if len(kv) == 2 {
			switch kv[0] {
			case "b":
				gcs.bucket = kv[1]
			case "f":
				gcs.filepath = kv[1]
			case "k":
				gcs.key = kv[1]
			}
		}
	}

	if gcs.bucket == "" {
		return fmt.Errorf("secret format error - 'b' for bucket is required")
	}
	if gcs.filepath == "" {
		return fmt.Errorf("secret format error - 'f' for file is required")
	}
	return nil
}

func (gcs *GcsDecrypter) fetchSecret(ctx context.Context) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to create GCS client: %s", err.Error())
	}
	bucket := client.Bucket(gcs.bucket)
	r, err := bucket.Object(gcs.filepath).NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get reader for bucket: %s, file: %s, error: %v", gcs.bucket, gcs.filepath, err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("unable to download file from bucket: %s, file: %s, error: %v", gcs.bucket, gcs.filepath, err)
	}
	if len(gcs.key) > 0 {
		return parseSecretFile(b, gcs.key)
	}
	return string(b), nil
}
