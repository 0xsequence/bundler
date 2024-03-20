TOP              := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SHELL            = bash -o pipefail
TEST_FLAGS       ?= -p 1 -v
# MOD_VENDOR       ?= -mod=vendor

GITTAG           ?= $(shell git describe --exact-match --tags HEAD 2>/dev/null || :)
GITBRANCH        ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || :)
LONGVERSION      ?= $(shell git describe --tags --long --abbrev=8 --always HEAD)$(echo -$GITBRANCH | tr / - | grep -v '\-master' || :)
VERSION          ?= $(if $(GITTAG),$(GITTAG),$(LONGVERSION))
GITCOMMIT        ?= $(shell git log -1 --date=iso --pretty=format:%H)
GITCOMMITDATE    ?= $(shell git log -1 --date=iso --pretty=format:%cd)
GITCOMMITAUTHOR  ?= $(shell git log -1 --date=iso --pretty="format:%an")

define run
	@GOGC=off go build -o ./bin/$(1) ./cmd/$(1)/main.go
	@./bin/$(1) --config=$(2)
endef

run:
	$(call run,bundler-node,./etc/bundler-1.conf)

run2:
	$(call run,bundler-node,./etc/bundler-2.conf)

run-template:
	$(call run,bundler-node,./etc/bundler-node.conf.sample)

define build
	GOGC=off GOBIN=$$PWD/bin \
	go install -v \
	  $(MOD_VENDOR) \
		-tags='$(BUILDTAGS)' \
		-gcflags='-e' \
		-ldflags='-X "github.com/0xsequence/bundler/bundler.VERSION=$(VERSION)" -X "github.com/0xsequence/bundler/bundler.GITBRANCH=$(GITBRANCH)" -X "github.com/0xsequence/bundler/bundler.GITCOMMIT=$(GITCOMMIT)" -X "github.com/0xsequence/bundler/bundler.GITCOMMITDATE=$(GITCOMMITDATE)" -X "github.com/0xsequence/bundler/bundler.GITCOMMITAUTHOR=$(GITCOMMITAUTHOR)"' \
		$(1)
endef

build: build-node

build-node:
	$(call build, ./cmd/bundler-node)

.PHONY: test
test:
	go clean -testcache && go test -v $$(go list ./... | grep -v /cmd/)

clean:
	rm -rf ./bin/*
	go clean -cache -testcache

.PHONY: proto
proto:
	go generate ./proto
