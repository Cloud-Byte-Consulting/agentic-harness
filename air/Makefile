BINARY     := air
VERSION    ?= dev
LDFLAGS    := -s -w -X github.com/cloud-byte/air/cmd.Version=$(VERSION)
PLATFORMS  := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
DISTBIN    := dist/bin

.PHONY: tidy test e2e ginkgo vet build cross clean

tidy:
	go mod tidy

vet:
	go vet ./...

# Run the Ginkgo specs via `go test` (works in CI, no extra binary).
test:
	go test ./...

e2e:
	go test -tags e2e ./e2e/...

# Run the Ginkgo specs via the Ginkgo CLI (parallel + randomized + richer output).
# e2e is build-tagged, so it has its own target (`make e2e`); skip it here.
ginkgo:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all -p --skip-package=e2e

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

cross:
	@mkdir -p $(DISTBIN)
	@for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; ext=; [ $$os = windows ] && ext=.exe; \
	  out=$(DISTBIN)/$(BINARY)_$${os}_$${arch}$$ext; \
	  echo "build $$os/$$arch -> $$out"; \
	  GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $$out . || exit 1; \
	done
	@cd $(DISTBIN) && { sha256sum * > SHA256SUMS 2>/dev/null || shasum -a 256 * > SHA256SUMS; } && echo "wrote $(DISTBIN)/SHA256SUMS"

clean:
	rm -rf dist $(BINARY)
