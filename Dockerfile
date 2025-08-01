FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN cd cmd/service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o auth-app


FROM alpine:latest
RUN apk add --no-cache curl
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
WORKDIR /app
COPY --from=builder /app/cmd/service/auth-app ./auth-app
COPY --from=builder /app/db/migrations ./db/migrations
RUN chown -R 1000:1000 /app && chmod 755 /app
USER 1000:1000
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1
ENTRYPOINT ["./auth-app"]