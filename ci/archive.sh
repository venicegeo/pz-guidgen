#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null
export GOPATH=$root/gogo

#----------------------------------------------------------------------

sh $root/ci/do_build.sh

#----------------------------------------------------------------------

app=$GOPATH/bin/pz-uuidgen

# gather some data about the repo
source $root/ci/vars.sh

# stage the artifact(s) for a mvn deploy
mv $app $root/$APP.$EXT

cd $root
tar cvzf $APP.tgz \
    $APP.$EXT \
    uuidgen.cov \
    lint.txt \
    glide.lock \
    glide.yaml
tar tzf $APP.tgz
