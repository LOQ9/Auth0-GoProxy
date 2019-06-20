VERSION?=$(shell git describe --tags --always)
CURRENT_DOCKER_IMAGE=loq9/auth0-goproxy:$(VERSION)
LATEST_DOCKER_IMAGE=loq9/auth0-goproxy:latest

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o build/auth0-goproxy cmd/auth0-goproxy/main.go
	docker build -t $(CURRENT_DOCKER_IMAGE) .

release: build
	docker push $(CURRENT_DOCKER_IMAGE)
	docker tag  $(CURRENT_DOCKER_IMAGE) $(LATEST_DOCKER_IMAGE)
	docker push $(LATEST_DOCKER_IMAGE)

.PHONY: build release
