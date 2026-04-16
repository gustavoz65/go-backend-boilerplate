FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/main.go

FROM alpine:3.21

WORKDIR /app
RUN adduser -D -g '' appuser

COPY --from=builder /app/server ./server

USER appuser
EXPOSE 3000

CMD ["./server"]
