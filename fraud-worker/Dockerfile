FROM golang:alpine as base

WORKDIR /app

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

### Copy code and dependencies
COPY . .
RUN go mod download
COPY . .

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