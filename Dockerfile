FROM golang:1.26-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/brenox-api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/brenox-api /usr/local/bin/brenox-api

ENV PORT=8080
EXPOSE 8080

USER nobody
ENTRYPOINT ["/usr/local/bin/brenox-api"]
