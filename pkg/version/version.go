package version

var (
	// local variable for version
	version string

	// key string to identify version property in manifest
	Key = "Version"
)

func GetOperatorVersion() string {

	// populate version and save it locally
	if version == "" {
		v, err := GetManifestValue(Key)
		if err == nil {
			version = v
		} else {
			version = "Unknown"
		}
	}

	return version
}
