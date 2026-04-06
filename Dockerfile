FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN go mod init oscar-proxy
COPY oscar-proxy.go .
RUN go build -o /oscar-proxy oscar-proxy.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /oscar-proxy /oscar-proxy
EXPOSE 5190
CMD ["/oscar-proxy"]