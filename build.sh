#!/bin/bash
mkdir -p bin
GOOS=darwin GOARCH=amd64 go build -o bin/cw .