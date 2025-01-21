#!/bin/bash
mkdir -p bin

# we compile specificaly for intel-compatibility, for older machines.
# this means Apple Silicon machines will need to have Rosetta2 isntalled:
#
# Run this to see if you have an M1+ CPU to see if Rosetta is installed:
#  arch -x86_64 zsh
#
# Run this to install it:
#  softwareupdate --install-rosetta
GOOS=darwin GOARCH=arm64 go build -trimpath -o bin/cw .