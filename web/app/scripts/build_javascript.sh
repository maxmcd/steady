#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname ${BASH_SOURCE[0]})"

cd ..

cp ../../steady/steadyrpc/steady.proto .

docker build -t steady-web .
rm steady.proto

id=$(docker create steady-web)
rm -rf ./dist
docker cp $id:/opt/dist ./
docker cp $id:/opt/steady.pb.ts .
docker rm -v $id



