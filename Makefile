main_package_path = ./cmd/gonc
binary_name = gonc


## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	go fmt ./...

## build: build the application
.PHONY: build
build:
	go build -o=/tmp/bin/${binary_name} ${main_package_path}

## test: run all tests
.PHONY: test
test:
	go test -v ./...

