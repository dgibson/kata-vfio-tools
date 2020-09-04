# Introduction

This document describes the steps required to use Kata guest agent for
rebinding the SRIOV device to VFIO using podman

## Assumptions

- Fedora 32+ or RHEL8.2+ distro
- These packages installed:
    `podman`
	`qemu-system-x86` (or `qemu-kvm` on RHEL)
	`busybox`
	`golang`
- GOPATH is setup
- Host system booted with `intel_iommu=on`
- Host system booted with `systemd.unified_cgroup_hierarchy=0`
- SELinux is set to permissive mode on the host
- The host device to be added to Kata container should be bound to VFIO driver

Note:
In certain cases you might have to use the unsafe_interrupts to allow passthrough of PCI devices to the VM
```
echo "options vfio_iommu_type1 allow_unsafe_interrupts=1" > /etc/modprobe.d/iommu_unsafe_interrupts.conf
```

## Build the Kata components from source

From the top-level directory of this tree, run:
```
$ make
```

This will do a number of things:
- Download and build the sources for a number of Kata components
- Install the components into `$KATAPREFIX` (by default `build/prefix`) so that it doesn't disrupt your main system
- Build some VFIO specific scripts
- Build a new Kata OS image configured for VFIO support, installing it in `$KATAPREFIX`
- Build a Kata `configuration.toml`  suitable for VFIO

The source-build components are built so that they will look for the
OS image and configuration built here.

## Configure podman to use the source-built runtime

This will make podman aware of a new runtime, called `kata-vfio` which
will use the components built locally above.  As root, run:
```
# make podman-conf-kata-vfio
```

You should only do this once.  If you move this tree, you'll need to
manually remove the stasnza from
`/usr/share/containers/containers.conf` and re-run it.

## Kata configuration.toml settings

The `configuation.toml` built from this tree should be suitable.
However, if you want to try this with your own `confugration.toml` you
will need to add these settings:

```
machine_type = "q35"
kernel_params = "systemd.unified_cgroup_hierarchy=0`
enable_iommu = true
```

In addition, if you want to run DPDK in the container, you will need
at least 2 cores assigned to the VM.  e.g.
```
default_vcpus = 2
```

## Rebinding host device

1. Set up SR-IOV virtual functions if desired
2. Pick a device to pass into your container, and find it's PCI address from `lspci`
3. If the address is `0000:02:00.0` use the `scripts/rebind.sh` to rebind it to the VFIO driver
```
# ./scripts/rebind.sh 0000:02:00.0
Device 0000:02:00.0 (0x10ec:0x522a) is in IOMMU group 9
Group 9 contains 1 devices [0000:02:00.0]
Loading VFIO driver...done
Unbinding 0000:02:00.0 from rtsx_pci...done
Adding IDs to VFIO driver...done
```

## Running

Assuming the host has IOMMU group 9 as above start the container with:

```
# podman --runtime=kata-runtime run -it --rm --cap-add=CAP_IPC_LOCK --device=/dev/vfio/9 fedora sh
```

Note: the `CAP_IPC_LOCK` is because programs need to lock memory in
order to create IOMMU mappings.  This is needed for regular
containers, not just Kata containers.

Inside the container shell you should see a VFIO device
```
# ls -l /dev/vfio
total 0
crw------- 1 root root 249, 0 Jun 22 11:19 2
crw-rw-rw- 1 root root 10, 196 Jun 22 11:19 vfio
```

## Trying with rootfs based initrd

A container image with Kata kernel and initrd having driver rebinding hook is available at docker.io/bpradipt/kata-initrd.
This can be used to get up and running quickly
Follow these steps to extract the kernel and initrd and place it in host path
```
CONTAINER_IMAGE=docker.io/bpradipt/kata-initrd
CONTAINER_RUNTIME=podman

echo "Copy initrd-vfio.img to /usr/share/kata-containers"
${CONTAINER_RUNTIME} run -v /usr/share/kata-containers:/data ${CONTAINER_IMAGE} cp /kata-containers-initrd.img /data/initrd-vfio.img

echo "Copy vmlinuz-vfio to /usr/share/kata-containers"
${CONTAINER_RUNTIME} run -v /usr/share/kata-containers:/data ${CONTAINER_IMAGE} cp /vmlinuz.container /data/vmlinuz-vfio
```
Then edit `configuration.toml` to specify the path for kernel and initrd
```
kernel="/usr/share/kata-containers/vmlinuz-vfio"
initrd="/usr/share/kata-containers/initrd-vfio"
```

## Debugging

If you access the Kata VM console using [these
instructions](../README.md#Debugging), the guest hook logs can be
found in `/tmp`.
