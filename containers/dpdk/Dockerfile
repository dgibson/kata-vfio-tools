FROM fedora

MAINTAINER David Gibson <dgibson@redhat.com>

# Add a layer which has downloaded the package lists, to speed up rebuild
RUN dnf -y makecache

# DPDK tools
RUN dnf install -y dpdk-tools

# Install debugging utils
RUN dnf install -y strace procps pciutils

RUN echo Dummy1

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
