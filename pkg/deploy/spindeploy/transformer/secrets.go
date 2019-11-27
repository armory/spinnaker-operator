package transformer

import (
	"context"
	"fmt"
	secrets2 "github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// secretsTransformer maps Kubernetes secrets onto the deployment of the service that requires it
// Either as a mounted file (encryptedFile) or an environment variable (tokens, passwords...)
type secretsTransformer struct {
	svc    v1alpha2.SpinnakerServiceInterface
	log    logr.Logger
	client client.Client
}

type secretsTransformerGenerator struct{}

func (s *secretsTransformerGenerator) NewTransformer(svc v1alpha2.SpinnakerServiceInterface,
	client client.Client, log logr.Logger) (Transformer, error) {
	tr := secretsTransformer{svc: svc, log: log, client: client}
	return &tr, nil
}

func (s *secretsTransformerGenerator) GetName() string {
	return "Secrets"
}

func (s *secretsTransformer) TransformConfig(ctx context.Context) error {
	return nil
}

func (s *secretsTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	for svc, cfg := range gen.Config {
		kCollector := &kubernetesSecretCollector{svc: svc, namespace: s.svc.GetNamespace()}

		for k := range cfg.Resources {
			sec, ok := cfg.Resources[k].(*v1.Secret)
			if ok {
				kCollector.mapSecrets(svc, sec)
			}
		}

		if err := kCollector.setInDeployment(cfg.Deployment); err != nil {
			return err
		}
	}
	return nil
}

type kubernetesSecretCollector struct {
	envVars      []v1.EnvVar
	volumes      []v1.Volume
	volumeMounts []v1.VolumeMount
	svc          string
	namespace    string
}

// transformSecrets goes through all secret data and replace references to passwords and files with env variables
// and file paths
func (k *kubernetesSecretCollector) mapSecrets(svc string, secret *v1.Secret) error {
	for key := range secret.Data {
		// Attempt to deserialize as YAML
		m := make(map[string]interface{})
		if err := yaml.Unmarshal(secret.Data[key], &m); err != nil {
			continue
		}
		// If it's YAML replace secret references
		ndata, err := k.sanitizeSecrets(svc, m)
		// This time, we harshly don't accept failure
		if err != nil {
			return err
		}

		// Replace the value
		b, err := yaml.Marshal(ndata)
		if err != nil {
			return err
		}
		secret.Data[key] = b
	}
	return nil
}

func (k *kubernetesSecretCollector) sanitizeSecrets(svc string, obj interface{}) (interface{}, error) {
	h := func(val string) (string, error) {
		e, f, p := secrets2.GetEngine(val)
		// If not Kubernetes secret, we just pass to the service as is
		if e != "k8s" {
			return val, nil
		}
		name, key := secrets.ParseKubernetesSecretParams(p)
		if f {
			return k.handleSecretFileReference(name, key)
		}
		s := k.handleSecretVarReference(name, key)
		return s, nil
	}
	return inspect.InspectStrings(obj, h)
}

func (k *kubernetesSecretCollector) getSecretFilePath() string {
	return fmt.Sprintf("/opt/%s/secrets", k.svc)
}

func (k *kubernetesSecretCollector) getSecretAbsolutePath(relative string) string {
	return path.Join(k.getSecretFilePath(), relative)
}

func (k *kubernetesSecretCollector) getRelativeSecretFilePath(secretName, key string) string {
	return path.Join(secretName, key)
}

func (k *kubernetesSecretCollector) getSecretMountPath(secretName string) string {
	return path.Join(k.getSecretFilePath(), secretName)
}

func (k *kubernetesSecretCollector) getVolumeName(secretName string) string {
	return fmt.Sprintf("volume-%s", secretName)
}

func (k *kubernetesSecretCollector) getEnvVarName(secretName, key string) string {
	return fmt.Sprintf("%s_%s_%s", safeEnvVarName(k.svc), safeEnvVarName(secretName), safeEnvVarName(key))
}

func (k *kubernetesSecretCollector) getEnvVarNameReference(varName string) string {
	return fmt.Sprintf("${%s}", varName)
}

// Environment name replacer: we know input will be valid secret name and keys or services
// Change when a new "Spin?!" service is added.
var envReplacer = strings.NewReplacer("-", "_", ".", "_", " ", "_")

func safeEnvVarName(val string) string {
	return strings.ToUpper(envReplacer.Replace(val))
}

func (k *kubernetesSecretCollector) handleSecretFileReference(secretName, key string) (string, error) {
	path, added := k.addVolume(secretName, key)
	// Add a volume mount
	if added {
		k.volumeMounts = append(k.volumeMounts, v1.VolumeMount{
			Name:      k.getVolumeName(secretName),
			ReadOnly:  false,
			MountPath: k.getSecretMountPath(secretName),
		})
	}
	absPath := k.getSecretAbsolutePath(path)
	return absPath, nil
}

// addVolume adds a volume from the secret referenced and return the path under which it should be mounted
// The boolean return value will be true if we've added a new volume
func (k *kubernetesSecretCollector) addVolume(secretName, key string) (string, bool) {
	p := k.getRelativeSecretFilePath(secretName, key)
	for i := range k.volumes {
		v := k.volumes[i]
		if v.Secret.SecretName == secretName {
			for j := range v.Secret.Items {
				// Secret already tracked?
				if v.Secret.Items[j].Key == key {
					// Let's reuse the same path
					return p, false
				}
			}
			// Let's add the key we want to mount
			v.Secret.Items = append(v.Secret.Items, v1.KeyToPath{
				Key:  key,
				Path: key,
			})
			return p, false
		}
	}

	// Secret has not been declared as a volume, we're adding it here
	k.volumes = append(k.volumes, v1.Volume{
		Name: k.getVolumeName(secretName),
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretName,
				Items: []v1.KeyToPath{
					{
						Key:  key,
						Path: key,
					},
				},
			},
		},
	})
	return p, true
}

func (k *kubernetesSecretCollector) handleSecretVarReference(secretName, key string) string {
	varName := k.getEnvVarName(secretName, key)
	for i := range k.envVars {
		if k.envVars[i].Name == varName {
			// The environment variable was already added, use it
			return k.getEnvVarNameReference(varName)
		}
	}
	k.envVars = append(k.envVars, v1.EnvVar{
		Name: varName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{secretName},
				Key:                  key,
			},
		},
	})
	return k.getEnvVarNameReference(varName)
}

func (k *kubernetesSecretCollector) setInDeployment(deployment *appsv1.Deployment) error {
	if len(k.envVars) == 0 && len(k.volumes) == 0 {
		return nil
	}

	container := util.GetContainerInDeployment(deployment, k.svc)
	if container == nil {
		return fmt.Errorf("unable to find container %s in deployment, cannot mount secrets", k.svc)
	}

	// Add volumes
	if len(k.volumes) > 0 {
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, k.volumes...)
		// Add volume mounts
		container.VolumeMounts = append(container.VolumeMounts, k.volumeMounts...)
	}

	// Add environment variables
	container.Env = append(container.Env, k.envVars...)
	return nil
}
