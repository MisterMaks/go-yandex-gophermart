.PHONY: build run

all: build run

build:
	@echo "-- building"
	go build -o ./bin/gophermart ./cmd/gophermart

run:
	@echo "-- running"
	./bin/gophermart
