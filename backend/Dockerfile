# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
# 添加go代理，解决镜像下载问题
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod ./
RUN go mod download || true

# Copy source code
COPY . .

# Download dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]