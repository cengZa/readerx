# 07. 版本路线图

## Phase 0：设计准备

目标：明确项目边界和架构。

产出：

- README
- PRD
- 架构设计
- 数据模型
- 命令设计
- Codex 任务拆分

验收：

- 开发人员能根据文档理解全貌
- 可以开始生成项目骨架

---

## Phase 1：MVP CLI

目标：能导入、能列出、能阅读。

功能：

- `reader import <file>`
- `reader list`
- `reader read <book-id> --chapter <n>`

验收：

- 可以导入一本 UTF-8 TXT
- 可以自动切章节
- 可以查看书架
- 可以打印某章节正文

---

## Phase 2：阅读进度

目标：支持继续阅读。

功能：

- `reader continue`
- 保存阅读进度
- 更新最近阅读时间

验收：

- 退出后再次进入能恢复到上次位置
- 不同书籍进度互不影响

---

## Phase 3：TUI 阅读器

目标：终端内沉浸阅读。

功能：

- Bubble Tea / Textual / Ink TUI
- 上下滚动
- 翻页
- 切章节
- 状态栏
- 退出保存

验收：

- 可以在 TUI 中读完整章
- 终端 resize 后不崩溃
- 中文显示正常

---

## Phase 4：书签与搜索

目标：提升阅读管理能力。

功能：

- 添加书签
- 查看书签
- 删除书签
- SQLite FTS5 搜索
- 搜索结果跳转

验收：

- 书签能跳转回指定位置
- 中文关键词能搜索
- 搜索结果有上下文摘要

---

## Phase 5：EPUB 支持

目标：支持更标准的电子书格式。

功能：

- EPUB 解析
- OPF / spine 读取
- XHTML 正文提取
- 章节顺序保持

验收：

- 能导入常见 EPUB
- 章节顺序正确
- 正文无明显 HTML 噪声

---

## Phase 6：Source 插件体系

目标：完成架构升级。

功能：

- Source interface
- Source registry
- LocalTxtSource
- LocalEpubSource
- Lazy Loading 支持

验收：

- Reader Core 不依赖 TXT / EPUB 具体实现
- 新增 Source 不修改 Reader Core
- 章节可支持 `empty/cached/failed` 状态

---

## Phase 7：AI 阅读助手

目标：加入智能能力。

功能：

- 章节摘要
- 全书摘要
- 人物关系提取
- 时间线
- 针对当前书籍问答
- 笔记导出

验收：

- 能对单章生成摘要
- 能基于本地内容回答问题
- 能控制 token 成本和上下文范围

---

## Phase 8：发布与分发

目标：让项目可安装、可使用。

功能：

- GitHub Release
- Homebrew Tap
- `go install`
- 配置目录规范
- 数据迁移说明
- 示例文件

验收：

- macOS / Linux 可安装
- 新用户 5 分钟内跑通导入和阅读
