# 02. 系统架构设计

## 1. 总体架构

Reader CLI 分为三大层：

```text
Source Layer
    ↓
Reader Core
    ↓
CLI / TUI Layer
```

## 2. 分层说明

### 2.1 Source Layer

负责内容来源。

可能来源：

- Local TXT
- Local EPUB
- Markdown
- RSS
- Web Article
- 合法在线 API
- 未来自定义插件

Source Layer 不关心阅读 UI，只负责提供标准化的 Book、Chapter、Content 数据。

### 2.2 Reader Core

负责阅读业务逻辑。

包括：

- 书籍管理
- 章节管理
- 阅读进度
- 分页
- 文本换行
- 书签
- 笔记
- 搜索
- 缓存
- 元数据管理

### 2.3 CLI / TUI Layer

负责用户交互。

包括：

- 命令行命令
- 交互式 TUI
- 快捷键
- 状态栏
- 布局
- 主题
- 用户输入

## 3. 推荐目录结构

```text
reader-cli/
├── cmd/
│   ├── root.go
│   ├── import.go
│   ├── list.go
│   ├── read.go
│   ├── continue.go
│   ├── search.go
│   ├── bookmark.go
│   └── config.go
│
├── internal/
│   ├── app/
│   │   ├── import_service.go
│   │   ├── read_service.go
│   │   ├── search_service.go
│   │   └── bookmark_service.go
│   │
│   ├── domain/
│   │   ├── book.go
│   │   ├── chapter.go
│   │   ├── progress.go
│   │   ├── bookmark.go
│   │   └── note.go
│   │
│   ├── source/
│   │   ├── source.go
│   │   ├── local_txt.go
│   │   ├── local_epub.go
│   │   └── registry.go
│   │
│   ├── parser/
│   │   ├── txt_parser.go
│   │   ├── epub_parser.go
│   │   └── markdown_parser.go
│   │
│   ├── reader/
│   │   ├── paginator.go
│   │   ├── layout.go
│   │   ├── progress.go
│   │   └── text_width.go
│   │
│   ├── storage/
│   │   ├── store.go
│   │   ├── sqlite_store.go
│   │   └── migrations.go
│   │
│   ├── search/
│   │   ├── index.go
│   │   └── sqlite_fts.go
│   │
│   ├── tui/
│   │   ├── app.go
│   │   ├── bookshelf.go
│   │   ├── reader.go
│   │   ├── search.go
│   │   └── styles.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── path.go
│   │
│   └── ai/
│       ├── summarize.go
│       └── ask.go
│
├── pkg/
├── docs/
├── testdata/
├── go.mod
└── README.md
```

## 4. 模块职责

| 模块 | 职责 |
|---|---|
| `cmd` | CLI 命令入口，只做参数解析和调用 app service |
| `app` | 应用服务层，编排业务流程 |
| `domain` | 核心实体，不依赖外部框架 |
| `source` | 数据源抽象与实现 |
| `parser` | 文件解析 |
| `reader` | 分页、排版、进度计算 |
| `storage` | 数据持久化 |
| `search` | 全文搜索 |
| `tui` | 终端交互界面 |
| `config` | 配置与路径 |
| `ai` | 未来 AI 阅读功能 |

## 5. 核心设计原则

### 5.1 UI 不直接访问数据库

错误：

```text
TUI → SQLite
```

正确：

```text
TUI → App Service → Storage
```

### 5.2 Reader Core 不关心 Source 类型

错误：

```text
Reader 直接解析 TXT
```

正确：

```text
Reader 只消费统一 Chapter 数据
```

### 5.3 Source 不关心 UI

Source 只负责内容获取，不处理分页、样式、快捷键。

### 5.4 本地优先

所有导入、阅读进度、书签、笔记都优先存在本地 SQLite。

## 6. 推荐技术栈

### Go 方案

| 需求 | 推荐 |
|---|---|
| CLI 命令 | Cobra |
| TUI | Bubble Tea |
| 样式 | Lip Gloss |
| SQLite | modernc.org/sqlite 或 mattn/go-sqlite3 |
| 配置 | Viper 或自研简单配置 |
| 字符宽度 | go-runewidth |
| 日志 | slog / zerolog |
| 测试 | testing + testify |

### Node 方案

| 需求 | 推荐 |
|---|---|
| CLI 命令 | Commander |
| TUI | Ink |
| SQLite | better-sqlite3 |
| 配置 | cosmiconfig |
| 打包 | pkg / tsup |

### Python 方案

| 需求 | 推荐 |
|---|---|
| CLI 命令 | Typer |
| TUI | Textual |
| SQLite | sqlite3 / SQLModel |
| 打包 | uv / pipx |

## 7. 推荐优先级

对当前项目，优先推荐：

```text
Go + Cobra + Bubble Tea + SQLite
```

理由：

- 适合单文件发布
- 适合 CLI
- 性能稳定
- 与后端工程师能力模型匹配
- 可跨平台
