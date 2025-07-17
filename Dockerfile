FROM golang:1.24.1 AS builder
WORKDIR /app
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/batch-saver -a ./cmd/batch-saver

FROM scratch
COPY --from=builder /app/bin/batch-saver /batch-saver
COPY --from=builder /app/internal/storage/migrations /internal/storage/migrations
EXPOSE 3000
CMD ["./batch-saver"]