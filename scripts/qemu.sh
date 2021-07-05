#! /bin/sh

/usr/bin/qemu-system-x86_64 "$@" -serial file:/tmp/qemu.log 2>/tmp/qemu_err.log

