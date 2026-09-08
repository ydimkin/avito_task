FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/core-service ./cmd/app/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates curl

COPY --from=builder /app/bin/core-service /app/core-service

EXPOSE 8080

ENTRYPOINT ["/app/core-service"]
