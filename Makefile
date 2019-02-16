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
	rm -rf /tmp/gomodel
	mkdir /tmp/gomodel
	docker run --rm -v /tmp/gomodel:/local openapitools/openapi-generator-cli generate --model-name-prefix engine \
		-i https://raw.githubusercontent.com/battlesnakeio/docs/master/apis/engine/spec.yaml \
		-g go \
		-o /local/ 
	docker run --rm -v /tmp/gomodel:/local openapitools/openapi-generator-cli generate --model-name-prefix snake \
		-i https://raw.githubusercontent.com/battlesnakeio/docs/master/apis/snake/spec.yaml \
		-g go \
		-o /local/
	rm -rf model
	mkdir model
	cp /tmp/gomodel/*model*.go model
.PHONY: proto

docker:
	docker build -t battlesnakeio/exporter .
.PHONY: build-docker

lint:
	golangci-lint run
.PHONY: lint