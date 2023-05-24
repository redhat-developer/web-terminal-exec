# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/ubi8/go-toolset
FROM registry.access.redhat.com/ubi8/go-toolset:1.19.6-4 as builder
ENV GOPATH=/go/
USER root
WORKDIR /web-terminal-exec
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go env GOPROXY && \
    go mod download

# Copy the go source
COPY . .

# Compile Web Terminal Exec binary
RUN make compile

# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/ubi8-minimal
FROM registry.access.redhat.com/ubi8-minimal:8.8-860
RUN microdnf -y update && microdnf clean all && rm -rf /var/cache/yum && echo "Installed Packages" && rpm -qa | sort -V && echo "End Of Installed Packages"
WORKDIR /
COPY --from=builder /web-terminal-exec/_output/bin/web-terminal-exec /usr/local/bin/web-terminal-exec
# For compatibility, Web Terminal binary must be available in /go/bin/che-machine-exec
RUN mkdir -p /go/bin/ && ln -s /usr/local/bin/web-terminal-exec /go/bin/che-machine-exec

ENV USER_UID=1001 \
    USER_NAME=web-terminal-exec

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
CMD /usr/local/bin/web-terminal-exec

# append Brew metadata here
