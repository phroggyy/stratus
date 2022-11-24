.PHONY: deps dev build topic

deps:
	docker compose up -d
dev:
	npx nodemon --exec "make build" -e "go mod"
build:
	go build .
topic:
	docker compose exec redpanda rpk topic create events
