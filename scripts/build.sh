#!/bin/sh

pushd `dirname $0` > /dev/null
base=$(pwd -P)
popd > /dev/null

export GOPATH=$base/gogo
mkdir -p $GOPATH

go get github.com/venicegeo/pz-uuidgen

go test -v github.com/venicegeo/pz-uuidgen

go install github.com/venicegeo/pz-uuidgen
