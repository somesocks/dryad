
BASE=$(shell pwd)

##
##	PROJECT DRYAD
##	-------------
##

##	make help - display the help
##
.PHONY: help
help:
	@grep -h "^##.*" ./Makefile

##	bootstrap-build - build a bootstrap version of dryad for the current os/arch to use to build the dev shell
##
.PHONY: bootstrap-build
bootstrap-build:
	@(cd $(BASE)/dyd/roots/dryad/src/dyd/assets && GO111MODULE=on go build -o $(BASE)/bootstrap/dryad)
# @(cd ./dryad/go/cli && go build)

##	bootstrap-install - install the bootstrap dryad as /usr/bin/dryad
##
.PHONY: bootstrap-install
bootstrap-install:
	@ sudo cp $(BASE)/bootstrap/dryad /usr/bin/dryad

##	dev-shell - use the bootstrap dryad to build and start the dev shell
##
.PHONY: dev-shell
dev-shell:
	@ dryad root build $(BASE)/dyd/roots/dev-shell && dryad stem exec $(BASE)/dyd/sprouts/dev-shell