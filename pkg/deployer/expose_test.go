package deployer

import (
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

func TestTransformConfigNoExposeNoOverrideBaseUrl(t *testing.T) {
	// no expose configuration, no overrideBaseUrl set in hal config
	g := exposeTransformerGenerator{}
	fakeClient := fake.NewFakeClient()
	spinSvc, _ := buildSpinSvc("1")
	tr, _ := g.NewTransformer(spinSvc, fakeClient, logf.Log.WithName("spinnakerservice"))
	hc := &halconfig.SpinnakerConfig{}

	err := tr.TransformConfig(hc)

	assert.Nil(t, err)
	url, _ := hc.GetHalConfigPropString("security.apiSecurity.overrideBaseUrl")
	assert.Equal(t, "", url)
	url, _ = hc.GetHalConfigPropString("security.uiSecurity.overrideBaseUrl")
	assert.Equal(t, "", url)
}

func TestTransformConfigExposedNoOverrideBaseUrlNoServices(t *testing.T) {
	// expose configuration, no overrideBaseUrl set in hal config, no running services to fetch their LB url
	g := exposeTransformerGenerator{}
	fakeClient := fake.NewFakeClient()
	spinSvc, _ := buildSpinSvc("1")
	spinSvc.Spec.Expose.Type = "LoadBalancer"
	tr, _ := g.NewTransformer(spinSvc, fakeClient, logf.Log.WithName("spinnakerservice"))
	hc := &halconfig.SpinnakerConfig{}

	err := tr.TransformConfig(hc)

	assert.Nil(t, err)
	url, _ := hc.GetHalConfigPropString("security.apiSecurity.overrideBaseUrl")
	assert.Equal(t, "", url)
	url, _ = hc.GetHalConfigPropString("security.uiSecurity.overrideBaseUrl")
	assert.Equal(t, "", url)
}
