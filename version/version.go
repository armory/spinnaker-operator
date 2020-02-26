package version

import (
	"io/ioutil"
	"strings"
)

var Version string

func init() {
	b, err := ioutil.ReadFile("./operator-version")
	if err != nil {
		Version = "Unknown"
	}
	Version = strings.TrimSpace(string(b))
}
