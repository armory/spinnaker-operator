package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
)

var _ interfaces.SpinnakerAccount = &SpinnakerAccount{}

func (s *SpinnakerAccount) GetSpec() *interfaces.SpinnakerAccountSpec {
	return &s.Spec
}
func (s *SpinnakerAccount) GetStatus() *interfaces.SpinnakerAccountStatus {
	return &s.Status
}
func (s *SpinnakerAccount) DeepCopyInterface() interfaces.SpinnakerAccount {
	return s.DeepCopy()
}
func (s *SpinnakerAccount) DeepCopySpinnakerAccount() interfaces.SpinnakerAccount {
	return s.DeepCopy()
}

var _ interfaces.SpinnakerAccountList = &SpinnakerAccountList{}

func (s *SpinnakerAccountList) GetItems() []interfaces.SpinnakerAccount {
	if interfaces.IsNil(s.Items) {
		return nil
	} else {
		var result []interfaces.SpinnakerAccount
		for _, i := range s.Items {
			result = append(result, &i)
		}
		return result
	}
}

func (s *SpinnakerAccountList) DeepCopySpinnakerAccountList() interfaces.SpinnakerAccountList {
	return s.DeepCopy()
}
