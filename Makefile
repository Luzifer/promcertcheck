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
		--probe="https://mondash.org/" \
		--log-level debug

container:
	docker build -t luzifer/promcertcheck .

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh
