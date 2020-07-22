#!/bin/bash

die() {
    echo FAILED "$@" >&2
    exit 1
}

echo "VFIO smoke test container starting"

ls -l /dev/vfio
lspci -D

(
    set -e

    if [ ! -c /dev/vfio/vfio ]; then
	die "Container doesn't see VFIO control device"
    fi

    VFIOGROUPS="$(ls /dev/vfio | grep -v vfio)"
    NGROUPS=$(ls /dev/vfio | grep -v vfio | wc -l)
    echo "Container sees $NGROUPS IOMMU groups [$VFIOGROUPS]"

    for group in $VFIOGROUPS; do
	echo "Testing group $group"
	DEVS=$(cd /sys/kernel/iommu_groups/$group/devices && echo *)
	for dev in $DEVS; do
	    # Ignore bridges
	    if [ -d /sys/bus/pci/devices/$dev/pci_bus ]; then
		continue
	    fi
	    echo "Testing group $group: $dev"
	    ./vfio-pci-device-open $group $dev
	    ./vfio-iommu-map-unmap $dev
	done
    done
)

exec /bin/bash
