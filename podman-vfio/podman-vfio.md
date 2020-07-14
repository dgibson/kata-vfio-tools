# Introduction
This document describes the steps required to use Kata guest agent
hook for rebinding the SRIOV device to VFIO using podman

## Assumptions
- Fedora 32+ distro
- Kata runtime packages installed from Fedora distro repo
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
- The host device to be added to Kata container should be bounded to VFIO driver


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
Perform the following steps as `root` user
1. Copy updated kata-agent binary `/usr/libexec/kata-containers/agent/usr/bin`
2. Copy hooks directory to `/usr/libexec/kata-containers`
3. Patch `/usr/libexec/kata-containers/osbuilder/kata-osbuilder.sh` using `0001-kata-osbuilder-Copy-hooks-to-initrd.patch`
4. Build initrd by running `/usr/libexec/kata-containers/osbuilder/kata-osbuilder.sh`

## Kata configuration.toml settings
Ensure the following settings are present in configuration.toml

```
machine_type = "q35"
sandbox_cgroup_only=true
guest_hook_path = "/usr/share/oci/hooks"
```

## Running
Assuming the host has a VFIO device `/dev/vfio/11` which you want to provide to Kata container

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

## Debugging
Enable Kata VM console by ensuring the following setting in `configuration.toml`

```
kernel_params = "agent.debug_console"
```

Run the Kata container, and get the kata container id *(assuming only one Kata container running on the system)*
```
CID=$(kata-runtime list -q)
```
Access Kata VM console for debugging
```
socat stdin,raw,echo=0,escape=0x11 unix-connect:"/run/vc/vm/${CID}/console.sock"
```

Access guest hook logs under `/tmp/`


