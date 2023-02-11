#!/usr/bin/env bash

set -e

cp ../steady/steadyrpc/steady.proto .
./node_modules/.bin/twirpscript

rm steady.proto

