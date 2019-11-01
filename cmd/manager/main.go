package main

import (
	"github.com/armory/spinnaker-operator/pkg/apis"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakeraccount"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakerservice"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakervalidating"
	"github.com/armory/spinnaker-operator/pkg/operator"
)

func main() {
	spinnakervalidating.SpinnakerServiceBuilder = &v1alpha2.SpinnakerServiceBuilder{}
	spinnakerservice.SpinnakerServiceBuilder = &v1alpha2.SpinnakerServiceBuilder{}
	spinnakeraccount.SpinnakerServiceBuilder = &v1alpha2.SpinnakerServiceBuilder{}
	operator.Start(apis.AddToScheme)
}
