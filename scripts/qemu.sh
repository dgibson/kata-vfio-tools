#! /bin/sh

QEMU=qemu-system-x86_64

$QEMU "$@" -serial file:/tmp/qemu.log 2>/tmp/qemu_err.log

