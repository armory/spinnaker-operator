package validate

import (
	"context"
	"github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// Transformers tracks the list of transformers
var Validators []SpinnakerValidator

func init() {
	Validators = append(Validators, &singleNamespaceValidator{})
}

type SpinnakerValidator interface {
	Validate(svc *v1alpha1.SpinnakerService, hc *halconfig.SpinnakerConfig, options Options) error
}

type Options struct {
	Ctx    context.Context
	Client client.Client
	Req    types.Request
}

func Validate(svc *v1alpha1.SpinnakerService, options Options) error {
	c, err := svc.GetConfigObject(options.Client)
	if err != nil {
		return err
	}

	hc := halconfig.NewSpinnakerConfig()
	if err := hc.FromConfigObject(c); err != nil {
		return err
	}

	for _, v := range Validators {
		if err := v.Validate(svc, hc, options); err != nil {
			return err
		}
	}
	return nil
}
