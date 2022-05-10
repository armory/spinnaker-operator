package main

import (
	"github.com/armory/spinnaker-operator/pkg/api"
	"github.com/armory/spinnaker-operator/pkg/api/accounts"
	"github.com/armory/spinnaker-operator/pkg/api/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/armory/spinnaker-operator/pkg/api/operator"
	"github.com/armory/spinnaker-operator/pkg/api/v1alpha2"

	"github.com/armory/spinnaker-operator/controllers/accountvalidating"
	"github.com/armory/spinnaker-operator/controllers/spinnakeraccount"
	"github.com/armory/spinnaker-operator/controllers/spinnakerservice"
	"github.com/armory/spinnaker-operator/controllers/spinnakervalidating"
)

func main() {
	v1alpha2.RegisterTypes()
	spinnakervalidating.TypesFactory = interfaces.DefaultTypesFactory
	accountvalidating.TypesFactory = interfaces.DefaultTypesFactory
	spinnakerservice.TypesFactory = interfaces.DefaultTypesFactory
	spinnakeraccount.TypesFactory = interfaces.DefaultTypesFactory
	accounts.TypesFactory = interfaces.DefaultTypesFactory
	kubernetes.TypesFactory = interfaces.DefaultTypesFactory
	operator.Start(api.AddToScheme)
}
