# Stage 1: Build binary
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server .

# Stage 2: Minimal runtime image
FROM alpine:3.19

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser
COPY --from=builder --chown=appuser:appgroup /app/server .

ENV PORT=8080 \
    APP_ENV=dev \
    APP_VERSION=v1.0.0

EXPOSE 8080

ENTRYPOINT ["./server"]
