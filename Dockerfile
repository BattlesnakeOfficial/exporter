FROM golang:1.11-alpine as builder
RUN apk --no-cache add build-base git

RUN git clone https://github.com/gobuffalo/packr.git /go/src/github.com/gobuffalo/packr
RUN cd /go/src/github.com/gobuffalo/packr/v2 && make install
WORKDIR /go/src/github.com/battlesnakeio/exporter/
COPY . .
RUN packr2 install 
# RUN CGO_ENABLED=0 GOOS=linux packr2 install -installsuffix cgo ...

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/ /bin/
CMD ["/bin/exporter"]
