# ReaderX User Manual

Chinese version: [USER_MANUAL.zh-CN.md](USER_MANUAL.zh-CN.md)

This manual explains how to install, configure, and use every current ReaderX command.

ReaderX is a local-first terminal reader. It stores all imported content and reading metadata in a local SQLite database.

## 1. Concepts

### Book

A book is an imported TXT or EPUB file. Each book receives a numeric `Book ID`.

### Chapter

ReaderX splits imported files into chapters.

- TXT chapters are detected from common headings such as `第一章`, `第 1 章`, `Chapter 1`, and `卷一`.
- TXT files without chapter headings are split into fallback pseudo chapters.
- EPUB chapters are read from the EPUB spine order.

### Progress

ReaderX saves the last read chapter and location for each book.

### Bookmarks

A bookmark stores a book, chapter, location, excerpt, and optional short note.

### Notes

A note stores your own text at the current reading location.

### Database

Default database:

```text
~/.readerx/reader.db
```

This is the normal database for daily use. You do not need to pass `--db` unless you want a temporary or project-local test database.

Use another database with the global `--db` flag:

```bash
readerx --db ./reader.db list
```

Important: `--db` must appear before the subcommand.

Correct:

```bash
readerx --db ./reader.db read 1
```

Incorrect:

```bash
readerx read 1 --db ./reader.db
```

To delete the default user database and start over:

```bash
make clean-user-data
```

This removes the default library at `~/.readerx/reader.db`, including imported books, progress, bookmarks, notes, and search indexes. `make clean` only removes project-local build and test files.

## 2. Installation

### Install From Source

Requirements:

- Go 1.26 or newer
- macOS or Linux terminal

Recommended install:

```bash
make install
```

By default this installs:

```text
~/.local/bin/readerx
```

Check:

```bash
readerx --help
```

Make sure `~/.local/bin` is in your `PATH`. For zsh:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
readerx --help
```

You can also install to another directory:

```bash
make install INSTALL_DIR=/usr/local/bin
```

For local development, `make build` creates a project-local `readerx` binary.

### Install From Release

Download a release archive from:

<https://github.com/cengZa/readerx/releases>

Extract:

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

## 3. Quick Start

Use the included sample book:

```bash
make install
readerx import ./book.txt
readerx list
readerx chapters 1
readerx read 1
```

Use your own book:

```bash
readerx import /path/to/book.txt
readerx import /path/to/book.epub
readerx list
readerx read 1
```

## 4. Global Command Format

```bash
readerx [global flags] <command> [command flags]
```

Global flags:

```text
--db string   SQLite database path
```

Example:

```bash
readerx --db ./reader.db search "剑气"
```

## 5. Commands

### 5.1 Help

Show root help:

```bash
readerx --help
```

Show command help:

```bash
readerx read --help
readerx bookmark --help
readerx config --help
```

### 5.2 Import

Import a local TXT or EPUB file.

```bash
readerx import <file>
readerx import <file> --replace
```

Examples:

```bash
readerx import ./book.txt
readerx import ./book.epub
readerx import ./book.epub --replace
```

Output example:

```text
导入成功：
书名：book
章节数：3
总字数：318
Book ID：1
```

Notes:

- TXT must be UTF-8.
- EPUB support targets common OPF/spine/XHTML books.
- Duplicate imports of the same content return the existing Book ID.
- Use `--replace` to reparse an already imported file and replace its stored chapters. This resets progress, bookmarks, and notes for that book.
- After import, ReaderX may print quality warnings for likely chapter numbering problems, such as duplicate chapter numbers, backwards numbering, or obvious skipped chapter numbers. These warnings do not modify imported content.

### 5.3 List Books

List imported books.

```bash
readerx list
```

Output example:

```text
ID  Title  Chapters  Progress  Last Read
1   book   3         33%       2026-05-06
```

### 5.4 List Chapters

List chapter metadata for a book.

```bash
readerx chapters <book-id>
```

Example:

```bash
readerx chapters 1
```

Output example:

```text
No  Title       Words
1   第一章 山雨欲来  93
2   第二章 灯下旧约  120
3   第三章 白鹿渡口  105
```

### 5.5 Read

Open a book in the TUI reader.

```bash
readerx read <book-id>
```

Open a specific chapter:

```bash
readerx read <book-id> --chapter <chapter-no>
```

Print plain text instead of opening the TUI:

```bash
readerx read <book-id> --chapter <chapter-no> --plain
```

Examples:

```bash
readerx read 1
readerx read 1 --chapter 2
readerx read 1 --chapter 2 --plain
```

Flags:

```text
--chapter int   chapter number
--plain         print chapter without entering TUI
```

### 5.6 Continue

Continue the most recently read book.

```bash
readerx continue
```

Plain output:

```bash
readerx continue --plain
```

Flags:

```text
--plain   print chapter without entering TUI
```

### 5.7 Search

Search imported books.

```bash
readerx search <keyword>
```

Examples:

```bash
readerx search "剑气"
readerx search "剑气" --book 1
readerx search "剑气" --limit 10
```

Flags:

```text
--book int    limit search to one book ID
--limit int   maximum number of results
```

Behavior:

- Search uses a lightweight SQLite ngram index.
- Results are verified with exact keyword matching.
- Very short one-character searches fall back to SQLite `LIKE`.
- If `--limit` is not set, ReaderX uses `search.limit` from config.

Output example:

```text
[Book 1] book / 第 1 章 第一章 山雨欲来
...一道剑气自山巅而起...
```

### 5.8 Bookmarks

#### Add Bookmark

Add a bookmark at the saved reading position of a book.

```bash
readerx bookmark add <book-id>
readerx bookmark add <book-id> --note "important moment"
```

Flags:

```text
--note string   bookmark note
```

You can also add a bookmark inside the TUI by pressing `b`.

#### List Bookmarks

```bash
readerx bookmark list
readerx bookmark list --book <book-id>
```

Flags:

```text
--book int   limit bookmarks to one book ID
```

#### Remove Bookmark

```bash
readerx bookmark remove <bookmark-id>
```

### 5.9 Notes

#### Add Note

Add a note at the saved reading position of a book.

```bash
readerx note add <book-id> --content "review this section"
```

Flags:

```text
--content string   note content
```

You can also add a note inside the TUI by pressing `m`.

#### List Notes

```bash
readerx note list
readerx note list --book <book-id>
```

Flags:

```text
--book int   limit notes to one book ID
```

#### Remove Note

```bash
readerx note remove <note-id>
```

### 5.10 Export

Export bookmarks or notes as Markdown.

#### Export Bookmarks

Print to stdout:

```bash
readerx export bookmarks
readerx export bookmarks --book 1
```

Write to file:

```bash
readerx export bookmarks --book 1 -o bookmarks.md
```

#### Export Notes

Print to stdout:

```bash
readerx export notes
readerx export notes --book 1
```

Write to file:

```bash
readerx export notes --book 1 -o notes.md
```

Flags:

```text
--book int            limit export to one book ID
-o, --output string   write Markdown to file
```

### 5.11 Config

List config:

```bash
readerx config list
```

Get one value:

```bash
readerx config get search.limit
```

Set one value:

```bash
readerx config set search.limit 25
readerx config set reader.width 100
readerx config set reader.theme dark
```

Supported keys:

```text
reader.width   positive integer
reader.theme   default, dark, light
search.limit   positive integer
```

Effects:

- `reader.width` controls the TUI reader maximum text width.
- `reader.theme` controls the TUI reader theme.
- `search.limit` controls default search result count when `search --limit` is not provided.

## 6. TUI Reader

Open:

```bash
readerx read 1
```

Keys:

```text
j / down   scroll down
k / up     scroll up
Space      next page
u          previous page
n          next chapter
p          previous chapter
g          jump to chapter
/          search current book
b          add bookmark
m          add note
s          save progress
q          quit and save
```

### Jump to Chapter

Press `g`, type a chapter number, then press Enter.

Cancel with Esc.

### Search Current Book

Press `/`, type a keyword, then press Enter.

When results appear:

```text
j / down   select next result
k / up     select previous result
Enter      jump to selected result
Esc        cancel
```

### Add Bookmark

Press `b`.

ReaderX saves the current book, chapter, line offset, character offset, and an excerpt.

### Add Note

Press `m`, type the note text, then press Enter.

Cancel with Esc.

## 7. Common Workflows

### Read a New TXT Book

```bash
readerx import ./book.txt
readerx list
readerx read 1
```

### Read an EPUB

```bash
readerx import ./book.epub
readerx list
readerx chapters 1
readerx read 1
```

### Use a Project-Local Test Database

```bash
readerx --db ./reader.db import ./book.txt
readerx --db ./reader.db list
readerx --db ./reader.db read 1
```

### Search and Jump Manually

```bash
readerx search "白鹿" --book 1
readerx read 1 --chapter 3
```

### Export Reading Notes

```bash
readerx export bookmarks --book 1 -o bookmarks.md
readerx export notes --book 1 -o notes.md
```

## 8. Development Commands

Build:

```bash
make build
```

Test:

```bash
make test
```

Clean:

```bash
make clean
```

This only removes project-local build and test artifacts such as `./readerx` and `./reader.db`. It does not remove your default user library at `~/.readerx/reader.db`.

Run without building:

```bash
go run . --db ./reader.db import ./book.txt
go run . --db ./reader.db read 1 --plain
```

Full validation:

```bash
go test ./...
go build ./...
```

## 9. Current Limitations

- EPUB support covers common OPF/spine/XHTML books, not every EPUB edge case.
- Search is an ngram-assisted exact keyword search, not a full language-aware tokenizer.
- Markdown export currently covers bookmarks and notes, not full books.
- Online sources and AI reading features are intentionally out of scope for the current version.

## 10. Safety and Data Ownership

ReaderX is local-first:

- Imported content is stored in your local SQLite database.
- ReaderX does not upload books.
- ReaderX does not include platform scraping or downloader logic.
- Use it with files you own or are allowed to read.
