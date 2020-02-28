package version

var SpinnakerOperatorVersion string

func init() {
	v := manifest["Spinnaker-Operator-Version"]
	if v == "" {
		SpinnakerOperatorVersion = "Unknown"
	}
	SpinnakerOperatorVersion = v
}
