# 10. 在线导入设计

本文档定义 ReaderX 在线导入的第一版边界。目标是让用户少做“先下载到本地再导入”的重复动作，同时保持项目仍然是本地阅读器，而不是下载器或爬虫。

## 1. 支持范围

第一阶段只支持用户明确提供的公开直链：

- `http://` / `https://`
- `.txt`
- `.epub`

命令形态：

```bash
readerx import-url https://example.com/book.epub
readerx import-url https://example.com/book.txt --replace
readerx import-url https://example.com/book.txt --title "自定义书名"
```

实现策略：

1. 校验 URL 协议，只允许 HTTP/HTTPS。
2. 识别文件类型，只接受 TXT 或 EPUB。
3. 使用超时请求下载到临时文件。
4. 限制最大下载大小，默认 100MB。
5. 复用现有本地导入流程入库。
6. 导入完成后删除临时文件。

## 2. 不支持范围

第一阶段不支持：

- z-library 等高版权风险站点。
- 登录、Cookie、验证码、反爬、镜像站发现。
- 绕过付费墙、访问控制或下载限制。
- 任意网页正文抽取。
- 把网页集合自动整理成书。

原因：

- 法律和版权风险不可控。
- 站点结构不稳定，维护成本高。
- 普通网页不是可靠的书籍格式，章节、分页、正文抽取质量不可预期。
- ReaderX 的核心定位是本地阅读、书库管理和阅读体验。

## 3. 后续方向

第二阶段可以加入合法开放目录源：

- OPDS 目录，例如 Project Gutenberg、Standard Ebooks、部分图书馆目录。
- `readerx source search <keyword>`
- `readerx source import <source-book-id>`

OPDS 适合作为 Source 插件接入，因为它提供机器可读的书籍元数据和下载链接，比解析网页正文稳定得多。

当前实现已经支持：

```bash
readerx source list
readerx source search <keyword>
readerx source search <keyword> --source gutenberg --limit 5
```

搜索结果会显示可导入的 EPUB / TXT URL。用户可以继续使用 `readerx import-url <url>` 导入选定结果。

## 4. 安全与体验约束

- 默认 User-Agent 使用 ReaderX 标识。
- 默认超时 30 秒。
- 默认最大下载 100MB。
- 错误信息需要说明是协议不支持、文件类型不支持、下载过大、HTTP 状态异常，还是导入解析失败。
- 单次 `import-url` 只处理一个明确 URL，不做批量抓取。
