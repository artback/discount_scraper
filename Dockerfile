# Build stage
FROM golang:1.25.3-alpine AS builder

ARG app_path

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/application $app_path

# Final stage
FROM alpine:latest

ARG app_path

WORKDIR /app

COPY --from=builder /app/application /app/application

CMD [ "/app/application" ]
