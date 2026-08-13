# ReaderX 领域模型与数据模型

本文档以 `internal/domain/domain.go` 和 `internal/storage/sqlite_store.go` 为准，区分核心领域对象、应用层 DTO 和 SQLite 持久化表。

## 1. 模型关系

```text
Book 1 --- N Chapter
Book 1 --- 0..1 Progress
Book 1 --- N Bookmark
Book 1 --- N Note
Chapter 1 --- N SearchTerm

SourceSearchResult --用户选择 URL--> ImportService --> Book + Chapters
```

删除 Book 时，SQLite 外键级联删除 Chapter、SearchTerm、Progress、Bookmark 和 Note。

## 2. 核心领域模型

### 2.1 Book

`domain.Book` 表示书库中的一本书。

| 字段 | 含义 | 持久化 |
|---|---|---|
| `ID` | 本地 Book ID | `books.id` |
| `SourceType` | 内容来源类型，目前主要是 `local_txt` / `local_epub` | `books.source_type` |
| `SourceBookID` | 外部来源中的书籍标识，当前导入链路通常为空 | `books.source_book_id` |
| `Title`、`Author`、`Description` | 书籍元数据 | `books` |
| `CoverURL` | 预留封面 URL，当前没有展示链路 | `books.cover_url` |
| `FilePath` | 原始本地文件路径；URL 导入时会记录临时路径，因此不能作为长期来源定位 | `books.file_path` |
| `ContentHash` | 原始文件 SHA-256，用于内容去重 | `books.content_hash` |
| `CreatedAt`、`UpdatedAt`、`LastReadAt` | 生命周期时间戳 | `books` |
| `ChapterCount` | 列表查询时聚合的章节数 | 不落在 `books` |
| `ProgressPercentage` | 列表查询时 JOIN 的进度 | 不落在 `books` |

`ContentHash` 有非空唯一索引，所以同一内容不能创建两本书。`--replace` 保留 Book ID，但替换元数据和章节，并清理旧进度、书签与笔记。

### 2.2 Chapter

`domain.Chapter` 是实际阅读和搜索的正文单元。

| 字段 | 含义 |
|---|---|
| `ID`、`BookID` | 本地章节和所属书籍标识 |
| `ChapterNo` | 书内顺序号，从 1 开始；同一本书内唯一 |
| `SourceChapterID` | EPUB 中对应的 spine 文件路径；TXT 通常为空 |
| `Title`、`Content` | 章节标题和纯文本正文 |
| `ContentStatus` | 当前导入章节统一为 `cached` |
| `ContentHash` | 章节正文 SHA-256 |
| `WordCount` | 当前实现实际统计非空白 rune 数，不是语言学意义的单词数 |
| `CreatedAt`、`UpdatedAt` | 时间戳 |

`empty`、`failed`、`partial` 等懒加载状态尚未进入业务链路，不应视为当前能力。

### 2.3 Progress

`domain.Progress` 表示一本书的唯一阅读位置，`book_id` 同时是主键。

| 字段 | 含义 |
|---|---|
| `BookID` | 所属书籍 |
| `ChapterNo` | 当前章节 |
| `LineOffset` | 当前排版结果中的首行偏移 |
| `CharOffset` | 当前正文中的 rune 位置近似值 |
| `Percentage` | 全书阅读百分比 |
| `UpdatedAt` | 最近保存时间 |

`SaveProgress` 使用 upsert，一本书始终只有一条进度。

### 2.4 Bookmark

`domain.Bookmark` 表示某个阅读位置的快照。

| 字段 | 含义 | 说明 |
|---|---|---|
| `ID`、`BookID` | 标识与所属书籍 | 持久化 |
| `ChapterNo`、`LineOffset`、`CharOffset` | 位置 | 持久化 |
| `Excerpt` | 从当前位置截取的正文 | 持久化 |
| `Note` | 书签附注 | 持久化 |
| `CreatedAt` | 创建时间 | 持久化 |
| `BookTitle`、`ChapterTitle` | 列表/导出展示字段 | JOIN 查询得到，不在 bookmarks 表 |

### 2.5 Note

`domain.Note` 表示用户在某个阅读位置写下的文本。

| 字段 | 含义 | 说明 |
|---|---|---|
| `ID`、`BookID` | 标识与所属书籍 | 持久化 |
| `ChapterNo`、`LineOffset`、`CharOffset` | 位置 | 持久化 |
| `Content` | 笔记正文，不能为空 | 持久化 |
| `CreatedAt`、`UpdatedAt` | 时间戳 | 持久化 |
| `BookTitle`、`ChapterTitle` | 展示字段 | JOIN 查询得到 |

当前没有“更新笔记”用例，`updated_at` 在创建时写入，列表按它排序。

### 2.6 SearchResult

`domain.SearchResult` 是本地搜索的只读投影，不单独持久化：

- `BookID`、`BookTitle`
- `ChapterNo`、`ChapterTitle`
- `Snippet`：围绕关键词生成的上下文摘要

### 2.7 Setting

`domain.Setting` 表示数据库内的键值配置：`Key`、`Value`、`UpdatedAt`。当前只允许 `reader.width`、`reader.theme`、`search.limit`。

## 3. 应用层模型

这些类型不是核心持久化实体，但构成业务用例的输入和输出。

| 类型 | 作用 | 生产者 / 消费者 |
|---|---|---|
| `ImportOptions` | `Replace`、标题覆盖、下载大小限制 | import / import-url 命令 -> ImportService |
| `ImportResult` | Book ID、章节/字数、已存在/替换状态、质量告警 | ImportService -> CLI |
| `ChapterView` | Book + Chapter + Progress 的阅读上下文 | ReadService -> CLI/TUI |
| `BookInfo` | 书籍、章节/字数、进度、书签/笔记计数 | LibraryService -> info/remove |
| `ListBooksOptions` | 书名过滤和排序方式 | list 命令 -> LibraryService |
| `SourceSearchOptions` | 来源、关键词和结果数 | source search -> SourceService |
| `SourceSearchResult` | OPDS 书名、作者、摘要、EPUB/TXT URL | SourceService -> CLI；不入库 |
| `ReaderOptions` | 最大阅读宽度和主题 | ConfigService -> TUI |

## 4. SQLite 表

| 表 | 用途 | 关键约束 |
|---|---|---|
| `books` | 书籍元数据 | `content_hash` 非空时唯一 |
| `chapters` | 章节正文 | `UNIQUE(book_id, chapter_no)`；删除书籍时级联 |
| `chapter_search_terms` | 章节 2/3-gram 辅助索引 | `PRIMARY KEY(chapter_id, term)`；删除章节时级联 |
| `reading_progress` | 每本书唯一进度 | `book_id` 主键；删除书籍时级联 |
| `bookmarks` | 阅读位置快照 | 删除书籍时级联 |
| `notes` | 阅读笔记 | 删除书籍时级联 |
| `settings` | 数据库级配置 | `key` 主键 |
| `sources` | 为未来可配置来源预留 | 当前没有 Store 方法和业务消费者 |

SQLite 打开时启用：

```sql
PRAGMA busy_timeout = 5000;
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
```

## 5. 搜索索引模型

当前不是 FTS5。每次插入章节时：

1. 合并章节标题与正文。
2. 转小写并保留字母、数字、连接符、下划线和非 ASCII 字符。
3. 生成长度为 2 和 3 的 ngram。
4. 写入 `chapter_search_terms(chapter_id, term)`。

搜索多字符关键词时先用 2-gram 交集找候选章节，再用 `LIKE` 验证完整关键词。单字符或无候选结果时直接回退 `LIKE`。

## 6. 写入事务与数据生命周期

- `InsertChapters`：章节与搜索词在一个事务中写入。
- `ReplaceBook`：更新书籍、删除进度/书签/笔记/旧章节、重建章节和索引，全部在一个事务中完成。
- `DeleteBook`：依赖外键级联清理关联数据。
- `SaveProgress`、`SetSetting`：使用 upsert。

## 7. 当前数据层限制

- schema 通过 `CREATE TABLE IF NOT EXISTS` 内嵌初始化，没有 schema version 或 migrations。
- `sources` 表与 `source_book_id`、`cover_url` 等字段主要是扩展预留。
- URL 导入复用临时文件，当前 `file_path` 不是可重放的远程来源地址。
- 没有 backup / restore / doctor / reindex 命令。
- 搜索索引会随正文长度增长，尚未提供单独重建和体积治理能力。
