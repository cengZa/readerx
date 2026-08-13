# ReaderX 维护者文档

本目录面向 ReaderX 的开发者和维护者。用户安装与命令用法请优先阅读仓库根目录的 `README.md`、`README.zh-CN.md` 和 `docs/USER_MANUAL*.md`。

## 当前基线

- 最新发布：`v0.5.1`
- 主要能力：本地 TXT / EPUB 导入、公开直链导入、Project Gutenberg OPDS 搜索、SQLite 书库、TUI 阅读、搜索、书签、笔记和配置
- 支持平台：macOS、Linux、Windows；发布包覆盖 amd64 / arm64
- 代码是行为事实来源；文档与代码冲突时以代码为准，并修正文档

## 文档职责

| 文件 | 唯一职责 |
|---|---|
| `01_PRD.md` | 产品定位、目标用户、核心价值和非目标 |
| `02_ARCHITECTURE.md` | 当前代码架构、业务链路、包职责和代码导读 |
| `03_FEATURE_SPEC.md` | 当前可用能力及验收口径 |
| `04_DATA_MODEL.md` | 领域模型、应用 DTO、SQLite 表和数据约束 |
| `05_SOURCE_PLUGIN.md` | 当前内容来源与导入链路，以及后续 Source 扩展边界 |
| `06_TUI_UX.md` | 当前已实现的 TUI 状态、布局和快捷键 |
| `07_ROADMAP.md` | 已发布能力、当前限制和下一阶段路线 |
| `09_RISKS.md` | 合规、数据、兼容性和架构风险 |

## 去重原则

- README 只做项目概览和快速开始，不复制完整命令手册。
- 用户手册负责“怎么用”，功能规格负责“系统能做什么”。
- 架构文档负责“代码怎么协作”，数据模型文档负责“数据是什么、存在哪里”。
- Source 文档同时承载在线导入边界，不再单独维护一份重复设计。
- 已完成的一次性开发任务不长期保留为任务提示词；后续工作统一进入路线图。

## 推荐阅读顺序

1. `01_PRD.md`
2. `02_ARCHITECTURE.md`
3. `04_DATA_MODEL.md`
4. `03_FEATURE_SPEC.md`
5. 按需阅读 Source、TUI、路线图和风险文档
