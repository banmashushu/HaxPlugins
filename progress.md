# Progress Log

## Session: 2026-04-20 ~ 2026-04-21

### Phase 1: 计划制定
- **Status:** complete
- **Started:** 2026-04-20 15:30
- Actions taken:
  - 读取技术规格文档（docs/lol-hexgates-plugin-tech-spec.md）
  - 检查现有代码状态（main.go, go.mod, .gitignore）
  - 检查环境（Go 1.25.8, Wails v2.11.0）
  - 创建 task_plan.md 开发计划
  - 创建 findings.md 研究发现
  - 创建 progress.md 进度日志
- Files created/modified:
  - task_plan.md (created)
  - findings.md (created)
  - progress.md (created)

### Phase 2: 数据层与爬虫 MVP
- **Status:** in_progress
- **Started:** 2026-04-20 16:00
- Actions taken:
  - 创建数据库操作层：champion.go, augment.go, build.go
  - 实现 DDragon API 客户端（ddragon.go）
  - 实现 OP.GG 爬虫框架（opgg.go, http.go, source.go）
  - 修复 HTTP Accept-Encoding 导致 Brotli 解码失败的问题
  - 创建 cmd/initdata：从 DDragon 导入英雄/装备基础数据
  - 创建 cmd/verify：验证数据源可用性
  - 成功导入 172 英雄 + 695 装备到 SQLite
  - **修复装备中英文名称相同 bug**（ItemInfo 拆 NameEN/NameCN）
  - **重写 OP.GG 英雄胜率爬取**：从 HTML 爬取改为内部 REST API
  - **发现 OP.GG MCP API**：提供 ARAM 海克斯数据（`lol_list_aram_augments`）
  - **创建 cmd/initaugments**：批量导入 195 海克斯 + 16043 条英雄组合数据
- Files created/modified:
  - internal/data/champion.go (created)
  - internal/data/augment.go (created)
  - internal/data/build.go (created)
  - internal/scraper/ddragon.go (created)
  - internal/scraper/opgg.go (created)
  - internal/scraper/http.go (created)
  - internal/scraper/source.go (created)
  - cmd/initdata/main.go (created)
  - cmd/verify/main.go (created)
  - cmd/initaugments/main.go (created)
- Issues:
  - ~~OP.GG 英雄胜率爬取失败~~ → **已解决**：改用内部 API `lol-api-champion.op.gg`
  - ~~海克斯数据来源不明~~ → **已解决**：使用 OP.GG MCP API
  - MCP API performance 分数非胜率，需前端标注为"强度评分"

## Test Results
| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| Go 版本检查 | go version | Go 1.25 | Go 1.25.8 | ✓ |
| Wails 版本检查 | wails version | Wails v2 | Wails v2.11.0 | ✓ |
| DDragon 英雄数据 | FetchChampions | 172 heroes | 172 heroes | ✓ |
| DDragon 装备数据 | FetchItems | 695 items | 695 items | ✓ |
| 数据库初始化 | go run ./cmd/initdata | 数据入库 | 172 heroes, 695 items | ✓ |
| OP.GG 版本检测 | GetCurrentPatch | 16.8.1 | 16.8.1 | ✓ |
| OP.GG 胜率爬取 | ScrapeChampionStats | 英雄列表 | page structure mismatch | ⚠ |

## Error Log
| Timestamp | Error | Attempt | Resolution |
|-----------|-------|---------|------------|
| 2026-04-20 19:45 | Accept-Encoding: br 导致 Brotli 解码失败 | 1 | 移除手动 Accept-Encoding 头 |
| 2026-04-20 20:00 | file already closed (resp.Body) | 1 | 先 io.ReadAll 再 goquery 解析 |

### Phase 1 补完: LCU 客户端集成
- **Status:** complete
- **Started:** 2026-04-21
- Actions taken:
  - 创建 `internal/lcu/client.go`：LCU 客户端封装（连接、事件监听、数据获取）
  - 创建 `internal/lcu/events.go`：事件总线（EventBus）实现
  - 集成 `lcu-gopher` 库（v0.0.3），修复 macOS 编译问题（`syscall.SysProcAttr.HideWindow`）
  - 创建 `cmd/lcudemo/main.go`：LCU 连接测试工具
- Files created/modified:
  - internal/lcu/client.go (created)
  - internal/lcu/events.go (created)
  - cmd/lcudemo/main.go (created)
  - internal/vendor/lcu-gopher/ (vendored + patched)
  - go.mod (added replace directive)
- Issues:
  - `lcu-gopher` Windows-only `HideWindow` 字段导致 macOS 编译失败 → **已解决**：本地 vendor 并移除该字段

## 5-Question Reboot Check
| Question | Answer |
|----------|--------|
| Where am I? | Phase 2 完成，Phase 1 LCU 集成补完，准备进入 Phase 3 |
| Where am I going? | Phase 3: 客户端选人阶段（LCU 连接 + UI） |
| What's the goal? | 开发 LOL 海克斯大乱斗辅助插件 |
| What have I learned? | OP.GG 有动态渲染，需 headless browser 或 API；lcu-gopher 有跨平台编译问题 |
### Phase 3: 客户端选人阶段 MVP
- **Status:** complete (MVP)
- **Started:** 2026-04-21
- Actions taken:
  - 重写 `app.go`：扩展 App 结构体，集成 DB + LCU，添加 4 个绑定 API
  - 实现 LCU 事件推送：游戏阶段变化、选人会话更新通过 `runtime.EventsEmit` 推送到前端
  - 创建 `frontend/src/components/ChampSelectView.tsx`：选人阶段主容器
  - 创建 `frontend/src/components/TeamMemberCard.tsx`：队友卡片（胜率、展开详情）
  - 创建 `frontend/src/components/AugmentList.tsx`：海克斯推荐列表（强度评分、选取率）
  - 创建 `frontend/src/components/BuildPanel.tsx`：出装推荐面板（装备+技能加点）
  - 重写 `frontend/src/App.tsx`：游戏阶段指示器 + 条件渲染
  - 更新 `frontend/src/App.css`：深色主题（LOL 风格），卡片式布局
  - 更新 Wails 绑定文件 `wailsjs/go/main/App.js` / `App.d.ts`
- Files created/modified:
  - app.go (rewritten)
  - main.go (OnShutdown 回调)
  - frontend/src/App.tsx (rewritten)
  - frontend/src/App.css (rewritten)
  - frontend/src/components/ChampSelectView.tsx (created)
  - frontend/src/components/TeamMemberCard.tsx (created)
  - frontend/src/components/AugmentList.tsx (created)
  - frontend/src/components/BuildPanel.tsx (created)
  - frontend/wailsjs/go/main/App.js (updated)
  - frontend/wailsjs/go/main/App.d.ts (updated)
- Test Results:
  - 前端构建：`npm run build` ✓ (tsc + vite build 通过)
  - Go 编译：`go build .` ✓ (包含前端 embed)
  - lcudemo 编译：`go build ./cmd/lcudemo/...` ✓

## 5-Question Reboot Check
| Question | Answer |
|----------|--------|
| Where am I? | Phase 3 MVP 完成，前后端编译通过 |
| Where am I going? | Phase 3 运行时测试（需 LOL 客户端），或 Phase 4 游戏内 Overlay |
| What's the goal? | 开发 LOL 海克斯大乱斗辅助插件 |
| What have I learned? | Wails EventsEmit/EventsOn 可实现 Go→React 实时推送；前端组件拆分保持代码整洁 |
| What have I done? | 数据库层完成，LCU 封装完成，选人阶段 UI 完成，前后端编译通过 |
