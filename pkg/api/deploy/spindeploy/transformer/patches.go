package transformer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/armory/spinnaker-operator/pkg/api/generated"
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/code-generator/_examples/apiserver/clientset/internalversion/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Applies patches
type patchTransformer struct {
	svc interfaces.SpinnakerService
	log logr.Logger
}

type PatchTransformerGenerator struct{}

func (g *PatchTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (Transformer, error) {

	tr := patchTransformer{svc: svc, log: log}
	return &tr, nil
}

func (g *PatchTransformerGenerator) GetName() string {
	return "Patches"
}

// nop
func (p *patchTransformer) TransformConfig(ctx context.Context) error {
	return nil
}

func (p *patchTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	for k, kust := range p.svc.GetKustomization() {
		s, ok := gen.Config[k]
		if ok {
			if err := p.applyKustomization(k, kust, &s); err != nil {
				return err
			}
			gen.Config[k] = s
		}
	}
	return nil
}

func (p *patchTransformer) applyKustomization(spinsvc string, kustomization interfaces.ServiceKustomization, config *generated.ServiceConfig) error {
	if kustomization.Deployment != nil {
		if config.Deployment == nil {
			return fmt.Errorf("deployment not generated for service %s, unable to apply kustomization", spinsvc)
		}
		data, err := p.applyPatches(spinsvc, kustomization.Deployment, config.Deployment)
		if err != nil {
			return err
		}
		dep, err := p.asDeployment(data)
		if err != nil {
			return err
		}
		config.Deployment = dep
	}
	if kustomization.Service != nil {
		if config.Service == nil {
			return fmt.Errorf("service not generated for service %s, unable to apply kustomization", spinsvc)
		}
		data, err := p.applyPatches(spinsvc, kustomization.Service, config.Service)
		if err != nil {
			return err
		}
		svc, err := p.asService(data)
		if err != nil {
			return err
		}
		config.Service = svc
	}
	return nil
}

func (p *patchTransformer) applyPatches(spinsvc string, kustomization *interfaces.Kustomization, obj runtime.Object) ([]byte, error) {
	originalObjJS, err := runtime.Encode(unstructured.UnstructuredJSONScheme, obj)
	if err != nil {
		return nil, err
	}
	c := &creator{}
	gvk := obj.GetObjectKind().GroupVersionKind()
	for i, patch := range kustomization.Patches {
		originalObjJS, err = getPatchedJSON(types.MergePatchType, originalObjJS, []byte(patch), gvk, c)
		if err != nil {
			return nil, fmt.Errorf("unable to apply json patch for %s at index %d: %v", spinsvc, i, err)
		}
	}
	if kustomization.PatchesJson6902 != "" {
		originalObjJS, err = getPatchedJSON(types.JSONPatchType, originalObjJS, []byte(kustomization.PatchesJson6902), gvk, c)
		if err != nil {
			return nil, fmt.Errorf("unable to apply json6902 patch for %s: %v", spinsvc, err)
		}
	}
	for i, patch := range kustomization.PatchesStrategicMerge {
		originalObjJS, err = getPatchedJSON(types.StrategicMergePatchType, originalObjJS, []byte(patch), gvk, c)
		if err != nil {
			return nil, fmt.Errorf("unable to apply strategic merge patch for %s at index %d: %v", spinsvc, i, err)
		}
	}
	return originalObjJS, nil
}

func (p *patchTransformer) asDeployment(data []byte) (*appsv1.Deployment, error) {
	dser := scheme.Codecs.UniversalDecoder()
	dep := &appsv1.Deployment{}
	obj, _, err := dser.Decode(data, &schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}, dep)
	if err != nil {
		return nil, err
	}
	if dep, ok := obj.(*appsv1.Deployment); ok {
		return dep, nil
	}
	return nil, errors.New("not parsed as a deployment, patching failed")
}

func (p *patchTransformer) asService(data []byte) (*v1.Service, error) {
	dser := scheme.Codecs.UniversalDecoder()
	svc := &v1.Service{}
	obj, _, err := dser.Decode(data, &schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	}, svc)
	if err != nil {
		return nil, err
	}
	if svc, ok := obj.(*v1.Service); ok {
		return svc, nil
	}
	return nil, errors.New("not parsed as a service, patching failed")
}

type creator struct{}

func (c *creator) New(gvk schema.GroupVersionKind) (runtime.Object, error) {
	if gvk.Kind == "Deployment" {
		return &appsv1.Deployment{}, nil
	}
	if gvk.Kind == "Service" {
		return &v1.Service{}, nil
	}
	return nil, fmt.Errorf("unknown gvk to patch %s", gvk.String())
}

func getPatchedJSON(patchType types.PatchType, originalJS, patchJS []byte, gvk schema.GroupVersionKind, creater runtime.ObjectCreater) ([]byte, error) {
	patchJS, err := yaml.ToJSON(patchJS)
	if err != nil {
		return nil, err
	}

	switch patchType {
	case types.JSONPatchType:
		patchObj, err := jsonpatch.DecodePatch(patchJS)
		if err != nil {
			return nil, err
		}
		bytes, err := patchObj.Apply(originalJS)
		// TODO: This is pretty hacky, we need a better structured error from the json-patch
		if err != nil && strings.Contains(err.Error(), "doc is missing key") {
			msg := err.Error()
			ix := strings.Index(msg, "key:")
			key := msg[ix+5:]
			return bytes, fmt.Errorf("object to be patched is missing field (%s)", key)
		}
		return bytes, err

	case types.MergePatchType:
		return jsonpatch.MergePatch(originalJS, patchJS)

	case types.StrategicMergePatchType:
		// get a typed object for this GVK if we need to apply a strategic merge patch
		obj, err := creater.New(gvk)
		if err != nil {
			return nil, fmt.Errorf("cannot apply strategic merge patch for %s locally", gvk.String())
		}
		return strategicpatch.StrategicMergePatch(originalJS, patchJS, obj)

	default:
		return nil, fmt.Errorf("unknown patching method: %v", patchType)
	}
}
