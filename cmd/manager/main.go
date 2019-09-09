package main

import (
	"github.com/armory/spinnaker-operator/pkg/apis"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakerservice"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakervalidating"
	"github.com/armory/spinnaker-operator/pkg/operator"
)

func main() {
	spinnakervalidating.SpinnakerKind = &v1alpha1.SpinnakerServiceKind{}
	spinnakerservice.SpinnakerServiceKind = &v1alpha1.SpinnakerServiceKind{}
	operator.Start(apis.AddToScheme)
}
