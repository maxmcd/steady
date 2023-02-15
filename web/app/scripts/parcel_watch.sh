#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname ${BASH_SOURCE[0]})"

cd ..


# TODO: this doesn't seem to be realiably rebuilding _asset.go.html, breaking go rebuilds
npx parcel watch --public-url=/assets ./_assets.go.html


