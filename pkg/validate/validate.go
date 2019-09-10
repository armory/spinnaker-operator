package validate

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// Transformers tracks the list of transformers
var Validators []SpinnakerValidator

func init() {
	Validators = append(Validators, &singleNamespaceValidator{})
}

type SpinnakerValidator interface {
	Validate(svc v1alpha1.SpinnakerServiceInterface, hc *halconfig.SpinnakerConfig, options Options) error
}

type Options struct {
	Ctx    context.Context
	Client client.Client
	Req    types.Request
}

func Validate(svc v1alpha1.SpinnakerServiceInterface, options Options) error {
	_, hc, err := v1alpha1.GetConfig(svc, options.Client)
	if err != nil {
		return err
	}

	for _, v := range Validators {
		if err := v.Validate(svc, hc, options); err != nil {
			return err
		}
	}
	return nil
}
