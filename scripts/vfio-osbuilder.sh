#! /bin/sh

KATASRC=$GOPATH/src/github.com/kata-containers
AGENTDIR=/usr/libexec/kata-containers/agent
HOOKPATH=/usr/share/oci/hooks
BUILDER=/usr/libexec/kata-containers/osbuilder/kata-osbuilder.sh

if [ ! -e $BUILDER ]; then
    BUILDER=/usr/libexec/kata-containers/osbuilder/fedora-kata-osbuilder.sh
fi

cp $KATASRC/agent/kata-agent $AGENTDIR/usr/bin
mkdir -p $AGENTDIR/$HOOKPATH/prestart
cp ./vfio-hook/vfio-hook $AGENTDIR/$HOOKPATH/prestart

$BUILDER
