#!/bin/bash
set -e

NAME=${1:-aquafarm}
PLATFORM=${2:-linux/amd64}
IMAGE="benzhi/${NAME}:latest"

echo "Building ${IMAGE} for ${PLATFORM}..."
docker build --platform ${PLATFORM} -t ${IMAGE} -f benzhi.Dockerfile .
echo "Done: ${IMAGE} (${PLATFORM})"
