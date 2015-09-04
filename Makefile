VERSION = $(shell git describe --tags)

test:
	go generate
	go run *.go \
		--probe="https://www.cloudkeys.de/" \
		--probe="https://foo.hub.luzifer.io/" \
		--probe="https://registry.luzifer.io/" \
		--probe="https://blog.knut.me/" \
		--probe="https://gobuilder.me/" \
		--probe="https://pwd.luzifer.io/" \
		--probe="https://www.itpad.de/" \
		--probe="https://mondash.org/"

container: build
	docker build -t luzifer/promcertcheck .

build:
	docker run -v $(CURDIR):/src -e LDFLAGS='-X main.version $(VERSION)' centurylink/golang-builder:latest
