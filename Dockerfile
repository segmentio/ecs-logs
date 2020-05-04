FROM golang:1.14-alpine as builder
RUN apk add --update curl ca-certificates make git gcc g++ python
# Enable go modules
ENV GO111MODULE=on
# enable go proxy for faster builds
ENV GOPROXY=https://proxy.golang.org
COPY . /go/src/github.com/segmentio/ecs-logs
WORKDIR $GOPATH/src/github.com/segmentio/ecs-logs
COPY . $GOPATH/src/github.com/segmentio/ecs-logs
# this is an auto-generated build command
# based upon the first argument of the entrypoint in the existing dockerfile.  
# This will work in most cases, but it is important to note
# that in some situations you may need to define a different build output with the -o flag
# This comment may be safely removed
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s -extldflags "-static"' -o /ecs-logs
FROM 528451384384.dkr.ecr.us-west-2.amazonaws.com/segment-scratch
COPY --from=builder ecs-logs ecs-logs
ENTRYPOINT ["ecs-logs"]
