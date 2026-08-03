#!/usr/bin/env bash

set -e

tag=$(git tag --points-at HEAD)

echo "current tag: $tag"

if [ -z "$tag" ]; then
    defaultTag="${GITHUB_REF_NAME}-${GITHUB_SHA:0:7}-$(date +%s)"
    echo "IMAGE_TAGS=main,$defaultTag" >>$GITHUB_OUTPUT
else
    # strip "v" prefix form tag
    tag=${tag#v}

    echo "IMAGE_TAGS=$tag,latest" >>$GITHUB_OUTPUT
fi
