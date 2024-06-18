#!/bin/bash

if [ ! -d "build" ]; then
  mkdir build
fi

if [ "$(ls -A build)" ]; then
  rm -rf build/*
fi

go build -o build/cloud-lb ../cmd/main.go

exit 0
