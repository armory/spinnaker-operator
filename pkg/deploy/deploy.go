package deploy

import (
	"context"
	spinnakerv1alpha2 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"k8s.io/apimachinery/pkg/runtime"
)

type ManifestGenerator interface {
	Generate(ctx context.Context, spinConfig *spinnakerv1alpha2.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error)
}

type Deployer interface {
	GetName() string
	// Deploy performs an action on the SpinnakerService. When an error is returned processing stops
	// When true, nil is returned, no other deployer is invoked and the reconcile request is requeued
	Deploy(ctx context.Context, svc spinnakerv1alpha2.SpinnakerServiceInterface, scheme *runtime.Scheme) (bool, error)
}
