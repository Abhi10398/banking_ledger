# Makefile

GOPATH=$(shell go env GOPATH)

export GOPATH
export GOROOT

.PHONY: run dev migrate seed test-concurrency \
        docker-up docker-down docker-logs docker-migrate docker-clean

# ── Local dev ─────────────────────────────────────────────────────────────────

# run — start the full stack via Docker Compose (matches Phase 6 spec)
run:
	docker-compose up --build

# dev — run the binary directly against a local postgres (no Docker)
dev:
	go run main.go api static

# migrate — run Liquibase migrations locally (requires liquibase on PATH)
# Override DB with: make migrate PG_HOST=… PG_USER=… PG_PASSWORD=…
migrate:
	go run ./cmd/migrate/

# seed — create 5 sample accounts with balances (app must be running)
# Override target with: make seed BASE_URL=http://localhost:3000/api
seed:
	go run ./cmd/seed/

# test-concurrency — fire 50 concurrent transfers and assert invariants
# Usage: make test-concurrency A=<uuid> B=<uuid>
test-concurrency:
	python3 scripts/concurrency_test.py $(A) $(B)

# ── Docker ────────────────────────────────────────────────────────────────────

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

docker-migrate:
	docker compose run --rm migrate

docker-clean:
	docker compose down -v --remove-orphans

build:
	go mod tidy
	go build -o myapp .

start: build
	./myapp api static

pg-install:
	brew install postgresql@18

pg-start:
	brew services start postgresql@18

pg-stop:
	brew services stop postgresql@18

mongo-install:
	brew tap mongodb/brew && brew install mongodb-community@7.0

mongo-start:
	brew services start mongodb-community@7.0

mongo-stop:
	brew services stop mongodb-community@7.0

list-services:
	brew services list

# Liquibase — install
liquibase-install:
	brew install liquibase

liquibase-mongo-ext:
	./scripts/install-liquibase-mongo-ext.sh

# Liquibase — PostgreSQL
PG_URL ?= jdbc:postgresql://localhost:5432/banking_ledger
PG_USER ?=
PG_PASSWORD ?=

pg-migrate:
	cd migrations/postgres && liquibase \
		--url="$(PG_URL)" \
		--username="$(PG_USER)" \
		--password="$(PG_PASSWORD)" \
		update

pg-migrate-status:
	cd migrations/postgres && liquibase \
		--url="$(PG_URL)" \
		--username="$(PG_USER)" \
		--password="$(PG_PASSWORD)" \
		status

pg-migrate-rollback:
	cd migrations/postgres && liquibase \
		--url="$(PG_URL)" \
		--username="$(PG_USER)" \
		--password="$(PG_PASSWORD)" \
		rollbackCount 1

pg-migrate-history:
	cd migrations/postgres && liquibase \
		--url="$(PG_URL)" \
		--username="$(PG_USER)" \
		--password="$(PG_PASSWORD)" \
		history

# Liquibase — MongoDB
MONGO_URL ?= mongodb://localhost:27017/awesome-database
MONGO_USER ?=
MONGO_PASSWORD ?=

mongo-migrate:
	cd migrations/mongo && liquibase \
		--url="$(MONGO_URL)" \
		--username="$(MONGO_USER)" \
		--password="$(MONGO_PASSWORD)" \
		update

mongo-migrate-status:
	cd migrations/mongo && liquibase \
		--url="$(MONGO_URL)" \
		--username="$(MONGO_USER)" \
		--password="$(MONGO_PASSWORD)" \
		status

mongo-migrate-rollback:
	cd migrations/mongo && liquibase \
		--url="$(MONGO_URL)" \
		--username="$(MONGO_USER)" \
		--password="$(MONGO_PASSWORD)" \
		rollbackCount 1

mongo-migrate-history:
	cd migrations/mongo && liquibase \
		--url="$(MONGO_URL)" \
		--username="$(MONGO_USER)" \
		--password="$(MONGO_PASSWORD)" \
		history