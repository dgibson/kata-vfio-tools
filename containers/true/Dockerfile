FROM fedora

MAINTAINER David Gibson <dgibson@redhat.com>

#ENV DNFFLAGS --disablerepo=updates

# Add a layer which has downloaded the package lists, to speed up rebuild
RUN dnf $DNFFLAGS -y makecache

WORKDIR /root

COPY entrypoint.sh /root
ENTRYPOINT ["/root/entrypoint.sh"]
