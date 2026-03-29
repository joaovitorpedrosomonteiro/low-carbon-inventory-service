FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /inventory-service ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /inventory-service .

ENV PORT=8085
ENV DATABASE_URL=postgres://lowcarbon:lowcarbon@postgres:5432/lowcarbon_inventory?sslmode=disable

EXPOSE 8085

ENTRYPOINT ["./inventory-service"]