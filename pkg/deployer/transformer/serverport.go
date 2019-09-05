package transformer

import (
	"fmt"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	"k8s.io/api/apps/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

// transformer used to support changing listening port of a service with "server.port" configuration
type serverPortTransformer struct {
	*defaultTransformer
	svc *spinnakerv1alpha1.SpinnakerService
	log logr.Logger
}

type serverPortTransformerGenerator struct{}

func (g *serverPortTransformerGenerator) NewTransformer(svc *spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error) {
	base := &defaultTransformer{}
	tr := serverPortTransformer{svc: svc, log: log, defaultTransformer: base}
	base.childTransformer = &tr
	return &tr, nil
}

func (t *serverPortTransformer) transformDeploymentManifest(deploymentName string, deployment *v1beta2.Deployment, hc *halconfig.SpinnakerConfig) error {
	if targetPort, _ := hc.GetServiceConfigPropString(deploymentName, "server.port"); targetPort != "" {
		intTargetPort, err := strconv.ParseInt(targetPort, 10, 32)
		if err != nil {
			return err
		}
		for _, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name != deploymentName {
				continue
			}
			if len(c.Ports) > 0 {
				c.Ports[0].ContainerPort = int32(intTargetPort)
			}
			for i, cmd := range c.ReadinessProbe.Exec.Command {
				if !strings.Contains(cmd, "http://localhost") {
					continue
				}
				c.ReadinessProbe.Exec.Command[i] = fmt.Sprintf("http://localhost:%d/health", intTargetPort)
			}
		}
	}
	return nil
}
