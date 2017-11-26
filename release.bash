#!/bin/bash

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
	echo "Usage: $0 <version>"
	exit 1
fi

set -u

sed -i "s/^version: .*/version: '${VERSION}'/" snap/snapcraft.yaml
sed -i "s/^RELEASE_VERSION=.*/RELEASE_VERSION=${VERSION}/" install.bash
git reset
git add snap/snapcraft.yaml install.bash
git commit -m "Release install.bash & snapcraft ${VERSION}" || true

git tag -s -a "v${VERSION}" -m "Release ${VERSION}" && git push origin "v${VERSION}" || true

goreleaser
