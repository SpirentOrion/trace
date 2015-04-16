PACKAGE_NAME 	= trace

.PHONY: all setup build clean

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

setup:
	@mkdir -p $(GOPATH)/src/$(ORG_PATH)
	@rm -f $(GOPATH)/src/$(REPO_PATH)
	@ln -s $(PWD) $(GOPATH)/src/$(REPO_PATH)
	@git config --global url."git@github.com:SpirentOrion".insteadOf https://github.com/SpirentOrion

else

# Local build setup
REPO_PATH       = .

setup:

endif

#
# Build
#

build: setup
	go get -t $(REPO_PATH)/...
	go build -a -v $(REPO_PATH)/...
	go test $(REPO_PATH)/...

clean:
