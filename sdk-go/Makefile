.PHONY: tools vet lint format test test-race coverage gen ci smoke-local

GO ?= go
GOFLAGS ?= -trimpath
PKGS ?= ./...

OAPI_CODEGEN_VERSION := v2.5.0

tools:
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)
	$(GO) install github.com/vladopajic/go-test-coverage/v2@latest

vet:
	$(GO) vet $(PKGS)

lint:
	golangci-lint run

format:
	gofumpt -l -w .

test:
	$(GO) test $(GOFLAGS) $(PKGS)

test-race:
	$(GO) test -race $(GOFLAGS) $(PKGS)

coverage:
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic $(PKGS)
	go-test-coverage --config=.testcoverage.yml

# Regenerate the typed OpenAPI bindings under generated/. Mirrors the Node
# SDK's `yarn generate:types` task; both consume the same canonical spec at
# ../openapi/spec.yaml. Only request types are generated (models mode); the
# response shapes are hand-curated under resources/<resource>/types.go for
# ergonomic naming. Re-run after every spec.yaml update.
gen:
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) \
		--config=generated/oapi-codegen.yaml ../openapi/spec.yaml

ci: vet lint test-race coverage

# Run the live smoke test against the host configured in ./.env (gitignored).
# Mirrors `yarn smoke:local` (sdk-node) and `make`-less `scripts/smoke-local.sh`
# (sdk-python): loads variables from .env, then runs the smoke harness. The CI
# path stays `go run ./scripts/livesmoke` (see .github/workflows/live-smoke.yml).
smoke-local:
	set -a && . ./.env && set +a && $(GO) run ./scripts/livesmoke
