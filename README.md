# HaxPlugins — LOL 海克斯大乱斗辅助插件

> 一个基于 Go + Wails 开发的 LOL（英雄联盟）海克斯大乱斗（Hexgates ARAM）辅助插件。

## 项目简介

本项目旨在为 LOL 海克斯大乱斗模式提供数据支持，帮助玩家在选人阶段和游戏内选海克斯时做出更优决策。

### 主要功能

| 阶段 | 场景 | 功能 |
|------|------|------|
| 阶段一 | **选人界面** | 显示队友英雄胜率排行，选中英雄后推荐最优海克斯 |
| 阶段二 | **游戏内选海克斯** | Overlay 悬浮窗展示海克斯选项的胜率、选取率、出装建议 |

### 技术栈

- **桌面框架**: Wails v3（Go + React）
- **LCU 连接**: lcu-gopher（与 LOL 客户端通信）
- **数据爬取**: Go net/http + goquery
- **本地缓存**: SQLite
- **截图 OCR**: kbinani/screenshot + gosseract（Tesseract）
- **全局热键**: robotn/gohook

### 核心原则

- 不触碰游戏内存（Vanguard 反作弊安全）
- 不注入 DLL / 不 Hook 渲染
- 仅通过截图 OCR + LCU API + 窗口叠加实现

## 项目文档

- [技术实现方案](docs/lol-hexgates-plugin-tech-spec.md)

## License

MIT