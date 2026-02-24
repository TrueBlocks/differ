BINARY = differ
INSTALL_DIR = $(HOME)/source

.PHONY: build clean

build:
	go build -o $(INSTALL_DIR)/$(BINARY) .

clean:
	rm -f $(INSTALL_DIR)/$(BINARY)
