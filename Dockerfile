# Development Stage (dev üçün)
FROM golang:1.24-alpine AS development
WORKDIR /app
RUN apk --no-cache add ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["go", "run", "./cmd/api"]

# Builder Stage (prod build üçün)
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# Runtime Stage (prod run)
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /api .
COPY --from=builder /worker .
CMD ["./api"]
