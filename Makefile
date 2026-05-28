.PHONY: build-api build-cli serve-web docker-build docker-up docker-down test test-integration

build-api:
	go build ./apps/api

build-cli:
	go build -o gigtape ./apps/cli

serve-web:
	cd apps/web && npm run dev

docker-build:
	docker compose build

docker-up:
	docker compose up --build

docker-down:
	docker compose down

test:
	go test ./packages/domain/... ./packages/usecases/... ./packages/adapters/setlistfm/... ./packages/adapters/spotify/... ./apps/api/... ./apps/cli/...

test-integration:
	RUN_INTEGRATION=true go test -tags integration ./packages/adapters/setlistfm/... ./packages/adapters/spotify/...
