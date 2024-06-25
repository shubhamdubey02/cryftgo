#!/usr/bin/env bash

set -euo pipefail

AVALANCHE_ROOT=$PKG_ROOT/cryftgo-$TAG

mkdir -p "$AVALANCHE_ROOT"

OK=$(cp ./build/cryftgo "$AVALANCHE_ROOT")
if [[ $OK -ne 0 ]]; then
  exit "$OK";
fi


echo "Build tgz package..."
cd "$PKG_ROOT"
echo "Tag: $TAG"
tar -czvf "cryftgo-linux-$ARCH-$TAG.tar.gz" "cryftgo-$TAG"
aws s3 cp "cryftgo-linux-$ARCH-$TAG.tar.gz" "s3://$BUCKET/linux/binaries/ubuntu/$RELEASE/$ARCH/"
