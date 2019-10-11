package v1alpha1

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

//func (s *SpinnakerConfigMap) GetHalConfigPropBool(prop string, defaultVal bool) (bool, error) {
//	return true, nil
//}

//+k8s:deepcopy-gen=false
//+k8s:openapi-gen=false
type Profiles interface {
	Get(string) (map[string]interface{}, bool)
	AsMap() map[string]map[string]interface{}
	DeepCopyProfiles() Profiles
}

//+k8s:deepcopy-gen=false
//+k8s:openapi-gen=false
type ProfilesMap map[string]map[string]interface{}

// +k8s:deepcopy-gen:interfaces=github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1.Profiles
// +k8s:deepcopy-gen:nonpointer-interfaces=true
func (p *ProfilesMap) DeepCopyProfiles() Profiles {
	return &ProfilesMap{}
}

func (p *ProfilesMap) Get(svc string) (map[string]interface{}, bool) {
	o, ok := (*p)[svc]
	return o, ok
}

func (p *ProfilesMap) AsMap() map[string]map[string]interface{} {
	return *p
}

//type ServiceSettings interface {
//	DeepCopyServiceSettings()
//}

// GetServiceSettingsPropString returns a service settings prop for a given service
func (s *SpinnakerConfig) GetServiceSettingsPropString(ctx context.Context, svc, prop string) (string, error) {
	//return inspect.GetObjectPropString(ctx, s.ServiceSettings, fmt.Sprintf("%s.%s", svc, prop))
	return inspect.GetObjectPropString(ctx, s.Profiles, fmt.Sprintf("%s.%s", svc, prop))
}

// GetHalConfigPropString returns a property stored in halconfig
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetHalConfigPropString(ctx context.Context, prop string) (string, error) {
	//return inspect.GetObjectPropString(ctx, s.HalConfig, prop)
	return inspect.GetObjectPropString(ctx, s.Profiles, prop)
}

// SetHalConfigProp sets a property in the config
func (s *SpinnakerConfig) SetHalConfigProp(prop string, value interface{}) error {
	//return inspect.SetObjectProp(s.HalConfig, prop, value)
	return inspect.SetObjectProp(s.Profiles, prop, value)
}

// GetHalConfigPropBool returns a boolean property in halconfig
func (s *SpinnakerConfig) GetHalConfigPropBool(prop string, defaultVal bool) (bool, error) {
	//return inspect.GetObjectPropBool(s.HalConfig, prop, defaultVal)
	return inspect.GetObjectPropBool(s.Profiles, prop, defaultVal)
}

// GetServiceConfigPropString returns the value of the prop in a service profile file
func (s *SpinnakerConfig) GetServiceConfigPropString(ctx context.Context, svc, prop string) (string, error) {
	p, ok := s.Profiles.Get(svc)
	if ok {
		return inspect.GetObjectPropString(ctx, p, prop)
	}
	return "", nil
}
