# ReaderX

ReaderX is a local-first terminal reader for long-form text. It imports books into a local SQLite library, lets you read from the terminal, and keeps progress, bookmarks, notes, and search data on your machine.

The project is intentionally not a downloader or platform scraper. It is designed for local files you own or are allowed to read.

## Features

- Import UTF-8 TXT files
- Import common spine-based EPUB files
- Parse chapters automatically
- Create fallback pseudo chapters for TXT files without headings
- Store books, chapters, progress, bookmarks, and notes in SQLite
- List books and chapter tables of contents
- Read in a Bubble Tea TUI
- Continue from saved progress
- Search imported content with simple SQLite `LIKE` matching
- Add/list/remove bookmarks
- Add/list/remove notes
- Run tests and builds in GitHub Actions CI

## Requirements

- Go 1.26 or newer
- macOS or Linux terminal

## Build

```bash
make build
```

This creates:

```bash
./readerx
```

Run tests:

```bash
make test
```

Clean local build/database artifacts:

```bash
make clean
```

## Quick Start

Use the included sample book:

```bash
make build
./readerx --db ./reader.db import ./book.txt
./readerx --db ./reader.db list
./readerx --db ./reader.db chapters 1
./readerx --db ./reader.db read 1
```

Use your own file:

```bash
./readerx import /path/to/book.txt
./readerx import /path/to/book.epub
```

## Database

By default, ReaderX stores data at:

```text
~/.readerx/reader.db
```

Use `--db` to choose another database:

```bash
./readerx --db ./reader.db list
```

`--db` is a global flag, so place it before the subcommand:

```bash
./readerx --db ./reader.db read 1
```

## Commands

### Import

```bash
./readerx import ./book.txt
./readerx import ./book.epub
```

### Bookshelf

```bash
./readerx list
```

### Chapters

```bash
./readerx chapters 1
```

### Read

Open the TUI reader:

```bash
./readerx read 1
```

Open a specific chapter:

```bash
./readerx read 1 --chapter 2
```

Print plain text instead of opening the TUI:

```bash
./readerx read 1 --chapter 2 --plain
```

Continue the most recently read book:

```bash
./readerx continue
./readerx continue --plain
```

### TUI Keys

```text
j / down   scroll down
k / up     scroll up
Space      next page
u          previous page
n          next chapter
p          previous chapter
b          add bookmark
s          save progress
q          quit and save
```

### Search

Search all books:

```bash
./readerx search "剑气"
```

Search one book:

```bash
./readerx search "剑气" --book 1
```

Current search is intentionally simple. It uses SQLite `LIKE`, which works for basic Chinese keyword search but is not a tokenizer-based full-text engine.

### Bookmarks

```bash
./readerx bookmark add 1 --note "important moment"
./readerx bookmark list
./readerx bookmark list --book 1
./readerx bookmark remove 1
```

You can also add a bookmark in the TUI by pressing `b`.

### Notes

```bash
./readerx note add 1 --content "review this section"
./readerx note list
./readerx note list --book 1
./readerx note remove 1
```

## Development

Run without building:

```bash
go run . --db ./reader.db import ./book.txt
go run . --db ./reader.db list
go run . --db ./reader.db read 1 --plain
```

Run the full validation suite:

```bash
go test ./...
go build ./...
```

## Architecture

```text
cmd/                    CLI command entrypoints
internal/app/           Application services
internal/domain/        Core domain types
internal/storage/       SQLite store
internal/parser/        TXT and EPUB parsers
internal/source/        Local file source dispatch
internal/reader/        Wrapping and pagination
internal/tui/           Terminal reader UI
reader_cli_docs/        Product and architecture docs
```

The CLI layer only parses command arguments and calls application services. Storage and parsing stay behind the application layer so the TUI does not talk directly to SQLite.

## Current Limitations

- EPUB support covers common OPF/spine/XHTML books, not every EPUB edge case.
- Search is basic keyword matching, not high-quality Chinese segmentation.
- Notes can be stored and listed, but export is not implemented yet.
- Online sources and AI reading features are intentionally out of scope for the current version.

## Roadmap

- Improve search with ngram or FTS-backed indexing
- Export notes and bookmarks to Markdown
- Add in-TUI chapter jump and search panels
- Harden EPUB compatibility with more fixtures
- Add GitHub releases and installation packages
- Explore optional AI summaries over local content

## Repository

GitHub: <https://github.com/cengZa/readerx>
