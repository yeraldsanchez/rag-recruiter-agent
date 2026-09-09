FROM golang:1.26-alpine AS base
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/worker ./cmd/worker

FROM alpine:latest AS api
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /bin/api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]

FROM alpine:latest AS worker
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /bin/worker /app/worker
ENTRYPOINT ["/app/worker"]