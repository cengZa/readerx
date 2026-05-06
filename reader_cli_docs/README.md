# Reader CLI 文档包

## 项目定位

Reader CLI 是一个本地优先、可扩展数据源的终端阅读器。它不是“某个平台小说下载器”，而是一个面向终端用户、开发者和重度文本阅读用户的 Terminal Reading Platform。

核心能力：

- 本地 TXT / EPUB / Markdown 导入
- 自动章节解析
- 终端 TUI 阅读
- 阅读进度保存
- 书签与笔记
- 全文搜索
- Source 插件体系
- 后续可扩展合法在线源、RSS、网页文章、AI 阅读助手

## 文档目录

| 文件 | 作用 |
|---|---|
| `01_PRD.md` | 产品需求文档，解释做什么、为什么做、给谁用 |
| `02_ARCHITECTURE.md` | 系统架构设计，解释整体分层和模块职责 |
| `03_FEATURE_SPEC.md` | 功能规格说明，拆解具体功能与验收标准 |
| `04_DATA_MODEL.md` | 数据模型与 SQLite 表设计 |
| `05_SOURCE_PLUGIN.md` | Source 插件体系设计 |
| `06_TUI_UX.md` | CLI/TUI 交互设计与快捷键规范 |
| `07_ROADMAP.md` | 版本路线图与迭代计划 |
| `08_CODEX_TASKS.md` | 给 Codex / AI Coding Agent 的任务拆分提示词 |
| `09_RISKS.md` | 合规、版权、技术与产品风险说明 |

## 推荐开发原则

1. 先做 Reader，不做 Downloader。
2. 先支持本地文件，再考虑在线源。
3. 核心逻辑与 UI 解耦。
4. Source 抽象先设计，具体 Source 后接入。
5. 每个阶段都要有可运行、可验收的版本。
