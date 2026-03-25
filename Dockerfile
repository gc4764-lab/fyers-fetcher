# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o fyers-fetcher main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest
WORKDIR /root/
# Install timezone data if your trading logic relies on specific timezones
RUN apk add --no-cache tzdata
ENV TZ=Asia/Kolkata 

COPY --from=builder /app/fyers-fetcher .
CMD ["./fyers-fetcher"]

