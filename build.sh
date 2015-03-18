#/bin/sh -e

if [ -n "$JOB_URL" ]; then
    # SpirentOrion Jenkins build environment setup
    if [ -z "$GOROOT" -a -d /usr/src/go ]; then
        export GOROOT=/usr/src/go
        PATH=/usr/src/go/bin:$PATH
    fi

    if [ -z "$GOPATH" -a -d /go ]; then
        export GOPATH=/go
        PATH=/go/bin:$PATH
    fi

    ORG_PATH=github.com/SpirentOrion
    REPO_PATH=${ORG_PATH}/trace

    mkdir -p ${GOPATH}/src/${ORG_PATH}
    rm -f ${GOPATH}/src/${REPO_PATH}
    ln -s ${PWD} ${GOPATH}/src/${REPO_PATH}
else
    # Local build
    REPO_PATH=.
fi

go get -t ${REPO_PATH}
go build -a -v ${REPO_PATH}
go build -a -v -o example/example ${REPO_PATH}/example
go test ${REPO_PATH}
