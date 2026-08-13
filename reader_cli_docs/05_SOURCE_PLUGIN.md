# ReaderX 内容来源与导入设计

本文档说明当前已经实现的 Source / Import 机制，并定义未来扩展边界。当前代码没有通用插件接口或 Source Registry，文件名沿用历史名称以避免文档链接失效。

## 1. 当前有三类输入

| 输入 | 命令 | 实现 |
|---|---|---|
| 本地文件或目录 | `readerx import <path>` | `ImportService` + `internal/source` + parser |
| 公开 TXT / EPUB 直链 | `readerx import-url <url>` | 下载到临时文件后复用本地导入 |
| 开放目录搜索 | `readerx source search <keyword>` | `SourceService` 查询 Project Gutenberg OPDS |

最终只有 TXT / EPUB 文件会进入统一导入链路并持久化为 `domain.Book` 和 `[]domain.Chapter`。

## 2. 本地来源

```text
source.ImportLocalFile(path)
  -> .txt  -> LocalTxtSource.ImportFile -> parser.ParseTXTFile
  -> .epub -> LocalEpubSource.ImportFile -> parser.ParseEPUBFile
  -> domain.Book + []domain.Chapter
```

`internal/source` 当前是格式分派层，不是动态插件系统。它隔离了应用层与具体 parser，使 `ImportService` 不需要判断 TXT / EPUB 的解析细节。

### TXT

- 书名默认取文件名。
- 文件字节 SHA-256 作为书籍内容哈希。
- 识别中文“第 N 章/节/回/卷/部/篇”、`卷 N` 和 `Chapter N`。
- 没有章节标题时每 4000 个 rune 创建伪章节。
- 章节正文清理连续空行，章节哈希按清理后正文计算。

### EPUB

- 读取 `META-INF/container.xml` 定位 OPF。
- 从 OPF 提取标题、作者、manifest 和 spine。
- 按 spine 顺序读取 XHTML，抽取第一个 heading 和正文。
- 跳过没有可读正文的 spine item。
- 整个 EPUB 文件字节 SHA-256 作为书籍内容哈希。

## 3. URL 导入

```text
ImportURLWithOptions(url)
  -> 只允许 http / https
  -> 30 秒超时
  -> 根据 URL 后缀或 Content-Type 判断 TXT / EPUB
  -> 默认最大 100MB
  -> 临时文件
  -> ImportFileWithOptions
  -> 本地解析、内容去重、入库、索引
  -> 清理临时目录
```

复用目标是 `app.ImportService.ImportFileWithOptions`。它读取 `domain.Book.ContentHash`，查询 `books.content_hash` 唯一索引，随后写入 `books`、`chapters` 和 `chapter_search_terms`；这些字段再被书库列表、阅读、搜索和标记功能消费。

当前限制：URL 本身没有保存在 `books`；`file_path` 来自临时文件路径，因此在线来源不可直接重放。

## 4. OPDS 开放目录搜索

```text
SourceService.Search
  -> https://www.gutenberg.org/ebooks/search.opds/
  -> Atom/XML entry
  -> acquisition link
  -> SourceSearchResult{Title, Author, EPUBURL, TextURL}
  -> CLI 输出
```

当前只注册了字符串常量 `gutenberg`。搜索结果是内存 DTO，不写 `sources`、`books` 或缓存；用户选择结果后需要执行：

```bash
readerx import-url <EPUB-or-TXT-url>
```

`sources` 表目前只是 schema 预留，没有对应 Store 方法。不要假设 `readerx source list` 来自数据库配置。

## 5. 为什么暂时没有通用插件接口

当前只有一个在线搜索源，直接实现比提前定义复杂接口更容易验证真实需求。等出现第二个来源，并且搜索、鉴权、分页、格式选择确实存在共同契约时，再抽象 Registry 更合适。

未来可能的最小接口应围绕当前真实能力，而不是一次覆盖所有设想：

```go
type CatalogSource interface {
    Name() string
    Search(ctx context.Context, query string, limit int) ([]SourceSearchResult, error)
}
```

文件导入与目录搜索应继续保持两个概念：Catalog Source 负责发现，ImportService 负责下载、解析和入库。

## 6. 下一步演进建议

1. 增加 `readerx source import`，用稳定结果 ID 或显式格式选择直接调用 ImportService。
2. 保存真实远程来源 URL，而不是临时文件路径。
3. 接入第二个合法 OPDS 来源后再引入 Registry。
4. 为请求超时、User-Agent、最大下载大小和重试提供结构化配置。
5. 明确是否启用 `sources` 表；若短期不用，后续 migration 可考虑移除预留表。

## 7. 合规边界

支持：

- 用户自己的本地文件
- 用户明确提供且有权访问的公开 TXT / EPUB 直链
- Project Gutenberg 等合法开放目录

不支持：

- z-library 等高版权风险来源
- 登录、Cookie、验证码、反爬和镜像发现
- 绕过付费墙、签名、风控或访问控制
- 任意网页正文抽取和批量抓取
- 未授权内容分发
