# ReaderX

ReaderX is a local-first terminal reader for long-form text. It imports TXT and EPUB files into a local SQLite library, provides a terminal reading UI, and keeps reading progress, search data, bookmarks, and notes on your machine.

ReaderX is not a downloader or scraper. It is designed for local files you own or are allowed to read.

## What It Can Do

- Import UTF-8 TXT files
- Import common OPF/spine-based EPUB files
- Parse chapters automatically
- Create fallback pseudo chapters for TXT files without headings
- Read books in a Bubble Tea TUI
- Continue from saved reading progress
- List books and chapters
- Search imported books with a lightweight SQLite ngram index
- Add, list, and remove bookmarks
- Add, list, and remove notes
- Export bookmarks and notes to Markdown
- Configure reader width, theme, and default search limit
- Build release archives through GitHub Actions

## Quick Start

```bash
make build
./readerx --db ./reader.db import ./book.txt
./readerx --db ./reader.db list
./readerx --db ./reader.db chapters 1
./readerx --db ./reader.db read 1
```

Use your own files:

```bash
./readerx import /path/to/book.txt
./readerx import /path/to/book.epub
```

## Install

Build from source:

```bash
make build
```

Download release archives:

<https://github.com/cengZa/readerx/releases>

After downloading:

```bash
tar -xzf readerx_vX.Y.Z_darwin_arm64.tar.gz
cd readerx_vX.Y.Z_darwin_arm64
chmod +x readerx
./readerx --help
```

Optional system-wide install:

```bash
sudo mv readerx /usr/local/bin/readerx
readerx --help
```

## User Manual

Read the full manual here:

[docs/USER_MANUAL.md](docs/USER_MANUAL.md)

中文版用户手册：

[docs/USER_MANUAL.zh-CN.md](docs/USER_MANUAL.zh-CN.md)

The manual covers every command, flag, TUI key, configuration value, and common workflow.

## Development

Run tests:

```bash
make test
```

Run the full validation suite:

```bash
go test ./...
go build ./...
```

Run without building:

```bash
go run . --db ./reader.db import ./book.txt
go run . --db ./reader.db read 1 --plain
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

The CLI layer parses arguments and calls application services. Storage and parsing stay behind the application layer so the TUI does not talk directly to SQLite.

## Current Limitations

- EPUB support covers common OPF/spine/XHTML books, not every EPUB edge case.
- Search uses a lightweight ngram index, not a full tokenizer-based search engine.
- Markdown export is available for bookmarks and notes, but full book export is not implemented yet.
- Online sources and AI reading features are intentionally out of scope for the current version.

## Repository

GitHub: <https://github.com/cengZa/readerx>
