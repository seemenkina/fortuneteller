FROM golang:alpine
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY ./ ./
RUN go build -o fortuneteller ./cmd/service
EXPOSE 8080
VOLUME ["/app/assets", "/app/books", "/app/books_keys"]

ENTRYPOINT ["/app/wait-for", "database:5432", "--", "/app/fortuneteller"]