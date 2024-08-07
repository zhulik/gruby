MRUBY_COMMIT ?= 3.3.0
MRUBY_VENDOR_DIR ?= mruby-build

GOLANGCI_LINT_VERSION := $(shell cat .golangci-lint-version)

all: libmruby.a lint test

clean:
	rm -rf ${MRUBY_VENDOR_DIR}
	rm -f libmruby.a.

libmruby.a: ${MRUBY_VENDOR_DIR}/mruby
	cd ${MRUBY_VENDOR_DIR}/mruby && ${MAKE}

${MRUBY_VENDOR_DIR}/mruby:
	mkdir -p ${MRUBY_VENDOR_DIR}
	git clone https://github.com/mruby/mruby.git ${MRUBY_VENDOR_DIR}/mruby
	cd ${MRUBY_VENDOR_DIR}/mruby && git reset --hard && git clean -fdx
	cd ${MRUBY_VENDOR_DIR}/mruby && git checkout ${MRUBY_COMMIT}

test: libmruby.a
	go test -race

.PHONY: all clean libmruby.a test

lint: bin/golangci-lint
	./bin/golangci-lint run

lint-fix: bin/golangci-lint
	./bin/golangci-lint run --fix

bin/golangci-lint: .golangci-lint-version
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- v$(GOLANGCI_LINT_VERSION)
