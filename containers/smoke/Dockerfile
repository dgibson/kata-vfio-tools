FROM fedora

MAINTAINER David Gibson <dgibson@redhat.com>

# Add a layer which has downloaded the package lists, to speed up rebuild
RUN dnf -y makecache

# Install things necessary to build the tests
RUN dnf install -y git gcc

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
