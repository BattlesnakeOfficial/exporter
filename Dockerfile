FROM golang:1.18.1-alpine as builder

COPY . /go/src/github.com/BattlesnakeOfficial/exporter/
WORKDIR /go/src/github.com/BattlesnakeOfficial/exporter

RUN apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux go install -installsuffix cgo ./cmd/...

# -----

FROM alpine:3.15.4

ARG APP_VERSION=0.0.0
ENV APP_VERSION=$APP_VERSION

RUN apk add --no-cache ca-certificates inkscape

WORKDIR /app

COPY --from=builder /go/bin/ /bin/
COPY ./media/assets/ ./media/assets/

CMD ["/bin/exporter"]
