test:
	go test -timeout 20s -race -coverprofile coverage.txt -covermode=atomic ./...
.PHONY: test

install:
	go build
.PHONY: install

run: install
	./exporter 
.PHONY: run

model:
	./model.sh
.PHONY: proto

docker:
	docker build -t battlesnakeio/exporter .
.PHONY: build-docker

lint:
	golangci-lint run
.PHONY: lint