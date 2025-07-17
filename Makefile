.PHONY: ci
ci: lint test build

.PHONY: gen
gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/service.proto

.PHONY: mockgen
mockgen:
	mockgen -package=mock -source=internal/service/service.go -destination=internal/service/mock/repository_mock.go

.PHONY: build
build:
	go build -o ./bin/batch-saver -a ./cmd/batch-saver

.PHONY: deps
deps:
	go mod tidy

.PHONY: test
test:
	go test -v -count=1 ./...

.PHONY: dockerise
dockerise:
	docker build -t batch-saver .

.PHONY: docker-up
docker-up: dockerise
	docker-compose up -d

.PHONY: docker-down
docker-down:
	docker-compose down