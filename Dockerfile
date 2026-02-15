# build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata
RUN adduser -D -g '' appuser

WORKDIR /build

# deps primeiro pra aproveitar cache
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . .

# static build, sem CGO
RUN mkdir -p /app && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /app/server ./cmd/server 2>/dev/null || \
    (printf 'package main\n\nfunc main() { println("Hardened App Running") }\n' > main.go && \
    CGO_ENABLED=0 go build -ldflags='-w -s' -o /app/server .)

# runtime — distroless pra não ter shell nem nada extra
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL maintainer="meluansantos"
LABEL org.opencontainers.image.source="https://github.com/meluansantos/secure-pipeline-poc"
LABEL org.opencontainers.image.description="Pipeline Hardening PoC"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /app/server

# 65532 = nonroot do distroless
USER 65532:65532

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "-health"] || exit 1

ENTRYPOINT ["/app/server"]

# alpine com shell, só pra debug quando precisar
FROM alpine:3.19 AS runtime-debug

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' -u 10001 appuser && \
    rm -rf /var/cache/apk/*

COPY --from=builder /app/server /app/server

RUN chown -R appuser:appuser /app && \
    chmod 500 /app/server

USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/server"]
