# ReaderX

English version: [README.md](README.md)

ReaderX 是一个本地优先的长文本终端阅读器。它可以把 TXT 和 EPUB 文件导入本地 SQLite 书库，提供终端阅读界面，并在你的机器上保存阅读进度、搜索索引、书签和笔记。

ReaderX 不是通用下载器，也不是爬虫工具。它面向你自己拥有或有权阅读的本地文件，以及你明确提供的公开 TXT / EPUB 直链。

## 能做什么

- 导入 UTF-8 TXT 文件
- 导入常见 OPF/spine 结构的 EPUB 文件
- 导入包含 TXT 和 EPUB 的目录
- 导入明确的公开 TXT / EPUB URL，并做文件类型和大小限制
- 通过 OPDS 搜索 Project Gutenberg，并导入搜索结果中的 URL
- 导入后提示疑似章节编号问题
- 自动解析章节
- 为没有章节标题的 TXT 文件创建伪章节
- 使用 Bubble Tea TUI 阅读书籍
- 从保存的阅读进度继续阅读
- 查看、排序、筛选书籍，查看详情，并带确认地删除书籍
- 使用轻量 SQLite ngram 索引搜索已导入书籍
- 添加、查看、删除书签
- 添加、查看、删除笔记
- 导出书签和笔记为 Markdown
- 配置阅读宽度、主题和默认搜索数量
- 查看已安装版本和发布包构建信息
- 通过 GitHub Actions 构建发布包

## 快速开始

```bash
make install
readerx import ./book.txt
readerx list
readerx chapters 1
readerx read 1
```

使用你自己的文件：

```bash
readerx import /path/to/book.txt
readerx import /path/to/book.epub
readerx import /path/to/books/
readerx import-url https://example.com/book.epub
readerx source search "alice"
```

默认情况下，ReaderX 会把书库保存在 `~/.readerx/reader.db`；正常使用时不需要传 `--db`。

## 安装

从源码安装到当前用户的命令目录：

```bash
make install
```

默认安装到：

```text
~/.local/bin/readerx
```

确保 `~/.local/bin` 已加入 `PATH`。如果你使用 zsh：

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
readerx --help
```

安装到其他目录：

```bash
make install INSTALL_DIR=/usr/local/bin
```

如果只是本地开发，`make build` 会在项目目录生成本地 `readerx` 二进制。

下载发布包：

<https://github.com/cengZa/readerx/releases>

下载后解压：

```bash
tar -xzf readerx_vX.Y.Z_darwin_arm64.tar.gz
cd readerx_vX.Y.Z_darwin_arm64
chmod +x readerx
```

把下载到的二进制安装为当前用户可全局使用的命令：

```bash
mkdir -p ~/.local/bin
cp readerx ~/.local/bin/readerx.tmp
xattr -c ~/.local/bin/readerx.tmp 2>/dev/null || true
chmod +x ~/.local/bin/readerx.tmp
mv -f ~/.local/bin/readerx.tmp ~/.local/bin/readerx
readerx --help
```

## 用户手册

完整英文手册：

[docs/USER_MANUAL.md](docs/USER_MANUAL.md)

中文版用户手册：

[docs/USER_MANUAL.zh-CN.md](docs/USER_MANUAL.zh-CN.md)

用户手册包含全部命令、参数、TUI 快捷键、配置项和常见工作流。

## 开发

运行测试：

```bash
make test
```

运行完整验证：

```bash
go test ./...
go build ./...
```

不构建直接运行：

```bash
go run . --db ./reader.db import ./book.txt
go run . --db ./reader.db read 1 --plain
```

清理项目内构建和测试产物：

```bash
make clean
```

这不会删除默认用户书库 `~/.readerx/reader.db`。

如果你确实想删除默认用户书库，执行：

```bash
make clean-user-data
```

这会删除 `~/.readerx/reader.db` 以及 SQLite 伴随文件。默认书库里的已导入书籍、阅读进度、书签、笔记和搜索索引都会被删除。

## 架构

```text
cmd/                    CLI 命令入口
internal/app/           应用服务
internal/domain/        核心领域类型
internal/storage/       SQLite 存储
internal/parser/        TXT 和 EPUB 解析器
internal/source/        本地文件来源分发
internal/reader/        换行和分页
internal/tui/           终端阅读 UI
reader_cli_docs/        产品与架构文档
```

CLI 层负责解析参数并调用应用服务。存储和解析逻辑位于应用层之后，因此 TUI 不会直接操作 SQLite。

## 当前限制

- EPUB 支持覆盖常见 OPF/spine/XHTML 书籍，但不保证覆盖所有 EPUB 边界情况。
- 搜索使用轻量 ngram 索引，不是完整分词搜索引擎。
- 目前支持导出书签和笔记为 Markdown，尚未实现整本书导出。
- 目录导入会递归导入支持的文件，但当前仍是顺序导入，还没有进度条。
- 通用网页抓取、z-library 类来源、登录绕过和 AI 阅读能力目前刻意不在当前版本范围内。

## 仓库

GitHub: <https://github.com/cengZa/readerx>
