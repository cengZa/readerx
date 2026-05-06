# 03. 功能规格说明

## 1. Import：导入书籍

### 命令

```bash
reader import ./book.txt
reader import ./book.epub
```

### 功能要求

- 根据文件后缀选择解析器
- 自动提取书名
- 自动识别章节
- 生成书籍记录
- 生成章节记录
- 建立搜索索引
- 输出导入统计

### 输出示例

```text
导入成功：
书名：剑来
章节数：1253
总字数：4,200,000
Book ID：1
```

### 验收标准

- 能导入 UTF-8 TXT
- 能识别常见“第 X 章”格式
- 无章节时能 fallback 为伪章节
- 重复导入同一文件时给出提示

---

## 2. List：书架

### 命令

```bash
reader list
```

### 功能要求

展示：

- Book ID
- 书名
- 作者
- 章节数
- 阅读进度
- 最近阅读时间

### 输出示例

```text
ID   Title      Chapters   Progress   Last Read
1    剑来       1253       32%        2026-05-06
2    三体       98         10%        2026-05-05
```

### 验收标准

- 没有书籍时提示用户导入
- 有书籍时按最近阅读或导入时间排序
- 阅读进度显示准确

---

## 3. Read：阅读书籍

### 命令

```bash
reader read 1
reader read 1 --chapter 12
```

### 功能要求

- 打开指定书籍
- 默认从阅读进度处继续
- 可指定章节
- 进入 TUI 阅读模式
- 退出时保存进度

### 快捷键

| 快捷键 | 功能 |
|---|---|
| `j` / `↓` | 向下滚动 |
| `k` / `↑` | 向上滚动 |
| `Space` | 下一页 |
| `n` | 下一章 |
| `p` | 上一章 |
| `b` | 添加书签 |
| `/` | 当前书搜索 |
| `g` | 跳转章节 |
| `q` | 退出并保存 |

### 验收标准

- 可以正常显示中文文本
- 终端窗口变化后能重新排版
- 翻页不丢内容
- 退出后进度保存

---

## 4. Continue：继续阅读

### 命令

```bash
reader continue
```

### 功能要求

- 打开最近阅读书籍
- 定位到上次章节和位置

### 验收标准

- 有最近记录时直接进入阅读器
- 没有记录时提示选择书籍

---

## 5. Search：搜索

### 命令

```bash
reader search "关键词"
reader search "关键词" --book 1
```

### 功能要求

- 支持全库搜索
- 支持指定书籍搜索
- 返回章节与上下文摘要
- 支持从搜索结果跳转阅读

### 输出示例

```text
[Book 1] 剑来 / 第十二章 山雨欲来
... 一道剑气自山巅而起 ...
```

### 验收标准

- 能搜索章节标题和正文
- 搜索结果包含 book_id、chapter_no、snippet
- 支持中文关键词

---

## 6. Bookmark：书签

### 命令

```bash
reader bookmark add
reader bookmark list
reader bookmark remove <bookmark-id>
```

### 功能要求

- 在当前阅读位置添加书签
- 保存章节、offset、摘录、备注
- 支持查看书签列表
- 支持删除书签

### 验收标准

- 书签可以从阅读器快捷键添加
- 书签能跳回对应章节和位置
- 删除书签后不再显示

---

## 7. Note：笔记

### 命令

```bash
reader note add
reader note list
reader note edit <note-id>
reader note remove <note-id>
```

### 功能要求

- 关联到书籍、章节和 offset
- 支持编辑和删除
- 后续可导出 Markdown

第一版可以暂缓。

---

## 8. Config：配置

### 命令

```bash
reader config get
reader config set theme dark
reader config set reader.width 80
```

### 推荐配置项

```yaml
reader:
  width: 80
  margin: 2
  theme: dark
  auto_save: true

storage:
  db_path: ~/.readerx/reader.db

ui:
  vim_keys: true
  show_status_bar: true
```

---

## 9. Export：导出

后续功能。

```bash
reader export bookmarks --book 1 --format markdown
reader export notes --book 1 --format markdown
```
