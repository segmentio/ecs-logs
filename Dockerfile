# We need a go compiler that's based on an image with libsystemd-dev installed,
# segment/golang give us just that.
FROM segment/golang:latest AS builder

# Copy the ecs-logs sources so they can be built within the container.
COPY . /go/src/github.com/segmentio/ecs-logs

# Build ecs-logs, then cleanup all unneeded packages.
RUN cd /go/src/github.com/segmentio/ecs-logs && \
    govendor sync && \
    go build -o /usr/local/bin/ecs-logs

FROM ubuntu:16.04
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    apt-get clean -y

COPY --from=builder /usr/local/bin/ecs-logs /usr/local/bin/ecs-logs

# Sets the container's entry point.
ENTRYPOINT ["ecs-logs"]
