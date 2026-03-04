FROM golang:1.25.7-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /finops-api ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /finops-api /usr/local/bin/finops-api
COPY --from=builder /app/internal/database/migrations /migrations

ENV MIGRATIONS_DIR=/migrations
ENV SERVER_PORT=8080
EXPOSE 8080
CMD ["finops-api"]
