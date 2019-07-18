package halconfig

import "gopkg.in/yaml.v2"

type SpinnakerCompleteConfig struct {
	Files       map[string]string `json:"files",omitempty`
	BinaryFiles map[string][]byte `json:"binary",omitempty`
	Profiles    map[string]string `json:"profiles",omitempty`
	HalConfig   *HalConfig        `json:"halConfig",omitempty`
}

type HalConfig struct {
	Version               string                `json:"version",omitempty`
	DeploymentEnvironment DeploymentEnvironment `json:"deploymentEnvironment",omitempty`
}

type DeploymentEnvironment struct {
	Type string `json:"type",omitempty`
}

func ParseHalConfig(data []byte) (HalConfig, error) {
	hc := HalConfig{}
	err := yaml.Unmarshal(data, &hc)
	return hc, err
}
