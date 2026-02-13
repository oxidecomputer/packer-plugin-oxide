PLUGIN_BINARY=packer-plugin-oxide
PLUGIN_GO_MODULE=$(shell go list -m)

PACKER_PLUGIN_SDK_VERSION=$(shell go list -f '{{ .Version }}' -m github.com/hashicorp/packer-plugin-sdk)

TEST_COUNT?=1
TEST_PACKAGES?=$(shell go list ./...)

OXIDE_PROJECT ?= packer-acc-test
OXIDE_BOOT_DISK_IMAGE_NAME ?= noble

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
plugin-check: install-packer-sdc build
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

.PHONY: fmt
fmt:
	@ go tool -modfile=tools/go.mod golangci-lint fmt

.PHONY: lint
lint:
	@ go tool -modfile=tools/go.mod golangci-lint run

.PHONY: testacc
testacc: dev
	PACKER_ACC=1 OXIDE_PROJECT=$(OXIDE_PROJECT) OXIDE_BOOT_DISK_IMAGE_NAME=$(OXIDE_BOOT_DISK_IMAGE_NAME) go test -count $(TEST_COUNT) -v $(TEST_PACKAGES) -timeout=120m
