# Copyright 2019, Armory
#
# Licensed under the Apache License, Version 2.0 (the "License")
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Inspired by MySQL Operator Makefile: https://github.com/oracle/mysql-operator/blob/master/Makefile

VERSION_TYPE    ?= "snapshot" # Must be one of: "snapshot", "rc", or "release"
BRANCH_OVERRIDE ?=
VERSION 	 	?= $(shell build-tools/version.sh $(VERSION_TYPE) $(BRANCH_OVERRIDE))
REGISTRY_ORG    ?= "armory"
OS      	 	?= $(shell go version | cut -d' ' -f 4 | cut -d'/' -f 1)
GOOS			:=${OS}
GOARCH			:=${ARCH}
ARCH    	 	?= $(shell go version | cut -d' ' -f 4 | cut -d'/' -f 2)
NAMESPACE 	 	?= "spinnaker-operator"
PWD 		  	= $(shell pwd)

REGISTRY        ?= docker.io
SRC_DIRS        := cmd pkg integration-tests
COMMAND         := cmd/manager/main
BUILD_HOME      := ${PWD}/build
BUILD_MF_DIR    := ${BUILD_HOME}/manifests
BUILD_BIN_DIR   := ${BUILD_HOME}/bin/$(OS)_$(ARCH)
BINARY 			:= ${BUILD_BIN_DIR}/spinnaker-operator
KUBECONFIG		?= ${HOME}/.kube/config
.DEFAULT_GOAL   := help
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/controller-gen
CG_WRAPPER := $(TOOLS_BIN_DIR)/cg_wrapper.sh


.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: docker-build docker-test docker-package

.PHONY: version
version: ## Prints the version of operator. Version type can be changed providing VERSION_TYPE param.
	@echo $(VERSION)

.PHONY: clean
clean: ## Deletes output directory
	rm -rf $(BUILD_HOME)

.PHONY: build
build: build-dirs manifest Makefile ## Compiles the code to produce binaries
	@echo "Operator version: $(VERSION)"
	@echo "Building: $(BINARY)"
	@go build -mod=vendor -o ${BINARY} cmd/manager/main.go

.PHONY: docker-build
docker-build: Makefile ## Runs "make build" in a docker container
	@echo "Running \"make build\" in docker"
	@docker buildx build \
	-t docker-local/$(REGISTRY_ORG)/spinnaker-operator-builder:$(VERSION) \
	--build-arg VERSION=${VERSION} \
	-f build-tools/Dockerfile .

.PHONY: test
test: Makefile ## Run unit tests. Doesn't need to compile the code.
	@go test -cover -mod=vendor ./...

.PHONY: docker-test
docker-test: Makefile ## Runs "make test" in a docker container
	@echo "Running \"make test\" in docker"
	@docker buildx build \
	--target tests \
	-f build-tools/Dockerfile .

.PHONY: integration-test
integration-test: build-dirs Makefile ## Run integration tests. See requirements in integration_tests/README.md
	@go test -tags=integration -mod=vendor -timeout=40m -run IntegrationTests/Operator= ./integration_tests/...

.PHONY: docker-package-push
docker-package-push: Makefile ## Builds and pushes the docker image to distribute
	@echo "Packaging and pushing final docker image"
	@docker  buildx build \
	--platform=linux/arm64,linux/amd64 \
	-t $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:$(VERSION) \
	--push \
	--build-arg CACHE_DATE=$(shell date +%s) \
	-f build-tools/Dockerfile .
	@echo "Successfully built and pushed image with tag $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:$(VERSION)"
	
.PHONY: docker-package-push-with-dev
docker-package-push-with-dev: Makefile ## Builds and pushes the docker image to distribute
	@echo "Packaging and pushing final docker image"
	@docker buildx build \
	--platform=linux/arm64,linux/amd64 \
	-t $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:$(VERSION) \
	-t $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:dev
	--push \
	--build-arg CACHE_DATE=$(shell date +%s) \
	-f build-tools/Dockerfile .
	@echo "Successfully built and pushed image with tags: $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:$(VERSION) $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:dev"

.PHONY: reverse-proxy
reverse-proxy: ## Installs a reverse proxy in Kubernetes to be able to debug locally
	kubectl --kubeconfig=${KUBECONFIG} create cm ssh-key --from-file=authorized_keys=${HOME}/.ssh/id_rsa.pub --dry-run -o yaml | kubectl apply -f -
	kubectl --kubeconfig=${KUBECONFIG} apply -f build-tools/deployment-reverseproxy.yml
	sleep 5
	kubectl --kubeconfig=${KUBECONFIG} port-forward deployment/spinnaker-operator-proxy 2222:22 & echo $$! > pf-pid
	sleep 5
	echo 'please set OPERATOR_NAME env var to spinnaker-operator-proxy' > /dev/stderr
	ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -TNngR 9876:localhost:9876 ssh://root@localhost:2222
	kill `cat pf-pid` && rm pf-pid

.PHONY: run-dev
run-dev: ## Runs operator locally
	@WATCH_NAMESPACE=$(NAMESPACE) && go run \
	    cmd/manager/main.go \
	    --kubeconfig=${KUBECONFIG}

.PHONY: debug
debug: ## Debugs operator locally
	OPERATOR_NAME=local-operator \
    WATCH_NAMESPACE=$(NAMESPACE) \
	dlv debug --headless --listen=:2345 --headless --log --api-version=2 cmd/manager/main.go -- \
	--kubeconfig ${KUBECONFIG} --disable-admission-controller

.PHONY: build-dirs
build-dirs:
	@echo "Creating build directories ${BUILD_BIN_DIR}, ${BUILD_MF_DIR}"
	@mkdir -p $(BUILD_BIN_DIR)
	@mkdir -p $(BUILD_MF_DIR)

.PHONY: manifest
manifest: build-dirs ## Copies and packages kubernetes manifest files with final docker image tags
	@echo "Generating operator MANIFEST file"
	@echo "Version="$(VERSION) > $(BUILD_BIN_DIR)/MANIFEST
	@echo "Built-By="$(shell whoami) >> $(BUILD_BIN_DIR)/MANIFEST
	@echo "Build-Date="$(shell date +'%Y-%m-%d_%H:%M:%S') >> $(BUILD_BIN_DIR)/MANIFEST
	@echo "Branch="$(shell git rev-parse --abbrev-ref HEAD) >> $(BUILD_BIN_DIR)/MANIFEST
	@echo "Revision="$(shell git describe --always) >> $(BUILD_BIN_DIR)/MANIFEST
	@echo "Build-Go-Version="$(shell go version) >> $(BUILD_BIN_DIR)/MANIFEST
	@echo "Copying kubernetes manifests"
	@cp -R deploy ${BUILD_MF_DIR}
	@if [[ -f ${BUILD_MF_DIR}/deploy/role.yaml ]] ; then rm ${BUILD_MF_DIR}/deploy/role.yaml ; fi
	@cat ${BUILD_MF_DIR}/deploy/operator/basic/deployment.yaml | sed "s|image: armory/spinnaker-operator:.*|image: $(REGISTRY_ORG)/spinnaker-operator:$(VERSION)|" | sed "s|image: armory/halyard:.*|image: armory/halyard:$(shell cat halyard-version | head -1)|" | sed "s|imagePullPolicy:.*|imagePullPolicy: IfNotPresent|" > ${BUILD_MF_DIR}/deploy/operator/basic/deployment.yaml.new
	@mv ${BUILD_MF_DIR}/deploy/operator/basic/deployment.yaml.new ${BUILD_MF_DIR}/deploy/operator/basic/deployment.yaml
	@cat ${BUILD_MF_DIR}/deploy/operator/cluster/deployment.yaml | sed "s|image: armory/spinnaker-operator:.*|image: $(REGISTRY_ORG)/spinnaker-operator:$(VERSION)|" | sed "s|image: armory/halyard:.*|image: armory/halyard:$(shell cat halyard-version | head -1)|" | sed "s|imagePullPolicy:.*|imagePullPolicy: IfNotPresent|" > ${BUILD_MF_DIR}/deploy/operator/cluster/deployment.yaml.new
	@mv ${BUILD_MF_DIR}/deploy/operator/cluster/deployment.yaml.new ${BUILD_MF_DIR}/deploy/operator/cluster/deployment.yaml
	@cd $(BUILD_MF_DIR) && tar -czf manifests.tgz deploy/ && mv manifests.tgz ..

.PHONY: lint
lint: ## Executes golint in all source files
	@find pkg cmd -name '*.go' | grep -v 'generated' | xargs -L 1 golint

.PHONY: k8s
k8s: ## Generates "deep copy" code from pkg/apis modules
	@go run tools/generate.go k8s

.PHONY: openapi
openapi: ## Generates the CRDs from pkg/apis modules
	@$(CG_WRAPPER)

.PHONY: openapi-internal
openapi-internal: build-dirs $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=deploy/crds

$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod # Build controller-gen from tools folder.
	cd $(TOOLS_DIR) && go build -tags=tools -o bin/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen

.PHONY: generate
generate: $(CONTROLLER_GEN) ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: addapi
addapi: ## Adds a new version of the CRD in pkg/apis
	@go run tools/add.go ${NEW_API_VERSION}
	rm deploy/crds/*${NEW_API_VERSION}*
	@echo "***** MANUAL TODO, YOU'RE NOT FINISHED YET ******"
	@echo "- Copy the contents of the previous version '_types.go' file into the new version"
	@echo "- Change storage version to new api by deleting '+kubebuilder:storageversion' comment above SpinnakerService struct from the previous version."
