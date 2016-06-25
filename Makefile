BUILD_PATH ?= .

.PHONY: all build test clean

all: build test

build:
	cd $(BUILD_PATH) && go build .

test:
	cd $(BUILD_PATH) && go test -race .

clean:
	go clean -i ./...
