# Introduction

This document describes the steps required to use Kata guest agent
hook for rebinding the SRIOV device to VFIO using podman

## Assumptions

- Fedora 32+ or RHEL8.2+ distro
- Kata packages installed distro repo
    `kata-runtime`
	`kata-osbuilder`
	`kata-shim`
	`kata-proxy`
	`kata-agent`
- Golang is installed
- GOPATH is setup
- A host system with supported PCI/SRIOV vendor:device -
    `8086:1521`
    `8086:1520`
    `8086:158b`
    `15b3:1015`
    `15b3:1017`

    List taken from OpenShift supported SRIOV device [list](https://docs.openshift.com/container-platform/4.2/networking/multiple_networks/configuring-sr-iov.html#supported-devices_configuring-sr-iov)
    with `8086:1521` added to handle the test infra
- Host system booted with `intel_iommu=on`
- Host system booted with `systemd.unified_cgroup_hierarchy=0`
- SELinux is set to permissive mode on the host
- The host device to be added to Kata container should be bound to VFIO driver

Note:
In certain cases you might have to use the unsafe_interrupts to allow passthrough of PCI devices to the VM
```
echo "options vfio_iommu_type1 allow_unsafe_interrupts=1" > /etc/modprobe.d/iommu_unsafe_interrupts.conf
```

## Build the host side Kata components

From the top-level directory of this tree, run:
```
$ make
```

This will download and build the sources for `kata-runtime`,
`kata-proxy` and `kata-shim`.  It will install them locally into
`$KATAPREFIX` - by default `build/prefix` within this working tree.

It will also build a Kata `configuration.toml` suitable for using VFIO
in `$KATAPREFIX/etc/configuration.toml` which the components are built
do use instead of the default system one.

## Configure podman to use the locally built runtime

This will make podman aware of a new runtime, called `kata-vfio` which
will use the components built locally above.  As root, run:
```
# make podman-conf-kata-vfio
```

You should only do this once.  If you move this tree, you'll need to
manually remove the stasnza from
`/usr/share/containers/containers.conf` and re-run it.

## Build kata-agent with fix for guest hooks

A new dracut based initrd needs to be created having the built `kata-agent` binary
```
$ mkdir -p ${GOPATH}/src/github.com/kata-containers
$ cd ${GOPATH}/src/github.com/kata-containers
$ git clone https://github.com/bpradipt/agent.git
$ cd agent
$ git checkout -b hook-fix origin/hook-fix
$ make
```
This will build the `kata-agent` binary

## Build vfio-hook

The source for the vfio hook program is in `vfio-hook/` in this repository.
```
$ cd vfio-hook
$ go build -v .
```

## Add additional VFIO modules to osbuilder dracut configuration

As `root`, from the top directory of this repository:

```
# (cd /usr/libexec/kata-containers/osbuilder/dracut/dracut.conf.d && patch -t ) < 0001-Add-additional-VFIO-modules-to-the-initrd.patch
```

Alternatively, you can manually edit
`/usr/libexec/kata-containers/osbuilder/dracut/dracut.conf.d/15-dracut-fedora.conf`
to add the modules `vfio_iommu_type1`, `irqbypass`, `vfio_virqfd` to
the `drivers` variable.

## Create dracut based initrd

Run as `root`:

```
# ./scripts/vfio-osbuilder.sh
```

## Kata configuration.toml settings
Ensure the following settings are present in configuration.toml

```
machine_type = "q35"
guest_hook_path = "/usr/share/oci/hooks"
kernel_params = "systemd.unified_cgroup_hierarchy=0`
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
# podman --runtime=kata-runtime run -it --rm -v /dev:/dev --cap-add=CAP_IPC_LOCK --device=/dev/vfio/9 fedora sh
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
