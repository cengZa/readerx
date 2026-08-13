# ReaderX 当前架构与业务链路

本文档描述 `v0.5.1` 对应代码的真实结构。它既是架构说明，也是从命令入口走到数据库和 TUI 的代码导读。

## 1. 总体结构

```text
main.go
  -> cmd/                 Cobra 命令、参数、终端输出
      -> internal/app/    用例编排、校验、结果聚合
          -> internal/source/ + internal/parser/
          -> internal/storage/
      -> internal/tui/    阅读交互
          -> internal/reader/
          -> storage 的窄接口

SQLite <-> internal/storage/sqlite_store.go
HTTP   <-> import-url / Project Gutenberg OPDS
```

主链路遵循 `cmd -> app -> storage`，但 TUI 为了在交互中高频保存进度、切章、搜索和添加标记，通过 `tui.ProgressSaver` 直接依赖一组存储能力。这是当前有意保持简单的实现，也是后续可以收紧的架构边界。

## 2. 包职责

| 包 | 主要职责 | 关键入口 |
|---|---|---|
| `main` | 启动 CLI | `main.go` |
| `cmd` | 命令注册、参数解析、打开数据库、输出结果 | `cmd/root.go`、各命令文件 |
| `internal/domain` | 核心领域数据结构 | `domain.Book`、`Chapter`、`Progress` 等 |
| `internal/app` | 业务用例、输入校验、跨存储操作编排 | `ImportService`、`ReadService`、`LibraryService` 等 |
| `internal/source` | 根据本地文件格式选择解析器 | `ImportLocalFile` |
| `internal/parser` | 把 TXT / EPUB 解析为统一 Book + Chapters | `ParseTXTFile`、`ParseEPUBFile` |
| `internal/storage` | Store 接口、SQLite schema、查询和事务 | `Store`、`SQLiteStore` |
| `internal/reader` | 按终端显示宽度换行和分页 | `Paginator` |
| `internal/tui` | Bubble Tea 阅读状态机和交互 | `ReaderModel`、`RunReader` |
| `internal/config` | 默认数据库路径 | `DefaultDBPath` |
| `internal/version` | 构建版本信息 | `Current` |

## 3. 启动与公共命令链路

所有命令都从下面的链路开始：

```text
main.main
  -> cmd.Execute
  -> Cobra 匹配子命令
  -> cmd.openStore
  -> config.DefaultDBPath（未传 --db 时）
  -> storage.OpenSQLite
  -> 初始化 PRAGMA 和 schema
```

默认数据库是 `~/.readerx/reader.db`；设置 `XDG_DATA_HOME` 时改为 `$XDG_DATA_HOME/readerx/reader.db`。`--db` 可以覆盖默认路径。

## 4. 导入链路

### 4.1 本地文件或目录

```text
readerx import <path>
  -> cmd/import.go
  -> app.ImportService.ImportPathWithOptions
  -> 文件：ImportFileWithOptions
  -> 目录：递归收集 .txt/.epub，排序后逐个导入
  -> source.ImportLocalFile
      -> LocalTxtSource -> parser.ParseTXTFile
      -> LocalEpubSource -> parser.ParseEPUBFile
  -> domain.Book + []domain.Chapter
  -> ChapterQualityWarnings
  -> Store.GetBookByContentHash
      -> 已存在：返回原书，或 --replace 事务替换
      -> 不存在：CreateBook + InsertChapters
  -> 插入章节时同时生成 chapter_search_terms
```

TXT 解析器识别中英文常见章节标题；没有标题时每 4000 个 rune 生成一个伪章节。EPUB 解析器读取 `container.xml`、OPF manifest/spine，再按 spine 顺序抽取 XHTML 正文。

### 4.2 公开 URL

```text
readerx import-url <url>
  -> ImportService.ImportURLWithOptions
  -> 校验 http/https、HTTP 状态、类型和最大大小
  -> 下载到临时目录
  -> ImportFileWithOptions
  -> 完全复用本地文件解析、去重、入库和索引链路
  -> 删除临时文件
```

这里最重要的复用点是 `ImportFileWithOptions`：在线导入没有第二套解析和持久化逻辑，因此本地与在线输入得到相同的领域模型、去重规则和搜索索引。

## 5. 开放书源搜索链路

```text
readerx source search <keyword>
  -> app.SourceService.Search
  -> Project Gutenberg search.opds
  -> XML 解析 Atom entry
  -> 只保留 acquisition 类型的 EPUB / TXT 链接
  -> []SourceSearchResult
  -> CLI 输出链接
  -> 用户选择链接后执行 readerx import-url
```

当前 `SourceService` 是内存中的单一来源实现，没有通用 Source 接口或 Registry；SQLite 的 `sources` 表也没有 Store 方法和业务消费者。换言之，OPDS 搜索结果不会入库，书籍只有在执行 `import-url` 后才进入 `books` 和 `chapters`。

## 6. 书库管理链路

### 列表与详情

```text
readerx list/info
  -> LibraryService
  -> Store 查询 books、chapters、progress、bookmarks、notes
  -> 应用层筛选、排序或聚合 BookInfo
  -> CLI 输出
```

`ListBooks` 的章节数和进度是 SQL 聚合出来的查询投影，不是 `books` 表中的列。`BookInfo` 的总字数、书签数和笔记数由应用层聚合。

### 删除

```text
readerx remove <book-id>
  -> 先读取 BookInfo 并要求确认
  -> LibraryService.RemoveBook
  -> DELETE books
  -> SQLite ON DELETE CASCADE 清理章节、索引、进度、书签和笔记
```

## 7. 阅读与进度链路

```text
readerx read <book-id> / readerx continue
  -> ReadService
  -> 读取 Book、Progress、Chapter 和章节总数
  -> 保存本次章节与基础进度，更新 last_read_at
  -> 非交互终端或 --plain：直接输出正文
  -> 交互终端：tui.RunReader
      -> reader.Paginator 按显示宽度换行分页
      -> ReaderModel 处理滚动、翻页、切章、跳转、搜索、标记
      -> 保存 Progress / Bookmark / Note
```

进度同时保存 `line_offset` 和 `char_offset`。当前恢复主要使用 `line_offset`；`char_offset` 用于稳定定位、书签摘录和未来更精确的 resize 恢复。

TUI 内的全书百分比按“已完成章节 + 当前章页内比例”计算。切章、手动保存、添加书签/笔记和退出都会保存进度。

## 8. 本地搜索链路

```text
导入章节
  -> 标题 + 正文规范化
  -> 生成 2-gram / 3-gram
  -> chapter_search_terms

readerx search <keyword>
  -> SearchService 校验
  -> Store.SearchChapters
  -> 多字符：ngram 找候选 + LIKE 精确确认
  -> 单字符或无候选：LIKE 回退
  -> SearchResult + 上下文 snippet
```

这不是 FTS5，也不是自然语言分词。索引用于缩小中文搜索候选集，最终仍以标题或正文包含原关键词为准。

## 9. 书签、笔记与导出

```text
当前位置 Progress
  -> BookmarkService / NoteService
  -> 读取章节和字符偏移
  -> bookmarks / notes
  -> list 时 JOIN books / chapters 补全展示标题
  -> ExportMarkdown 生成 Markdown
```

CLI 添加书签或笔记时使用已保存进度；TUI 添加时先保存当前实时位置。书签保存摘录和可选备注，笔记保存用户正文。

## 10. 配置链路

`ConfigService` 只允许三个键：

| 键 | 默认值 | 消费位置 |
|---|---:|---|
| `reader.width` | `100` | `cmd/read.go`、`cmd/continue.go` 构造 TUI 选项 |
| `reader.theme` | `default` | TUI 样式 |
| `search.limit` | `50` | CLI 本地搜索默认结果数 |

设置写入 SQLite `settings` 表，因此跟随所选数据库，而不是写入单独配置文件。

## 11. 当前架构债务

- schema 直接内嵌在 `sqlite_store.go`，没有版本号和迁移机制。
- 首次导入时 `CreateBook` 与 `InsertChapters` 是两个独立事务；章节写入失败可能留下没有章节的书籍记录。
- `sources` 表暂未被使用，当前 Source 搜索也没有统一接口或 Registry。
- TUI 通过窄接口直接调用存储，尚未完全遵守 `UI -> App Service -> Storage`。
- 目录导入顺序执行，单文件失败会终止整个批次。
- URL 下载使用包级 HTTP client 和固定 30 秒超时，尚未暴露更完整的网络配置。
- 超大章节会整章加载并重新分页，仍有内存和启动延迟优化空间。
