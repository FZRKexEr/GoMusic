# GoMusic

GoMusic 是一个轻量的歌单解析服务。当前 fork 版本已经精简为只支持汽水音乐歌单/收藏页解析：输入汽水音乐分享文案或链接，服务会提取歌单页中的曲目，返回统一的 JSON 歌单结构。

## 当前能力

- 解析汽水音乐短分享链接和 `music.douyin.com/qishui/...` 歌单页链接。
- 支持从分享文案中自动提取 URL。
- 优先读取页面 SSR 注入的 `_ROUTER_DATA`，拿到更稳定的歌单标题、作者、曲目和视频条目信息。
- 当 SSR 数据不可用时，回退到 DOM 解析。
- 支持返回原始歌名或标准化歌名。
- 支持输出顺序和基础格式转换。

## 不再支持的内容

这个 fork 做了明确的收敛，不再维护多平台和持久化逻辑：

- 删除网易云音乐解析。
- 删除 QQ 音乐解析、签名、加密脚本和相关模型。
- 删除 DB/Redis 依赖，服务现在是无状态的，不需要持久化。
- 删除旧前端和静态资源路由，后续前端会重新设计。

## API

启动服务：

```bash
go run .
```

默认监听：

```text
:8081
```

解析歌单：

```http
POST /songlist
Content-Type: application/x-www-form-urlencoded

url=<汽水音乐分享文案或歌单链接>
```

查询参数：

- `detailed=true`：保留页面里的原始歌名；不传时会做基础标准化。
- `format=song-singer`：默认格式，返回 `歌名 - 歌手`。
- `format=singer-song`：返回 `歌手 - 歌名`。
- `format=song`：只返回歌名。
- `order=reverse`：倒序返回。

响应示例：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "name": "歌单标题-创建者",
    "songs": [
      "Song A - Singer A",
      "Song B - Singer B"
    ],
    "songs_count": 2
  }
}
```

错误响应示例：

```json
{
  "code": 400,
  "msg": "不支持的音乐链接格式",
  "data": null
}
```

## 主要改动

相对上游 fork，本版本做了这些整理：

- Web 框架从 Gin 切换到 Hertz。
- 移除数据库、Redis 和历史缓存逻辑。
- 移除网易云音乐支持。
- 移除 QQ 音乐支持，包括 JS/native sign、加密逻辑、请求模型和测试。
- 移除自实现 `misc/log`，改用 Go 标准库 `log/slog`。
- 移除旧 Vue 前端和静态文件服务。
- 汽水音乐解析改为“SSR JSON 优先，DOM fallback”的结构。
- HTTP 工具统一设置 User-Agent 和请求超时。
- 返回码改为语义化的 `ResultCodeOK = 0`、`ResultCodeBadRequest = 400`。
- 补充 PatchConvey 风格单元测试，所有外部依赖都用 mock/fake 数据，不依赖真实用户链接。

## 开发

运行测试：

```bash
go test ./...
```

查看覆盖率：

```bash
go test ./... -cover
```

项目当前没有前端，也没有持久化组件；核心代码集中在：

- `handler/`：Hertz 路由和请求处理。
- `logic/`：汽水音乐链接解析、页面解析和歌单提取。
- `misc/httputil/`：HTTP 请求工具。
- `misc/models/`：响应结构和歌单模型。
- `misc/utils/`：歌名标准化工具。
