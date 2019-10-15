// +build !ignore_autogenerated

// Code generated by generate. DO NOT EDIT.

package v1alpha2

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposeConfig) DeepCopyInto(out *ExposeConfig) {
	*out = *in
	in.Service.DeepCopyInto(&out.Service)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposeConfig.
func (in *ExposeConfig) DeepCopy() *ExposeConfig {
	if in == nil {
		return nil
	}
	out := new(ExposeConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposeConfigService) DeepCopyInto(out *ExposeConfigService) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Overrides != nil {
		in, out := &in.Overrides, &out.Overrides
		*out = make(map[string]ExposeConfigServiceOverrides, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposeConfigService.
func (in *ExposeConfigService) DeepCopy() *ExposeConfigService {
	if in == nil {
		return nil
	}
	out := new(ExposeConfigService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposeConfigServiceOverrides) DeepCopyInto(out *ExposeConfigServiceOverrides) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposeConfigServiceOverrides.
func (in *ExposeConfigServiceOverrides) DeepCopy() *ExposeConfigServiceOverrides {
	if in == nil {
		return nil
	}
	out := new(ExposeConfigServiceOverrides)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in FreeForm) DeepCopyInto(out *FreeForm) {
	{
		in := &in
		clone := in.DeepCopy()
		*out = *clone
		return
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerConfig) DeepCopyInto(out *SpinnakerConfig) {
	*out = *in
	if in.Files != nil {
		in, out := &in.Files, &out.Files
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.ServiceSettings.DeepCopyInto(&out.ServiceSettings)
	if in.Profiles != nil {
		in, out := &in.Profiles, &out.Profiles
		*out = make(map[string]FreeForm, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	in.Config.DeepCopyInto(&out.Config)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerConfig.
func (in *SpinnakerConfig) DeepCopy() *SpinnakerConfig {
	if in == nil {
		return nil
	}
	out := new(SpinnakerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerDeploymentStatus) DeepCopyInto(out *SpinnakerDeploymentStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerDeploymentStatus.
func (in *SpinnakerDeploymentStatus) DeepCopy() *SpinnakerDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(SpinnakerDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerService) DeepCopyInto(out *SpinnakerService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerService.
func (in *SpinnakerService) DeepCopy() *SpinnakerService {
	if in == nil {
		return nil
	}
	out := new(SpinnakerService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SpinnakerService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerServiceBuilder) DeepCopyInto(out *SpinnakerServiceBuilder) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerServiceBuilder.
func (in *SpinnakerServiceBuilder) DeepCopy() *SpinnakerServiceBuilder {
	if in == nil {
		return nil
	}
	out := new(SpinnakerServiceBuilder)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerServiceList) DeepCopyInto(out *SpinnakerServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SpinnakerService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerServiceList.
func (in *SpinnakerServiceList) DeepCopy() *SpinnakerServiceList {
	if in == nil {
		return nil
	}
	out := new(SpinnakerServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SpinnakerServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerServiceSpec) DeepCopyInto(out *SpinnakerServiceSpec) {
	*out = *in
	in.SpinnakerConfig.DeepCopyInto(&out.SpinnakerConfig)
	in.Expose.DeepCopyInto(&out.Expose)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerServiceSpec.
func (in *SpinnakerServiceSpec) DeepCopy() *SpinnakerServiceSpec {
	if in == nil {
		return nil
	}
	out := new(SpinnakerServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerServiceStatus) DeepCopyInto(out *SpinnakerServiceStatus) {
	*out = *in
	in.LastConfigurationTime.DeepCopyInto(&out.LastConfigurationTime)
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]SpinnakerDeploymentStatus, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerServiceStatus.
func (in *SpinnakerServiceStatus) DeepCopy() *SpinnakerServiceStatus {
	if in == nil {
		return nil
	}
	out := new(SpinnakerServiceStatus)
	in.DeepCopyInto(out)
	return out
}
