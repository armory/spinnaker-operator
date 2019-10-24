package spinnakervalidating

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/validate"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-spinnakerservice,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// spinnakerValidatingController performs preflight checks
type spinnakerValidatingController struct {
	client client.Client
}

// NewSpinnakerService instantiates the type we're going to validate
var SpinnakerServiceBuilder v1alpha2.SpinnakerServiceBuilderInterface

// Implement admission.Handler so the controller can handle admission request.
var _ admission.Handler = &spinnakerValidatingController{}
var log = logf.Log.WithName("spinvalidate")

// Add adds the validating admission controller
func Add(m manager.Manager) error {
	// Determine environment
	ns, name, err := getOperatorNameAndNamespace()
	if err != nil {
		return err
	}

	// Create Kubernetes service for listening to requests from API server
	rawClient := kubernetes.NewForConfigOrDie(m.GetConfig())
	port := 9876
	err = deployWebhookService(ns, name, port, rawClient)
	if err != nil {
		return err
	}

	// Create or get certificates
	c, err := getCertContext(ns, name)
	if err != nil {
		return err
	}

	// Register webhook server
	hookServer := m.GetWebhookServer()
	hookServer.CertDir = c.certDir
	hookServer.Port = port
	gv := SpinnakerServiceBuilder.GetGroupVersion()
	path := "/validate-v1alpha2-spinnakerservice"
	hookConfigName := fmt.Sprintf("spinnakervalidatingwebhook.%s", gv.Group)
	hookServer.Register(path, &webhook.Admission{Handler: &spinnakerValidatingController{}})

	// Create validating webhook configuration for registering our webhook with the API server
	w := getWebhookConfig(hookConfigName, name, ns, path, c)
	return deployValidatingWebhookConfiguration(hookConfigName, ns, w, rawClient)
}

func getOperatorNameAndNamespace() (string, string, error) {
	name, err := k8sutil.GetOperatorName()
	if err != nil {
		return "", "", err
	}
	ns, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		envNs := os.Getenv("ADMISSION_PROXY_NAMESPACE")
		if envNs == "" {
			return "", "", fmt.Errorf("unable to determine operator namespace. Error: %s and ADMISSION_PROXY_NAMESPACE env var not set", err.Error())
		}
		ns = envNs
	}
	return ns, name, nil
}

func deployWebhookService(ns string, name string, port int, client *kubernetes.Clientset) error {
	// Always recreate the service
	_ = client.CoreV1().Services(ns).Delete(name, &metav1.DeleteOptions{})

	// Create the service
	selectorLabels := map[string]string{"name": name}
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Labels:    selectorLabels,
		},
		Spec: v1.ServiceSpec{
			Selector: selectorLabels,
			Ports: []v1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       443,
					TargetPort: intstr.FromInt(port),
				},
			},
		},
	}
	_, err := client.CoreV1().Services(ns).Create(service)
	return err
}

func deployValidatingWebhookConfiguration(configName, ns string, webhook v1beta1.Webhook, client *kubernetes.Clientset) error {
	// Always recreate the configuration
	_ = client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Delete(configName, &metav1.DeleteOptions{})
	_, err := client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(&v1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: ns,
		},
		Webhooks: []v1beta1.Webhook{webhook},
	})
	return err
}

func getWebhookConfig(configName, operatorName, ns, path string, c *certContext) v1beta1.Webhook {
	return v1beta1.Webhook{
		Name: configName,
		ClientConfig: v1beta1.WebhookClientConfig{
			Service: &v1beta1.ServiceReference{
				Namespace: ns,
				Name:      operatorName,
				Path:      &path,
			},
			CABundle: c.signingCert,
		},
		Rules: []v1beta1.RuleWithOperations{getSpinnakerServiceRule()},
	}
}

func getSpinnakerServiceRule() v1beta1.RuleWithOperations {
	gv := SpinnakerServiceBuilder.GetGroupVersion()
	return v1beta1.RuleWithOperations{
		Operations: []v1beta1.OperationType{
			v1beta1.Create,
			v1beta1.Update,
		},
		Rule: v1beta1.Rule{
			APIGroups:   []string{gv.Group},
			APIVersions: []string{gv.Version},
			Resources:   []string{"spinnakerservices"},
		},
	}
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
