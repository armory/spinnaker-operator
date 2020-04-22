package accounts

import (
	"github.com/blang/semver"
	"strings"
)

func IsDynamicAccountSupported(version string) bool {
	v, err := semver.Make(version)
	if err != nil {
		return false
	}
	min, err := semver.Make(MinSpinnakerVersionAccountCRD)
	if err != nil {
		return false
	}
	return min.LTE(v)
}

func IsDynamicFileSupported(svc, version string) bool {
	return strings.HasPrefix(svc, "clouddriver") && IsDynamicAccountSupported(version)
}
