# ReaderX

ReaderX is a local-first terminal reader. The first runnable version supports importing UTF-8 TXT files, listing the local bookshelf, and printing chapters from SQLite.

## Run

```bash
make build
./readerx import ./book.txt
./readerx import ./book.epub
./readerx list
./readerx chapters 1
./readerx read 1 --chapter 1
./readerx continue
./readerx search "keyword"
./readerx bookmark add 1
./readerx bookmark list
./readerx note add 1 --content "review this section"
./readerx note list
```

During development you can also use `go run` directly:

```bash
go run . import ./book.txt
go run . import ./book.epub
go run . list
go run . chapters 1
go run . read 1 --chapter 1
go run . continue
go run . search "keyword"
go run . bookmark add 1
go run . bookmark list
go run . note add 1 --content "review this section"
go run . note list
```

By default the SQLite database is stored at `~/.readerx/reader.db`. Use `--db <path>` to choose another database:

```bash
go run . --db ./reader.db import ./book.txt
```

`read` enters the TUI reader when stdin/stdout are terminals. Use `--plain` for script-friendly output:

```bash
go run . read 1 --chapter 1 --plain
```

## MVP Scope

- TXT import
- EPUB import for common spine-based books
- Chapter parsing for common Chinese and English headings
- Fallback pseudo chapters when no headings are found
- SQLite persistence
- CLI bookshelf and chapter reading
- Basic TUI reading with `j/k`, `Space`, `u`, `n/p`, `b`, `s`, and `q`
- Bookmarks from CLI and TUI
- Notes from CLI
- Basic Chinese-friendly LIKE search

Online sources, note export, and AI features are intentionally left for later phases.
