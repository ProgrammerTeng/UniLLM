.PHONY: dev build test deploy clean

# ---- Local Development ----
dev:
	@echo "Starting PostgreSQL and Redis (assumes local install)..."
	@echo "Starting Go backend..."
	cd . && go run ./cmd/server &
	@echo "Starting Next.js frontend..."
	cd web && npm run dev &

build:
	go build -o server ./cmd/server
	cd web && npm run build

test:
	go test ./...

# ---- Docker ----
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# ---- Production Deploy ----
deploy:
	@echo "Deploying UniLLM..."
	@test -f .env || (echo "ERROR: .env file missing. Copy .env.example to .env and fill in values." && exit 1)
	docker-compose -f docker-compose.yml up -d --build
	@echo "Deployed! Check http://localhost"

# ---- Utilities ----
clean:
	rm -f server unillm
	cd web && rm -rf .next node_modules

gen-secrets:
	@echo "JWT_SECRET=$$(openssl rand -hex 32)"
	@echo "ENCRYPTION_KEY=$$(openssl rand -hex 32)"
	@echo "POSTGRES_PASSWORD=$$(openssl rand -base64 24)"
