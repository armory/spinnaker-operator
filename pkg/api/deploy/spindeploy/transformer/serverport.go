package transformer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// transformer used to support changing listening port of a service with "server.port" configuration
type serverPortTransformer struct {
	*DefaultTransformer
	svc interfaces.SpinnakerService
	log logr.Logger
}

type ServerPortTransformerGenerator struct{}

func (g *ServerPortTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := serverPortTransformer{svc: svc, log: log, DefaultTransformer: base}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *ServerPortTransformerGenerator) GetName() string {
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
