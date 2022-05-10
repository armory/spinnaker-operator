package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
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

func (s *SpinnakerAccountList) GetResourceVersion() string        { return s.ResourceVersion }
func (s *SpinnakerAccountList) SetResourceVersion(version string) { s.ResourceVersion = version }
func (s *SpinnakerAccountList) GetSelfLink() string               { return s.SelfLink }
func (s *SpinnakerAccountList) SetSelfLink(selfLink string)       { s.SelfLink = selfLink }
func (s *SpinnakerAccountList) GetContinue() string               { return s.Continue }
func (s *SpinnakerAccountList) SetContinue(c string)              { s.Continue = c }
func (s *SpinnakerAccountList) GetRemainingItemCount() *int64     { return s.RemainingItemCount }
func (s *SpinnakerAccountList) SetRemainingItemCount(c *int64)    { s.RemainingItemCount = c }
