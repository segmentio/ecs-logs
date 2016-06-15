FROM ubuntu:16.04
MAINTAINER engineering@segment.com

RUN apt-get update -y && apt-get install -y ca-certificates

COPY ecs-logs /usr/local/bin/ecs-logs

ENTRYPOINT ["ecs-logs"]
