VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: version

fast:
	go build $(LDFLAGS) -o maccer

static:
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) -o maccer .

local: fast
	DEBUG=1 \
	BIND=localhost:8080 \
	BOT_ID="285421343594512384" \
	GUILD_ID="231799104731217931" \
	VERIFIED_ROLE="285459413882634241" \
	DEBUG_USER="86435690711093248" \
	ADMINISTRATIVE_CHANNEL="282581078643048448" \
	PRIMARY_CHANNEL="231799104731217931" \
	FORUM_ENDPOINT="https://forum.bayarearoleplay.com" \
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
	docker build --no-cache -t southclaws/maccer:$(VERSION) -f Dockerfile.dev .

build-prod:
	docker build --no-cache -t southclaws/maccer:$(VERSION) .

build-test:
	docker build --no-cache -t southclaws/maccer-test:$(VERSION) -f Dockerfile.testing .

push: build-prod
	docker push southclaws/maccer:$(VERSION)
	
run:
	-docker rm maccer-test
	docker run \
		--name maccer-test \
		--network host \
		-e BIND=0.0.0.0:8080 \
		-e BIND=localhost:8080 \
		-e BOT_ID="285421343594512384" \
		-e GUILD_ID="231799104731217931" \
		-e VERIFIED_ROLE="285459413882634241" \
		-e DEBUG_USER="86435690711093248" \
		-e ADMINISTRATIVE_CHANNEL="282581078643048448" \
		-e PRIMARY_CHANNEL="231799104731217931" \
		-e FORUM_ENDPOINT="https://forum.bayarearoleplay.com" \
		-e DEBUG=1 \
		southclaws/maccer:$(VERSION)

enter:
	docker run -it --entrypoint=bash southclaws/maccer:$(VERSION)

enter-mount:
	docker run -v $(shell pwd)/testspace:/samp -it --entrypoint=bash southclaws/maccer:$(VERSION)

# Test stuff

test-container: build-test
	docker run --network host southclaws/maccer-test:$(VERSION)

mongodb:
	-docker stop mongodb
	-docker rm mongodb
	docker run --name mongodb -p 27017:27017 -d mongo
