FROM segment/golang:latest
MAINTAINER engineering@segment.com

COPY . /go/src/github.com/segmentio/ecs-logs
RUN go build -o /usr/local/bin/ecs-logs github.com/segmentio/ecs-logs && \
    rm -rf /go/* /usr/local/go

ENTRYPOINT ["ecs-logs"]
