FROM golang:1.24-alpine AS builder

WORKDIR /rss_exporter
COPY . /rss_exporter

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

RUN go mod tidy && go test ./... && go build -trimpath -ldflags="-w -s" -o rss_exporter

FROM alpine:latest AS runner

WORKDIR /rss_exporter
COPY --from=builder /rss_exporter/rss_exporter .
COPY --from=builder /rss_exporter/config.example.yml config.yml

EXPOSE 9091/tcp

ENTRYPOINT ["./rss_exporter"]
