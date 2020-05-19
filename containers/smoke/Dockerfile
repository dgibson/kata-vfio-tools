FROM fedora

MAINTAINER Adrian Moreno <amorenoz@redhat.com>

RUN dnf groupinstall -y "Development Tools"
RUN dnf install -y wget numactl-devel git make numactl numatop strace ltrace pciutils which driverctl gcc 

# Install debugging utils
RUN dnf -y install strace ltrace gdb procps util-linux coreutils

ARG repo="https://github.com/awilliam/tests.git"
ARG version=master

ENV REPO $repo
ENV VER $version

WORKDIR /usr/src
RUN git clone $REPO

WORKDIR /usr/src/tests
RUN for file in `ls -x vfio*.c`;  do \
    bin=$(echo $file | cut -d '.' -f 1); \
    gcc -o $bin $file; \
    cp $bin /usr/local/bin || echo "oops"; \
    done

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
