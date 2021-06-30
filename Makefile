BUILD=$(CURDIR)/build
KATASRC = ~/src/kata-containers

KATAPREFIX = $(BUILD)/prefix
KATACONFIG = $(BUILD)/configuration.toml

KERNEL = $(BUILD)/kata-vmlinuz
INITRD = $(BUILD)/kata-initrd.img

CRIO_CONF = /etc/crio/crio.conf.d/kata-vfio.conf

AGENT_BIN = $(KATASRC)/src/agent/target/x86_64-unknown-linux-gnu/release/kata-agent
MBUTO = $(CURDIR)/mbuto

OSBUILDER_SCRIPT = $(BUILD)/vfio-kata-osbuilder.sh
AGENT_TREE = $(BUILD)/agent
OSBUILDER = $(KATASRC)/osbuilder
DRACUTDIR = $(OSBUILDER)/dracut/dracut.conf.d

QEMU := /usr/libexec/qemu-kvm
ifneq ($(wildcard $(QEMU)),$(QEMU))
QEMU := /usr/bin/qemu-system-x86_64
endif

VIRTIOFSD := /usr/libexec/virtiofsd

export GOPATH = $(BUILD)/go

all: runtime $(INITRD)

runtime: runtime-install $(KATACONFIG)

runtime-install: runtime-build
	make -C $(KATASRC)/src/runtime install PREFIX=$(KATAPREFIX) SYSCONFIG=$(KATACONFIG)

runtime-build:
	make -C $(KATASRC)/src/runtime PREFIX=$(KATAPREFIX) SYSCONFIG=$(KATACONFIG)

$(KATACONFIG): configuration.toml.template Makefile
	mkdir -p $(dir $@)
	sed 's!%BUILD%!$(BUILD)!;s!%QEMU%!$(QEMU)!;s!%VIRTIOFSD%!$(VIRTIOFSD)!' < $< > $@

agent: $(AGENT_BIN)

$(AGENT_BIN): FORCE
	make -C $(KATASRC)/src/agent LIBC=gnu

initrd: $(INITRD)

$(INITRD): $(MBUTO) $(AGENT_BIN)
	$< -c gzip -f $@

clean:
	chmod -R u+w $(BUILD) || true
	rm -rf $(BUILD)
	rm -f *~

crio-conf-kata-vfio: kata-vfio-crio.conf.template
	sed 's!%KATASHIMV2%!$(KATAPREFIX)/bin/containerd-shim-kata-v2!' < $< > $(CRIO_CONF)

.PHONY: FORCE

FORCE:
