BIN_DIR = $(CURDIR)

all: build

build: FORCE
	CGO_ENABLED=0 GOBIN=$(BIN_DIR) go install ./...

check: vet

vet:
	go vet ./...

test:
	go test -race -count 1 ./...

clean:
	$(RM) $(wildcard $(BIN_DIR)/*)

FORCE:

.PHONY: all build check vet test
