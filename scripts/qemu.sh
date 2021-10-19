#! /bin/sh

QEMU=/home/dwg/src/qemu/build/kata/x86_64-softmmu/qemu-system-x86_64

$QEMU "$@" -serial file:/tmp/qemu.log 2>/tmp/qemu_err.log

