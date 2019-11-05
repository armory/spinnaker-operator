package transformer

import (
	"context"
	"fmt"
	spinnakerv1alpha2 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/go-logr/logr"
	"k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

// transformer used to support changing listening port of a service with "server.port" configuration
type serverPortTransformer struct {
	*DefaultTransformer
	svc spinnakerv1alpha2.SpinnakerServiceInterface
	log logr.Logger
}

type serverPortTransformerGenerator struct{}

func (g *serverPortTransformerGenerator) NewTransformer(svc spinnakerv1alpha2.SpinnakerServiceInterface,
	client client.Client, log logr.Logger) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := serverPortTransformer{svc: svc, log: log, DefaultTransformer: base}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *serverPortTransformerGenerator) GetName() string {
	return "ServerPort"
}

func (t *serverPortTransformer) transformDeploymentManifest(ctx context.Context, deploymentName string, deployment *v1.Deployment) error {
	if targetPort, _ := t.svc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, deploymentName, "server.port"); targetPort != "" {
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
