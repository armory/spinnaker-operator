package controller

import (
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakeraccount"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakerservice"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakerstatus"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, spinnakerservice.Add, spinnakeraccount.Add, spinnakerstatus.Add)
}
