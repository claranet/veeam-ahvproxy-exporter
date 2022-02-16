# Dockerfile builds an image for a client_golang example.
#
# Use as (from the root for the client_golang repository):
#    docker build -f examples/$name/Dockerfile -t prometheus/golang-example-$name .

# Builder image, where we build the example.

FROM golang:1.17 AS builder

ENV GOPATH /go

WORKDIR /veeam-ahvproxy-exporter
COPY . .
RUN echo "> GOPATH: " $GOPATH
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags -w

# Final image.
FROM quay.io/prometheus/busybox:latest

LABEL description "Prometheus exporter for VEEAM AHV Proxies" \
      version "v0.4.0" \
      maintainer "Martin Weber <martin.weber@de.clara.net>"

WORKDIR /
COPY --from=builder /veeam-ahvproxy-exporter/veeam-ahvproxy-exporter /usr/local/bin/veeam-ahvproxy-exporter
RUN touch /config.yml

EXPOSE 9760
ENTRYPOINT ["/usr/local/bin/veeam-ahvproxy-exporter"]
CMD [ "-exporter.conf", "/config.yml" ]
