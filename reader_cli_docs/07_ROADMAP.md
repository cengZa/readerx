# 07. 版本路线图

本文档记录 ReaderX 当前进度和下一阶段路线。命令名前缀统一为 `readerx`。

## 最新发布：v0.4.0

目标：成为可安装、可发布、可日常使用的本地 TXT / EPUB 终端阅读器。

已完成：

- 全局安装：`make install` 安装到 `~/.local/bin/readerx`。
- 默认数据库：日常使用默认写入 `~/.readerx/reader.db`。
- 导入能力：支持 UTF-8 TXT、常见 OPF/spine EPUB、目录递归导入。
- 重复处理：相同内容重复导入返回已有 Book ID，`--replace` 可重新解析。
- 导入质量提示：提示疑似章节编号重复、跳号、回退，不自动修改内容。
- 书库管理：`list`、`info`、`chapters`、`remove`。
- 列表筛选排序：`readerx list --sort recent|created|title`，`readerx list --filter <keyword>`。
- 删除保护：`readerx remove` 交互确认，非交互环境必须传 `--yes`。
- 阅读：`read`、`continue`、`--plain`、指定章节。
- TUI：滚动、翻页、左右键、Home/End、章节切换、帮助页、窄屏状态栏、进度显示。
- 搜索：轻量 SQLite ngram 索引，支持中文关键词、书籍限定和结果跳转。
- 书签与笔记：添加、查看、删除，TUI 内快捷添加。
- 导出：书签和笔记导出 Markdown。
- 配置：阅读宽度、主题、默认搜索数量。
- 版本信息：`readerx version`，Release 包注入 tag、commit、构建时间。
- 发布：GitHub Actions 构建 macOS / Linux / Windows amd64 / arm64 发布包。
- 文档：英文 README、中文 README、英文用户手册、中文用户手册。

当前限制：

- EPUB 只覆盖常见结构，不保证覆盖所有 EPUB 边界情况。
- 搜索不是完整分词搜索，复杂中文语义检索不是当前目标。
- 大文件或目录导入没有进度条，用户可能误以为卡住。
- 目录导入是顺序导入，失败时会中断当前批次。
- TUI 每次仍以章节为读取单位，超大章节的启动和排版还有优化空间。
- 尚未提供整本书导出、数据备份/恢复、数据库迁移命令。
- OPDS 目录搜索、任意网页正文抽取、平台爬虫和 AI 阅读助手暂不进入当前发布版。

## 当前主干：v0.5.0 在线导入基础

目标：减少“手动下载到本地再导入”的重复步骤，但不把 ReaderX 做成通用下载器或爬虫。

已完成：

- `readerx import-url <url>`：导入明确的公开 HTTP/HTTPS TXT / EPUB 直链。
- `readerx import-url --replace`：相同内容重新解析并替换。
- `readerx import-url --title <title>`：覆盖导入后的书名。
- `readerx import-url --max-bytes <bytes>`：限制最大下载字节数，默认 100MB。
- `readerx source list`：列出当前支持的开放书源。
- `readerx source search <keyword>`：通过 Project Gutenberg OPDS 搜索可导入的 EPUB / TXT URL。
- 在线导入设计文档：[10_ONLINE_IMPORT_DESIGN.md](10_ONLINE_IMPORT_DESIGN.md)。

范围外：

- 不接入 z-library 等高版权风险来源。
- 不做登录、Cookie、验证码、反爬、镜像发现、付费墙绕过。
- 不解析任意网页正文作为书籍。

下一步：

- 增加 `readerx source import`，让用户可以从搜索结果直接导入，不必手动复制 URL。
- 评估接入 Standard Ebooks OPDS。
- 增加搜索结果缓存或历史选择能力。

## 下一阶段候选：v0.6.0 稳定性与批量导入体验

目标：降低大书和多书导入时的不确定感，提高错误恢复能力。

范围内：

- 批量导入统计增强：新增、已存在、替换、失败分别统计。
- 目录导入容错：单个文件失败不终止整个批次，最后汇总失败清单。
- 导入进度提示：至少显示当前正在处理的文件；优先做轻量文本进度。
- 导入耗时统计：输出总耗时和每本书关键统计。
- `readerx import --dry-run`：只扫描可导入文件并输出计划，不写数据库。
- 测试覆盖：目录导入部分失败、空目录、重复导入、`--dry-run`。

范围外：

- 不做 EPUB 自动修复。
- 不引入 AI。
- 不重写搜索引擎。

验收：

- `readerx import books/` 在多文件目录中能清楚显示处理进度和最终汇总。
- 一个坏文件不会阻止其他好文件入库。
- 用户能在导入前用 `--dry-run` 确认会导入哪些文件。

## 后续阶段候选

### v0.7.0 阅读体验

- 搜索结果面板从单行状态栏升级为可浏览列表。
- TUI 内支持章节目录面板。
- 书签/笔记列表支持从 TUI 内打开。
- 当前章节内搜索高亮。
- 可配置自动保存频率。

### v0.8.0 数据管理

- `readerx backup` / `readerx restore`。
- 数据库 schema version 和迁移机制。
- `readerx doctor` 检查数据库、索引和书籍内容状态。
- `readerx reindex` 重建搜索索引。

### v0.9.0 发布与安装体验

- Homebrew Tap。
- Release 安装脚本。
- CI 增加 release archive smoke test。

### 暂不规划

- EPUB 自动修复。
- z-library 类来源、登录绕过、任意网页抓书。
- AI 摘要、问答、人物关系。
