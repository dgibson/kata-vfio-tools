# Introduction

Steps to run pods using a locally built Kata v2.x, with cri-o and
crictl, but not a full Kubernetes install.

## Assumptions

TBD


## Setting up crictl & CRI-O

Install the packages you'll need.  In Fedora:

> # dnf install cri-tools containernetworking-plugins

## Build and install CRI-O

Unfortunately as of Jun 30 2021 Fedora's latest cri-o version is
`cri-o-2:1.19.0-43.gitc00d425.module_f34+10209+fb610d12.x86_64`.  This
doesn't have the fix for `https://github.com/cri-o/cri-o/issues/4589`,
which will break the way this repo tries to build things. 

```
$ git clone https://github.com/cri-o/cri-o.git
$ cd cri-o
$ make
$ sudo make install
$ sudo cp 10-crio-bridge.conf /etc/cni/net.d
```

And you'll need to point it at the Fedora packaged CNI plugins, by
creating `/etc/crio/crio.conf.d/rpm-cni.conf` with the contents
```
[crio.network]

plugin_dirs = [
	"/usr/libexec/cni",
]
```

### Running a test container with crictl + runc

Run the CRI-O daemon as root:
```
# crio -l debug
```

The following commans run in another shell, in the top directory of
the `kata-vfio-tools` repo.  This uses the sample configuration files
in `crictl/*.json`.

Start a pod/sandbox:
> # p = $(crictl runp pod.json)

Create the container in the pod:
> # c = $(crictl create $p container-busybox.json pod.json)

Start the container:
> # crictl start

Run a shell within the container to test:
> # crictl exec -it $c /bin/sh

### Running a test container (helper shortcuts)

From the top directory of the `kata-vfio-tools` repo.  Not necessarily
as root, but with sudo permission:

NB: Using these you can only run one container at a time.

To run the `busybox` sample container:
```
$ make -C crictl run-busybox
```

To run a shell in the `busybox` container:
```
$ make sh-busybox
```

To shut down all running pods and containers:
```
$ make cleaner
```

## Using a local Kata build

From the top directory of the `kata-vfio-tools` repo:

1. Update `KATASRC` in the `Makefile` with the path to your Kata 2.x
   tree (by default `~/src/kata-containers`).
2. Run `make`

This will build and (minimally) install Kata underneath `build/` in
the `kata-vfio-tools` tree, so that it doesn't mess up things on your
host system (and doesn't need root).

You'll also need to point Kata at the right kernel.  Look for the `mbuto` output from the `make` above, e.g.:
```
Kata Containers [hypervisor.qemu] configuration:

	kernel = "/boot/vmlinuz-5.12.12-300.fc34.x86_64"
	initrd = "/home/dwg/src/kata-vfio-tools/build/kata-initrd.img"
```

Then
```
$ ln -s /boot/vmlinuz-5.12.12-300.fc34.x86_64 build/kata-vmlinuz
```

### Configuring CRI-O to use your local Kata build

As root, run:

```
# make crio-conf-kata-vfio
```

That will rewrite `/etc/crio/crio.conf.d/kata-vfio.conf` with the
correct configuration for your local Kata build under the name
`kata-vfio`.

