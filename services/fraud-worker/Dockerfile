FROM golang:1.14-alpine as base

RUN apk add --no-cache git

WORKDIR /app

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

### Copy code and dependencies
COPY . .
RUN go mod download
COPY . .

RUN go version

### Build
RUN go build -ldflags '-s -w' -o app
RUN ls

### Certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

### App
FROM scratch as app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=base app /
ENTRYPOINT ["/app"]