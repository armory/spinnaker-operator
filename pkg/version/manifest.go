package version

import (
	"io/ioutil"
	"os"
	"strings"
)

const (
	operatorHomePath = "OPERATOR_HOME"
	manifestFile     = "MANIFEST"
)

var manifest = make(map[string]string)

func init() {
	// read MANIFEST file, that contains Operator Version Information
	// absolute path: $(OPERATOR_HOME)/MANIFEST
	path := os.Getenv(operatorHomePath)
	body, err := ioutil.ReadFile(path + "/" + manifestFile)

	if err != nil {
		return
	}

	raws := strings.Split(string(body), "\n")

	for _, raw := range raws {
		if raw != "" {
			values := strings.Split(raw, "=")
			if len(values) == 2 {
				manifest[values[0]] = values[1]
			}
		}
	}
}
