// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha2

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"./pkg/apis/spinnaker/v1alpha2.ExposeConfig":                 schema_pkg_apis_spinnaker_v1alpha2_ExposeConfig(ref),
		"./pkg/apis/spinnaker/v1alpha2.ExposeConfigService":          schema_pkg_apis_spinnaker_v1alpha2_ExposeConfigService(ref),
		"./pkg/apis/spinnaker/v1alpha2.ExposeConfigServiceOverrides": schema_pkg_apis_spinnaker_v1alpha2_ExposeConfigServiceOverrides(ref),
		"./pkg/apis/spinnaker/v1alpha2.SpinnakerAccount":             schema_pkg_apis_spinnaker_v1alpha2_SpinnakerAccount(ref),
		"./pkg/apis/spinnaker/v1alpha2.SpinnakerAccountSpec":         schema_pkg_apis_spinnaker_v1alpha2_SpinnakerAccountSpec(ref),
		"./pkg/apis/spinnaker/v1alpha2.SpinnakerAccountStatus":       schema_pkg_apis_spinnaker_v1alpha2_SpinnakerAccountStatus(ref),
		"./pkg/apis/spinnaker/v1alpha2.SpinnakerService":             schema_pkg_apis_spinnaker_v1alpha2_SpinnakerService(ref),
		"./pkg/apis/spinnaker/v1alpha2.SpinnakerServiceSpec":         schema_pkg_apis_spinnaker_v1alpha2_SpinnakerServiceSpec(ref),
		"./pkg/apis/spinnaker/v1alpha2.SpinnakerServiceStatus":       schema_pkg_apis_spinnaker_v1alpha2_SpinnakerServiceStatus(ref),
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_ExposeConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ExposeConfig represents the configuration for exposing Spinnaker",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"service": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.ExposeConfigService"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/spinnaker/v1alpha2.ExposeConfigService"},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_ExposeConfigService(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ExposeConfigService represents the configuration for exposing Spinnaker using k8s services",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"annotations": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
					"publicPort": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"overrides": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("./pkg/apis/spinnaker/v1alpha2.ExposeConfigServiceOverrides"),
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/spinnaker/v1alpha2.ExposeConfigServiceOverrides"},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_ExposeConfigServiceOverrides(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ExposeConfigServiceOverrides represents expose configurations of type service, overriden by specific services",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"publicPort": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"annotations": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_SpinnakerAccount(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpinnakerAccount is the Schema for the spinnakeraccounts API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.SpinnakerAccountSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.SpinnakerAccountStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/spinnaker/v1alpha2.SpinnakerAccountSpec", "./pkg/apis/spinnaker/v1alpha2.SpinnakerAccountStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_SpinnakerAccountSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpinnakerAccountSpec defines the desired state of SpinnakerAccount",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"enabled": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"validate": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"permissions": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"array"},
										Items: &spec.SchemaOrArray{
											Schema: &spec.Schema{
												SchemaProps: spec.SchemaProps{
													Type:   []string{"string"},
													Format: "",
												},
											},
										},
									},
								},
							},
						},
					},
					"auth": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"object"},
										Format: "",
									},
								},
							},
						},
					},
					"Env": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"object"},
										Format: "",
									},
								},
							},
						},
					},
					"settings": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"object"},
										Format: "",
									},
								},
							},
						},
					},
				},
				Required: []string{"enabled", "type", "validate", "permissions", "auth", "Env", "settings"},
			},
		},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_SpinnakerAccountStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpinnakerAccountStatus defines the observed state of SpinnakerAccount",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"valid": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"invalidReason": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"lastValidatedAt": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.Timestamp"),
						},
					},
				},
				Required: []string{"valid", "invalidReason", "lastValidatedAt"},
			},
		},
		Dependencies: []string{
			"k8s.io/apimachinery/pkg/apis/meta/v1.Timestamp"},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_SpinnakerService(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpinnakerService is the Schema for the spinnakerservices API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.SpinnakerServiceSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.SpinnakerServiceStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/spinnaker/v1alpha2.SpinnakerServiceSpec", "./pkg/apis/spinnaker/v1alpha2.SpinnakerServiceStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_SpinnakerServiceSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpinnakerServiceSpec defines the desired state of SpinnakerService",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"spinnakerConfig": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.SpinnakerConfig"),
						},
					},
					"expose": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.ExposeConfig"),
						},
					},
					"accounts": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/spinnaker/v1alpha2.AccountConfig"),
						},
					},
				},
				Required: []string{"spinnakerConfig"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/spinnaker/v1alpha2.AccountConfig", "./pkg/apis/spinnaker/v1alpha2.ExposeConfig", "./pkg/apis/spinnaker/v1alpha2.SpinnakerConfig"},
	}
}

func schema_pkg_apis_spinnaker_v1alpha2_SpinnakerServiceStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpinnakerServiceStatus defines the observed state of SpinnakerService",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"version": {
						SchemaProps: spec.SchemaProps{
							Description: "Current deployed version of Spinnaker",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"lastConfigurationTime": {
						SchemaProps: spec.SchemaProps{
							Description: "Last time the configuration was updated",
							Ref:         ref("k8s.io/apimachinery/pkg/apis/meta/v1.Time"),
						},
					},
					"lastConfigHash": {
						SchemaProps: spec.SchemaProps{
							Description: "Last deployed hash",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"services": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-map-keys": "name",
								"x-kubernetes-list-type":     "map",
							},
						},
						SchemaProps: spec.SchemaProps{
							Description: "Services deployment information",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("./pkg/apis/spinnaker/v1alpha2.SpinnakerDeploymentStatus"),
									},
								},
							},
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Description: "Overall Spinnaker status",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"serviceCount": {
						SchemaProps: spec.SchemaProps{
							Description: "Number of services in Spinnaker",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"uiUrl": {
						SchemaProps: spec.SchemaProps{
							Description: "Exposed Deck URL",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiUrl": {
						SchemaProps: spec.SchemaProps{
							Description: "Exposed Gate URL",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"accountCount": {
						SchemaProps: spec.SchemaProps{
							Description: "Number of accounts",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/spinnaker/v1alpha2.SpinnakerDeploymentStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.Time"},
	}
}
