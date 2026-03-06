# ── build stage ──────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Cache dependency downloads separately from the build.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /bin/app \
    ./main

# ── final stage ──────────────────────────────────────────────────────────────
# scratch + CA certificates gives the smallest possible image while still
# allowing outbound HTTPS (arXiv API, Telegram API).
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /app

ENTRYPOINT ["/app"]
