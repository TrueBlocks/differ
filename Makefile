BINARY = differ
INSTALL_DIR = $(HOME)/source

.PHONY: build test clean

build:
	go build -o $(INSTALL_DIR)/$(BINARY) .

test:
	go test ./...

clean:
	rm -f $(INSTALL_DIR)/$(BINARY)
