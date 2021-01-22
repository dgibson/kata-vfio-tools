#! /bin/sh

set -ex

cat >> ~/.profile <<EOF
export GOPATH=$HOME/go
#export PATH=$PATH:$GOPATH/bin:/usr/local/go/bin
EOF

. ~/.profile

wget https://golang.org/dl/go1.14.4.linux-amd64.tar.gz
sudo tar xzvpf go1.14.4.linux-amd64.tar.gz -C /usr/local
sudo cp /usr/local/go/bin/go* /usr/local/bin/

KATASRC=$GOPATH/src/github.com/kata-containers

go get github.com/kata-containers/kata-containers || true
(
    cd $KATASRC/kata-containers
    git remote add dwg-github https://github.com/dgibson/kata-containers
    git fetch dwg-github
    git checkout -b test dwg-github/pcipath
)
go get github.com/kata-containers/tests || true
(
    cd $KATASRC/tests
    git checkout -b test origin/2.0-dev
    #CRIO="no" CRI_CONTAINERD="yes" OPENSHIFT="no" .ci/setup.sh
)
