# Findings & Decisions

## OP.GG Internal API Discovery

发现于 2026-04-21，解决 OP.GG HTML 爬取失效问题。

### Discovery Process

1. **HTML scraping failed** — OP.GG 使用 Next.js streaming SSR，初始 HTML 表格为空，数据通过 `__next_f.push` 流式注入，goquery 解析不到任何内容。
2. **Headless browser blocked** — 尝试 `go-rod/rod` 模拟浏览器，被 Cloudflare 返回 403。
3. **Reference project** — 用户提供的 OPGG.py 源码中出现了 `https://lol-api-champion.op.gg/api` 域名，证实 OP.GG 前端不走 SSR，而是直接调用内部 REST API。
4. **Direct validation** — curl 验证确认 API 可访问、返回结构化 JSON。

### API Endpoints

| Endpoint | URL | Purpose |
|----------|-----|---------|
| Champion Meta | `GET https://lol-api-champion.op.gg/api/meta/champions?hl=zh_CN` | 英雄元数据（ID、英文 key、中文名） |
| Champion Stats | `GET https://lol-api-champion.op.gg/api/{region}/champions/{mode}?tier=all&hl=zh_CN` | 英雄胜率、登场率、禁用率、Tier |

**Response format (stats):**
```json
{
  "data": [
    {
      "id": 86,
      "tier": "3",
      "rank": 51,
      "average_stats": {
        "play": 3309,
        "win_rate": 0.511333,
        "pick_rate": 0.0660215,
        "ban_rate": null,
        "kda": 2.892337,
        "tier": 3,
        "rank": 51
      }
    }
  ]
}
```

### Request Requirements

| Header | Value | Required |
|--------|-------|----------|
| `Accept` | `application/json` | Yes |
| `Referer` | `https://op.gg/` | Yes (for stats endpoint) |
| `User-Agent` | Browser UA | Yes (via `newRequest`) |

### Known Limitations

| Issue | Detail |
|-------|--------|
| `version` param | `version=16.8.1` causes **422 Unprocessable Entity**. Do NOT pass version to stats API. |
| `aram-mayhem` mode | Returns `{"message":"Mode was invalid"}`. Must use `aram` mode. |
| Region | `NA` works globally. Other regions not tested. |
| Stability | Internal API, OP.GG may change without notice. |

### Mode Mapping

```
hexgates      → aram
aram-mayhem   → aram
```

> Note: Hexgates (海克斯传送门) has been a permanent ARAM mechanic since S14. OP.GG no longer distinguishes between "classic ARAM" and "aram-mayhem" at the API level.

### Implementation

See `internal/scraper/opgg.go`:
- `fetchChampionMeta()` — calls meta endpoint, builds `id → {key, name}` map
- `ScrapeChampionStats()` — calls stats endpoint, merges with meta map
- No headless browser needed; direct HTTP client works

## Requirements
- 开发 LOL 海克斯大乱斗辅助插件
- 阶段一：客户端选人界面显示英雄胜率排行 + 海克斯推荐 + 出装建议
- 阶段二：游戏内 Overlay 悬浮窗展示海克斯选项胜率
- 核心约束：不触碰游戏内存，仅通过截图 OCR + LCU API + 窗口叠加实现
- 主语言：Go，桌面框架 Wails

## Research Findings
- 当前环境：Go 1.25.8 + Wails v2.11.0（macOS/arm64）
- 技术规格文档使用 Wails v3 API，但 v3 为 alpha 阶段，实际应使用 v2
- Wails v2 与 v3 API 差异较大，代码需适配
- lcu-gopher 是 Go 社区维护的 LCU 客户端库
- OP.GG 有独立的 aram-mayhem 中文页面，但有 Cloudflare 反爬风险
- 截图 OCR 方案（kbinani/screenshot + gosseract）在 macOS 上可行

## Technical Decisions
| Decision | Rationale |
|----------|-----------|
| Wails v2 | 环境已安装 v2.11.0，v3 alpha 不稳定 |
| MVP 手动查询 | OCR 开发成本高，手动查询可立即提供价值 |
| 先 macOS | 开发者当前平台，Wails 跨平台打包容易 |
| SQLite 本地缓存 | 第三方无公开 API，数据变化频率低 |
| 多数据源冗余 | OP.GG 中文 + U.GG 精确胜率 |

## Issues Encountered
| Issue | Resolution |
|-------|------------|
| Wails v3 vs v2 | 使用 v2，文档中 v3 代码需适配 |

## Resources
- Wails v2 Docs: https://wails.io/docs/gettingstarted/installation
- lcu-gopher: https://github.com/Its-Haze/lcu-gopher
- OP.GG aram-mayhem: https://op.gg/zh-cn/lol/modes/aram-mayhem
- 技术规格文档: docs/lol-hexgates-plugin-tech-spec.md

## Visual/Browser Findings
- 技术规格文档包含完整的数据库 Schema、代码架构、开发阶段划分
- 项目目录结构设计清晰（internal/app, lcu, overlay, data, scraper, service 等）
