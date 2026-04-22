# HaxPlugins — LOL 海克斯大乱斗辅助插件

> 一个基于 Go + Wails v2 开发的 LOL（英雄联盟）海克斯大乱斗（Hexgates ARAM）辅助插件。

## 项目简介

本项目旨在为 LOL 海克斯大乱斗模式提供数据支持，帮助玩家在选人阶段和游戏内选海克斯时做出更优决策。

### 主要功能

| 阶段 | 场景 | 功能 |
|------|------|------|
| 选人界面 | **客户端选人阶段** | 显示队友英雄胜率、 tier 等级、最优海克斯推荐、出装+符文+技能加点、最佳协同英雄 |
| 游戏内 | **选海克斯时** | Overlay 悬浮窗展示海克斯选项的胜率、选取率、出装建议（待实现）|

### 已集成的数据

- **英雄胜率**：172 英雄 ARAM 胜率 / 选取率 / Tier 等级
- **海克斯推荐**：195 海克斯 × 英雄组合，强度评分 + 胜率 + 选取率
- **出装推荐**：核心装备 + 鞋子 + 符文 + 技能加点顺序
- **协同推荐**：英雄最佳搭档（胜率、评分、场次）
- **装备数据**：247 件嚎哭深渊装备

### 技术栈

- **桌面框架**: Wails v2（Go + React + TypeScript）
- **LCU 连接**: lcu-gopher（与 LOL 客户端通信）
- **数据爬取**: OP.GG MCP API + OP.GG 内部 REST API
- **本地缓存**: SQLite（WAL 模式）
- **图标资源**: DDragon CDN（英雄头像、装备图标）
- **截图 OCR**: kbinani/screenshot + gosseract（Tesseract）（规划中）
- **全局热键**: robotn/gohook（规划中）
- **样式**: Tailwind CSS

### 核心原则

- 不触碰游戏内存（Vanguard 反作弊安全）
- 不注入 DLL / 不 Hook 渲染
- 仅通过截图 OCR + LCU API + 窗口叠加实现

## 快速开始

### 数据初始化

首次使用或补丁更新后，需要运行以下命令初始化数据：

```bash
# 1. 英雄基础数据（DDragon API）
go run ./cmd/initdata

# 2. 海克斯数据（OP.GG MCP API）
go run ./cmd/initaugments

# 3. 出装数据（OP.GG MCP API）
go run ./cmd/initbuilds

# 4. 协同数据（OP.GG MCP API）
go run ./cmd/initsynergies

# 5. 装备数据（OP.GG MCP API）
go run ./cmd/inititems
```

### 运行应用

```bash
# 开发模式（带热重载）
wails dev

# 生产构建
wails build
```

应用启动时会自动连接 LOL 客户端。如果客户端未运行，将自动进入 **Mock 模式**，显示预设英雄数据用于 UI 测试。

## 项目文档

- [技术实现方案](docs/lol-hexgates-plugin-tech-spec.md)
- [开发计划](task_plan.md)
- [进度日志](progress.md)
- [UI 重设计文档](docs/superpowers/specs/2026-04-22-ui-redesign-design.md)

## License

MIT
