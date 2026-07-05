FROM golang:1.26-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=1.0.0
ARG COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w \
    -X github.com/brainart16/brenox/internal/version.Engine=${VERSION} \
    -X github.com/brainart16/brenox/internal/version.Commit=${COMMIT}" \
  -o /out/brenox-engine ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app

COPY --from=builder /out/brenox-engine /usr/local/bin/brenox-engine

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD curl -fsS "http://127.0.0.1:${PORT}/health" || exit 1

USER nobody
ENTRYPOINT ["/usr/local/bin/brenox-engine"]
