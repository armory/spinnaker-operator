package spinnakervalidating

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/validate"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-spinnakerservice,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// spinnakerValidatingController performs preflight checks
type spinnakerValidatingController struct {
	client  client.Client
	decoder admission.Decoder
}

// NewSpinnakerService instantiates the type we're going to validate
var SpinnakerServiceBuilder v1alpha2.SpinnakerServiceBuilderInterface

// Implement admission.Handler so the controller can handle admission request.
var _ admission.Handler = &spinnakerValidatingController{}
var log = logf.Log.WithName("spinvalidate")

// Add adds the validating admission controller
func Add(m manager.Manager) error {
	//ns, err := k8sutil.GetOperatorNamespace()
	_, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}

	hookServer := m.GetWebhookServer()
	hookServer.Register("/validate-v1alpha2-spinnakerservice", &webhook.Admission{Handler: &spinnakerValidatingController{}})
	return nil
}

// Handle is the entry point for spinnaker preflight validations
func (v *spinnakerValidatingController) Handle(ctx context.Context, req admission.Request) admission.Response {
	svc, err := v.getSpinnakerService(req)
	if err != nil {
		log.Error(err, "Unable to retrieve Spinnaker service from request")
		return admission.Errored(http.StatusExpectationFailed, err)
	}
	if svc == nil {
		log.Info("No SpinnakerService found in request")
		return admission.ValidationResponse(true, "")
	}
	opts := validate.Options{
		Ctx:    ctx,
		Client: v.client,
		Req:    req,
		Log:    log,
	}
	log.Info("Starting validation")
	validationResults := validate.ValidateAll(svc, opts)
	err = v.collectErrors(validationResults)
	if err != nil {
		log.Error(err, err.Error(), "metadata.name", svc)
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.Info("SpinnakerService is valid", "metadata.name", svc)
	return admission.ValidationResponse(true, "")
}

// InjectClient injects the client.
func (v *spinnakerValidatingController) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// InjectDecoder injects the decoder.
func (v *spinnakerValidatingController) InjectDecoder(d admission.Decoder) error {
	v.decoder = d
	return nil
}

func (v *spinnakerValidatingController) collectErrors(results []validate.ValidationResult) error {
	errorMsg := "SpinnakerService validation failed:\n"
	hasErrors := false
	for _, r := range results {
		if r.Error != nil {
			hasErrors = true
			errorMsg = fmt.Sprintf("%s%s\n", errorMsg, r.Error.Error())
		}
	}
	if hasErrors {
		return errors.New(errorMsg)
	} else {
		return nil
	}
}
