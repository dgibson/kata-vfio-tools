#! /bin/sh

set -x

DISK="$1"
SCRIPT="$2"

${QEMU} -name ${VMNAME} \
	-machine q35,kernel_irqchip=split \
	-m ${MEMORY} \
	-cpu host \
	-smp 2 \
	-kernel ${KERNEL} \
	-initrd ${INITRAMFS} \
	-append "console=ttyS0,115200n8 systemd.unified_cgroup_hierarchy=0 root=/dev/sda1" \
	-netdev user,id=net0,hostfwd=tcp:127.0.0.1:${SSH_PORT}-:22 \
	-device virtio-net-pci,netdev=net0 \
	-vga none \
	-nographic \
	-hda ${DISK} &
PID="$!"

die() {
    kill $PID
    exit 1
}

while ! ssh -o StrictHostKeyChecking=accept-new kata@localhost -p ${SSH_PORT} true; do
    sleep 1
done

scp -P ${SSH_PORT} ${SCRIPT} kata@localhost: || die
ssh kata@localhost -p ${SSH_PORT} sh ${SCRIPT} || die
ssh kata@localhost -p ${SSH_PORT} sudo poweroff

wait $PID
