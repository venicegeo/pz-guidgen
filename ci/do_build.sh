#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null
export GOPATH=$root/gogo

#----------------------------------------------------------------------

mkdir -p "$GOPATH" "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

go version

# install metalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install

# build ourself, and go there
go get github.com/venicegeo/pz-uuidgen
cd $GOPATH/src/github.com/venicegeo/pz-uuidgen

# run unit tests w/ coverage collection
go test -v -coverprofile=$root/uuidgen.cov github.com/venicegeo/pz-uuidgen/uuidgen

# run lint
sh ci/metalinter.sh | tee $root/lint.txt
wc -l $root/lint.txt
