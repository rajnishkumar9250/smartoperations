.PHONY: build test clean prepare update docker

GO = CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=cmd/device-mqtt

.PHONY: $(MICROSERVICES)

DOCKERS=docker_device_mqtt_go

.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse HEAD)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-mqtt-go.Version=$(VERSION)"

build: $(MICROSERVICES)
	$(GO) build ./...

cmd/device-mqtt:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

test:
	$(GO) test ./... -cover

clean:
	rm -f $(MICROSERVICES)

prepare:

run:
	cd bin && ./edgex-launch.sh

docker: $(DOCKERS)

docker_device_mqtt_go:
	docker build \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-device-mqtt-go:$(GIT_SHA) \
		-t edgexfoundry/docker-device-mqtt-go:$(VERSION)-dev \
		.
