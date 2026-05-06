# 05. Source 插件体系设计

## 1. 为什么需要 Source 抽象

如果项目只支持 TXT，那么很容易写成：

```text
Reader → TXT Parser → SQLite
```

但未来如果要支持：

- EPUB
- Markdown
- RSS
- Web Article
- 合法在线 API
- 自定义插件

就必须把“内容来源”抽象出来。

目标：

> Reader Core 不关心内容从哪里来，只消费统一的 Book、Chapter、Content。

## 2. Source 的核心职责

Source 负责：

- 搜索书籍
- 获取书籍元信息
- 获取章节目录
- 获取章节正文
- 判断是否支持 Lazy Loading
- 返回统一数据结构

## 3. 概念接口

```go
type Source interface {
    Name() string
    Type() string

    SearchBooks(ctx context.Context, keyword string) ([]BookMeta, error)
    GetBook(ctx context.Context, sourceBookID string) (BookMeta, error)
    ListChapters(ctx context.Context, sourceBookID string) ([]ChapterMeta, error)
    GetChapterContent(ctx context.Context, sourceChapterID string) (ChapterContent, error)

    Capabilities() SourceCapabilities
}
```

## 4. SourceCapabilities

```go
type SourceCapabilities struct {
    Search      bool
    LazyLoad    bool
    Offline     bool
    Export      bool
    AuthRequired bool
}
```

示例：

| Source | Search | LazyLoad | Offline |
|---|---:|---:|---:|
| Local TXT | false | false | true |
| Local EPUB | false | false | true |
| RSS | true | true | false |
| Web Article | true | true | false |
| Online API | true | true | false |

## 5. 数据结构

### BookMeta

```go
type BookMeta struct {
    SourceType   string
    SourceBookID string
    Title        string
    Author       string
    Description  string
    CoverURL     string
}
```

### ChapterMeta

```go
type ChapterMeta struct {
    SourceChapterID string
    ChapterNo       int
    Title           string
    WordCount       int
}
```

### ChapterContent

```go
type ChapterContent struct {
    SourceChapterID string
    Title           string
    Content         string
    ContentHash     string
}
```

## 6. Lazy Loading 机制

### 全量导入模式

适合：

- TXT
- EPUB
- Markdown

流程：

```text
Parse File
  ↓
Book + All Chapters
  ↓
SQLite
```

### 懒加载模式

适合：

- 在线源
- RSS
- 网页文章集合

流程：

```text
Import Book Meta
  ↓
Import Chapter Catalog
  ↓
Chapter content_status = empty
  ↓
Read Chapter
  ↓
Fetch Content
  ↓
Cache Content
```

## 7. Source Registry

需要一个 Source 注册器：

```go
type Registry struct {
    sources map[string]Source
}

func (r *Registry) Register(source Source)
func (r *Registry) Get(sourceType string) (Source, bool)
func (r *Registry) List() []Source
```

## 8. Source 接入流程

新增一个 Source 时，只需要：

1. 实现 Source 接口
2. 注册到 Registry
3. 配置 source_type
4. 测试 import/read/search 流程

Reader Core 不需要修改。

## 9. 第一阶段 Source

### LocalTxtSource

职责：

- 读取 TXT 文件
- 解析章节
- 返回 BookMeta + ChapterContent

### LocalEpubSource

第二阶段实现。

### MarkdownSource

第三阶段实现。

## 10. 关于番茄小说等在线平台

项目本身不应内置任何未经授权的逆向接口。

如果未来存在合法 API 或用户授权方式，可以以 Source 插件形式接入。

原则：

- 不绕过登录
- 不破解签名
- 不批量分发内容
- 不缓存未授权内容
- 不作为下载器宣传

## 11. Source 配置

```yaml
sources:
  local:
    enabled: true

  rss:
    enabled: true
    feeds:
      - https://example.com/feed.xml

  web:
    enabled: false
```

配置可以落入 `sources.config_json`。
