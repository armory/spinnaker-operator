package spinnakervalidating

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/validate"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// +kubebuilder:webhook:path=/validate-v1-spinnakerservice,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// spinnakerValidatingController annotates Pods
type spinnakerValidatingController struct {
	client  client.Client
	decoder types.Decoder
}

// NewSpinnakerService instantiates the type we're going to validate
var SpinnakerServiceBuilder v1alpha1.SpinnakerServiceBuilderInterface

// Implement admission.Handler so the controller can handle admission request.
var _ admission.Handler = &spinnakerValidatingController{}
var log = logf.Log.WithName("spinvalidate")

// Add adds the validating admission controller
func Add(m manager.Manager) error {
	validatingWebhook, err := builder.NewWebhookBuilder().
		Name("validating.k8s.io").
		Validating().
		WithManager(m).
		Rules(getSpinnakerServiceRule(), getConfigMapRule()).
		Handlers(&spinnakerValidatingController{}).
		Build()

	if err != nil {
		return err
	}

	disableWebhookConfigInstaller := false

	options, err := getWebhookBootstrapOptions()
	if err != nil {
		return err
	}
	as, err := webhook.NewServer("spinnaker-admission-server", m, webhook.ServerOptions{
		Port:                          9876,
		CertDir:                       "/tmp/cert",
		DisableWebhookConfigInstaller: &disableWebhookConfigInstaller,
		BootstrapOptions:              options,
	})

	if err != nil {
		return err
	}
	return as.Register(validatingWebhook)
}

func getSpinnakerServiceRule() admissionregistrationv1beta1.RuleWithOperations {
	gv := SpinnakerServiceBuilder.GetGroupVersion()
	return admissionregistrationv1beta1.RuleWithOperations{
		Operations: []admissionregistrationv1beta1.OperationType{
			admissionregistrationv1beta1.Create,
			admissionregistrationv1beta1.Update,
		},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{gv.Group},
			APIVersions: []string{gv.Version},
			Resources:   []string{"spinnakerservices"},
		},
	}
}

func getConfigMapRule() admissionregistrationv1beta1.RuleWithOperations {
	gv := corev1.SchemeGroupVersion
	return admissionregistrationv1beta1.RuleWithOperations{
		Operations: []admissionregistrationv1beta1.OperationType{
			admissionregistrationv1beta1.Create,
			admissionregistrationv1beta1.Update,
		},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{gv.Group},
			APIVersions: []string{gv.Version},
			Resources:   []string{"configmaps"},
		},
	}
}

func getWebhookBootstrapOptions() (*webhook.BootstrapOptions, error) {
	selector := "spinnaker-operator"
	name := "spinnaker-admission-service"
	ns, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		envNs := os.Getenv("ADMISSION_PROXY_NAMESPACE")
		if envNs == "" {
			return nil, fmt.Errorf("Error getting operator namespace: %s  and env var ADMISSION_PROXY_NAMESPACE, not set", err.Error())
		}
		ns = envNs
		selector = "spinnaker-operator-proxy"
		name = "spinnaker-admission-service-proxy"
	}

	return &webhook.BootstrapOptions{
		Service: &webhook.Service{
			Namespace: ns,
			Name:      name,
			// Selectors should select the pods that runs this webhook server.
			Selectors: map[string]string{
				"name": selector,
			},
		},
	}, nil
}

// spinnakerValidatingController adds an annotation to every incoming pods.
func (v *spinnakerValidatingController) Handle(ctx context.Context, req types.Request) types.Response {
	svc, err := v.getSpinnakerService(req)
	if err != nil {
		log.Error(err, "Unable to retrieve Spinnaker service from request")
		return admission.ErrorResponse(http.StatusExpectationFailed, err)
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
	validationErrors := validate.Validate(svc, opts)
	if len(validationErrors) > 0 {
		errorMsg := "SpinnakerService validation failed:\n"
		for _, e := range validationErrors {
			errorMsg = fmt.Sprintf("%s%s\n", errorMsg, e.Error())
		}
		aggregatedError := errors.New(errorMsg)
		log.Error(aggregatedError, aggregatedError.Error(), "metadata.name", svc)
		return admission.ErrorResponse(http.StatusBadRequest, aggregatedError)
	}
	log.Info("SpinnakerService is valid", "metadata.name", svc)
	return admission.ValidationResponse(true, "")
}

// spinnakerValidatingController implements inject.Client.
// A client will be automatically injected.
var _ inject.Client = &spinnakerValidatingController{}

// InjectClient injects the client.
func (v *spinnakerValidatingController) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// spinnakerValidatingController implements inject.Decoder.
// A decoder will be automatically injected.
var _ inject.Decoder = &spinnakerValidatingController{}

// InjectDecoder injects the decoder.
func (v *spinnakerValidatingController) InjectDecoder(d types.Decoder) error {
	v.decoder = d
	return nil
}
