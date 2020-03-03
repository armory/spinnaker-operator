package version

var (
	// local variable for version
	version string

	// key string to identify version property in manifest
	Key = "Version"
)

func GetOperatorVersion() string {

	// initialize manifest lazily
	if len(manifest) == 0 {
		_ = read()
	}

	// populate version and save it locally
	if version == "" {
		v, ok := manifest[Key]
		if ok {
			version = v
		} else {
			version = "Unknown"
		}
	}

	return version
}
