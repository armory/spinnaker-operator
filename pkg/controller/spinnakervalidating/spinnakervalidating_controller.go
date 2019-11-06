package spinnakervalidating

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	webhook "github.com/armory/spinnaker-operator/pkg/controller/webhook"
	"github.com/armory/spinnaker-operator/pkg/halyard"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/validate"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-spinnakerservice,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// spinnakerValidatingController performs preflight checks
type spinnakerValidatingController struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewSpinnakerService instantiates the type we're going to validate
var SpinnakerServiceBuilder v1alpha2.SpinnakerServiceBuilderInterface

// Implement admission.Handler so the controller can handle admission request.
var _ admission.Handler = &spinnakerValidatingController{}
var log = logf.Log.WithName("spinvalidate")

// Add adds the validating admission controller
func Add(m manager.Manager) error {
	spinSvc := SpinnakerServiceBuilder.New()
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
		Ctx:         secrets.NewContext(ctx),
		Client:      v.client,
		Req:         req,
		Log:         log,
		Halyard:     halyard.NewService(),
		SpinBuilder: SpinnakerServiceBuilder,
	}
	log.Info("Starting validation")
	validationResult := validate.ValidateAll(svc, opts)
	if validationResult.HasErrors() {
		errorMsg := validationResult.GetErrorMessage()
		err := fmt.Errorf(errorMsg)
		log.Error(err, errorMsg, "metadata.name", svc)
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
func (v *spinnakerValidatingController) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
