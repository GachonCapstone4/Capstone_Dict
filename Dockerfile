FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o diag-tool ./cmd/diag-tool/

FROM alpine:3.19
RUN apk add --no-cache iproute2 iputils procps
COPY --from=builder /app/diag-tool /usr/local/bin/diag-tool
ENTRYPOINT ["diag-tool"]
