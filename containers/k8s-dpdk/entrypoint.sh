#!/bin/bash

die() {
    echo FAILED "$@" >&2
    exit 1
}

echo "k8s VFIO dpdk test container"

ls -l /dev/vfio
env | grep '^PCIDEVICE'

(
    set -e

    if [ ! -c /dev/vfio/vfio ]; then
	die "Container doesn't see VFIO control device"
    fi

    CMD="dpdk-testpmd -l0,1 --log-level=lib.eal:8"

    DEV="$PCIDEVICE_OPENSHIFT_IO_INTELNICS"
    echo "Using device: $DEV"
    echo "Forward mode is: $FORWARD_MODE"
    echo "Peer is: $PEER_MAC"

    lspci -D -v -s $DEV

    CMD="$CMD -a $DEV"

    CMD="$CMD -- --stats-period=2 --forward-mode=$FORWARD_MODE --eth-peer=0,$PEER_MAC -a"

    echo "About to run: $CMD"

    $CMD
)

exec sleep infinity