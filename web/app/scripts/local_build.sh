#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname ${BASH_SOURCE[0]})"

cd ..

bun install

cp ../../steady/steadyrpc/steady.proto .
npx twirpscript
rm steady.proto
mv steady.pb.ts ./src

npx parcel build --public-url=/assets ./_assets.go.html

