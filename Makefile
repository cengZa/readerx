.PHONY: build test clean

build:
	go build -o readerx .

test:
	go test ./...

clean:
	rm -f readerx reader.db reader.db-shm reader.db-wal
