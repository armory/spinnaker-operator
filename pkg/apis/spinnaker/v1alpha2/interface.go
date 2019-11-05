package v1alpha2

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type SpinnakerServiceInterface interface {
	v1.Object
	runtime.Object
	GetStatus() *SpinnakerServiceStatus
	GetExpose() ExposeConfig
	GetSpinnakerConfig() *SpinnakerConfig
	DeepCopyInterface() SpinnakerServiceInterface
	GetAccountsConfig() AccountConfig
	GetValidation() SpinnakerValidation
}

type SpinnakerServiceListInterface interface {
	runtime.Object
	GetItems() []SpinnakerServiceInterface
}

type SpinnakerServiceBuilderInterface interface {
	New() SpinnakerServiceInterface
	NewList() SpinnakerServiceListInterface
	GetGroupVersion() schema.GroupVersion
}

func (s *SpinnakerService) DeepCopyInterface() SpinnakerServiceInterface {
	return s.DeepCopy()
}

func (s *SpinnakerService) GetExpose() ExposeConfig {
	return s.Spec.Expose
}

func (s *SpinnakerService) GetStatus() *SpinnakerServiceStatus {
	return &s.Status
}

func (s *SpinnakerService) GetSpinnakerConfig() *SpinnakerConfig {
	return &s.Spec.SpinnakerConfig
}

func (s *SpinnakerService) GetValidation() SpinnakerValidation {
	return s.Spec.Validation
}

func (s *SpinnakerServiceList) GetItems() []SpinnakerServiceInterface {
	r := make([]SpinnakerServiceInterface, 0)
	for _, i := range s.Items {
		r = append(r, &i)
	}
	return r
}

type SpinnakerServiceBuilder struct{}

func (s *SpinnakerServiceBuilder) New() SpinnakerServiceInterface {
	return &SpinnakerService{}
}

func (s *SpinnakerServiceBuilder) NewList() SpinnakerServiceListInterface {
	return &SpinnakerServiceList{}
}

func (s *SpinnakerService) GetAccountsConfig() AccountConfig {
	return s.Spec.Accounts
}

func (s *SpinnakerServiceBuilder) GetGroupVersion() schema.GroupVersion {
	return SchemeGroupVersion
}
