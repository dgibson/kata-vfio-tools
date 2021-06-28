Running containers with crictl
==============================

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

> # c=$(crictl start $(crictl create $(crictl runp --runtime=kata pod.json) container-busybox.json pod.json))

Then to enter the container to test:

> # crictl exec -it $c /bin/sh
