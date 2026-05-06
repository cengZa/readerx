# ReaderX 用户手册

英文版：[USER_MANUAL.md](USER_MANUAL.md)

本文档说明如何安装、配置和使用 ReaderX 当前支持的全部命令。

ReaderX 是一个本地优先的终端阅读器。它会把导入的书籍内容、阅读进度、书签、笔记和搜索索引保存在本地 SQLite 数据库中。

ReaderX 不是下载器，也不是爬虫工具。它面向你自己拥有或有权阅读的本地 TXT / EPUB 文件。

## 1. 基本概念

### 书籍

书籍是导入后的 TXT 或 EPUB 文件。每本书都会获得一个数字 `Book ID`。

### 章节

ReaderX 会把导入文件切分为章节。

- TXT 会识别常见章节标题，例如 `第一章`、`第 1 章`、`Chapter 1`、`卷一`。
- 如果 TXT 没有章节标题，ReaderX 会自动生成伪章节。
- EPUB 会按照 EPUB spine 顺序读取章节。

### 阅读进度

ReaderX 会保存每本书最后阅读到的章节和位置。

### 书签

书签会保存书籍、章节、位置、摘录和可选备注。

### 笔记

笔记是你在当前阅读位置写下的文本内容。

### 数据库

默认数据库位置：

```text
~/.readerx/reader.db
```

使用全局参数 `--db` 可以指定其他数据库：

```bash
./readerx --db ./reader.db list
```

注意：`--db` 必须放在子命令前面。

正确：

```bash
./readerx --db ./reader.db read 1
```

错误：

```bash
./readerx read 1 --db ./reader.db
```

## 2. 安装

### 从源码构建

环境要求：

- Go 1.26 或更新版本
- macOS 或 Linux 终端

构建：

```bash
make build
```

生成可执行文件：

```bash
./readerx
```

检查命令：

```bash
./readerx --help
```

### 从 Release 安装

从这里下载发布包：

<https://github.com/cengZa/readerx/releases>

解压并运行：

```bash
tar -xzf readerx_vX.Y.Z_darwin_arm64.tar.gz
cd readerx_vX.Y.Z_darwin_arm64
chmod +x readerx
./readerx --help
```

可选：安装到系统路径：

```bash
sudo mv readerx /usr/local/bin/readerx
readerx --help
```

## 3. 快速开始

使用仓库内置测试书：

```bash
make build
./readerx --db ./reader.db import ./book.txt
./readerx --db ./reader.db list
./readerx --db ./reader.db chapters 1
./readerx --db ./reader.db read 1
```

使用你自己的书籍：

```bash
./readerx import /path/to/book.txt
./readerx import /path/to/book.epub
./readerx list
./readerx read 1
```

## 4. 全局命令格式

```bash
./readerx [全局参数] <命令> [命令参数]
```

全局参数：

```text
--db string   SQLite 数据库路径
```

示例：

```bash
./readerx --db ./reader.db search "剑气"
```

## 5. 命令说明

### 5.1 帮助

查看根命令帮助：

```bash
./readerx --help
```

查看某个命令帮助：

```bash
./readerx read --help
./readerx bookmark --help
./readerx config --help
```

### 5.2 导入书籍

导入本地 TXT 或 EPUB 文件。

```bash
./readerx import <file>
```

示例：

```bash
./readerx import ./book.txt
./readerx import ./book.epub
```

输出示例：

```text
导入成功：
书名：book
章节数：3
总字数：318
Book ID：1
```

说明：

- TXT 需要是 UTF-8 编码。
- EPUB 支持常见 OPF / spine / XHTML 结构。
- 重复导入相同内容会被 content hash 拒绝，避免重复入库。

### 5.3 查看书架

列出已导入书籍。

```bash
./readerx list
```

输出示例：

```text
ID  Title  Chapters  Progress  Last Read
1   book   3         33%       2026-05-06
```

### 5.4 查看章节目录

列出某本书的章节元数据。

```bash
./readerx chapters <book-id>
```

示例：

```bash
./readerx chapters 1
```

输出示例：

```text
No  Title       Words
1   第一章 山雨欲来  93
2   第二章 灯下旧约  120
3   第三章 白鹿渡口  105
```

### 5.5 阅读

打开 TUI 阅读器：

```bash
./readerx read <book-id>
```

打开指定章节：

```bash
./readerx read <book-id> --chapter <chapter-no>
```

不进入 TUI，直接打印纯文本：

```bash
./readerx read <book-id> --chapter <chapter-no> --plain
```

示例：

```bash
./readerx read 1
./readerx read 1 --chapter 2
./readerx read 1 --chapter 2 --plain
```

参数：

```text
--chapter int   章节号
--plain         不进入 TUI，直接输出章节正文
```

### 5.6 继续阅读

继续最近阅读的书籍。

```bash
./readerx continue
```

纯文本输出：

```bash
./readerx continue --plain
```

参数：

```text
--plain   不进入 TUI，直接输出章节正文
```

### 5.7 搜索

搜索已导入书籍。

```bash
./readerx search <keyword>
```

示例：

```bash
./readerx search "剑气"
./readerx search "剑气" --book 1
./readerx search "剑气" --limit 10
```

参数：

```text
--book int    限定搜索某一本书
--limit int   最多返回多少条结果
```

行为说明：

- 搜索使用轻量 SQLite ngram 索引。
- 候选章节会再用精确关键词匹配校验。
- 单字搜索会 fallback 到 SQLite `LIKE`。
- 如果没有传 `--limit`，会使用配置项 `search.limit`。

输出示例：

```text
[Book 1] book / 第 1 章 第一章 山雨欲来
...一道剑气自山巅而起...
```

### 5.8 书签

#### 添加书签

在某本书当前保存的阅读位置添加书签。

```bash
./readerx bookmark add <book-id>
./readerx bookmark add <book-id> --note "important moment"
```

参数：

```text
--note string   书签备注
```

也可以在 TUI 中按 `b` 添加书签。

#### 查看书签

```bash
./readerx bookmark list
./readerx bookmark list --book <book-id>
```

参数：

```text
--book int   只查看某本书的书签
```

#### 删除书签

```bash
./readerx bookmark remove <bookmark-id>
```

### 5.9 笔记

#### 添加笔记

在某本书当前保存的阅读位置添加笔记。

```bash
./readerx note add <book-id> --content "review this section"
```

参数：

```text
--content string   笔记内容
```

也可以在 TUI 中按 `m` 添加笔记。

#### 查看笔记

```bash
./readerx note list
./readerx note list --book <book-id>
```

参数：

```text
--book int   只查看某本书的笔记
```

#### 删除笔记

```bash
./readerx note remove <note-id>
```

### 5.10 导出

把书签或笔记导出为 Markdown。

#### 导出书签

打印到 stdout：

```bash
./readerx export bookmarks
./readerx export bookmarks --book 1
```

写入文件：

```bash
./readerx export bookmarks --book 1 -o bookmarks.md
```

#### 导出笔记

打印到 stdout：

```bash
./readerx export notes
./readerx export notes --book 1
```

写入文件：

```bash
./readerx export notes --book 1 -o notes.md
```

参数：

```text
--book int            只导出某本书
-o, --output string   写入 Markdown 文件
```

### 5.11 配置

查看全部配置：

```bash
./readerx config list
```

读取单个配置：

```bash
./readerx config get search.limit
```

设置配置：

```bash
./readerx config set search.limit 25
./readerx config set reader.width 100
./readerx config set reader.theme dark
```

支持的配置项：

```text
reader.width   正整数
reader.theme   default, dark, light
search.limit   正整数
```

配置效果：

- `reader.width` 控制 TUI 阅读器最大文本宽度。
- `reader.theme` 控制 TUI 阅读器主题。
- `search.limit` 控制 `search` 默认返回结果数量。

## 6. TUI 阅读器

打开：

```bash
./readerx read 1
```

快捷键：

```text
j / down   向下滚动
k / up     向上滚动
Space      下一页
u          上一页
n          下一章
p          上一章
g          跳转章节
/          搜索当前书
b          添加书签
m          添加笔记
s          保存进度
q          退出并保存
```

### 跳转章节

按 `g`，输入章节号，然后按 Enter。

按 Esc 取消。

### 搜索当前书

按 `/`，输入关键词，然后按 Enter。

出现搜索结果后：

```text
j / down   选择下一条结果
k / up     选择上一条结果
Enter      跳转到选中的结果
Esc        取消
```

### 添加书签

按 `b`。

ReaderX 会保存当前书籍、章节、行偏移、字符偏移和摘录。

### 添加笔记

按 `m`，输入笔记内容，然后按 Enter。

按 Esc 取消。

## 7. 常见工作流

### 阅读一本新的 TXT 书

```bash
./readerx import ./book.txt
./readerx list
./readerx read 1
```

### 阅读 EPUB

```bash
./readerx import ./book.epub
./readerx list
./readerx chapters 1
./readerx read 1
```

### 使用项目内测试数据库

```bash
./readerx --db ./reader.db import ./book.txt
./readerx --db ./reader.db list
./readerx --db ./reader.db read 1
```

### 搜索后手动跳转

```bash
./readerx search "白鹿" --book 1
./readerx read 1 --chapter 3
```

### 导出阅读沉淀

```bash
./readerx export bookmarks --book 1 -o bookmarks.md
./readerx export notes --book 1 -o notes.md
```

## 8. 开发命令

构建：

```bash
make build
```

测试：

```bash
make test
```

清理：

```bash
make clean
```

不构建直接运行：

```bash
go run . --db ./reader.db import ./book.txt
go run . --db ./reader.db read 1 --plain
```

完整验证：

```bash
go test ./...
go build ./...
```

## 9. 当前限制

- EPUB 支持常见 OPF / spine / XHTML 结构，但不保证覆盖所有 EPUB 边界情况。
- 搜索是 ngram 辅助的精确关键词搜索，不是完整自然语言分词搜索。
- Markdown 导出目前覆盖书签和笔记，不支持整本书导出。
- 当前版本不包含在线源和 AI 阅读能力。

## 10. 安全和数据归属

ReaderX 是本地优先工具：

- 导入内容保存在你的本地 SQLite 数据库。
- ReaderX 不上传书籍。
- ReaderX 不包含平台爬虫或下载器逻辑。
- 请只导入你拥有或有权阅读的文件。
