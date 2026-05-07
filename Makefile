INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: build install uninstall test clean clean-user-data

build:
	go build -o readerx .

install: build
	mkdir -p "$(INSTALL_DIR)"
	cp readerx "$(INSTALL_DIR)/readerx.tmp"
	xattr -c "$(INSTALL_DIR)/readerx.tmp" 2>/dev/null || true
	chmod +x "$(INSTALL_DIR)/readerx.tmp"
	mv -f "$(INSTALL_DIR)/readerx.tmp" "$(INSTALL_DIR)/readerx"
	@echo "Installed readerx to $(INSTALL_DIR)/readerx"
	@echo "Make sure $(INSTALL_DIR) is in your PATH."

uninstall:
	rm -f "$(INSTALL_DIR)/readerx"
	@echo "Removed $(INSTALL_DIR)/readerx"

test:
	go test ./...

clean:
	rm -f readerx reader.db reader.db-shm reader.db-wal

clean-user-data:
	rm -f "$(HOME)/.readerx/reader.db" "$(HOME)/.readerx/reader.db-shm" "$(HOME)/.readerx/reader.db-wal"
	@echo "Removed ReaderX default user database from $(HOME)/.readerx"
