package secrets

import (
	"context"
	"fmt"
	yamlParse "gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"strings"
)

const (
	encryptedPrefix     = "encrypted:"
	encryptedFilePrefix = "encryptedFile:"
)

var Engines = map[string]func(context.Context, bool, string) (Decrypter, error){
	"gcs":             NewGcsDecrypter,
	"noop":            NewNoopDecrypter,
	"s3":              NewS3Decrypter,
	"secrets-manager": NewAwsSecretsManagerDecrypter,
}

type Decrypter interface {
	Decrypt() (string, error)
	IsFile() bool
}

func IsEncryptedSecret(val string) bool {
	return strings.HasPrefix(val, encryptedPrefix) ||
		strings.HasPrefix(val, encryptedFilePrefix)
}

func NewDecrypter(ctx context.Context, encryptedSecret string) (Decrypter, error) {
	e, isFile, params := GetEngine(encryptedSecret)
	if e == "" {
		return &NoopDecrypter{value: encryptedSecret}, nil
	}
	engine, ok := Engines[e]
	if !ok {
		return nil, fmt.Errorf("secret engine %s not registered", e)
	}
	return engine(ctx, isFile, params)
}

// GetEngine returns the name of the engine if recognized,
// the remainder of the parameters (unparsed) and a boolean that indicates
// if the user requested a file.
func GetEngine(encryptedSecret string) (string, bool, string) {
	isFile := false
	prefixLen := 0
	if strings.HasPrefix(encryptedSecret, encryptedPrefix) {
		prefixLen = len(encryptedPrefix)
	} else if strings.HasPrefix(encryptedSecret, encryptedFilePrefix) {
		prefixLen = len(encryptedFilePrefix)
		isFile = true
	}
	idx := strings.Index(encryptedSecret, "!")
	if idx == -1 {
		return "", false, ""
	}
	return encryptedSecret[prefixLen:idx], isFile, encryptedSecret[idx+1:]
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

func ToTempFile(content []byte) (string, error) {
	f, err := ioutil.TempFile("", "secret-")
	if err != nil {
		return "", err
	}
	defer f.Close()

	f.Write([]byte(content))
	return f.Name(), nil
}
