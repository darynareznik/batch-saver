# Batch Saver

A high-concurrency Golang gRPC service for batching and persisting events.

## Features

**gRPC API**: Simple endpoint to stream events to the batcher.

**Batcher**: Buffers incoming events by `groupID` and flushes them to the database based on:
- Max event count
- Max wait duration

**Writer Pool**: Limits concurrent database writes to avoid overload.

**PostgreSQL**: Persistent event storage.

**Docker**: Fully containerized setup with `docker-compose`.

## Running the Service

To run the service: `make docker-up`