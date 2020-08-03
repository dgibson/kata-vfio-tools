#!/bin/bash

die() {
    echo FAILED "$@" >&2
    exit 1
}

echo "VFIO dpdk test container starting"

ls -l /dev/vfio
lspci -D

(
    set -e

    if [ ! -c /dev/vfio/vfio ]; then
	die "Container doesn't see VFIO control device"
    fi

    VFIOGROUPS="$(cd /dev/vfio && echo [0-9]*)"
    NGROUPS=$(ls /dev/vfio | grep -v vfio | wc -l)
    echo "Container sees $NGROUPS IOMMU groups [$VFIOGROUPS]"

    # Mount Hugepages. FIXME: This should be done by kata-agent
    #mkdir -p /dev/hugepages; mount -t hugetlbfs nodev /dev/hugepages

    CMD="testpmd -l0,1 --log-level=lib.eal:8"

    group=$(echo $VFIOGROUPS | cut -f1 -d' ')
    echo "Using group [$group]"
    devices=$(cd /sys/kernel/iommu_groups/$group/devices && echo *)

    echo "Using group $group ($devices)"

    for dev in "$devices"; do
	    # Ignore bridges
	    if [ -d /sys/bus/pci/devices/$dev/pci_bus ]; then
		continue
	    fi
	CMD="$CMD -w$dev"
    done

    CMD="$CMD -- --stats-period 2 --forward-mode=txonly -a"

    # RUN Debug commands
    for dev in "$devices"; do
	lspci -vv -s $dev
    done

    echo "About to run: $CMD"

    $CMD
)

exec /bin/bash
