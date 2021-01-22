#! /bin/sh

set -ex

. ~/.profile

KATASRC=$GOPATH/src/github.com/kata-containers

cd $KATASRC/tests
CRIO="no" CRI_CONTAINERD="yes" OPENSHIFT="no" .ci/setup.sh
