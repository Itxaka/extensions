BUNDLED_EXTENSIONS ?= k8s-build-scheduler migrate-emerge remote-exec geninitramfs kernel-switcher autobump-github autobump-github-ng geniso genimage qa-artefacts migrate-entropy package-browser parallel-tools portage apkbuildconverter repo-devkit
BUNDLED_EXTENSIONS_TEST ?= autobump-github
UBINDIR ?= /usr/bin
DESTDIR ?=

all: build

build:
	for d in $(BUNDLED_EXTENSIONS); do $(MAKE) -C extensions/$$d build; done

install: build
	for d in $(BUNDLED_EXTENSIONS); do $(MAKE) -C extensions/$$d install; done

install_luet:
	curl https://get.mocaccino.org/luet/get_luet_root.sh | sh

test:
	for d in $(BUNDLED_EXTENSIONS_TEST); do $(MAKE) -C extensions/$$d test; done