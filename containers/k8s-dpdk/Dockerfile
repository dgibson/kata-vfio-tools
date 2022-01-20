FROM fedora:35

MAINTAINER David Gibson <dgibson@redhat.com>

#ENV DNFFLAGS --disablerepo=updates

# Add a layer which has downloaded the package lists, to speed up rebuild
RUN dnf $DNFFLAGS -y makecache

# DPDK tools
RUN dnf $DNFFLAGS -y install dpdk-tools

# Firmware, needed for some devices
#RUN dnf $DNFFLAGS -y install linux-firmware

# Install debugging utils
RUN dnf $DNFFLAGS -y install strace procps pciutils

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
