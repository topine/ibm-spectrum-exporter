TAG?=$(shell git tag|tail -1)
COMMIT=$(shell git rev-parse HEAD)
VERSION?=$(TAG)
DATE=$(shell date +%Y-%m-%d/%H:%M:%S )
BUILDINFOPKG=main
LDFLAGS= -ldflags "-w -X ${BUILDINFOPKG}.TAG=${TAG} -X ${BUILDINFOPKG}.COMMIT=${COMMIT} -X ${BUILDINFOPKG}.VERSION=${VERSION} -X ${BUILDINFOPKG}.BUILDTIME=${DATE} -s"

all: build

build:
	CGO_ENABLED=0 go build  -installsuffix cgo ${LDFLAGS} -o bin/ibm-spectrum-exporter ./main.go

buildlinux:
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo ${LDFLAGS} -o bin/ibm-spectrum-exporter ./main.go

container: buildlinux
	docker build -t topine/ibm-spectrum-exporter:$(TAG) .

.PHONY: test
test:
	go test -coverprofile=coverage.out -covermode=count ./...

.PHONY: coverage-html
coverage-html: test
	go tool cover -html=coverage-all.out

.PHONY: coverage
coverage:
	go tool cover -html=cover.out -o coverage.html

.PHONY: container
push: container
	docker push topine/ibm-spectrum-exporter:$(TAG)

.PHONY: clean
clean:
	rm -f ibm-spectrum-exporter

# Run all the linters
.PHONY: lint
lint:
	golangci-lint run ./... --print-issued-lines=false

.PHONY: check
check: fmt lint
