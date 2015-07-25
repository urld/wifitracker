#!/bin/bash -e

REPO_PATH="go-wifi-tracker"

export GOPATH=${PWD}/gopath

rm -f $GOPATH/src/${REPO_PATH}
mkdir -p $GOPATH/src
ln -s ${PWD} $GOPATH/src/${REPO_PATH}

eval $(go env)

go get $(go list -f "{{range .Imports}}{{ .  }} {{end}}")
# Static compilation is useful when etcd is run in a container
CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-s" -o bin/wifitracker ${REPO_PATH}
