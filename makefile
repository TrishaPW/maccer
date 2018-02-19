VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
-include .env

.PHONY: version

fast:
	go build $(LDFLAGS) -o maccer

static:
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) -o maccer .

local: fast
	DEBUG=1 \
	./maccer

version:
	git tag $(VERSION)
	git push
	git push origin $(VERSION)

test:
	go test -v -race

# Docker

build:
	docker build --no-cache -t southclaws/maccer:$(VERSION) .

push: build
	docker push southclaws/maccer:$(VERSION)
	
run:
	-docker kill maccer-test
	-docker rm maccer-test
	docker run \
		--name maccer-test \
		--network host \
		--env-file .env \
		-e DEBUG=1 \
		southclaws/maccer:$(VERSION)

run-prod:
	-docker kill maccer
	-docker rm maccer
	docker run \
		--name maccer \
		-d \
		--env-file .env \
		-e DEBUG=1 \
		southclaws/maccer:$(VERSION)


# -
# Testing
# -


mongodb:
	-docker stop mongodb
	-docker rm mongodb
	docker run \
		--name mongodb \
		-p 27017:27017 \
		-d \
		mongo

