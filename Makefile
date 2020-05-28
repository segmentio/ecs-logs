VERSION := unmaintained
LDFLAGS := "-X main.version=$(VERSION)"
REPO := github.com/segmentio/ecs-logs
SOURCES := $(git ls-files *.go)
DOCKER_TAG := segment/ecs-logs:v$(VERSION)

default: bin/ecs-logs-linux-amd64

bin/ecs-logs-linux-amd64: $(SOURCES)
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o $@ $(REPO)

vendor:
	go mod vendor

test:
	go test $(shell go list ./...)

image:
	docker build -t $(DOCKER_TAG) -t segment/ecs-logs:latest .

push_image:
	docker push $(DOCKER_TAG)
	docker push segment/ecs-logs:latest

clean:
	-rm -f bin/* *.deb

.PHONY: test clean deb upload_deb image push_image
