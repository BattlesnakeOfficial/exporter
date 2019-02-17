test:
	go test -coverprofile coverage.txt -covermode=atomic ./...
.PHONY: test

install:
	go build
.PHONY: install

run: install
	./exporter 
.PHONY: install

model:
	./scripts/model.sh
.PHONY: model

images:
	./scripts/images.sh
.PHONY: images

docker:
	docker build -t battlesnakeio/exporter .
.PHONY: docker