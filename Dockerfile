FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o subscriptions-aggregator ./cmd/app

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/subscriptions-aggregator /app/subscriptions-aggregator
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/docs /app/docs
COPY .env.example /app/.env.example

EXPOSE 8080

CMD ["/app/subscriptions-aggregator"]