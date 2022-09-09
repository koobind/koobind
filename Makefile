include Makefile.conf

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: version
version: ## Set version in code
	@echo "// Generated by Makefile\n\npackage misc\n\nvar version = \"v$(VERSION)\"\n" >koocli/cmd/misc/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"v$(VERSION)\"\n" >koomgr/internal/config/version_.go

.PHONY: doc
doc: ## Generate doc index
	# doctoc README.md --github --title '## Index'

.PHONY: precommit
precommit: doc version ## To ensure uptodate generated stuff.
	cd koomgr && make precommit