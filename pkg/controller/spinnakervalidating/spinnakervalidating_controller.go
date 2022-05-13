package spinnakervalidating

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/controller/webhook"
	"github.com/armory/spinnaker-operator/pkg/halyard"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/validate"
	"gomodules.xyz/jsonpatch/v2"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	ValidationConfigHashKey      = "validation"
	DefaultValidationFreqSeconds = 10
)

// +kubebuilder:webhook:path=/validate-v1-spinnakerservice,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io,admissionReviewVersions=v1,sideEffects=none

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

	hc := svc.GetStatus().GetHash(ValidationConfigHashKey)
	if hc != nil {
		if !v.NeedsValidation(hc.LastUpdatedAt) {
			return admission.Allowed("")
		}
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

	log.Info("Validating SpinnakerService", "metadata.name", svc.GetName())
	validationResult := validate.ValidateAll(svc, opts)
	if validationResult.HasErrors() {
		errorMsg := validationResult.GetErrorMessage()
		err := fmt.Errorf(errorMsg)
		log.Error(err, errorMsg, "metadata.name", svc)
		return admission.Denied(errorMsg)
	}
	// Update the status with any admission status change, only if there's already an existing SpinnakerService
	if req.AdmissionRequest.Operation == v1.Update {
		if len(validationResult.StatusPatches) > 0 {
			validationResult.StatusPatches = append(validationResult.StatusPatches, v.addLastValidation(svc))
			log.Info(fmt.Sprintf("patching SpinnakerService status with %v", validationResult.StatusPatches), "metadata.name", svc.GetName())
			if err := v.client.Status().Patch(ctx, svc, &precomputedPatch{validationResult}); err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
		}
	}
	log.Info("SpinnakerService is valid", "metadata.name", svc.GetName())
	return admission.ValidationResponse(true, "")
}

func (v *spinnakerValidatingController) NeedsValidation(lastValid metav1.Time) bool {
	if lastValid.IsZero() {
		return true
	}
	n := lastValid.Time.Add(time.Duration(DefaultValidationFreqSeconds) * time.Second)
	return time.Now().After(n)
}

func (v *spinnakerValidatingController) getHash(config interface{}) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}

func (v *spinnakerValidatingController) addLastValidation(svc interfaces.SpinnakerService) jsonpatch.JsonPatchOperation {
	hash, _ := v.getHash(svc.GetStatus())
	return jsonpatch.NewOperation("replace", fmt.Sprintf("/status/lastDeployed/%s", ValidationConfigHashKey), interfaces.HashStatus{
		Hash:          hash,
		LastUpdatedAt: metav1.NewTime(time.Now()),
	})
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

type precomputedPatch struct {
	validationResult validate.ValidationResult
}

func (p *precomputedPatch) Type() types.PatchType {
	return types.JSONPatchType
}

func (p *precomputedPatch) Data(obj client.Object) ([]byte, error) {
	return json.Marshal(p.validationResult.StatusPatches)
}
