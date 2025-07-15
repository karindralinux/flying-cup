FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/flying-cup .

FROM alpine:latest

# Install ca-certificates for SSL support, Git, and wget for cloudflared
RUN apk --no-cache add ca-certificates git wget

# Install cloudflared
RUN wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -O /usr/local/bin/cloudflared && \
    chmod +x /usr/local/bin/cloudflared && \
    cloudflared version

WORKDIR /app

COPY --from=builder /app/flying-cup .

EXPOSE 8080

CMD ["./flying-cup"]