package util

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// AddEnvVarToDeployment adds an environment variable to the given deployment containers for which the filter function returns true.
// merge is used to indicate how to merge the value with an already existing value for the env var.
func AddEnvVarToDeployment(d *appsv1.Deployment, e v1.EnvVar, merge func(old, new string) string, filter func(c v1.Container) bool) {
	ctrs := make([]v1.Container, 0)
	for _, c := range d.Spec.Template.Spec.Containers {
		if !filter(c) {
			ctrs = append(ctrs, c)
			continue
		}
		found := false
		vars := make([]v1.EnvVar, 0)
		for _, ce := range c.Env {
			if ce.Name == e.Name {
				found = true
				ce.Value = merge(ce.Value, e.Value)
			}
			vars = append(vars, ce)
		}
		if !found {
			vars = append(vars, e)
		}
		c.Env = vars
		ctrs = append(ctrs, c)
	}
	d.Spec.Template.Spec.Containers = ctrs
}
