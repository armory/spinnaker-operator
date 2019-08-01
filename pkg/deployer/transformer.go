package deployer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Transformers tracks the list of transformers
var Transformers []TransformerGenerator

func init() {
	Transformers = append(Transformers, &ownerTransformerGenerator{}, &targetTransformerGenerator{}, &exposeTransformerGenerator{})
}

// Transformer affects how Spinnaker is deployed.
// It can change the Spinnaker configuration itself with TransformConfig.
// It can also change the manifests before they are updated.
type Transformer interface {
	TransformConfig(hc *halconfig.SpinnakerConfig) error
	TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error
}

// TransformerGenerator generates transformers for the given SpinnakerService
type TransformerGenerator interface {
	NewTransformer(svc spinnakerv1alpha1.SpinnakerService, client client.Client) (Transformer, error)
}
