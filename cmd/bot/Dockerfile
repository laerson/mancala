# syntax=docker/dockerfile:1

FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY cmd/bot ./cmd/bot
COPY internal ./internal
COPY proto ./proto

RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot/main.go

FROM builder AS test
RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS release

WORKDIR /

COPY --from=builder /app/bot .

EXPOSE 50057

USER nonroot:nonroot
CMD ["./bot"]