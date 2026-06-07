ifeq ($(OS),Windows_NT)
CUR_DIR=$(shell echo %CD%)
else
CUR_DIR=$(shell pwd)
endif

.PHONY: run migrate

# Локальный запуск без Docker (нужен PostgreSQL и файл .env).
run:
	go run ./cmd/service

# Миграции в локальный PostgreSQL (Windows: powershell -File scripts/migrate.ps1).
migrate:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/migrate.ps1
else
	@bash scripts/migrate.sh
endif

# Генерация pkg/api и api/service.swagger.json (protoc + локальные third_party; без buf.build).
# Требуются Docker и сеть при первом запуске (git clone googleapis / grpc-gateway).
proto.gen:
	docker run --rm \
		-v ${CUR_DIR}:/workspace \
		-w /workspace \
		golang:1.24-bookworm \
		bash /workspace/scripts/proto-gen.sh


# proto.deps.update создает buf.lock
proto.deps.update:
	docker run --rm \
		-v ${CUR_DIR}:/workspace \
		-w /workspace \
		bufbuild/buf:1.57.0 dep update

lint:
	docker run --rm -v `pwd`:/app -w /app golangci/golangci-lint:v2.4.0 golangci-lint run --timeout 5m0s -v
