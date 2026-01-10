
FROM golang:1.24-alpine AS development
WORKDIR /app
RUN apk --no-cache add ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Development üçün port
EXPOSE 8080 
CMD ["go", "run", "./cmd/api"]

# Builder Stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# Runtime Stage (Render bura baxacaq)
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
# Builder-dən binary-ləri götürürük
COPY --from=builder /api .
COPY --from=builder /worker .

COPY --from=builder /app/migrations ./migrations

# Render üçün portu açırıq
EXPOSE 8080

# API serverini başladırıq
CMD ["./api"]
