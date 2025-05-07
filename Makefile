PLUGIN_BINARY=packer-plugin-oxide
PLUGIN_GO_MODULE=$(shell go list -m)

PACKER_PLUGIN_SDK_VERSION=$(shell go list -f '{{ .Version }}' -m github.com/hashicorp/packer-plugin-sdk)

TEST_COUNT?=1
TEST_PACKAGES?=$(shell go list ./...)

.PHONY: build
build:
	go build -o ${PLUGIN_BINARY}

.PHONY: dev
dev:
	go build -ldflags="-X 'main.VersionPreRelease=dev'" -o ${PLUGIN_BINARY}
	packer plugins install --path ${PLUGIN_BINARY} "$(shell echo "${PLUGIN_GO_MODULE}" | sed 's/packer-plugin-//')"

.PHONY: install-packer-sdc
install-packer-sdc:
	go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${PACKER_PLUGIN_SDK_VERSION}

.PHONY: plugin-check
plugin-check: install-packer-sdc
	packer-sdc plugin-check ${PLUGIN_BINARY}

.PHONY: generate
generate: install-packer-sdc
	go generate ./...
	rm -rf .docs
	packer-sdc renderdocs -src docs -partials docs-partials/ -dst .docs/
	./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "hashicorp"
	rm -r ".docs"

.PHONY: test
test:
	go test -race -count $(TEST_COUNT) $(TEST_PACKAGES) -timeout=3m

.PHONY: testacc
testacc: dev
	PACKER_ACC=1 go test -count $(TEST_COUNT) -v $(TEST_PACKAGES) -timeout=120m
