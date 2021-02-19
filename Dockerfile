FROM golang:alpine as builder

COPY . /go/src/github.com/Luzifer/promcertcheck
WORKDIR /go/src/github.com/Luzifer/promcertcheck

RUN set -ex \
 && apk add --update git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      -mod=readonly

FROM alpine:latest

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      ca-certificates

COPY --from=builder /go/bin/promcertcheck /usr/local/bin/promcertcheck

EXPOSE 3000
VOLUME ["/data/certs"]

ENTRYPOINT ["/usr/local/bin/promcertcheck"]
CMD ["--probe=https://www.google.com/", "--probe=https://www.facebook.com/"]

# vim: set ft=Dockerfile:
