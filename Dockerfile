FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN cd cmd/service && CGO_ENABLED=0 go build -ldflags "-s -w" -o auth-app
FROM golang:1.24-alpine
WORKDIR /app
COPY --from=builder /app/cmd/service/auth-app ./auth-app

COPY db/migrations /app/db/migrations
COPY cmd /app/cmd
COPY pkg /app/pkg
COPY internal /app/internal
COPY go.mod go.sum ./
EXPOSE 8080
ENTRYPOINT ["./auth-app"]