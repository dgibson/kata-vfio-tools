BASEIMG = ~/mirror/bionic-server-cloudimg-amd64.img
TZ = Australia/Sydney
KEYFILE = ~/.ssh/id_ed25519.pub
export VMNAME = kata2-test
export MEMORY = 4096
export SSH_PORT = 11023
KTAG = 4.15.0-132-generic

export KERNEL = vmlinuz-$(KTAG)
export INITRAMFS = initrd.img-$(KTAG)
CLOUD_ISO = cloud-init.iso

GENISOIMAGE = genisoimage
QEMU_IMG = qemu-img
export QEMU = qemu-kvm
SSH = ssh
SSH_KEYGEN = ssh-keygen
VIRT_COPY_OUT = virt-copy-out

VMSCRIPT = ./vmscript.sh
QEMU_CMD = $(QEMU) -name $(VMNAME) \
		-machine q35,kernel_irqchip=split \
		-m $(MEMORY) \
		-cpu host \
		-smp 4 \
		-kernel $(KERNEL) \
		-initrd $(INITRAMFS) \
		-append "console=ttyS0,115200n8 systemd.unified_cgroup_hierarchy=0 root=/dev/sda1" \
		-netdev user,id=net0,hostfwd=tcp:127.0.0.1:$(SSH_PORT)-:22 \
		-device virtio-net-pci,netdev=net0 \
		-vga none \
		-nographic \
		-hda

REMOVE_HOST_KEY = $(SSH_KEYGEN) -R '[localhost]:$(SSH_PORT)' || true

KEY = $(shell cat $(KEYFILE))

all: ready.qcow2

run: ready.qcow2 $(KERNEL) $(INITRAMFS)
	$(QEMU_CMD) $<


distro.qcow2: $(BASEIMG) $(CLOUD_ISO) $(KERNEL) $(INITRAMFS)
	$(REMOVE_HOST_KEY)
	$(QEMU_IMG) create -f qcow2 -F qcow2 -b $< $@ 20G
	$(QEMU_CMD) $@ -cdrom $(CLOUD_ISO)

code.qcow2: distro.qcow2 $(KERNEL) $(INITRAMFS) $(VMSCRIPT) code.sh
	$(QEMU_IMG) create -f qcow2 -F qcow2 -b $< $@
	$(VMSCRIPT) $@ code.sh

ci-setup.qcow2: code.qcow2 $(KERNEL) $(INITRAMFS) $(VMSCRIPT) ci-setup.sh
	$(QEMU_IMG) create -f qcow2 -F qcow2 -b $< $@
	$(VMSCRIPT) $@ ci-setup.sh

ready.qcow2: ci-setup.qcow2 $(KERNEL) $(INITRAMFS)
	$(QEMU_IMG) create -f qcow2 -F qcow2 -b $< $@

$(KERNEL): $(BASEIMG)
	$(VIRT_COPY_OUT) -a $< /boot/$(KERNEL) .

$(INITRAMFS): $(BASEIMG)
	$(VIRT_COPY_OUT) -a $< /boot/$(INITRAMFS) .

user-data: user-data.in
	sed 's!@TZ@!$(TZ)!;s!@KEY@!$(KEY)!' < $< > $@

$(CLOUD_ISO): meta-data user-data
	$(GENISOIMAGE) -output $@ -input-charset utf8 -volid cidata -joliet -rock $^


ssh:
	$(SSH) -o StrictHostKeyChecking=accept-new kata@localhost -p $(SSH_PORT) 


remove-host-key:
	$(SSH_KEYGEN) -R '[localhost]:$(SSH_PORT)' || true

clean:
	$(REMOVE_HOST_KEY)
	rm -f *~
	rm -f *.qcow2
	rm -f user-data $(CLOUD_ISO)
	rm -f $(KERNEL) $(INITRAMFS)
