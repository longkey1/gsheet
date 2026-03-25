.DEFAULT_GOAL := help

export GO_VERSION=$(shell grep "^go " go.mod | sed 's/^go //')
export PRODUCT_NAME=$(shell cat .product_name 2>/dev/null || echo "unknown")

.PHONY: build
build: ## Build the binary to ./bin/
	@mkdir -p bin
	go build -o bin/$(PRODUCT_NAME)

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: vet
vet: ## Vet code
	go vet ./...

.PHONY: tidy
tidy: ## Tidy dependencies
	go mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/

.PHONY: tools
tools: ## Install tools
	go install github.com/goreleaser/goreleaser@latest

.PHONY: release

# Get current version from git tag
VERSION := $(shell v=$$(git tag --sort=-v:refname | head -n1 2>/dev/null); [ -n "$$v" ] && echo "$$v" || echo "v0.0.0")
MAJOR := $(shell echo $(VERSION) | cut -d. -f1 | tr -d 'v')
MINOR := $(shell echo $(VERSION) | cut -d. -f2)
PATCH := $(shell echo $(VERSION) | cut -d. -f3)

# Calculate next version based on release type
define next_version
$(if $(filter patch,$1),v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1),\
$(if $(filter minor,$1),v$(MAJOR).$(shell expr $(MINOR) + 1).0,\
$(if $(filter major,$1),v$(shell expr $(MAJOR) + 1).0.0,\
v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1))))
endef

# Variables for release target
dryrun ?= true
type ?=

release: ## Release target with type argument. Usage: make release type=patch|minor|major dryrun=false
	@if [ "$(type)" = "" ]; then \
		echo "Usage: make release type=<type> [dryrun=false]"; \
		echo ""; \
		echo "Types:"; \
		echo "  patch  - Increment patch version (e.g., v1.2.3 -> v1.2.4)"; \
		echo "  minor  - Increment minor version (e.g., v1.2.3 -> v1.3.0)"; \
		echo "  major  - Increment major version (e.g., v1.2.3 -> v2.0.0)"; \
		echo ""; \
		echo "Options:"; \
		echo "  dryrun - Set to false to actually create and push the tag (default: true)"; \
		echo ""; \
		echo "Current version: $(VERSION)"; \
		exit 0; \
	elif [ "$(type)" = "patch" ] || [ "$(type)" = "minor" ] || [ "$(type)" = "major" ]; then \
		NEXT_VERSION=$$(if [ "$(type)" = "patch" ]; then \
			echo "v$(MAJOR).$(MINOR).$$(expr $(PATCH) + 1)"; \
		elif [ "$(type)" = "minor" ]; then \
			echo "v$(MAJOR).$$(expr $(MINOR) + 1).0"; \
		elif [ "$(type)" = "major" ]; then \
			echo "v$$(expr $(MAJOR) + 1).0.0"; \
		fi); \
		echo "Current version: $(VERSION)"; \
		echo "Next version: $$NEXT_VERSION"; \
		if [ "$(dryrun)" = "false" ]; then \
			echo "Creating new tag $$NEXT_VERSION..."; \
			git push origin master --no-verify --force-with-lease; \
			git tag -a $$NEXT_VERSION -m "Release of $$NEXT_VERSION"; \
			git push origin $$NEXT_VERSION --no-verify --force-with-lease; \
			echo "Tag $$NEXT_VERSION has been created and pushed"; \
			echo "GitHub Actions will build the release binary automatically"; \
		else \
			echo "[DRY RUN] Showing what would be done..."; \
			echo "Would push to origin/master"; \
			echo "Would create tag: $$NEXT_VERSION"; \
			echo "Would push tag to origin: $$NEXT_VERSION"; \
			echo ""; \
			echo "To execute this release, run:"; \
			echo "  make release type=$(type) dryrun=false"; \
			echo "Dry run complete."; \
		fi \
	else \
		echo "Error: Invalid release type. Use 'patch', 'minor', or 'major'"; \
		exit 1; \
	fi

.PHONY: re-release

# Variables for re-release target
dryrun ?= true
tag ?=

re-release: ## Rerelease target with tag argument. Usage: make re-release tag=<tag> dryrun=false
	@TAG="$(tag)"; \
	if [ -z "$$TAG" ]; then \
		TAG=$$(git describe --tags --abbrev=0); \
	fi; \
	if [ -z "$$TAG" ]; then \
		echo "Error: No tag found near HEAD and no tag specified."; \
		exit 1; \
	fi; \
	echo "Target tag: $$TAG"; \
	if [ "$(dryrun)" = "false" ]; then \
		echo "Deleting GitHub release..."; \
		gh release delete "$$TAG" -y; \
		echo "Deleting local tag..."; \
		git tag -d "$$TAG"; \
		echo "Deleting remote tag..."; \
		git push origin ":refs/tags/$$TAG" --no-verify --force; \
		echo "Recreating tag on HEAD..."; \
		git tag -a "$$TAG" -m "Release $$TAG"; \
		echo "Pushing tag to origin..."; \
		git push origin "$$TAG" --no-verify --force-with-lease; \
		echo "Recreating GitHub release..."; \
		gh release create "$$TAG" --title "$$TAG" --notes "Re-release of $$TAG"; \
		echo "GitHub Actions will build the release binary automatically"; \
		echo "Done!"; \
	else \
		echo "[DRY RUN] Showing what would be done..."; \
		echo "Would delete release: $$TAG"; \
		echo "Would delete local tag: $$TAG"; \
		echo "Would delete remote tag: $$TAG"; \
		echo "Would create new tag at HEAD: $$TAG"; \
		echo "Would push tag to origin: $$TAG"; \
		echo "Would create new release for: $$TAG"; \
		echo ""; \
		echo "To execute this re-release, run:"; \
		if [ -n "$(tag)" ]; then \
			echo "  make re-release tag=$$TAG dryrun=false"; \
		else \
			echo "  make re-release dryrun=false"; \
		fi; \
		echo "Dry run complete."; \
	fi

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
