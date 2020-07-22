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
- The host device to be added to Kata container should be bound to VFIO driver

Note:
In certain cases you might have to use the unsafe_interrupts to allow passthrough of PCI devices to the VM
```
echo "options vfio_iommu_type1 allow_unsafe_interrupts=1" > /etc/modprobe.d/iommu_unsafe_interrupts.conf
```

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

## Create dracut based initrd

Perform the following steps as `root`

1. Copy updated kata-agent binary `/usr/libexec/kata-containers/agent/usr/bin`
2. Copy hooks directory to `/usr/libexec/kata-containers/agent/usr/share/oci/`
3. Apply the patch to add vfio modules  `cd /usr/libexec/kata-containers/osbuilder/dracut/dracut.conf.d && patch -t < 0001-Add-additional-VFIO-modules-to-the-initrd.patch`
4. Build initrd by running `/usr/libexec/kata-containers/osbuilder/kata-osbuilder.sh`

## Kata configuration.toml settings
Ensure the following settings are present in configuration.toml

```
machine_type = "q35"
guest_hook_path = "/usr/share/oci/hooks"
kernel_params = "systemd.unified_cgroup_hierarchy=0`
```

## Running

Assuming the host has a VFIO device `/dev/vfio/11` which you want to
provide to Kata container

```
podman run -it --rm -v /dev:/dev --device=/dev/vfio/11 --runtime=/usr/bin/kata-runtime fedora sh
```

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

## Hook Source Code
https://github.com/bpradipt/kata-hooks

## Debugging

If you access the Kata VM console using [these
instructions](../README.md#Debugging), the guest hook logs can be
found in `/tmp`.
