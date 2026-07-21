FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o open_oscar_server ./cmd/server

FROM alpine:latest

RUN addgroup -S -g 65532 oscar && \
    adduser -S -D -H -u 65532 -G oscar oscar && \
    install -d -o oscar -g oscar /data

COPY --from=builder --chown=65532:65532 /app/open_oscar_server /app/open_oscar_server

USER 65532:65532
WORKDIR /data

EXPOSE 5190 8080 9898 1088 4000/udp

CMD ["/app/open_oscar_server"]
