package main

import (
	"github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/controller/accountvalidating"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakeraccount"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakerservice"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakervalidating"
	"github.com/armory/spinnaker-operator/pkg/operator"
)

func main() {
	v1alpha2.RegisterTypes()
	spinnakervalidating.TypesFactory = interfaces.DefaultTypesFactory
	accountvalidating.TypesFactory = interfaces.DefaultTypesFactory
	spinnakerservice.TypesFactory = interfaces.DefaultTypesFactory
	spinnakeraccount.TypesFactory = interfaces.DefaultTypesFactory
	accounts.TypesFactory = interfaces.DefaultTypesFactory
	kubernetes.TypesFactory = interfaces.DefaultTypesFactory
	operator.Start(apis.AddToScheme)
}
