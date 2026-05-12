# Stage 1

FROM golang:1.26.3-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o sub-service ./cmd/server

# Stage 2

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata curl

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder --chown=appuser:appuser /app/sub-service .
COPY --from=builder --chown=appuser:appuser /app/migrations ./migrations

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

EXPOSE 8080

CMD ["./sub-service"]
