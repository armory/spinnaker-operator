package validate

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Transformers tracks the list of transformers
var Validators []SpinnakerValidator

func init() {
	Validators = append(Validators, &singleNamespaceValidator{})
}

type SpinnakerValidator interface {
	Validate(svc v1alpha2.SpinnakerServiceInterface, options Options) error
}

type Options struct {
	Ctx    context.Context
	Client client.Client
	Req    admission.Request
}

func Validate(svc v1alpha2.SpinnakerServiceInterface, options Options) error {
	for _, v := range Validators {
		if err := v.Validate(svc, options); err != nil {
			return err
		}
	}
	return nil
}
