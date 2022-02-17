ARG GOARCH=arm64
ARG GITHASH=unset
ARG VERSIONSTRING=unset

FROM golang:1.17-alpine AS builder

ARG GOARCH
ARG GITHASH
ARG VERSIONSTRING

RUN mkdir -p /workdir
WORKDIR /workdir
COPY go.mod go.sum ./
RUN apk --no-cache add ca-certificates
RUN go mod download
COPY . .
RUN \
  CGO_ENABLED=0 GOARCH=${GOARCH} go build -ldflags "-extldflags '-static' -X github.com/lisa/speedtest-go-exporter/pkg/version.GitHash=${GITHASH} -X github.com/lisa/speedtest-go-exporter/pkg/version.Version=${VERSIONSTRING}" -a -o /workdir/speedtest-exporter . && chmod +x /workdir/speedtest-exporter

FROM scratch

COPY --from=builder /workdir/speedtest-exporter /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD [ "/speedtest-exporter" ]
