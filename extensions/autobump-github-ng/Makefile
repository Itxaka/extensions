UBINDIR ?= /usr/bin
SHAREDIR ?= /usr/share
DESTDIR ?=
GOOS?=linux
GOARCH?=amd64

all: build install

build:
	CGO_ENABLED=0 go build -o luet-autobump-github-ng  main.go strategies.go

install: build
	install -d $(DESTDIR)/$(UBINDIR)
	install -m 0755 luet-autobump-github-ng $(DESTDIR)/$(UBINDIR)/
