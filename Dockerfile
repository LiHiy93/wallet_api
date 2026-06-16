FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -trimpath -ldflags="-s -w" -o /bin/wallet-api ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/wallet-api /usr/local/bin/wallet-api
COPY migrations ./migrations
EXPOSE 8080
CMD ["wallet-api"]
