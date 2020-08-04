BUILD = $(CURDIR)/build
KATAPREFIX = $(BUILD)/prefix
KATACONFIG = $(KATAPREFIX)/etc/configuration.toml

GO = go
PODMAN_CONF = /usr/share/containers/containers.conf
KATA_UPSTREAM = https://github.com/kata-containers

export GOPATH = $(BUILD)/go
KATASRC = $(GOPATH)/src/github.com/kata-containers
RUNTIME_PKGS = runtime proxy shim

GOSOURCES = $(RUNTIME_PKGS:%=$(KATASRC)/%)

all: runtime

runtime: $(RUNTIME_PKGS:%=%-install) $(KATACONFIG)

runtime-sources: $(GOSOURCES)

$(RUNTIME_PKGS:%=%-build): %-build: $(KATASRC)/%
	make -C $< PREFIX=$(KATAPREFIX) SYSCONFIG=$(KATACONFIG)

$(RUNTIME_PKGS:%=%-install): %-install: $(KATASRC)/% %-build
	make -C $< PREFIX=$(KATAPREFIX) SYSCONFIG=$(KATACONFIG) install

$(KATACONFIG): configuration.toml.template
	mkdir -p $(dir $@)
	sed 's!%KATAPREFIX%!$(KATAPREFIX)!' < $< > $@

$(GOSOURCES): %:
	mkdir -p $(KATASRC)
	cd $(KATASRC) && git clone $(KATA_UPSTREAM)/$(notdir $*)

clean:
	rm -rf build
	rm -f *~

podman-conf-kata-vfio: podman-kata-vfio.conf.template
	sed 's!%KATARUNTIME%!$(KATAPREFIX)/bin/kata-runtime!' < $< >> $(PODMAN_CONF)
