package spinnakervalidating

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/controller/webhook"
	"github.com/armory/spinnaker-operator/pkg/halyard"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/validate"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/rest"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-spinnakerservice,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// spinnakerValidatingController performs preflight checks
type spinnakerValidatingController struct {
	client     client.Client
	decoder    *admission.Decoder
	restConfig *rest.Config
}

// TypesFactory instantiates the type we're going to validate
var TypesFactory interfaces.TypesFactory

// Implement all intended interfaces.
var _ admission.Handler = &spinnakerValidatingController{}
var _ inject.Config = &spinnakerValidatingController{}
var _ inject.Client = &spinnakerValidatingController{}
var _ admission.DecoderInjector = &spinnakerValidatingController{}
var log = logf.Log.WithName("spinvalidate")

// Add adds the validating admission controller
func Add(m manager.Manager) error {
	spinSvc := TypesFactory.NewService()
	gvk, err := apiutil.GVKForObject(spinSvc, m.GetScheme())
	if err != nil {
		return err
	}
	webhook.Register(gvk, "spinnakerservices", &spinnakerValidatingController{})
	return nil
}

// Handle is the entry point for spinnaker preflight validations
func (v *spinnakerValidatingController) Handle(ctx context.Context, req admission.Request) admission.Response {
	log.Info(fmt.Sprintf("Handling admission request for: %s", req.AdmissionRequest.Kind.Kind))
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
		Ctx:          secrets.NewContext(ctx, v.restConfig, req.Namespace),
		Client:       v.client,
		Req:          req,
		Log:          log,
		Halyard:      halyard.NewService(),
		TypesFactory: TypesFactory,
	}
	defer secrets.Cleanup(opts.Ctx)

	log.Info("Starting validation")
	validationResult := validate.ValidateAll(svc, opts)
	if validationResult.HasErrors() {
		errorMsg := validationResult.GetErrorMessage()
		err := fmt.Errorf(errorMsg)
		log.Error(err, errorMsg, "metadata.name", svc)
		return admission.Errored(http.StatusBadRequest, err)
	}
	// Update the status with any admission status change, only if there's already an existing SpinnakerService
	if req.AdmissionRequest.Operation == v1beta1.Update {
		if err := v.client.Status().Update(ctx, svc); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
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
func (v *spinnakerValidatingController) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// InjectConfig injects the rest config for creating raw kubernetes clients.
func (v *spinnakerValidatingController) InjectConfig(c *rest.Config) error {
	v.restConfig = c
	return nil
}
