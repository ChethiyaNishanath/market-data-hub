#!/bin/bash

set -e

APP_NAME="market-data-hub-app"
ECR_REPO_NAME="chethiya-training-aws-ecr-repository"

AWS_REGION="${AWS_REGION:?AWS_REGION not set}"
AWS_PROFILE="aws_training"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:?AWS_ACCOUNT_ID not set}"

VERSION=$(git describe --tags --abbrev=0)
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
VERSION=$(git describe --tags --always)
ECR_IMAGE_URI="${ECR_REGISTRY}/${ECR_REPO_NAME}:${VERSION}"

echo "Logging into Amazon ECR"
aws ecr get-login-password --region "$AWS_REGION" --profile "$AWS_PROFILE" | docker login --username AWS --password-stdin "$ECR_REGISTRY"

echo "Building Docker Image"
docker build --build-arg VERSION="$VERSION" --build-arg COMMIT="$COMMIT" --build-arg BUILD_DATE="$BUILD_DATE" -t "$APP_NAME:${VERSION}" .

echo "Tagging Docker Image"
docker tag "$APP_NAME:${VERSION}" "$ECR_IMAGE_URI"
docker tag "$APP_NAME:${VERSION}" "${ECR_REGISTRY}/${ECR_REPO_NAME}:latest"

echo "Pushing Image to ECR"
if ! docker push "$ECR_IMAGE_URI"; then
  echo "Push failed for versioned tag, attempting cleanup"
  docker rmi "$ECR_IMAGE_URI" || true
  exit 1
fi

if ! docker push "${ECR_REGISTRY}/${ECR_REPO_NAME}:latest"; then
  echo "Push failed for latest tag, attempting cleanup"
  docker rmi "${ECR_REGISTRY}/${ECR_REPO_NAME}:latest" || true
  exit 1
fi

echo "Successfully pushed to ECR ImageURI: $ECR_IMAGE_URI"