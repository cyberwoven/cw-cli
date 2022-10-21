#!/bin/bash
mkdir -p bin
GOOS=darwin GOARCH=amd64 go build -trimpath -o bin/cw .