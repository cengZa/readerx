# 08. Codex / AI Coding Agent 任务拆分

状态：本文档保留早期任务拆分作为历史参考。ReaderX 当前实现已覆盖 Task 1-11 的主要能力，并额外完成目录导入、书库管理、删除确认、版本命令、发布包和中英文用户手册。下一阶段任务以 [07_ROADMAP.md](07_ROADMAP.md) 为准。

## 使用原则

不要一次让 Codex 生成整个项目。应该按阶段投喂，每次任务边界清晰、验收明确。

## Task 1：生成项目骨架

```text
请生成一个 Go CLI 项目骨架，项目名 readerx。

要求：
1. 使用 cobra 作为 CLI 框架。
2. 目录结构包括：
   - cmd
   - internal/app
   - internal/domain
   - internal/storage
   - internal/parser
   - internal/reader
   - internal/source
   - internal/tui
   - internal/config
3. 实现 root、import、list、read、continue、search、bookmark 命令的空壳。
4. 每个命令只做参数解析，并调用 app service。
5. 保持代码简洁、可测试、模块边界清晰。
6. 添加 README，说明如何运行。
```

## Task 2：实现 SQLite 存储层

```text
请为 readerx 实现 SQLite 存储层。

要求：
1. 使用 database/sql。
2. 使用 SQLite。
3. 表包括：
   - books
   - chapters
   - reading_progress
   - bookmarks
   - notes
   - sources
   - settings
4. 提供 Store 接口和 SQLiteStore 实现。
5. 支持 InitSchema。
6. 支持：
   - CreateBook
   - InsertChapters
   - ListBooks
   - GetBook
   - GetChapter
   - CountChapters
   - SaveProgress
   - GetProgress
   - AddBookmark
   - ListBookmarks
7. 所有方法都返回明确 error。
8. 为核心方法写单元测试。
```

## Task 3：实现 TXT Parser

```text
请实现 TXT 小说解析器。

要求：
1. 输入文件路径，输出 BookMeta 和 []Chapter。
2. 支持 UTF-8。
3. 章节标题识别支持：
   - 第1章
   - 第一章
   - 第 1 章
   - Chapter 1
   - 卷一
4. 如果无法识别章节，则按每 4000 字生成伪章节。
5. 保留段落结构。
6. 清理连续多余空行。
7. 自动根据文件名生成默认书名。
8. 为解析器编写单元测试，覆盖中文章节、英文章节、无章节三种情况。
```

## Task 4：实现 Import 命令

```text
请实现 readerx import <file>。

要求：
1. 根据文件后缀选择 parser。
2. 解析后写入 SQLite。
3. 同时写入 books 和 chapters。
4. 打印导入结果：
   - 书名
   - 章节数
   - 总字数
   - book_id
5. 同名或同 hash 书籍重复导入时给出提示，不直接覆盖。
6. 错误信息要清楚。
```

## Task 5：实现 List 命令

```text
请实现 readerx list。

要求：
1. 从 SQLite 查询所有书籍。
2. 展示：
   - ID
   - Title
   - Author
   - Chapter count
   - Progress
   - Last read time
3. 没有书籍时提示用户使用 readerx import。
4. 输出格式适合终端阅读。
```

## Task 6：实现基础 Read 命令

```text
请实现 readerx read <book-id> --chapter <n>。

要求：
1. 从数据库读取指定章节。
2. 如果没有指定 chapter，则读取阅读进度。
3. 如果没有进度，则从第 1 章开始。
4. 第一版可以直接打印正文，不必进入 TUI。
5. 读取后更新 last_read_at。
```

## Task 7：实现 TUI 阅读器

```text
请使用 Bubble Tea 实现 readerx 的 TUI 阅读器。

要求：
1. 显示章节标题和正文。
2. 支持 j/k 上下滚动。
3. 支持 Space 下一页。
4. 支持 u 上一页。
5. 支持 n/p 切换章节。
6. 支持 q 退出并保存进度。
7. 支持终端尺寸变化。
8. 中文宽度处理使用 go-runewidth。
9. 样式使用 lipgloss。
10. UI 保持简洁，不要过度装饰。
```

## Task 8：实现书签

```text
请实现书签功能。

要求：
1. 支持 readerx bookmark add。
2. 支持 readerx bookmark list。
3. 支持 readerx bookmark remove <id>。
4. TUI 中按 b 可以添加当前位置为书签。
5. 书签保存 book_id、chapter_no、char_offset、excerpt、note。
6. 书签列表可以展示上下文摘录。
```

## Task 9：实现搜索

```text
请实现全文搜索。

要求：
1. 使用 SQLite ngram 辅助索引。
2. 导入章节时写入搜索索引表。
3. 支持 readerx search <keyword>。
4. 支持 --book 限制某本书。
5. 搜索结果展示：
   - book_id
   - book title
   - chapter_no
   - chapter title
   - snippet
6. 支持中文搜索。
```

## Task 10：实现 Source 抽象

```text
请重构 readerx，引入 Source 插件体系。

要求：
1. 定义 Source 接口。
2. 定义 BookMeta、ChapterMeta、ChapterContent。
3. 实现 SourceRegistry。
4. 将 TXT 导入改造成 LocalTxtSource。
5. Reader Core 不直接依赖 TXT Parser。
6. 为未来 EPUB、RSS、OnlineSource 预留扩展点。
```

## Task 11：实现 EPUB

```text
请实现 EPUB 解析器。

要求：
1. EPUB 本质按 zip 读取。
2. 解析 META-INF/container.xml。
3. 找到 OPF 文件。
4. 解析 manifest 和 spine。
5. 按 spine 顺序读取 XHTML。
6. 提取正文。
7. 生成章节。
8. 处理常见编码和 HTML entity。
```

## Task 12：实现 AI 摘要原型

状态：暂不规划。ReaderX 当前重点是本地阅读体验、导入稳定性和数据维护能力。

```text
请实现 readerx ai summarize <book-id> --chapter <n> 的原型。

要求：
1. 从本地数据库读取章节内容。
2. 通过抽象 LLMClient 调用模型。
3. 不要把具体模型 SDK 写死在业务逻辑里。
4. 支持环境变量配置 API Key。
5. 输出章节摘要。
6. 注意控制输入长度。
```
