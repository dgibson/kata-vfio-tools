BASEIMG = ~/mirror/Fedora-Cloud-Base-33-1.2.x86_64.qcow2
TZ = Australia/Sydney
KEYFILE = ~/.ssh/id_ed25519.pub
NAME = kata-vfio-test-vm
MEMORY = 8192
SSH_PORT = 11022
SOCKNET_PORT = 11099
KTAG = 5.8.15-301.fc33.x86_64

KERNEL = vmlinuz-$(KTAG)
INITRAMFS = initramfs-$(KTAG).img
CLOUD_ISO = cloud-init.iso
DISK = disk.qcow2

GENISOIMAGE = genisoimage
QEMU_IMG = qemu-img
QEMU = qemu-kvm
SSH = ssh
SSH_KEYGEN = ssh-keygen
VIRT_COPY_OUT = virt-copy-out

QEMU_CMD = $(QEMU) -name $(NAME) \
		-machine q35,kernel_irqchip=split \
		-m $(MEMORY) \
		-cpu host \
		-smp 2 \
		-kernel $(KERNEL) \
		-initrd $(INITRAMFS) \
		-append "console=ttyS0,115200n8 systemd.unified_cgroup_hierarchy=0 intel_iommu=on root=/dev/sda1" \
		-hda $(DISK) \
		-netdev user,id=net0,hostfwd=tcp:127.0.0.1:$(SSH_PORT)-:22 \
		-device virtio-net-pci,netdev=net0 \
		-netdev socket,id=net1,listen=:$(SOCKNET_PORT) \
		-device virtio-net-pci,netdev=net1 \
		-device intel-iommu \
		-vga none \
		-nographic

KEY = $(shell cat $(KEYFILE))

all: run

user-data: user-data.in Makefile
	sed 's!@TZ@!$(TZ)!;s!@KEY@!$(KEY)!' < $< > $@

$(CLOUD_ISO): meta-data user-data
	$(GENISOIMAGE) -output $@ -input-charset utf8 -volid cidata -joliet -rock $^

disk: $(DISK)

$(DISK): $(BASEIMG) $(CLOUD_ISO) extract remove-host-key
	$(SSH_KEYGEN) -R '[localhost]:$(SSH_PORT)' || true
	$(QEMU_IMG) create -f qcow2 -F qcow2 -b $< $@ 20G
	$(QEMU_CMD) -cdrom $(CLOUD_ISO)

run: $(DISK) extract
	$(QEMU_CMD)

ssh:
	$(SSH) -o StrictHostKeyChecking=accept-new root@localhost -p $(SSH_PORT) 

remove-host-key:
	$(SSH_KEYGEN) -R '[localhost]:$(SSH_PORT)' || true

$(KERNEL): $(BASEIMG)
	$(VIRT_COPY_OUT) -a $< /boot/$(KERNEL) .

$(INITRAMFS): $(BASEIMG)
	$(VIRT_COPY_OUT) -a $< /boot/$(INITRAMFS) .

extract: $(KERNEL) $(INITRAMFS)

clean: remove-host-key
	rm -f *~
	rm -f user-data $(CLOUD_ISO) $(DISK)
	rm -f $(KERNEL) $(INITRAMFS)
