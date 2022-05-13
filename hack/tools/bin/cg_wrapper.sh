#!/usr/bin/env bash

# Wrap the openapi generate command to ignore a failure that
# does not impact CRD generation. When this issue is resolved we
# may have a better approach:
# https://github.com/kubernetes-sigs/controller-tools/issues/636

function capture {
  if [[ "$?" -eq 2 ]]; then
    # This is a known failure in the make generate command
    exit 0
  fi
}
trap capture EXIT

make openapi-internal
