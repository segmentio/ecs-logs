# We need a go compiler that's based on an image with libsystemd-dev installed,
# segment/golang give us just that.
FROM segment/golang:latest

# Copy the ecs-logs sources so they can be built within the container.
COPY . /go/src/github.com/segmentio/ecs-logs

# Build ecs-logs, then cleanup all unneeded packages.
RUN cd /go/src/github.com/segmentio/ecs-logs && \
    govendor sync && \
    go build -o /usr/local/bin/ecs-logs && \
    apt-get remove -y apt-transport-https build-essential git curl docker-engine && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/* /go/* /usr/local/go /usr/src/Makefile*

# Sets the container's entry point.
ENTRYPOINT ["ecs-logs"]
