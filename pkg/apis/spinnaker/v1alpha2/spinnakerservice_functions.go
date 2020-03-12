package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func RegisterTypes() {
	interfaces.DefaultTypesFactory.Factories[interfaces.V1alpha2Version] = &TypesFactory{}
}

var _ interfaces.SpinnakerService = &SpinnakerService{}

func (s *SpinnakerService) GetSpec() *interfaces.SpinnakerServiceSpec {
	return &s.Spec
}
func (s *SpinnakerService) GetStatus() *interfaces.SpinnakerServiceStatus {
	return &s.Status
}
func (s *SpinnakerService) DeepCopyInterface() interfaces.SpinnakerService {
	return s.DeepCopy()
}
func (s *SpinnakerService) DeepCopySpinnakerService() interfaces.SpinnakerService {
	return s.DeepCopy()
}

var _ interfaces.SpinnakerServiceList = &SpinnakerServiceList{}

func (s *SpinnakerServiceList) GetItems() []interfaces.SpinnakerService {
	if interfaces.IsNil(s.Items) {
		return nil
	} else {
		var result []interfaces.SpinnakerService
		for _, i := range s.Items {
			result = append(result, &i)
		}
		return result
	}
}
func (s *SpinnakerServiceList) DeepCopySpinnakerServiceList() interfaces.SpinnakerServiceList {
	return s.DeepCopy()
}

type TypesFactory struct{}

var _ interfaces.TypesFactory = &TypesFactory{}

func (f *TypesFactory) NewService() interfaces.SpinnakerService {
	return &SpinnakerService{}
}
func (f *TypesFactory) NewServiceList() interfaces.SpinnakerServiceList {
	return &SpinnakerServiceList{}
}
func (f *TypesFactory) NewAccount() interfaces.SpinnakerAccount {
	return &SpinnakerAccount{}
}
func (f *TypesFactory) NewAccountList() interfaces.SpinnakerAccountList {
	return &SpinnakerAccountList{}
}
func (f *TypesFactory) GetGroupVersion() schema.GroupVersion {
	return SchemeGroupVersion
}
func (f *TypesFactory) DeepCopyLatestTypesFactory() interfaces.TypesFactory {
	return f.DeepCopy()
}
