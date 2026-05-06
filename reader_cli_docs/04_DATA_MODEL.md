# 04. 数据模型设计

## 1. 设计目标

本地存储采用 SQLite。目标：

- 支持书籍、章节、进度、书签、笔记
- 支持全文搜索
- 支持未来在线源 Lazy Loading
- 支持数据迁移
- 支持离线阅读

## 2. 实体关系

```text
Book 1 ── N Chapter
Book 1 ── 1 Progress
Book 1 ── N Bookmark
Book 1 ── N Note
Source 1 ── N Book
```

## 3. books 表

```sql
CREATE TABLE books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type TEXT NOT NULL DEFAULT 'local',
    source_book_id TEXT DEFAULT '',
    title TEXT NOT NULL,
    author TEXT DEFAULT '',
    description TEXT DEFAULT '',
    cover_url TEXT DEFAULT '',
    file_path TEXT DEFAULT '',
    content_hash TEXT DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_read_at INTEGER DEFAULT 0
);
```

### 字段说明

| 字段 | 说明 |
|---|---|
| `source_type` | local_txt / local_epub / markdown / rss / online |
| `source_book_id` | 外部数据源中的 book id |
| `content_hash` | 用于重复导入检测 |
| `last_read_at` | 最近阅读时间 |

---

## 4. chapters 表

```sql
CREATE TABLE chapters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL,
    chapter_no INTEGER NOT NULL,
    source_chapter_id TEXT DEFAULT '',
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    content_status TEXT NOT NULL DEFAULT 'cached',
    content_hash TEXT DEFAULT '',
    word_count INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    UNIQUE(book_id, chapter_no),
    FOREIGN KEY(book_id) REFERENCES books(id)
);
```

### content_status

| 状态 | 说明 |
|---|---|
| `empty` | 只有目录，没有正文 |
| `cached` | 正文已缓存 |
| `failed` | 拉取失败 |
| `partial` | 部分缓存 |

这个字段为未来 Lazy Loading 设计。

---

## 5. reading_progress 表

```sql
CREATE TABLE reading_progress (
    book_id INTEGER PRIMARY KEY,
    chapter_no INTEGER NOT NULL,
    line_offset INTEGER NOT NULL DEFAULT 0,
    char_offset INTEGER NOT NULL DEFAULT 0,
    percentage REAL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY(book_id) REFERENCES books(id)
);
```

### 为什么同时保留 line_offset 和 char_offset？

- `line_offset` 适合 TUI 分页
- `char_offset` 更适合文本定位
- 终端宽度变化时，line_offset 可能失真，char_offset 可用于重算

---

## 6. bookmarks 表

```sql
CREATE TABLE bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL,
    chapter_no INTEGER NOT NULL,
    line_offset INTEGER DEFAULT 0,
    char_offset INTEGER DEFAULT 0,
    excerpt TEXT DEFAULT '',
    note TEXT DEFAULT '',
    created_at INTEGER NOT NULL,
    FOREIGN KEY(book_id) REFERENCES books(id)
);
```

---

## 7. notes 表

```sql
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL,
    chapter_no INTEGER NOT NULL,
    line_offset INTEGER DEFAULT 0,
    char_offset INTEGER DEFAULT 0,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY(book_id) REFERENCES books(id)
);
```

---

## 8. sources 表

```sql
CREATE TABLE sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config_json TEXT DEFAULT '{}',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
```

---

## 9. settings 表

```sql
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER NOT NULL
);
```

---

## 10. 全文搜索 FTS5

```sql
CREATE VIRTUAL TABLE chapters_fts USING fts5(
    title,
    content,
    content='chapters',
    content_rowid='id'
);
```

### 同步策略

第一阶段可以在插入章节时手动写入 FTS：

```sql
INSERT INTO chapters_fts(rowid, title, content)
VALUES (?, ?, ?);
```

后续可以用 trigger 自动同步。

---

## 11. 推荐索引

```sql
CREATE INDEX idx_books_last_read_at ON books(last_read_at);
CREATE INDEX idx_chapters_book_no ON chapters(book_id, chapter_no);
CREATE INDEX idx_bookmarks_book ON bookmarks(book_id);
CREATE INDEX idx_notes_book ON notes(book_id);
```

---

## 12. 数据迁移

建议从第一版开始保留 migrations：

```text
migrations/
  001_init.sql
  002_add_notes.sql
  003_add_sources.sql
```

不要把 schema 写死在业务代码里长期维护。
