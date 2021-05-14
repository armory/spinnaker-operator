# Development Guide for Spinnaker Operator

This document gives a breakdown of the various build processes and options for building Operator from source.

<!-- TOC depthFrom:2 -->

- [Developer Quick Start](#developer-quick-start)
- [Build Pre-Requisites](#build-pre-requisites)
- [Build and deploy Operator from source](#build-and-deploy-from-source)
- [Build details](#build-details)
    - [Make targets](#make-targets) 

<!-- /TOC -->

## Developer Quick Start

To build Operator from a source you need a Kubernetes cluster. If you do not have an existing
Kubernetes cluster, you can install [kind](https://kind.sigs.k8s.io/) to have access
to a cluster on your local machine.

## Build Pre-Requisites

### Command line tools

To build this project you must first install several command line utilities and a Kubernetes cluster.

- [`make`](https://www.gnu.org/software/make/) - Make build system
- [`docker`](https://docs.docker.com/install/) - Docker command line client
- [`kind`](https://docs.docker.com/install/) (Optional) - Tool for running local Kubernetes clusters using Docker container

### Kubernetes Cluster

In order to run the integration tests and test any changes made to the operator you will need a functioning Kubernetes cluster. This can be a remote cluster, or a local development cluster.

#### Kind

The default kind setup should allow the integration tests to be run without additional configuration changes.

## Build and deploy from source

To build Spinnaker Operator from a source the operator code needs to be compiled into a container image and deploy it
in a Kubernetes cluster. The easiest way to make your custom Operator build
accessible, is to publish docker image on [Docker Hub](https://hub.docker.com/) or your private docker registry. The instructions below use Docker Hub.

1. If you don't have one already, create an account on [Docker Hub](https://hub.docker.com/). Then log your local
   Docker client into Docker Hub using:

        docker login

2. Make sure that the `REGISTRY_ORG` and `REGISTRY` environment variables are set to the same value as your
   username on the Docker Registry, and the Docker Registry you are using.

        export REGISTRY_ORG=docker_hub_username
        export REGISTRY=docker_registry_name  #defaults to docker.io if unset

3. Now build the Docker image and push them to your repository on Docker Hub:

        make docker-build
        make docker-package
        make docker-push-dev
   \
   When the Docker image is packaged, it will be tagged in the following
   format: `${REGISTRY}/${REGISTRY_ORG}/spinnaker-operator:dev` in your local repository.


4. To use the new built image, update
   the `deploy/operator/cluster/deployment.yaml`file replacing the image reference (in `image`
   property) of `spinnaker-operator` container, with the one with the same name but with the repository changed, You also can update the halyard container image.


5. The installation files assume you're installing into the namespace `spinnaker-operator`. If you want to use a different one,
   you'll need to replace it in the installation files.

       deploy/operator/cluster/role_binding.yaml

6. Then deploy the Cluster Operator by running the following (replace `spinnaker-operator` with your namespace if
   necessary):

        kubectl -n spinnaker-operator apply -f deploy/operator/cluster

7. Finally, you can deploy the SpinnakerService resource running:

        kubectl -n spinnaker apply -f deploy/spinnaker/basic/spinnakerservice.yml       

## Build details

### Make targets

Spinnaker Operator includes a `Makefile` with various Make targets to build the project.

Commonly used Make targets:

- `docker_build` for [building Docker image](#building-docker-image)
- `docker-package` for [package final Docker image](#package-final-docker-image)
- `docker-push-dev` for [pushing images to a Docker registry](#tagging-and-pushing-docker-image)

### Building Docker image

The `docker_build` target will build the base Docker image. 

### Package final Docker image

Target `docker-package` will take base image created from `docker_build` and build the final image.

### Tagging and pushing Docker image

Once image is built target `docker-push-dev` will push it to the defined docker registry.