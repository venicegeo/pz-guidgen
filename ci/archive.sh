#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

#----------------------------------------------------------------------

export GOPATH=$root/gogo
mkdir -p "$GOPATH"

# glide expects this to already exist
mkdir "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

export GO15VENDOREXPERIMENT="1"

curl https://glide.sh/get | sh

# get ourself, and go there
go get github.com/venicegeo/pz-uuidgen
cd $GOPATH/src/github.com/venicegeo/pz-uuidgen

glide install
glide update

#----------------------------------------------------------------------

src=$GOPATH/bin/pz-uuidgen

# gather some data about the repo
source $root/ci/vars.sh

# stage the artifact for a mvn deploy
mv $src $root/$APP.$EXT
