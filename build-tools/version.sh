#!/bin/bash

#-----------------------------------------------------------------------------
# Calculates the next version based on git tags and current branch.
#-----------------------------------------------------------------------------

VERSION_TYPE=$1
[[ ! $VERSION_TYPE =~ snapshot|rc|release ]] && echo "Usage: $(basename "$0") snapshot|rc|release" && exit 1

BRANCH=$(git rev-parse --abbrev-ref HEAD)

case $BRANCH in
  ^release-*)
    tmp=$(echo "$BRANCH" | cut -d'-' -f 2)
    read -r br_major br_minor br_patch <<< "${tmp//./ }"
    v=$(git tag --sort=-v:refname | grep "v$br_major.$br_minor" | head -1 | sed 's|v||g') ;;
  *)
    v=$(git tag --sort=-v:refname | head -1 | sed 's|v||g') ;;
esac

read -r major minor patch rc <<< "${v//./ }" && patch=$(echo "$patch" | sed 's|[^0-9]*||g')

case $VERSION_TYPE in
    snapshot)
        if [ "x$(git status --porcelain)" != "x" ] ; then u=".uncommitted"; fi
        br=$(echo ".$BRANCH" | sed 's|[-/_]|.|g')
        commit=$(git rev-parse --short HEAD)
        if [[ "x$rc" = "x" ]]
        then
            if [[ $BRANCH =~ ^release-* ]] ; then ((patch++)) ; else ((minor++)) && patch=0 ; fi
        fi
        NEW_VERSION="${major}.${minor}.${patch}-snapshot$u$br.$commit"
        ;;
    rc)
        if [[ "x$rc" = "x" ]]
        then
            rc=1
            if [[ $BRANCH =~ ^release-* ]] ; then ((patch++)) ; else ((minor++)) && patch=0 ; fi
        else
            ((rc++))
        fi
        NEW_VERSION="${major}.${minor}.${patch}-rc.${rc}"
        ;;
    release)
        if [[ "x$rc" = "x" ]]
        then
            if [[ $BRANCH =~ ^release-* ]] ; then ((patch++)) ; else ((minor++)) && patch=0 ; fi
        fi
        NEW_VERSION="${major}.${minor}.${patch}"
        ;;
esac

echo "$NEW_VERSION"
