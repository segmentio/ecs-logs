GIT_DIRTY := $(shell test -n "`git status --porcelain`" && echo "-CHANGES" || true)
GIT_DESCRIBE := $(shell git describe --tags --always)
VERSION := $(patsubst v%,%,$(GIT_DESCRIBE)$(GIT_DIRTY))
LDFLAGS := "-X main.version=$(VERSION)"
REPO := github.com/segmentio/ecs-logs
DEBFILE := ecs-logs_$(VERSION)_amd64.deb
SOURCES := $(git ls-files *.go)

default: bin/ecs-logs-linux-amd64

bin/ecs-logs-linux-amd64: $(SOURCES)
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o $@ $(REPO)

depend:
	go get -u github.com/kardianos/govendor
	govendor sync
	gem install --no-ri --no-rdoc fpm package_cloud

dep: depend

$(DEBFILE): bin/ecs-logs-linux-amd64
	@if [ -z "$(VERSION)" ]; then echo "VERSION not defined"; false; fi
	fpm -s dir  -t deb -n ecs-logs -v $(VERSION) -m sre-team@segment.com --vendor "Segment.io, Inc." \
		./bin/ecs-logs-linux-amd64=/usr/bin/ecs-logs

deb: $(DEBFILE)

upload_deb: $(DEBFILE)
	package_cloud push segment/infra/ubuntu/xenial $(DEBFILE)

clean:
	-rm -f bin/* *.deb

.PHONY: depend dep clean deb upload_deb
