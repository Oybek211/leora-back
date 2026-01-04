FROM golang:1.25.3-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /leora-server cmd/server/main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /leora-server /app/leora-server
COPY scripts/wait-for-db.sh /app/wait-for-db.sh
RUN chmod +x /app/wait-for-db.sh
COPY configs /app/configs
COPY migrations /app/migrations
COPY .env.example /app/.env
EXPOSE 9090
ENTRYPOINT ["/app/leora-server"]
