FROM registry.access.redhat.com/ubi9-minimal:9.7-1771346502 AS build

ARG vendor_version
ARG TARGETARCH

WORKDIR /

RUN microdnf --assumeyes update \
    && microdnf --assumeyes --nodocs --setopt=install_weak_deps=0 --setopt=tsflags=nodocs install \
    gzip \
    tar \
    && microdnf clean all

RUN curl -L -o coredns.tgz "https://github.com/coredns/coredns/releases/download/v$vendor_version/coredns_${vendor_version}_linux_$TARGETARCH.tgz" \
    && tar zxvf coredns.tgz \
    && rm -f coredns.tgz

RUN setcap cap_net_bind_service=+ep /coredns


FROM registry.access.redhat.com/ubi9-minimal:9.7-1771346502

ARG commit_id
ARG vendor_version
ARG image_version

LABEL com.instana.commit.id=$commit_id
LABEL com.instana.vendor.version=$vendor_version
LABEL com.instana.image.version=$image_version
LABEL com.instana.name=coredns

WORKDIR /

COPY --from=build /coredns /coredns

USER nobody:nobody

EXPOSE 53 53/udp

ENTRYPOINT ["/coredns"]