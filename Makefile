.PHONY: build run

all: build run

build:
	@echo "-- building"
	go build -o ./bin/gophermart ./cmd/gophermart

run:
	@echo "-- running"
	./bin/gophermart

.PHONY: docker
docker:
	@echo "-- building docker container"
	docker build -f build/Dockerfile -t gophermart .

.PHONY: docker_run
docker_run:
	@echo "-- starting docker container"
	docker run -it -p 8080:8080 gophermart

.PHONY: dcb
dcb:
	@echo "-- starting docker compose"
	docker-compose -f ./deployments/docker-compose.yml up --build

.PHONY: test
test:
	@echo "-- testing"
	go test ./... -coverprofile cover.out

.PHONY: cover
cover:
	@echo "-- opening coverage"
	go tool cover -html cover.out
