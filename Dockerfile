FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o open_oscar_server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o oscar-admin ./cmd/oscar_admin

FROM alpine:latest

RUN apk add --no-cache bash curl && \
    addgroup -S -g 65532 oscar && \
    adduser -S -D -H -u 65532 -G oscar oscar && \
    install -d -o oscar -g oscar /data

COPY --from=builder --chown=65532:65532 /app/open_oscar_server /app/open_oscar_server
COPY --from=builder --chmod=755 /app/oscar-admin /usr/local/bin/oscar-admin
COPY --chmod=755 scripts/import_bart.sh scripts/populate_icq_users.sh /usr/local/bin/

USER 65532:65532
WORKDIR /data

EXPOSE 5190 8080 9898 1088 4000/udp

CMD ["/app/open_oscar_server"]
