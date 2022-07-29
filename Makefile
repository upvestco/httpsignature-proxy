BUILDTOOL=go
VERSION=1.1.0-dev
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
COMMIT=$(shell git rev-parse HEAD)
COMPILED=$(shell date -u '+%Y%m%d-%H%M%S')
BUILTBY=$(shell id -un)
LDFLAGS="-X github.com/upvestco/httpsignature-proxy/cmd.date=$(COMPILED) -X github.com/upvestco/httpsignature-proxy/cmd.commit=$(COMMIT) -X github.com/upvestco/httpsignature-proxy/cmd.version=$(VERSION) -X github.com/upvestco/httpsignature-proxy/cmd.builtBy=$(BUILTBY)"

default: macos

clean:
	rm -rf httpsignature-proxy

macos: clean
	GOOS=darwin $(BUILDTOOL) build -ldflags $(LDFLAGS)
linux: clean
	GOOS=linux $(BUILDTOOL) build -ldflags $(LDFLAGS)
win: clean
	GOOS=windows $(BUILDTOOL) build -ldflags $(LDFLAGS)
