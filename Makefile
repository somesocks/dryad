
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
	./scripts/nix-shell/run.sh ./scripts/tasks/bootstrap-build.sh

# @(cd ./dryad/go/cli && go build)

##	bootstrap-install - install the bootstrap dryad as /usr/bin/dryad
##
.PHONY: bootstrap-install
bootstrap-install:
	./scripts/nix-shell/run.sh ./scripts/tasks/bootstrap-install.sh

##	dev-shell - use the bootstrap dryad to build and start the dev shell
##
.PHONY: dev-shell
dev-shell:
	./scripts/nix-shell/run.sh ./scripts/tasks/dryad-shell.sh
