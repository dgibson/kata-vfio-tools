#!/bin/bash

# TODO: Write some some tests

dmesg
lspci -vv

ls -l /sys/kernel/iommu_groups/*/devices
ls -l /dev/vfio/

sleep infinity
