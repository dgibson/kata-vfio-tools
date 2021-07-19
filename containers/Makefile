CONTAINERS = smoke dpdk true
IMGPREFIX = quay.io/dwg

.PHONY: $(CONTAINERS)

all: $(CONTAINERS)

$(CONTAINERS): %: %/Dockerfile
	podman build -t vfio-$* $*

push: $(CONTAINERS:%=push-%)

push-%: %
	podman tag localhost/vfio-$* $(IMGPREFIX)/vfio-$*:latest
	podman push $(IMGPREFIX)/vfio-$*:latest

clean:
	rm -f *~

cleaner: clean
	podman rmi $(CONTAINERS:%=vfio-%)
