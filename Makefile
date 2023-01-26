
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

##	make build - build the cli versions of dryad
##
.PHONY: bootstrap-build
bootstrap-build:
	@(cd $(BASE)/dyd/roots/core/dryad/dyd/assets/src && go build -ldflags '-s -w' -o $(BASE)/bootstrap/dryad)
# @(cd ./dryad/go/cli && go build)

.PHONY: dev
dev:
	@ dryad root build ./dyd/roots/core/dev-shell && dryad stem exec ./dyd/sprouts/core/dev-shell