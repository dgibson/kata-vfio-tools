FROM fedora

MAINTAINER David Gibson <dgibson@redhat.com>

# Add a layer which has downloaded the package lists, to speed up rebuild
RUN dnf -y makecache

# Tools to build DPDK
RUN dnf install -y git-core make gcc numactl-devel

# Install debugging utils
#RUN dnf -y install strace ltrace gdb procps pciutils

ARG repo=https://github.com/DPDK/dpdk.git
ARG version=master

ENV REPO $repo
ENV VER $version

WORKDIR /usr/src
RUN git clone $REPO dpdk

ENV DPDK_DIR=/usr/src/dpdk
WORKDIR ${DPDK_DIR}
RUN git checkout ${VER}

# Build DPDK

ENV RTE_TARGET=x86_64-native-linuxapp-gcc
ENV RTE_SDK=${DPDK_DIR}

# Disable kernel-related stuff
RUN sed -i -e 's/EAL_IGB_UIO=y/EAL_IGB_UIO=n/' config/common_linux
RUN sed -i -e 's/KNI_KMOD=y/KNI_KMOD=n/' config/common_linux
RUN sed -i -e 's/LIBRTE_KNI=y/LIBRTE_KNI=n/' config/common_linux
RUN sed -i -e 's/LIBRTE_PMD_KNI=y/LIBRTE_PMD_KNI=n/' config/common_linux
RUN make -j4 EXTRA_CFLAGS="-g -O0" install T=${RTE_TARGET} DESTDIR=install MAKE_PAUSE=n
RUN cp ${DPDK_DIR}/install/bin/testpmd /usr/bin
RUN cp install/bin/testpmd /usr/bin

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
