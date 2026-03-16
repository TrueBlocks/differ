BINARY = differ
INSTALL_DIR = $(HOME)/source
MSG ?= update

.PHONY: build test clean add commit push

build:
	go build -o $(INSTALL_DIR)/$(BINARY) .

test:
	go test ./...

clean:
	rm -f $(INSTALL_DIR)/$(BINARY)

add:
	@git add -A

commit: build
	@git add -A
	@git commit -m "$(MSG)" || true

push: build
	@git add -A
	@git commit -m "$(MSG)" || true
	@git push
