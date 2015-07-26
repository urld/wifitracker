#!/bin/bash

# check if docker is available:
docker version >/dev/null 2>&1
DOCKER_STATUS=$?
KERNEL=$(uname -s)
if [[ $DOCKER_STATUS -ne 0 ]] || [[ $KERNEL != Linux ]]; then
        set -e

        # setup build environment:
        REPO_PATH="go-wifi-tracker"

        export GOPATH=${PWD}/gopath

        rm -f $GOPATH/src/${REPO_PATH}
        mkdir -p $GOPATH/src
        ln -s ${PWD} $GOPATH/src/${REPO_PATH}

        eval $(go env)

        # get dependencies:
        go get $(go list -f "{{range .Imports}}{{ .  }} {{end}}")
        # build static binary:
        CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-s" -o bin/wifitracker ${REPO_PATH}
else
        echo "docker available. building in container..."
        ARCH=$(uname -m)
        if [[ $ARCH == arm*  ]]; then
                docker run -it -v $(pwd):$(pwd) -w $(pwd) hypriot/rpi-golang:1.4.2 /bin/bash -c $(pwd)/build.sh
        else
                docker run -it -v $(pwd):$(pwd) -w $(pwd) golang:1.4.2 /bin/bash -c $(pwd)/build.sh
        fi
fi

