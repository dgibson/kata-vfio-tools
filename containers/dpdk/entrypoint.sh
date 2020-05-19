#!/bin/bash

# Mount Hugepages. FIXME: This should be done by kata-agent
mkdir -p /dev/hugepages; mount -t hugetlbfs nodev /dev/hugepages

CMD="testmpd -l0,1 --log-level=lib.eal:8"


#FIXME: We should be able to use an ENV var here but the PCI addreses won't match
devices=("0000:01:01.0")

for dev in "${devices[@]}"
do
    CMD="$CMD -w$dev"
done

CMD="$CMD -- --stats-period 2 --forward-mode=txonly -a"

# RUN Debug commands
for dev in "${devices[@]}"
do
    lspci -vv -s $dev
done
dmesg

$CMD

