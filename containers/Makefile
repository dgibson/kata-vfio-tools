CONTAINERS = smoke dpdk true
IMGPREFIX = quay.io/dwg

.PHONY: $(CONTAINERS)

all: $(CONTAINERS)

$(CONTAINERS): %: %/Dockerfile
	podman build -t vfio-$* $*

push: $(CONTAINERS:%=push-%)

tag: $(CONTAINERS:%=tag-%)

push-%: tag-%
	podman push $(IMGPREFIX)/vfio-$*:latest

tag-%: %
	podman tag localhost/vfio-$* $(IMGPREFIX)/vfio-$*:latest

clean:
	rm -f *~

cleaner: clean
	podman rmi $(CONTAINERS:%=vfio-%)
