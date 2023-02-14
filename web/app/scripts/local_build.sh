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

export ROOT="./node_modules/monaco-editor/esm/vs"
npx parcel build "$ROOT/language/json/json.worker.js" --no-source-maps
npx parcel build "$ROOT/language/css/css.worker.js" --no-source-maps
npx parcel build "$ROOT/language/html/html.worker.js" --no-source-maps
npx parcel build "$ROOT/language/typescript/ts.worker.js" --no-source-maps
npx parcel build "$ROOT/editor/editor.worker.js" --no-source-maps
