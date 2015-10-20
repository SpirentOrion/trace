BUILD_PATH	?= .

.PHONY: all restore build clean

all: build

restore:
	cd $(BUILD_PATH) && godep restore

build:
	(cd $(BUILD_PATH) && \
	 go get -t ./... && \
	 go build -a -v ./... && \
	 go test ./...)

rebase:
	godep update `cat Godeps/Godeps.json | jq -r .Deps[].ImportPath`

clean:
