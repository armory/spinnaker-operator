package secrets

type nopDecrypter struct {
	params map[string]string
}

func (n *nopDecrypter) Decrypt() (string, error) {
	return n.params["v"], nil
}

func NewNop(params map[string]string) *nopDecrypter {
	return &nopDecrypter{params}
}
