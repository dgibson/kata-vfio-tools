BUILD=$(CURDIR)/build
KATASRC = ~/src/kata-containers

KATAPREFIX = $(BUILD)/prefix
KATACONFIG = $(BUILD)/configuration.toml

CRIO_CONF = /etc/crio/crio.conf
INITRD = $(KATAPREFIX)/var/cache/kata-containers/kata-containers-initrd.img
OSBUILDER_SCRIPT = $(BUILD)/vfio-kata-osbuilder.sh
AGENT_TREE = $(BUILD)/agent
AGENT_BIN = $(AGENT_TREE)/usr/bin/kata-agent
OSBUILDER = $(KATASRC)/osbuilder
DRACUTDIR = $(OSBUILDER)/dracut/dracut.conf.d

QEMU := /usr/libexec/qemu-kvm
ifneq ($(wildcard $(QEMU)),$(QEMU))
QEMU := /usr/bin/qemu-system-x86_64
endif

GO = go
KATA_UPSTREAM = https://github.com/kata-containers
VFIO_REPO = https://github.com/dgibson
VFIO_REF = vfio

export GOPATH = $(BUILD)/go
DRACUTFILES = 15-dracut-fedora.conf 99-vfio.conf
OSBUILDER_DRACUTFILES = $(DRACUTFILES:%=$(DRACUTDIR)/%)

UPSTREAM_SOURCES = $(KATASRC)/proxy $(KATASRC)/shim \
	$(KATASRC)/osbuilder
VFIO_SOURCES = $(KATASRC)/agent $(KATASRC)/runtime

all: runtime #$(INITRD)

runtime: runtime-install $(KATACONFIG)

runtime-install: runtime-build
	make -C $(KATASRC)/src/runtime install PREFIX=$(KATAPREFIX) SYSCONFIG=$(KATACONFIG)

runtime-build:
	make -C $(KATASRC)/src/runtime PREFIX=$(KATAPREFIX) SYSCONFIG=$(KATACONFIG)

$(KATACONFIG): configuration.toml.template Makefile
	mkdir -p $(dir $@)
	sed 's!%BUILD%!$(BUILD)!;s!%QEMU%!$(QEMU)!' < $< > $@

agent: $(KATASRC)/agent
	make -C $<

initrd: $(INITRD)

$(INITRD): $(OSBUILDER_SCRIPT) $(AGENT_BIN) $(OSBUILDER_DRACUTFILES)
	$<

$(DRACUTDIR)/%: dracut/% $(OSBUILDER)
	cp $< $@

$(AGENT_BIN): agent
	make -C $(KATASRC)/agent DESTDIR=$(AGENT_TREE) install

$(OSBUILDER_SCRIPT): vfio-kata-osbuilder.sh.template
	sed 's!%KATAPREFIX%!$(KATAPREFIX)!;s!%AGENT_TREE%!$(AGENT_TREE)!;s!%OSBUILDER%!$(OSBUILDER)!' < $< > $@
	chmod +x $@

$(UPSTREAM_SOURCES): %:
	mkdir -p $(KATASRC)
	cd $(KATASRC) && git clone $(KATA_UPSTREAM)/$(notdir $*)

$(VFIO_SOURCES): %:
	mkdir -p $(KATASRC)
	cd $(KATASRC) && git clone -b $(VFIO_REF) $(VFIO_REPO)/kata-$(notdir $*) $(notdir $*)

clean:
	chmod -R u+w $(BUILD) || true
	rm -rf $(BUILD)
	rm -f *~

crio-conf-kata-vfio: kata-vfio-crio.conf.template
	sed 's!%KATASHIMV2%!$(KATAPREFIX)/bin/containerd-shim-kata-v2!' < $< >> $(CRIO_CONF)
