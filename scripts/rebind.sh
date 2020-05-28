#! /bin/sh

set -e

die () {
    echo "$@" >&2
    exit 1
}

PCI="$1"

if [ -z "$PCI" ]; then
    die "Usage: $0 <PCI address>"
fi

if [ ! -L /sys/bus/pci/devices/$PCI ] ; then
    PCI="0000:$PCI"
fi

DEVDIR="/sys/bus/pci/devices/$PCI"

if [ ! -L $DEVDIR ]; then
    die "No such PCI device $PCI"
fi

VENDOR=$(cat "$DEVDIR/vendor")
PRODUCT=$(cat "$DEVDIR/device")
GROUP=$(readlink "$DEVDIR/iommu_group" | sed 's!.*iommu_groups/!!')

echo "Device $PCI ($VENDOR:$PRODUCT) is in IOMMU group $GROUP"

GROUPDIR="/sys/kernel/iommu_groups/$GROUP"

DEVS=$(ls $GROUPDIR/devices)
NDEV=$(ls $GROUPDIR/devices | wc -l)

echo "Group $GROUP contains $NDEV devices [$DEVS]"

if [ "$NDEV" != "1" ]; then
    die "Can't rebind multi-device groups yet"
fi

VFIODIR=/sys/bus/pci/drivers/vfio-pci

if [ ! -d "$VFIODIR" ]; then
    echo -n "Loading VFIO driver..."
    modprobe vfio-pci
    echo "done"
fi

if [ -e "$DEVDIR/driver" ]; then
   DRIVER=$(basename $(readlink $DEVDIR/driver))

   echo -n "Unbinding $PCI from $DRIVER..."

   echo "$PCI" > "$DEVDIR/driver/unbind"

   echo "done"
fi

echo -n "Adding IDs to VFIO driver..."
echo $VENDOR $PRODUCT > $VFIODIR/new_id
echo done

if [ $(basename $(readlink $DEVDIR/driver)) != vfio-pci ]; then
   die "Didn't bind to VFIO driver :("
fi
