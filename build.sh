#!/bin/bash

echo "Generating binaries"
mkdir -p dist
cd dist
rm *.exe
gox -os="windows linux macos" ..
cd ..

echo "Creating docker image"

docker build --tag cachet-monitor:${1:-latest} .