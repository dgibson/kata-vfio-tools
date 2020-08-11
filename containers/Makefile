CONTAINERS = smoke dpdk true

.PHONY: $(CONTAINERS)

all: $(CONTAINERS)

$(CONTAINERS): %: %/Dockerfile
	podman build -t vfio-$* $*

clean:
	rm -f *~
