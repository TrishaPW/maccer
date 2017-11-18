VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
include .env

.PHONY: version

fast:
	go build $(LDFLAGS) -o maccer

static:
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) -o maccer .

local: fast
	BIND=localhost:8080 \
	FORUM_ENDPOINT="https://forum.bayarearoleplay.com" \
	BOT_ID=380354914326413322 \
	LOG_CHANNEL=381193287958265866 \
	GUILD_ID=334457680972218368 \
	VERIFIED_ROLE=381201593913311252 \
	DEBUG_USER=86435690711093248 \
	ADMINISTRATIVE_CHANNEL=282581078643048448 \
	PRIMARY_CHANNEL=231799104731217931 \
	DISCORD_TOKEN=$(DISCORD_TOKEN) \
	FORUM_KEY=$(FORUM_KEY) \
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
		-e BIND=0.0.0.0:8080 \
		-e FORUM_ENDPOINT="https://forum.bayarearoleplay.com" \
		-e BOT_ID=380354914326413322 \
		-e LOG_CHANNEL=381193287958265866 \
		-e GUILD_ID=334457680972218368 \
		-e VERIFIED_ROLE=381201593913311252 \
		-e DEBUG_USER=86435690711093248 \
		-e ADMINISTRATIVE_CHANNEL=282581078643048448 \
		-e PRIMARY_CHANNEL=231799104731217931 \
		-e DISCORD_TOKEN=$(DISCORD_TOKEN) \
		-e FORUM_KEY=$(FORUM_KEY) \
		-e DEBUG=1 \
		southclaws/maccer:$(VERSION)

run-prod:
	-docker kill maccer
	-docker rm maccer
	docker run \
		--name maccer \
		-d \
		-e BIND=0.0.0.0:8080 \
		-e FORUM_ENDPOINT="https://forum.bayarearoleplay.com" \
		-e BOT_ID=380354914326413322 \
		-e LOG_CHANNEL=381193287958265866 \
		-e GUILD_ID=334457680972218368 \
		-e VERIFIED_ROLE=381201593913311252 \
		-e DEBUG_USER=86435690711093248 \
		-e ADMINISTRATIVE_CHANNEL=282581078643048448 \
		-e PRIMARY_CHANNEL=231799104731217931 \
		-e DISCORD_TOKEN=$(DISCORD_TOKEN) \
		-e FORUM_KEY=$(FORUM_KEY) \
		-e DEBUG=1 \
		southclaws/maccer:$(VERSION)
