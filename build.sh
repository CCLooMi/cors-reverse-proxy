#!/bin/bash

echo ""

echo "copy config"
if [ -f "./conf/config.yaml" ]; then
    if [ ! -d "./dist/conf" ]; then
        mkdir -p "./dist/conf"
    fi
    cp -r "./conf/config.yaml" "./dist/conf/"
else
    echo "config.yaml does not exist. Skipping copy operation."
fi

echo "Building"
go build -o "./dist/cors-revers-proxy" "./main"
echo "Done"