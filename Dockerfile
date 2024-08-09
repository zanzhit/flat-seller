FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o flat-seller cmd/flat-seller/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client postgresql

WORKDIR /root/

COPY --from=builder /app/flat-seller .
COPY ./config /config
COPY wait-for-postgres.sh .

RUN chmod +x wait-for-postgres.sh

CMD ["./wait-for-postgres.sh", "db", "5432", "--", "./flat-seller"]
