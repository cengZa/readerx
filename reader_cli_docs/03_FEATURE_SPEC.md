# 03. 功能规格说明

本文档描述 ReaderX 当前可用能力。用户级命令以 `readerx` 为前缀。

## 1. Import：导入书籍

### 命令

```bash
readerx import ./book.txt
readerx import ./book.epub
readerx import ./books
readerx import ./book.epub --replace
readerx import-url https://example.com/book.epub
readerx import-url https://example.com/book.txt --title "自定义书名"
```

### 功能要求

- 根据文件后缀选择 TXT 或 EPUB 解析器。
- 支持递归导入目录中的 `.txt` 和 `.epub` 文件。
- 自动提取书名。
- 自动识别章节。
- TXT 无章节时创建伪章节。
- EPUB 按 OPF/spine 顺序读取 XHTML 内容。
- `import-url` 支持明确的公开 HTTP/HTTPS TXT 或 EPUB 直链。
- `import-url` 默认最大下载 100MB，可通过 `--max-bytes` 调整。
- `import-url` 支持 `--replace` 和 `--title`。
- 写入 books、chapters 和搜索索引。
- 相同内容重复导入时返回已有 Book ID。
- `--replace` 替换同内容书籍的章节，并清理该书进度、书签、笔记。
- 导入后输出章节数、总字数、Book ID。
- 对疑似章节编号重复、跳号、回退给出质量提示，不自动修改内容。

### 验收标准

- 能导入 UTF-8 TXT。
- 能导入常见 OPF/spine EPUB。
- 能导入目录。
- 能导入公开 TXT / EPUB 直链。
- 重复导入不会创建重复书籍。
- `--replace` 能重新解析并覆盖旧章节。
- 导入质量提示不会阻断正常导入。

## 2. Source：开放书源搜索

### 命令

```bash
readerx source list
readerx source search <keyword>
readerx source search <keyword> --source gutenberg --limit 5
```

### 功能要求

- 列出当前支持的合法开放书源。
- 通过 Project Gutenberg OPDS 搜索书籍。
- 解析 OPDS acquisition links。
- 搜索结果只展示可导入的 EPUB / TXT URL。
- 用户选择 URL 后，通过 `readerx import-url <url>` 导入。
- 不接入 z-library 类来源，不解析任意网页正文。

### 验收标准

- `readerx source list` 能列出 `gutenberg`。
- `readerx source search <keyword>` 能输出书名、作者、Source ID、EPUB/TXT URL。
- OPDS 解析只保留有 TXT 或 EPUB 获取链接的条目。

## 3. List / Info / Remove：书库管理

### 命令

```bash
readerx list
readerx list --sort recent
readerx list --sort created
readerx list --sort title
readerx list --filter <keyword>
readerx info <book-id>
readerx remove <book-id>
readerx remove <book-id> --yes
```

### 功能要求

- `list` 展示 Book ID、书名、章节数、阅读进度、最近阅读时间。
- `list --sort` 支持最近阅读、导入时间、标题排序。
- `list --filter` 支持按书名筛选。
- `info` 展示来源、文件路径、章节数、总字数、进度、书签数、笔记数。
- `remove` 删除书籍、章节、阅读进度、书签、笔记和搜索索引。
- 交互式删除要求输入 Book ID 确认。
- 非交互删除必须传 `--yes`。

### 验收标准

- 空书库提示用户导入。
- 排序和筛选结果稳定。
- 删除操作不会误删其他书籍。
- 删除后关联数据级联清理。

## 4. Read / Continue：阅读

### 命令

```bash
readerx read <book-id>
readerx read <book-id> --chapter <chapter-no>
readerx read <book-id> --plain
readerx continue
readerx continue --plain
```

### 功能要求

- 默认进入 TUI 阅读模式。
- 非终端或 `--plain` 时输出纯文本。
- 未指定章节时从保存进度继续。
- 阅读后更新最近阅读时间。
- 退出时保存章节、行偏移、字符偏移和全书百分比。

### TUI 快捷键

| 快捷键 | 功能 |
|---|---|
| `j` / `↓` | 向下滚动 |
| `k` / `↑` | 向上滚动 |
| `Space` / `→` | 下一页 |
| `u` / `←` | 上一页 |
| `Home` | 章节开头 |
| `End` | 章节结尾 |
| `n` | 下一章 |
| `p` | 上一章 |
| `g` | 跳转章节 |
| `/` | 当前书搜索 |
| `b` | 添加书签 |
| `m` | 添加笔记 |
| `s` | 保存进度 |
| `?` | 显示或关闭帮助 |
| `q` | 退出并保存 |

### 验收标准

- 中文宽度显示正确。
- 终端尺寸变化后重新排版。
- 最后一页显示为完整最后一屏。
- 窄屏状态栏保留关键信息。
- 退出后进度可恢复。

## 5. Search：搜索

### 命令

```bash
readerx search <keyword>
readerx search <keyword> --book <book-id>
readerx search <keyword> --limit 10
```

### 功能要求

- 使用 SQLite ngram 索引辅助搜索。
- 结果再进行精确关键词校验。
- 单字搜索 fallback 到 SQLite `LIKE`。
- 支持全库搜索和单书搜索。
- 输出书籍、章节和上下文摘要。
- TUI 内 `/` 支持搜索当前书并跳转结果。

### 验收标准

- 中文关键词能搜索。
- 结果包含 book_id、book title、chapter_no、chapter title、snippet。
- TUI 搜索结果展示章节标题和清理后的单行上下文。

## 6. Bookmark：书签

### 命令

```bash
readerx bookmark add <book-id>
readerx bookmark add <book-id> --note "important"
readerx bookmark list
readerx bookmark list --book <book-id>
readerx bookmark remove <bookmark-id>
```

### 功能要求

- 保存书籍、章节、行偏移、字符偏移、摘录和可选备注。
- 支持按书过滤。
- TUI 中按 `b` 添加当前位置书签。
- 支持导出为 Markdown。

## 7. Note：笔记

### 命令

```bash
readerx note add <book-id> --content "review this section"
readerx note list
readerx note list --book <book-id>
readerx note remove <note-id>
```

### 功能要求

- 保存书籍、章节、行偏移、字符偏移和笔记内容。
- 支持按书过滤。
- TUI 中按 `m` 添加当前位置笔记。
- 支持导出为 Markdown。

## 8. Export：导出

### 命令

```bash
readerx export bookmarks
readerx export bookmarks --book 1 -o bookmarks.md
readerx export notes
readerx export notes --book 1 -o notes.md
```

### 功能要求

- 支持书签和笔记导出 Markdown。
- 不传 `-o` 时输出到 stdout。
- 传 `--book` 时只导出单本书。

## 9. Config：配置

### 命令

```bash
readerx config list
readerx config get reader.width
readerx config set reader.width 100
readerx config set reader.theme dark
readerx config set search.limit 25
```

### 支持配置

```text
reader.width   正整数
reader.theme   default, dark, light
search.limit   正整数
```

## 10. Version：版本

### 命令

```bash
readerx version
```

### 功能要求

- 源码构建默认显示 `dev` 和 `unknown`。
- Release 包显示 tag、commit 和构建时间。

## 11. 清理与维护

### 命令

```bash
make clean
make clean-user-data
```

### 功能要求

- `make clean` 只清理项目内构建和测试文件。
- `make clean-user-data` 删除默认用户数据库 `~/.readerx/reader.db` 及 SQLite 伴随文件。
