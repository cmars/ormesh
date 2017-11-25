#!/bin/bash

set -eu

VERSION=$1
if [ -z "$VERSION" ]; then
	echo "Usage: $0 <version>"
	exit 1
fi

sed -i 's/^version: .*/version: \''${VERSION}'\'' snap/snapcraft.yaml
git reset
git add snap/snapcraft.yaml
git commit -m "Release snapcraft ${VERSION}"

git tag -s -a "v${VERSION}" -m "Release ${VERSION}"
git push origin "v${VERSION}"

goreleaser
