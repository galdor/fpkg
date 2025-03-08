BUILD_ID = $(shell git describe --tags HEAD)
ifndef BUILD_ID
$(error Cannot identify build id from Git repository data)
endif

BIN_DIR = $(CURDIR)

all: build

build: FORCE
	CGO_ENABLED=0 \
	go build -o $(BIN_DIR) -ldflags="-X 'main.buildId=$(BUILD_ID)'" .

check: vet

vet:
	go vet ./...

test:
	go test -race -count 1 ./...

clean:
	$(RM) $(wildcard $(BIN_DIR)/*)

FORCE:

.PHONY: all build check vet test
