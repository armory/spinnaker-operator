package transformer

import (
	"context"
	"fmt"
	secups "github.com/armory/go-yaml-tools/pkg/secrets"
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

const (
	awsProviderAccessKey    = "providers.aws.accessKeyId"
	awsProviderSecretKey    = "providers.aws.secretAccessKey"
	awsPersistenceAccessKey = "persistentStorage.s3.accessKeyId"
	awsPersistenceSecretKey = "persistentStorage.s3.secretAccessKey"
	awsArtifactsRootKey     = "artifacts.s3.accounts"
	awsArtifactsAccessKey   = "awsAccessKeyId"
	awsArtifactsSecretKey   = "awsSecretAccessKey"
	awsCanary               = "canary"
)

// secretsTransformer maps Kubernetes secrets onto the deployment of the service that requires it
// Either as a mounted file (encryptedFile) or an environment variable (tokens, passwords...)
type secretsTransformer struct {
	svc        v1alpha2.SpinnakerServiceInterface
	log        logr.Logger
	client     client.Client
	k8sSecrets *k8sSecretHolder
}

type secretsTransformerGenerator struct{}

// k8sSecretHolder keeps track of kubernetes secret references that appear in selected fields of the config.
type k8sSecretHolder struct {
	awsCredsByService map[string]*awsCredentials
}

type awsCredentials struct {
	accessKeyId     string
	secretAccessKey string
}

func (s *secretsTransformerGenerator) NewTransformer(svc v1alpha2.SpinnakerServiceInterface,
	client client.Client, log logr.Logger) (Transformer, error) {
	tr := secretsTransformer{svc: svc, log: log, client: client, k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}
	return &tr, nil
}

func (s *secretsTransformerGenerator) GetName() string {
	return "Secrets"
}

func (s *secretsTransformer) TransformConfig(ctx context.Context) error {
	spinCfg := s.svc.GetSpinnakerConfig()
	return s.replaceK8sSecretsFromAwsKeys(spinCfg, ctx)
}

// replaceK8sSecretsFromAwsKeys replaces any kubernetes secret references from aws credentials fields and saves them for later processing
func (s *secretsTransformer) replaceK8sSecretsFromAwsKeys(spinCfg *v1alpha2.SpinnakerConfig, ctx context.Context) error {
	persistenceKeys, err := s.getAndReplace(awsPersistenceAccessKey, awsPersistenceSecretKey, spinCfg, ctx)
	if err != nil {
		return err
	}
	if persistenceKeys != nil {
		s.k8sSecrets.awsCredsByService["front50"] = persistenceKeys
	}
	providerKeys, err := s.getAndReplace(awsProviderAccessKey, awsProviderSecretKey, spinCfg, ctx)
	if err != nil {
		return err
	}
	if providerKeys != nil {
		s.k8sSecrets.awsCredsByService["clouddriver"] = providerKeys
	}
	artifactKeys, err := s.getAndReplaceArray(awsArtifactsRootKey, awsArtifactsAccessKey, awsArtifactsSecretKey, spinCfg, ctx)
	if err != nil {
		return err
	}
	if artifactKeys != nil {
		s.k8sSecrets.awsCredsByService["clouddriver"] = artifactKeys
	}
	can, ok := spinCfg.Config[awsCanary]
	if !ok {
		return nil
	}
	newCan, err := s.sanitizeK8sSecret(can, ctx)
	if err != nil {
		return err
	}
	spinCfg.Config[awsCanary] = newCan
	return nil
}

func (s *secretsTransformer) getAndReplace(accessKeyProp, secretKeyProp string, spinCfg *v1alpha2.SpinnakerConfig, ctx context.Context) (*awsCredentials, error) {
	secretRaw, err := spinCfg.GetRawHalConfigPropString(secretKeyProp)
	if err != nil {
		// ignore error if key doesn't exist
		return nil, nil
	}
	e, _, _ := secups.GetEngine(secretRaw)
	if e != "k8s" {
		return nil, nil
	}
	err = spinCfg.SetHalConfigProp(secretKeyProp, "OVERRIDDEN_BY_ENV_VARS")
	if err != nil {
		return nil, err
	}
	accessKey, err := spinCfg.GetHalConfigPropString(ctx, accessKeyProp)
	if err != nil {
		return nil, fmt.Errorf("aws secret key configured without access key for property %s", secretKeyProp)
	}
	return &awsCredentials{
		accessKeyId:     accessKey,
		secretAccessKey: secretRaw,
	}, nil
}

// getAndReplaceArray retrieves a single aws access and secret key pair from an input array (last one wins)
func (s *secretsTransformer) getAndReplaceArray(rootProp, accessKeyProp, secretKeyProp string, spinCfg *v1alpha2.SpinnakerConfig, ctx context.Context) (*awsCredentials, error) {
	root, err := spinCfg.GetHalConfigObjectArray(ctx, rootProp)
	if err != nil {
		// ignore error if key doesn't exist
		return nil, nil
	}
	var secretKey string
	var accessKey string
	ok := false
	for _, i := range root {
		secretKey, ok = i[secretKeyProp].(string)
		if !ok {
			continue
		}
		e, _, _ := secups.GetEngine(secretKey)
		if e != "k8s" {
			continue
		}
		i[secretKeyProp] = "OVERRIDDEN_BY_ENV_VARS"
		accessKey, ok = i[accessKeyProp].(string)
		if !ok {
			return nil, fmt.Errorf("aws secret access key specified without access key under %s", root)
		}
	}
	err = spinCfg.SetHalConfigProp(rootProp, root)
	if err != nil {
		return nil, err
	}
	return &awsCredentials{
		accessKeyId:     accessKey,
		secretAccessKey: secretKey,
	}, nil
}

func (s *secretsTransformer) sanitizeK8sSecret(object interface{}, ctx context.Context) (interface{}, error) {
	h := func(val string) (string, error) {
		if !secups.IsEncryptedSecret(val) {
			return val, nil
		}
		e, _, _ := secups.GetEngine(val)
		if e != "k8s" {
			return val, nil
		}
		s, f, err := secrets.Decode(ctx, val)
		if err != nil {
			return "", err
		}
		if f {
			return "", fmt.Errorf("\"encryptedFile...\" specified for a non file property (%s), should be \"encrypted...\" instead", val)
		}
		return s, nil
	}
	return inspect.InspectStrings(object, h)
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
		err := kCollector.mapAwsKeys(s.k8sSecrets)
		if err != nil {
			return err
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

// mapSecrets goes through all secret data and replace references to passwords and files with env variables
// and file paths
func (k *kubernetesSecretCollector) mapSecrets(svc string, secret *v1.Secret) error {
	for key := range secret.Data {
		// Is this a json string?
		v := secret.Data[key]
		s := strings.TrimSpace(string(v))
		if (s[0] == '{' && s[len(s)-1] == '}') || (s[0] == '[' && s[len(s)-1] == ']') {
			continue
		}
		// Attempt to deserialize as YAML
		m := make(map[string]interface{})
		if err := yaml.Unmarshal(v, &m); err != nil {
			continue
		}
		// If it's YAML replace secret references
		ndata, err := k.sanitizeSecrets(m)
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

// mapAwsKeys adds env vars for AWS keys if needed
func (k *kubernetesSecretCollector) mapAwsKeys(keys *k8sSecretHolder) error {
	svcKeys, ok := keys.awsCredsByService[k.svc]
	if !ok {
		return nil
	}
	// if keys were already added, skip
	for i := range k.envVars {
		if k.envVars[i].Name == "AWS_ACCESS_KEY_ID" {
			return nil
		}
	}
	n, secretKey := secrets.ParseKubernetesSecretParams(svcKeys.secretAccessKey)
	k.envVars = append(k.envVars, v1.EnvVar{
		Name:  "AWS_ACCESS_KEY_ID",
		Value: svcKeys.accessKeyId,
	})
	k.envVars = append(k.envVars, v1.EnvVar{
		Name: "AWS_SECRET_ACCESS_KEY",
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{n},
				Key:                  secretKey,
			},
		},
	})

	return nil
}

func (k *kubernetesSecretCollector) sanitizeSecrets(obj interface{}) (interface{}, error) {
	h := func(val string) (string, error) {
		e, f, p := secups.GetEngine(val)
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
