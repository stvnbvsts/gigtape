.PHONY: build-api build-cli serve-web test test-integration

build-api:
	go build ./apps/api

build-cli:
	go build -o gigtape ./apps/cli

serve-web:
	cd apps/web && npm run dev

test:
	go test ./packages/...

test-integration:
	RUN_INTEGRATION=true go test -tags integration ./packages/adapters/...
