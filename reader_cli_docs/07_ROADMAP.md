# ReaderX 版本路线图

## 当前发布：v0.5.1

`v0.5.1` 已具备可安装、可发布和可日常使用的本地终端阅读闭环。

已发布能力：

- TXT / EPUB 文件与目录导入
- 内容哈希去重、`--replace` 和章节质量提示
- 公开 HTTP/HTTPS TXT / EPUB 直链导入
- Project Gutenberg OPDS 搜索
- SQLite 书库、列表/详情/删除、阅读进度
- TUI 分页、切章、跳转、书内搜索、帮助和主题
- 本地 ngram 搜索、书签、笔记和 Markdown 导出
- macOS / Linux / Windows 的 amd64 / arm64 发布包

## 当前限制

- 目录导入顺序执行，单文件失败会中断批次，没有进度和 dry-run。
- EPUB 只覆盖常见 OPF/spine/XHTML 结构。
- 搜索是 ngram 候选 + LIKE 校验，不是完整中文分词或语义搜索。
- Source 搜索只支持 Project Gutenberg，结果不能一条命令直接导入。
- 在线导入没有持久化原始 URL，`file_path` 会记录临时文件路径。
- TUI 没有目录、书签/笔记面板和关键词高亮。
- schema 没有版本号、迁移、备份和修复工具。
- 超大章节仍然整章加载和重新分页。

## 下一阶段：v0.6.0 导入可靠性与来源闭环

目标：让多书导入可预期，并让开放书源搜索到导入形成完整闭环。

建议顺序：

1. 目录导入容错：单个文件失败不终止其他文件，最终汇总成功/失败。
2. `readerx import --dry-run`：展示会处理的文件，不写数据库。
3. 文本进度与耗时：显示当前文件、总数和最终统计。
4. 保存远程来源 URL，修正在线书籍 `file_path` 语义。
5. `readerx source import`：从明确来源结果直接选择 EPUB/TXT 并复用 ImportService。
6. 加入 release archive smoke test，至少检查每个包的二进制名和 `version` 输出。

验收重点：

- 坏文件不阻止同目录中的好文件入库。
- 用户能在写数据库前预览批次。
- OPDS 结果可以无需手工复制 URL 完成导入。
- 在线书籍能展示可追踪的来源 URL。

## v0.7.0 阅读体验

- TUI 章节目录面板
- 搜索结果列表与正文高亮
- 书签/笔记列表和跳转
- resize 时基于字符偏移更稳定地恢复位置
- 超大章节分页性能优化

## v0.8.0 数据维护

- schema version 和 migration
- `readerx backup` / `readerx restore`
- `readerx doctor` 检查数据库、索引和孤立数据
- `readerx reindex` 重建 `chapter_search_terms`
- 明确保留或移除当前未使用的 `sources` 表

## 后续候选

- 第二个合法 OPDS 来源和最小 CatalogSource Registry
- Homebrew Tap 与 Windows 安装脚本
- 整本书或阅读资料的可移植导出

## 明确不规划

- z-library 类来源
- 登录、验证码、反爬、付费墙或访问控制绕过
- 任意网页批量抓书和未授权内容分发
- EPUB 内容自动篡改式“修复”
