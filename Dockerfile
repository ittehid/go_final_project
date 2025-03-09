FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o go_final_project ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/go_final_project /app/go_final_project
COPY web /app/web

ENV TODO_DBFILE=/app/data/scheduler.db
ENV TODO_PASSWORD=12345

CMD ["/app/go_final_project"]