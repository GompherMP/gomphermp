BINARY_NAME=gompher
MAIN_PATH=./cmd/gompher/main.go

all: build test

deps:
	go mod tidy
	go mod download

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

test:
	go test -v ./...

install: build
	install -m 0755 $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)

benchmark: build
	bash benchmarks/run.sh ./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean
