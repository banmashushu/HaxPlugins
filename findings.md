# Findings & Decisions

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
