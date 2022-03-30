module github.com/armory/spinnaker-operator

require (
	github.com/armory/go-yaml-tools v0.0.0-20200316192928-75770481ad01
	github.com/aws/aws-sdk-go v1.31.9
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/coreos/prometheus-operator v0.41.1 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v1.2.0
	github.com/golang/mock v1.5.0
	github.com/jinzhu/copier v0.3.5
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/openshift/origin v0.0.0-20160503220234-8f127d736703
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	golang.org/x/tools v0.1.9 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.5
	k8s.io/apiextensions-apiserver v0.23.5 // indirect
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.23.5
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20220324211241-9f9c01d62a3a
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/yaml v1.3.0
)

// Pinned to kubernetes-1.14.1
replace (
	k8s.io/api => k8s.io/api v0.23.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	k8s.io/client-go => k8s.io/client-go v0.22.8
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
	k8s.io/component-base => k8s.io/component-base v0.20.15
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190409022021-00b8e31abe9d
	// k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20220324211241-9f9c01d62a3a
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20211104224443-923526ac052c
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.1
)

// Remove when controller-tools v0.2.2 is released
// Required for the bugfix https://github.com/kubernetes-sigs/controller-tools/pull/322
replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.2-0.20190919011008-6ed4ff330711

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.19.4

replace github.com/operator-framework/operator-registry => github.com/operator-framework/operator-registry v1.2.0

go 1.13
