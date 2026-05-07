# ReaderX

Chinese version: [README.zh-CN.md](README.zh-CN.md)

ReaderX is a local-first terminal reader for long-form text. It imports TXT and EPUB files into a local SQLite library, provides a terminal reading UI, and keeps reading progress, search data, bookmarks, and notes on your machine.

ReaderX is not a downloader or scraper. It is designed for local files you own or are allowed to read.

## What It Can Do

- Import UTF-8 TXT files
- Import common OPF/spine-based EPUB files
- Warn about likely chapter numbering problems after import
- Parse chapters automatically
- Create fallback pseudo chapters for TXT files without headings
- Read books in a Bubble Tea TUI
- Continue from saved reading progress
- List books, inspect book details, remove books, and list chapters
- Search imported books with a lightweight SQLite ngram index
- Add, list, and remove bookmarks
- Add, list, and remove notes
- Export bookmarks and notes to Markdown
- Configure reader width, theme, and default search limit
- Show installed version and release build metadata
- Build release archives through GitHub Actions

## Quick Start

```bash
make install
readerx import ./book.txt
readerx list
readerx chapters 1
readerx read 1
```

Use your own files:

```bash
readerx import /path/to/book.txt
readerx import /path/to/book.epub
```

By default ReaderX stores its library at `~/.readerx/reader.db`; you do not need to pass `--db` for normal use.

## Install

Install from source into your user binary directory:

```bash
make install
```

By default this installs:

```text
~/.local/bin/readerx
```

Make sure `~/.local/bin` is in your `PATH`. For zsh:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
readerx --help
```

Install somewhere else:

```bash
make install INSTALL_DIR=/usr/local/bin
```

For local development, `make build` creates a project-local `readerx` binary.

Download release archives:

<https://github.com/cengZa/readerx/releases>

After downloading:

```bash
tar -xzf readerx_vX.Y.Z_darwin_arm64.tar.gz
cd readerx_vX.Y.Z_darwin_arm64
chmod +x readerx
```

Install the downloaded binary globally for your user:

```bash
mkdir -p ~/.local/bin
cp readerx ~/.local/bin/readerx.tmp
xattr -c ~/.local/bin/readerx.tmp 2>/dev/null || true
chmod +x ~/.local/bin/readerx.tmp
mv -f ~/.local/bin/readerx.tmp ~/.local/bin/readerx
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

Clean project-local build and test artifacts:

```bash
make clean
```

This does not remove your default user library at `~/.readerx/reader.db`.

To remove the default user library, run:

```bash
make clean-user-data
```

This deletes `~/.readerx/reader.db` and its SQLite sidecar files. Imported books, reading progress, bookmarks, notes, and search indexes in the default library will be removed.

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
