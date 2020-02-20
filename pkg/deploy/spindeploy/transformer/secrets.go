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
	genAccessKey  v1.EnvVar   // AWS_ACCESS_KEY_ID
	genSecretKey  v1.EnvVar   // AWS_SECRET_ACCESS_KEY
	svcSecretKeys []v1.EnvVar // SERVICE_SECRETNAME_SECRETKEY
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
	persistenceKeys, err := s.getAndReplace("front50", awsPersistenceAccessKey, awsPersistenceSecretKey, spinCfg)
	if err != nil {
		return err
	}
	if persistenceKeys != nil {
		s.k8sSecrets.awsCredsByService["front50"] = persistenceKeys
	}
	providerKeys, err := s.getAndReplace("clouddriver", awsProviderAccessKey, awsProviderSecretKey, spinCfg)
	if err != nil {
		return err
	}
	if providerKeys != nil {
		s.k8sSecrets.awsCredsByService["clouddriver"] = providerKeys
	}
	artifactKeys, err := s.getAndReplaceArray("clouddriver", awsArtifactsRootKey, awsArtifactsAccessKey, awsArtifactsSecretKey, spinCfg)
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

func (s *secretsTransformer) getAndReplace(svc, accessKeyProp, secretKeyProp string, spinCfg *v1alpha2.SpinnakerConfig) (*awsCredentials, error) {
	secretRaw, source, err := spinCfg.GetRawConfigPropString(svc, secretKeyProp)
	if err != nil {
		return nil, nil
	}
	if !isK8sEngine(secretRaw) {
		return nil, nil
	}
	accessKeyRaw, _, err := spinCfg.GetRawConfigPropString(svc, accessKeyProp)
	if err != nil || accessKeyRaw == "" {
		return nil, fmt.Errorf("aws secret key configured without access key for property %s", secretKeyProp)
	}
	envVarName, secretName, secretKey := getEnvVarNameFromSecretRaw(svc, secretRaw)
	switch source {
	case v1alpha2.HalConfigSource:
		err = spinCfg.SetHalConfigProp(secretKeyProp, getEnvVarNameReference(envVarName))
		break
	case v1alpha2.ProfileConfigSource:
		err = spinCfg.SetServiceConfigProp(svc, secretKeyProp, getEnvVarNameReference(envVarName))
	}
	if err != nil {
		return nil, err
	}
	return &awsCredentials{
		genAccessKey:  envVarFromRawString("AWS_ACCESS_KEY_ID", accessKeyRaw),
		genSecretKey:  envVarFromSecretReference("AWS_SECRET_ACCESS_KEY", secretName, secretKey),
		svcSecretKeys: []v1.EnvVar{envVarFromSecretReference(envVarName, secretName, secretKey)},
	}, nil
}

// getAndReplaceArray retrieves a single aws access and secret key pair from an input array (last one wins)
func (s *secretsTransformer) getAndReplaceArray(svc, rootProp, accessKeyProp, secretKeyProp string, spinCfg *v1alpha2.SpinnakerConfig) (*awsCredentials, error) {
	root, source, err := spinCfg.GetConfigObjectArray(svc, rootProp)
	if err != nil {
		// ignore error if key doesn't exist
		return nil, nil
	}
	var genAccessKey v1.EnvVar
	var genSecretKey v1.EnvVar
	var svcSecretKeys []v1.EnvVar
	for _, i := range root {
		secretRaw, ok := i[secretKeyProp].(string)
		if !ok {
			continue
		}
		if !isK8sEngine(secretRaw) {
			continue
		}
		envVarName, secretName, secretKey := getEnvVarNameFromSecretRaw(svc, secretRaw)
		i[secretKeyProp] = getEnvVarNameReference(envVarName)
		accessKey, ok := i[accessKeyProp].(string)
		if !ok {
			return nil, fmt.Errorf("aws secret access key specified without access key under %s", root)
		}
		genAccessKey = envVarFromRawString("AWS_ACCESS_KEY_ID", accessKey)
		genSecretKey = envVarFromSecretReference("AWS_SECRET_ACCESS_KEY", secretName, secretKey)
		svcSecretKeys = append(svcSecretKeys, envVarFromSecretReference(envVarName, secretName, secretKey))
	}
	switch source {
	case v1alpha2.HalConfigSource:
		err = spinCfg.SetHalConfigProp(rootProp, root)
		break
	case v1alpha2.ProfileConfigSource:
		err = spinCfg.SetServiceConfigProp(svc, rootProp, root)
	}
	if err != nil {
		return nil, err
	}
	if len(svcSecretKeys) == 0 {
		return nil, nil
	}
	return &awsCredentials{
		genAccessKey:  genAccessKey,
		genSecretKey:  genSecretKey,
		svcSecretKeys: svcSecretKeys,
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
				err := kCollector.mapSecrets(sec)
				if err != nil {
					return err
				}
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
func (k *kubernetesSecretCollector) mapSecrets(secret *v1.Secret) error {
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
	k.addEnvVarIfNotAdded(svcKeys.genAccessKey)
	k.addEnvVarIfNotAdded(svcKeys.genSecretKey)
	for _, e := range svcKeys.svcSecretKeys {
		k.addEnvVarIfNotAdded(e)
	}
	return nil
}

func (k *kubernetesSecretCollector) addEnvVarIfNotAdded(envVar v1.EnvVar) {
	for i := range k.envVars {
		if k.envVars[i].Name == envVar.Name {
			return
		}
	}
	k.envVars = append(k.envVars, envVar)
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

// Environment name replacer: we know input will be valid secret name and keys or services
// Change when a new "Spin?!" service is added.
var envReplacer = strings.NewReplacer("-", "_", ".", "_", " ", "_")

func safeEnvVarName(val string) string {
	return strings.ToUpper(envReplacer.Replace(val))
}

func (k *kubernetesSecretCollector) handleSecretFileReference(secretName, key string) (string, error) {
	vPath, added := k.addVolume(secretName, key)
	// Add a volume mount
	if added {
		k.volumeMounts = append(k.volumeMounts, v1.VolumeMount{
			Name:      k.getVolumeName(secretName),
			ReadOnly:  false,
			MountPath: k.getSecretMountPath(secretName),
		})
	}
	absPath := k.getSecretAbsolutePath(vPath)
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
	varName := getEnvVarName(k.svc, secretName, key)
	envVar := envVarFromSecretReference(varName, secretName, key)
	k.addEnvVarIfNotAdded(envVar)
	return getEnvVarNameReference(varName)
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

func getEnvVarName(svc, secretName, key string) string {
	return fmt.Sprintf("%s_%s_%s", safeEnvVarName(svc), safeEnvVarName(secretName), safeEnvVarName(key))
}

func getEnvVarNameFromSecretRaw(svc, secretRaw string) (envVarName, secretName, secretKey string) {
	secretName, secretKey = secrets.ParseKubernetesSecretParams(secretRaw)
	envVarName = getEnvVarName(svc, secretName, secretKey)
	return
}

func getEnvVarNameReference(varName string) string {
	return fmt.Sprintf("${%s}", varName)
}

func envVarFromSecretReference(varName, secretName, secretKey string) v1.EnvVar {
	return v1.EnvVar{
		Name: varName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{secretName},
				Key:                  secretKey,
			},
		},
	}
}

func envVarFromRawString(varName, varValue string) v1.EnvVar {
	return v1.EnvVar{Name: varName, Value: varValue}
}

func isK8sEngine(s string) bool {
	e, _, _ := secups.GetEngine(s)
	return e == "k8s"
}
