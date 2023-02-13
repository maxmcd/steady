#!/usr/bin/env bash

set -ex

cp ../steady/steadyrpc/steady.proto .

docker build -t steady-web .
id=$(docker create steady-web)
rm -rf ./dist
docker cp $id:/opt/dist ./
docker cp $id:/opt/steady.pb.ts .
docker rm -v $id

rm steady.proto

