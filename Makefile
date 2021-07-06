BUILD=$(CURDIR)/build
KATASRC = $(HOME)/src/kata-containers

KATAPREFIX = $(BUILD)/prefix
KATACONFIG = $(BUILD)/configuration.toml

KERNEL = $(BUILD)/kata-vmlinuz
INITRD = $(BUILD)/kata-initrd.img

CRIO_CONF = /etc/crio/crio.conf.d/kata-vfio.conf


OSBUILDER_SCRIPT = $(BUILD)/vfio-kata-osbuilder.sh
AGENT_TREE = $(BUILD)/agent
AGENT_BIN = $(AGENT_TREE)/usr/bin/kata-agent
OSBUILDER = $(KATASRC)/tools/osbuilder
DRACUTDIR = $(OSBUILDER)/dracut/dracut.conf.d
DRACUTFILES = 15-dracut-fedora.conf 99-vfio.conf
OSBUILDER_DRACUTFILES = $(DRACUTFILES:%=$(DRACUTDIR)/%)

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

$(AGENT_BIN): agent-build
	make -C $(KATASRC)/src/agent LIBC=gnu DESTDIR=$(AGENT_TREE) install

agent-build: FORCE
	make -C $(KATASRC)/src/agent LIBC=gnu

initrd: $(INITRD)

$(INITRD): $(OSBUILDER_SCRIPT) $(AGENT_BIN) $(OSBUILDER_DRACUTFILES)
	$<

$(DRACUTDIR)/%: dracut/% $(OSBUILDER)
	cp $< $@

$(OSBUILDER_SCRIPT): vfio-kata-osbuilder.sh.template
	sed 's!%BUILD%!$(BUILD)!;s!%AGENT_TREE%!$(AGENT_TREE)!;s!%OSBUILDER%!$(OSBUILDER)!' < $< > $@
	chmod +x $@
clean:
	chmod -R u+w $(BUILD) || true
	rm -rf $(BUILD)
	rm -f *~

crio-conf-kata-vfio: kata-vfio-crio.conf.template
	sed 's!%KATASHIMV2%!$(KATAPREFIX)/bin/containerd-shim-kata-v2!' < $< > $(CRIO_CONF)

.PHONY: FORCE

FORCE:
