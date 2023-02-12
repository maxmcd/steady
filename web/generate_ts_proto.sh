#!/usr/bin/env bash

set -ex

ls -lah ./node_modules/.bin

cp ../steady/steadyrpc/steady.proto .
./node_modules/.bin/twirpscript

rm steady.proto

