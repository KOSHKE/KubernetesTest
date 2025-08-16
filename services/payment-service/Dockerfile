FROM golang:1.25-alpine

WORKDIR /app

ENV GOBIN=/go/bin

# Install Air once at build time
RUN go install github.com/air-verse/air@latest

CMD ["air", "-c", ".air.toml"]


