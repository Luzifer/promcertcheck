FROM golang:alpine

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

ADD . /go/src/github.com/Luzifer/promcertcheck
WORKDIR /go/src/github.com/Luzifer/promcertcheck

RUN set -ex \
 && apk add --update git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge git

EXPOSE 3000

VOLUME ["/data/certs"]

ENTRYPOINT ["/go/bin/promcertcheck"]
CMD ["--probe=https://www.google.com/", "--probe=https://www.facebook.com/"]
