#!/bin/sh

# force go modules
export GOPATH=""

# disable cgo
export CGO_ENABLED=0

set -e
set -x

# linux
GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/drone-findfiles
GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/drone-findfiles
GOOS=linux GOARCH=arm   go build -o release/linux/arm/drone-findfiles

# windows
GOOS=windows go build -o release/windows/amd64/drone-findfiles.exe
