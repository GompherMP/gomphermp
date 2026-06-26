# Variables to facilitate changes
BINARY_NAME=gompher
MAIN_PATH=./cmd/gompher/main.go

# Default command when running 'make'
all: build test

# Install dependencies
deps:
	go mod tidy
	go mod download

# Compile the transpiler
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run unit tests
test:
	go test -v ./...

install: build
	install -m 0755 $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)

# Clean generated binary files
clean:
	rm -f $(BINARY_NAME)
	go clean
