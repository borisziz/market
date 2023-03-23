CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.51.1
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
PACKAGE=route256/loms/cmd/app

all: format build test lint

build: bindir
	go build -race -o ${BINDIR}/app ${PACKAGE}

test:
	go test -cover ./...

run:
	go run ${PACKAGE}

lint: install-lint
	${LINTBIN} run

precommit: format build test lint
	echo "OK"

bindir:
	mkdir -p ${BINDIR}

format: install-smartimports
	${SMARTIMPORTS} -exclude internal/mocks

install-lint: bindir
	test -f ${LINTBIN} || \
		(GOBIN=${BINDIR} go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINTVER} && \
		mv ${BINDIR}/golangci-lint ${LINTBIN})

install-smartimports: bindir
	test -f ${SMARTIMPORTS} || \
		(GOBIN=${BINDIR} go install github.com/pav5000/smartimports/cmd/smartimports@latest && \
		mv ${BINDIR}/smartimports ${SMARTIMPORTS})

generate:
	mkdir -p pkg/loms/v1
	protoc -I api/loms/v1 -I ../vendor-proto \
	--go_out=pkg/loms/v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/loms/v1 --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pkg/loms/v1 \
    --grpc-gateway_opt=logtostderr=true --grpc-gateway_opt=paths=source_relative \
	api/loms/v1/service.proto --validate_out lang=go:pkg/loms/v1 \

	mv pkg/loms/v1/route256/loms/pkg/loms_v1/* pkg/loms/v1
	rm -r  ./pkg/loms/v1/route256