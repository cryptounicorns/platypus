.DEFAULT_GOAL = all

numcpus  := $(shell cat /proc/cpuinfo | grep '^processor\s*:' | wc -l)
version  := $(shell git rev-list --count HEAD).$(shell git rev-parse --short HEAD)

name     := platypus
package  := github.com/cryptounicorns/$(name)

build       := ./build
build_id    := 0x$(shell echo $(version) | sha1sum | awk '{print $$1}')
ldflags     := -X $(package)/cli.version=$(version) -B $(build_id)
build_flags := -a -ldflags "$(ldflags)" -o build/$(name)

.PHONY: all
all:: dependencies

.PHONY: dependencies
dependencies::
	glide install

.PHONY: test
test:: dependencies
	go test -v $(shell glide novendor)

.PHONY: bench
bench:: dependencies
	go test -bench=. -v $(shell glide novendor)

.PHONY: lint
lint:: dependencies
	go vet -v $(shell glide novendor)

.PHONY: check
check:: lint test

.PHONY: all
all:: $(name)

.PHONY: $(name)
$(name):: dependencies
	mkdir -p $(build)
	@echo "Build id: $(build_id)"
	go build $(build_flags) -v $(package)/$(name)

.PHONY: build
build:: $(name)

.PHONY: clean
clean::
	git clean -xddff

include nix.mk
include config.mk
