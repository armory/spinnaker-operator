package version

import (
	"io/ioutil"
)

var (
	Version = getVersion()
)

func getVersion() string {
	version, err := ioutil.ReadFile("./operator-version")
	if err != nil {
		return "Unknown"
	}
	return string(version)
}
