test:
	go test -coverprofile coverage.txt -covermode=atomic ./...
.PHONY: test

build:
	packr2 build
.PHONY: build

install:
	packr2 install
.PHONY: install

run: build
	./exporter 
.PHONY: run

model:
	./scripts/model.sh
.PHONY: model

images:
	./scripts/images.sh
.PHONY: images

docker:
	docker build -t battlesnakeio/exporter .
.PHONY: docker