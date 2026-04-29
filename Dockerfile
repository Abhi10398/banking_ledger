# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk --no-cache add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o myapp .

# ── Run stage ──────────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/myapp        ./myapp
COPY --from=builder /app/config       ./config
COPY --from=builder /app/static       ./static

EXPOSE 8080

ENTRYPOINT ["./myapp", "api"]
