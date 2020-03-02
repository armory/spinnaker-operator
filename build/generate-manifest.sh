#!/usr/bin/env bash

set -e

SPINNAKER_OPERATOR_VERSION=$1
VERSION=$2
RELEASE_VERSION=$3

if [ -z "$RELEASE_VERSION" ]; then
  SPINNAKER_OPERATOR_VERSION="$SPINNAKER_OPERATOR_VERSION-$VERSION"
fi

echo "Version="$SPINNAKER_OPERATOR_VERSION > build/MANIFEST
echo "Built-By="$(whoami) >> build/MANIFEST
echo "Build-Date="$(date +'%Y-%m-%d_%H:%M:%S') >> build/MANIFEST
echo "Branch="$(git rev-parse --abbrev-ref HEAD) >> build/MANIFEST
echo "Revision="$(git describe --always) >> build/MANIFEST
echo "Build-Go-Version="$(go version) >> build/MANIFEST
