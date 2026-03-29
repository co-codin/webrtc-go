APP      := webrtc-go
CMD      := ./cmd
BIN      := ./bin/$(APP)
DEV_COMPOSE  := containers/composes/dc.dev.yml
PROD_COMPOSE := containers/composes/dc.prod.yml

.PHONY: help build run test lint tidy \
        docker-build \
        dev dev-down dev-logs \
        prod prod-stop prod-down prod-logs

# ─── default ──────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "  Local development"
	@echo "    make build        build the binary to ./bin/$(APP)"
	@echo "    make run          run the server locally (go run)"
	@echo "    make test         run all tests"
	@echo "    make lint         vet + staticcheck (install staticcheck if needed)"
	@echo "    make tidy         go mod tidy"
	@echo ""
	@echo "  Docker (dev)"
	@echo "    make docker-build  build the app Docker image"
	@echo "    make dev           docker compose up (build + coturn)"
	@echo "    make dev-down      docker compose down"
	@echo "    make dev-logs      tail compose logs"
	@echo ""
	@echo "  Docker (prod)"
	@echo "    make prod          docker compose up -d"
	@echo "    make prod-stop     graceful stop (120 s timeout)"
	@echo "    make prod-down     docker compose down"
	@echo "    make prod-logs     tail compose logs"
	@echo ""

# ─── local ────────────────────────────────────────────────────────────────────

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -o $(BIN) $(CMD)

run:
	go run $(CMD)/main.go

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

tidy:
	go mod tidy

# ─── docker (dev) ─────────────────────────────────────────────────────────────

docker-build:
	docker build -t $(APP) -f containers/images/Dockerfile .

dev:
	docker compose -f $(DEV_COMPOSE) up --build

dev-down:
	docker compose -f $(DEV_COMPOSE) down

dev-logs:
	docker compose -f $(DEV_COMPOSE) logs -f --tail 100

# ─── docker (prod) ────────────────────────────────────────────────────────────

prod:
	docker compose -f $(PROD_COMPOSE) up -d --build

prod-stop:
	docker compose -f $(PROD_COMPOSE) stop --timeout 120

prod-down:
	docker compose -f $(PROD_COMPOSE) down --timeout 120

prod-logs:
	docker compose -f $(PROD_COMPOSE) logs -f --tail 100
