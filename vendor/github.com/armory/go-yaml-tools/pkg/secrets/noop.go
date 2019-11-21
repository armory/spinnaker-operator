package secrets

import "context"

func NewNoopDecrypter(ctx context.Context, isFile bool, params string) (Decrypter, error) {
	return &NoopDecrypter{
		value:  params,
		isFile: isFile,
	}, nil
}

type NoopDecrypter struct {
	value  string
	isFile bool
}

func (n *NoopDecrypter) Decrypt() (string, error) {
	if n.isFile {
		return ToTempFile([]byte(n.value))
	}
	return n.value, nil
}

func (n *NoopDecrypter) ParseTokens(secret string) {
	n.value = secret[len("encrypted:noop!v:"):]
}

func (n *NoopDecrypter) IsFile() bool {
	return n.isFile
}
