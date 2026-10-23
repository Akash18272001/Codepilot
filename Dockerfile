FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o agent-server ./cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/agent-server .
EXPOSE 8080
CMD ["./agent-server"]
