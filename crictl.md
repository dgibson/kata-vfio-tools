# Introduction

Steps to run pods using a locally built Kata v2.x, with cri-o and
crictl, but not a full Kubernetes install.

## Assumptions

TBD


## Setting up crictl & CRI-O

Install the packages you'll need.  In Fedora:

> # dnf module enable cri-o:1.20
> # dnf install cri-o cri-tools containernetworking-plugins


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

### Configuring CRI-O to use your local Kata build

As root, run:

```
# make crio-conf-kata-vfio
```

That will append a stanza to `/etc/crio/crio.conf` with the correct
configuration for your local Kata build under the name `kata-vfio`.
You only need to do this once, if it needs updating, you'll have to
manually remove it from `crio.conf` and re-run.
