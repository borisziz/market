CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.51.1
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
PACKAGE=route256/checkout/cmd/app

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
	${SMARTIMPORTS} -exclude internal/domain/zzz*

install-lint: bindir
	test -f ${LINTBIN} || \
		(GOBIN=${BINDIR} go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINTVER} && \
		mv ${BINDIR}/golangci-lint ${LINTBIN})

install-smartimports: bindir
	test -f ${SMARTIMPORTS} || \
		(GOBIN=${BINDIR} go install github.com/pav5000/smartimports/cmd/smartimports@latest && \
		mv ${BINDIR}/smartimports ${SMARTIMPORTS})

generate:
	mkdir -p pkg/products
	protoc -I api/products/ \
		--go_out=pkg/products --go_opt=paths=source_relative \
    	--go-grpc_out=pkg/products --go-grpc_opt=paths=source_relative \
    	api/products/product-service.proto

	mkdir -p pkg/checkout/v1
	protoc -I api/checkout/v1 -I ../vendor-proto \
	--go_out=pkg/checkout/v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/checkout/v1 --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pkg/checkout/v1 \
    --grpc-gateway_opt=logtostderr=true --grpc-gateway_opt=paths=source_relative \
	api/checkout/v1/domain.proto --validate_out lang=go:pkg/checkout/v1 \

	mv pkg/checkout/v1/route256/checkout/pkg/checkout_v1/* pkg/checkout/v1
	rm -r  ./pkg/checkout/v1/route256