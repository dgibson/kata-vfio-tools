Running containers locally with crictl
======================================

Installing CRI-O and crictl
---------------------------

In Fedora:

> # dnf module enable cri-o:1.20
> # dnf install cri-o cri-tools containernetworking-plugins

Run the CRI-O daemon
--------------------

> # crio -l debug

Run your container
------------------

Using the sample configuration files in `crictl/*.json`:

Start a pod/sandbox:
> # p = $(crictl runp pod.json)

Create the container in the pod:
> # c = $(crictl create $p container-busybox.json pod.json)

Start the container:
> # crictl start

Run a shell within the container to test:
> # crictl exec -it $c /bin/sh

Run sample containers (shortcuts)
---------------------------------

Using the scripts and Makefile in `crictl/`:

NB: Using these you can only run one container at a time.

To run the `busybox` sample container:
> $ make run-busybox

To run a shell in the `busybox container:
> $ make sh-busybox

To shut down all running pods and containers:
> $ make cleaner

Using Kata 2.x with crictl
==========================

Build Kata 2.x
--------------


