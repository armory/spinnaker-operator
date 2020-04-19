#!/bin/bash

#------------------------------------------------------------------------------------------------------------------
# Calculates the next version based on git tags and current branch.
#
# Examples:
# * Snapshot version: 0.1.0-snapshot.[uncommitted].chore.foo.bar.test.050f9cd
# * RC version:       0.1.0-rc.9
# * Release version:  0.1.0
#
# Step logic:
# * On release branch, only patch version is stepped depending on the latest git tag matching the branch name
# * On all other branches, if latest tag is not rc, step minor and set patch=0. Otherwise, step rc number
#------------------------------------------------------------------------------------------------------------------

VERSION_TYPE=$1
BRANCH_OVERRIDE=$2
[[ ! $VERSION_TYPE =~ snapshot|rc|release ]] && echo "Usage: $(basename "$0") snapshot|rc|release [branch-name]" && exit 1

if [[ -z $BRANCH_OVERRIDE ]] ; then BRANCH=$(git rev-parse --abbrev-ref HEAD) ; else BRANCH=$BRANCH_OVERRIDE ; fi

case $BRANCH in
  release-*)
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
