#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

#----------------------------------------------------------------------

export GOPATH=$root/gogo
mkdir -p "$GOPATH"

# glide expects these to already exist
mkdir "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

# install metalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install

# build ourself, and go there
go get github.com/venicegeo/pz-uuidgen
cd $GOPATH/src/github.com/venicegeo/pz-uuidgen

#----------------------------------------------------------------------

go test -v -coverprofile=uuidgen.cov github.com/venicegeo/pz-uuidgen/uuidgen

sh ci/metalinter.sh | tee lint.txt
wc -l lint.txt

###
