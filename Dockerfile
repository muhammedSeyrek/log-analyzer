# syntax=docker/dockerfile:1.4


FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN cd plugins && go run ../tools/gen/main.go

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
        -trimpath \
        -ldflags "-s -w -X log-analyzer/internal/cli.version=docker" \
        -o /out/log-analyzer \
        ./cmd

FROM scratch AS bin
COPY --from=builder /out/log-analyzer /log-analyzer

FROM alpine:3.19 AS runner
