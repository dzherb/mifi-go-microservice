.PHONY: run-dev
run-dev:
	@go run cmd/main.go

.PHONY: run
run:
	@docker compose up -d
