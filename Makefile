.PHONY: build run clean test fmt web admin docker

APP=goshop

# ---- 后端 ----
build:
	go build -o bin/$(APP) cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test ./... -v

fmt:
	go fmt ./...
	go vet ./...

tidy:
	go mod tidy

# ---- 前端 ----
web:
	cd web && npm install && npm run dev

web-build:
	cd web && npm ci && npm run build

admin:
	cd admin && npm install && npm run dev

admin-build:
	cd admin && npm ci && npm run build

# ---- Docker ----
docker:
	docker compose up -d

docker-build:
	docker compose build

docker-down:
	docker compose down

# ---- 工具 ----
clean:
	rm -rf bin/

seed:
	@echo "重置数据库并重新 seed..."
	mysql -h127.0.0.1 -uroot -p -e "DROP DATABASE IF EXISTS goshop; CREATE DATABASE goshop CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
	go run cmd/server/main.go &
	@sleep 3 && kill $$! 2>/dev/null; echo "Seed 完成"
