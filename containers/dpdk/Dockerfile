FROM fedora

MAINTAINER Adrian Moreno <amorenoz@redhat.com>

RUN dnf groupinstall -y "Development Tools"
RUN dnf install -y wget numactl-devel git make numactl numatop strace ltrace pciutils dpdk-tools which driverctl gcc vim

# Install debugging utils
RUN dnf -y install strace ltrace gdb procps util-linux coreutils

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
RUN make -j4 EXTRA_CFLAGS="-g -O0" install T=${RTE_TARGET} DESTDIR=install
RUN cp ${DPDK_DIR}/install/bin/testpmd /usr/bin
RUN cp install/bin/testpmd /usr/bin

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
