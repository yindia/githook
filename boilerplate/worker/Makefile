SHELL := /bin/sh

ROOT ?= ../..
CONFIG ?= ./config.yaml
SCRIPTS ?= $(ROOT)/scripts
SERVER ?= $(ROOT)/main.go

.PHONY: run-worker run-server docker-build docker-run deps-up deps-down

run-worker:
	go run . -config $(CONFIG)

run-server:
	go run $(SERVER) -config $(CONFIG)

deps-up:
	docker compose -f ./docker-compose.yaml up -d

deps-down:
	docker compose -f ./docker-compose.yaml down

docker-build:
	docker build -f ./Dockerfile -t githooks-worker ../..

docker-run:
	docker run --rm \
		-e GITHUB_WEBHOOK_SECRET \
		-e GITHUB_APP_ID \
		-e GITHUB_PRIVATE_KEY_PATH \
		-e GITLAB_ACCESS_TOKEN \
		-e BITBUCKET_ACCESS_TOKEN \
		githooks-worker -config /app/config.yaml
