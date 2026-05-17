
FROM golang:1.24-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/api


FROM alpine:3.20

RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN adduser -D -H -s /sbin/nologin appuser

WORKDIR /app

COPY --from=builder /out/app /app/app


COPY --from=builder /src/database/migrations /app/database/migrations

USER appuser

EXPOSE 8080
ENTRYPOINT ["/app/app"]