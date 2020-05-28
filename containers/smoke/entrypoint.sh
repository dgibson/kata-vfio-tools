#!/bin/bash

die() {
    echo FAILED "$@" >&2
    exit 1
}

echo "VFIO smoke test container starting"

ls -l /dev/vfio

(
    set -e

    if [ ! -c /dev/vfio/vfio ]; then
	die "Container doesn't see VFIO control device"
    fi

    VFIOGROUPS="$(ls /dev/vfio | grep -v vfio)"
    NGROUPS=$(ls /dev/vfio | grep -v vfio | wc -l)
    echo "Container sees $NGROUPS IOMMU groups [$VFIOGROUPS]"
    false
) || exec /bin/bash
