FROM alpine:3.19 AS certs-src

FROM golang:1.24-alpine AS builder

WORKDIR /usr/local/src/pantry-api

COPY . .

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go build -o bin/pantry-api .

FROM scratch

COPY --from=certs-src /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/local/src/pantry-api/bin/pantry-api /usr/local/bin/

CMD ["/usr/local/bin/pantry-api"]
