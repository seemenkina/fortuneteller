FROM golang:alpine AS builder
WORKDIR /app
COPY ./ ./
VOLUME ["/go", "/root/.cache"]
RUN go build -o fortuneteller ./cmd/service

FROM alpine:latest
WORKDIR /app
COPY ./wait-for ./wait-for
COPY --from=builder /app/fortuneteller ./fortuneteller
EXPOSE 8080
VOLUME ["/app/assets", "/app/books"]
ENTRYPOINT ["/app/wait-for", "database:5432", "--", "/app/fortuneteller"]
