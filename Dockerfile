FROM segment/golang:latest
MAINTAINER engineering@segment.com

COPY . /go/src/github.com/segmentio/ecs-logs
RUN go build -o /usr/local/bin/ecs-logs github.com/segmentio/ecs-logs && \
    rm -rf /go/* /usr/local/go /usr/src/Makefile* && \
    apt-get remove -y apt-transport-https build-essential git curl docker-engine && \
    apt-get autoremove -y && \
    apt-get clean -y

ENTRYPOINT ["ecs-logs"]
