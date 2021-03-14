BUILDTOOL=go
VERSION=0.1.1
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
COMMIT=$(shell git rev-parse HEAD)
COMPILED=$(shell date -u '+%Y%m%d-%H%M%S')
LDFLAGS="-X github.com/upvestco/httpsignature-proxy/cmd.compiled=$(COMPILED) -X github.com/upvestco/httpsignature-proxy/cmd.commit=$(COMMIT) -X github.com/upvestco/httpsignature-proxy/cmd.version=$(VERSION)"

default: macos

clean: 
	rm -rf httpsignature-*

macos: clean
	GOOS=darwin $(BUILDTOOL) build -ldflags $(LDFLAGS)
