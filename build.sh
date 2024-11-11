#!/bin/bash

DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
REVISION=$(git rev-list --count $(git branch --show-current))
VERSION=v1.${REVISION}

docker build --build-arg BUILD_DATE="${DATE}" --build-arg IMAGE_VERSION=${VERSION} --format docker -t vm75/vpn-sandbox .

IMAGE_ID=$(docker images | grep vpn-sandbox | grep latest | awk '{ print $3}')

echo "tagging with ${IMAGE_ID} vm75/vpn-sandbox:latest vm75/vpn-sandbox:${VERSION}"

docker tag ${IMAGE_ID} vm75/vpn-sandbox:latest vm75/vpn-sandbox:${VERSION}

if [ "$1" == push ] ; then
	docker login
	if [[ $? -eq 0 ]] ; then
		docker push vm75/vpn-sandbox
	fi
fi
