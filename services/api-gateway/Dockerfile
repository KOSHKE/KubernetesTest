FROM golang:1.25-alpine

WORKDIR /app

ENV GOBIN=/go/bin

# Install Air once at build time
RUN go install github.com/air-verse/air@latest

# Default command runs Air with service config
CMD ["air", "-c", ".air.toml"]