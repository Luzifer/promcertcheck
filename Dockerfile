FROM alpine

ENV GOPATH /go:/go/src/github.com/Luzifer/promcertcheck/Godeps/_workspace

ADD . /go/src/github.com/Luzifer/promcertcheck
WORKDIR /go/src/github.com/Luzifer/promcertcheck

RUN apk --update add git go ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short || echo dev)" \
 && apk --purge del git go

EXPOSE 3000
ENTRYPOINT ["/go/bin/promcertcheck"]
CMD ["--probe=https://www.google.com/", "--probe=https://www.facebook.com/"]
