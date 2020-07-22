FROM fedora

MAINTAINER David Gibson <dgibson@redhat.com>

# Add a layer which has downloaded the package lists, to speed up rebuild
RUN dnf -y makecache

# Install things necessary to build the tests
RUN dnf install -y git-core gcc make

# Install debugging utils
RUN dnf -y install strace ltrace gdb procps-ng pciutils

ARG repo="https://github.com/dgibson/vfio-tests.git"

ENV REPO $repo

WORKDIR /usr/src
RUN git clone $REPO

WORKDIR /usr/src/vfio-tests
RUN make

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
