# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/todo-api ./cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/todo-worker ./cmd/worker

FROM alpine:3.20
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /bin/todo-api /app/todo-api
COPY --from=builder /bin/todo-worker /app/todo-worker
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080
USER app

CMD ["/app/todo-api"]
