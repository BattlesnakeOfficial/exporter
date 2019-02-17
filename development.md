# Development

You will need to install:

- Golang
- Docker
- gin (`go get github.com/codegangsta/gin`)

## Run the service

```shell
gin run serve.go
```

## Build the executable

```shell
make intsall
```

## Regenate the model from the openapi yaml

```shell
make model
```

## Regenate the custom image heads/tails from board repo

```shell
make images
```

## Run the tests

```shell
make test
```
