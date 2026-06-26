.PHONY: dev dev-backend dev-frontend check build-backend build-frontend check-release

dev:
	./scripts/dev-refactor.sh

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && pnpm dev

check:
	cd backend && go test ./...
	cd backend && go vet ./...
	cd frontend && pnpm type-check
	cd frontend && pnpm build

build-backend:
	mkdir -p backend/.cache
	cd backend && go build -o .cache/bophotos ./cmd/server

build-frontend:
	cd frontend && pnpm build

check-release: check build-backend
	test -f frontend/dist/index.html
	grep -q 'COPY frontend/' Dockerfile
	grep -q 'COPY backend/' Dockerfile
	test "$$(find . -maxdepth 1 -type d ! -name . ! -name .git ! -name backend ! -name frontend ! -name docs | wc -l | tr -d ' ')" = "0"
