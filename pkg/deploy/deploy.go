package deploy

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"k8s.io/apimachinery/pkg/runtime"
)

type ManifestGenerator interface {
	Generate(ctx context.Context, spinConfig *interfaces.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error)
}

type Deployer interface {
	GetName() string
	// Deploy performs an action on the SpinnakerService. When an error is returned processing stops
	// When true, nil is returned, no other deployer is invoked and the reconcile request is requeued
	Deploy(ctx context.Context, svc interfaces.SpinnakerService, scheme *runtime.Scheme) (bool, error)
}
