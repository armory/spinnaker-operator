package settings

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Configurers = []AccountConfigurer{
	&KubernetesAccountConfigurer{},
}

type ServiceSettings struct {
	Service  string
	Settings map[string]interface{}
}

type AccountConfigurer interface {
	Accept(account v1alpha2.SpinnakerAccount) bool
	Add(account v1alpha2.SpinnakerAccount, serviceSettings ServiceSettings) error
	GetService() string
}

func GetAffectedServices(account v1alpha2.SpinnakerAccount) []string {
	svcs := make([]string, 0)
	for _, c := range Configurers {
		if c.Accept(account) && !foundIn(c.GetService(), svcs) {
			svcs = append(svcs, c.GetService())
		}
	}
	return svcs
}

func foundIn(obj string, list []string) bool {
	for _, s := range list {
		if s == obj {
			return true
		}
	}
	return false
}

func PrepareSettings(c client.Client, namespace string, svcs []string) ([]ServiceSettings, error) {
	l := &v1alpha2.SpinnakerAccountList{}
	if err := c.List(context.TODO(), l, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	var s []ServiceSettings
	var bySvc = make(map[string]ServiceSettings)

	for _, c := range Configurers {
		// Only consider services passed or any if no service passed
		if len(svcs) == 0 || foundIn(c.GetService(), svcs) {
			svc, ok := bySvc[c.GetService()]
			if !ok {
				svc = ServiceSettings{
					Service:  c.GetService(),
					Settings: make(map[string]interface{}),
				}
				bySvc[c.GetService()] = svc
			}
			for _, a := range l.Items {
				if c.Accept(a) {
					//if a.Status.Valid && c.Accept(a) {
					if err := c.Add(a, svc); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	for _, ss := range bySvc {
		s = append(s, ss)
	}
	return s, nil
}

func UpdateSecret(secret *v1.Secret, settings ServiceSettings, profileName string) error {
	k := fmt.Sprintf("%s-%s.yml", settings.Service, profileName)
	b, err := yaml.Marshal(settings.Settings)
	if err != nil {
		return err
	}
	secret.Data[k] = b
	return nil
}
