#!/bin/sh

export GOPATH=.

go get github.com/venicegeo/pz-uuidgen

go install github.com/venicegeo/pz-uuidgen
