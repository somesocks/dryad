

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
.PHONY: build
build:
	@(cd ./dryad/go/cli && go build)