PACKAGE_NAME 	= trace

.PHONY: all setup restore build clean

all: build

#
# Setup
#

ifdef JOB_URL

# SpirentOrion Jenkins build environment setup
export GOROOT	= /usr/src/go
export GOPATH	= /go
export PATH	:= /go/bin:/usr/src/go/bin:$(PATH)
ORG_PATH	= github.com/SpirentOrion
REPO_PATH	= $(ORG_PATH)/$(PACKAGE_NAME)
BUILD_PATH	= $(GOPATH)/src/$(REPO_PATH)

setup:
	@mkdir -p $(GOPATH)/src/$(ORG_PATH)
	@rm -f $(GOPATH)/src/$(REPO_PATH)
	@ln -s $(PWD) $(GOPATH)/src/$(REPO_PATH)
	@git config --global url."git@github.com:SpirentOrion".insteadOf https://github.com/SpirentOrion

else

# Local build setup
BUILD_PATH	= .

setup:

endif

restore: setup
	cd $(BUILD_PATH) && godep restore

build: setup
	(cd $(BUILD_PATH) && \
	 go get -t ./... && \
	 go build -a -v ./... && \
	 go test ./...)

rebase:
	godep update `cat Godeps/Godeps.json | jq -r .Deps[].ImportPath`

clean:
