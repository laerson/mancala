# syntax=docker/dockerfile:1

FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY cmd/notifications ./cmd/notifications
COPY internal ./internal
COPY proto ./proto

RUN CGO_ENABLED=0 GOOS=linux go build -o notifications ./cmd/notifications/main.go

FROM builder AS test
RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS release

WORKDIR /

COPY --from=builder /app/notifications .

EXPOSE 50056

USER nonroot:nonroot
CMD ["./notifications"]