# syntax=docker/dockerfile:1

FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY cmd/mancala ./cmd/mancala
COPY internal ./internal
COPY proto ./proto

RUN CGO_ENABLED=0 GOOS=linux go build -o mancala ./cmd/mancala/main.go

FROM builder AS test
RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS release

WORKDIR /

COPY --from=builder /app/mancala .

USER nonroot:nonroot
ENTRYPOINT ["./mancala"]