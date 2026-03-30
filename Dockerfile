# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/todo ./cmd

FROM alpine:3.20
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /bin/todo /app/todo

EXPOSE 8080
USER app

CMD ["/app/todo"]
