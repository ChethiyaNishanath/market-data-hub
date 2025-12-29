#!/bin/bash

VERSION="${1:-v1.0.0}"

echo "Checking out master branch"
git checkout master

echo "Creating tag: $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"

echo "Pushing tag to remote"
git push origin "$VERSION"

echo "Running ECR push script"

chmod +x ./ecr-push.sh

./ecr-push.sh