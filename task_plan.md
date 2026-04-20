# Task Plan: LOL 海克斯大乱斗辅助插件开发

## Goal
开发一个 LOL 海克斯大乱斗（Hexgates ARAM）辅助插件，通过 LCU API + 截图 OCR + 透明 Overlay 为玩家提供英雄胜率、海克斯推荐、出装建议等数据支持。

## Current Phase
Phase 1

## Phases

### Phase 1: 环境验证与基础骨架（1-2 周）
- [ ] Wails v2 项目初始化（当前已安装 v2.11.0，v3 为 alpha 暂不采用）
- [ ] SQLite 数据库初始化 + 迁移文件
- [ ] lcu-gopher 集成验证（能否连接 LOL 客户端）
- [ ] 前端 React + TypeScript 骨架搭建
- [ ] 基础目录结构创建（internal/ 各包）
- [ ] Zap 日志初始化
- [ ] Viper 配置管理初始化
- **Status:** pending

### Phase 2: 数据层与爬虫 MVP（2-3 周）
- [ ] 设计并实现 SQLite Schema（7 张表）
- [ ] 英雄基础数据初始化（从 DDragon 获取）
- [ ] 海克斯基础数据初始化
- [ ] OP.GG 爬虫实现（英雄胜率、海克斯数据）
- [ ] 手动触发一次数据爬取，填充测试数据
- [ ] 数据版本管理（按补丁版本存储）
- [ ] 验证数据源稳定性（OP.GG 页面结构是否可用）
- **Status:** pending

### Phase 3: 客户端选人阶段（2 周）
- [ ] LCU 游戏阶段监听（Lobby → ChampSelect → InProgress）
- [ ] 获取队友英雄列表（/lol-champ-select/v1/session）
- [ ] 英雄胜率查询 API（Go → SQLite → React）
- [ ] 前端选人界面：队友英雄胜率排行卡片
- [ ] 海克斯推荐 API（按英雄 ID 查询最优海克斯）
- [ ] 出装推荐 API
- [ ] 前端海克斯列表 + 出装树组件
- **Status:** pending

### Phase 4: 游戏内 Overlay（2 周）
- [ ] 透明无边框窗口（Wails window 配置）
- [ ] 窗口置顶 + 定位到 LOL 游戏窗口上方
- [ ] 全局热键 F1（robotn/gohook）
- [ ] 手动查询模式：搜索框 + 海克斯列表
- [ ] 英雄信息缓存（进游戏前缓存当前英雄）
- [ ] 截图 OCR 自动识别海克斯（进阶，可选）
- [ ] 不同分辨率适配配置
- **Status:** pending

### Phase 5: 优化与发布（1-2 周）
- [ ] UI 美化（参考 LOL 游戏风格）
- [ ] 配置面板（热键修改、数据源切换、显示设置）
- [ ] 数据自动更新（定时爬虫）
- [ ] 错误处理与降级方案（连接失败、数据缺失）
- [ ] 打包发布（Windows / macOS）
- **Status:** pending

## Key Questions
1. Wails v3 是 alpha 阶段，当前环境已安装 v2.11.0，是否用 v2？（决策：用 v2，稳定成熟，社区生态完善）
2. OP.GG 页面结构是否稳定？是否有 Cloudflare 反爬？（决策：先实现 MVP 爬虫验证，若失败则考虑其他数据源或手动导入）
3. OCR 中文识别准确率是否足够？（决策：MVP 阶段先用手动查询模式，OCR 作为进阶功能）
4. 是否需要支持 Windows 和 macOS 双平台？（决策：先 macOS，后 Windows）

## Decisions Made
| Decision | Rationale |
|----------|-----------|
| Wails v2 | 当前环境已安装 v2.11.0，v3 为 alpha 不稳定，v2 社区成熟 |
| MVP 先实现手动查询模式 | OCR 准确率和开发成本不确定，手动查询可立即提供价值 |
| 先 macOS 后 Windows | 开发者当前在 macOS，Wails 跨平台打包容易 |
| SQLite 本地缓存 | 第三方数据站无公开 API，海克斯数据变化频率低（每补丁一次） |
| 多数据源冗余 | OP.GG 中文 + U.GG 精确胜率，互为备份 |
| 不读内存、不注入 DLL | Vanguard 反作弊严格，截图+Overlay 是安全方案 |

## Risks & Mitigation
| Risk | Impact | Mitigation |
|------|--------|------------|
| OP.GG 改版/反爬 | 爬虫失效 | 多数据源冗余；先验证 MVP 爬虫 |
| OCR 中文识别率低 | 游戏内体验差 | MVP 先用手动查询；OCR 作为加分项 |
| 不同分辨率 UI 位置不同 | Overlay 错位 | 支持常见分辨率配置；玩家手动校准 |
| LCU API 变更 | 客户端功能失效 | lcu-gopher 社区维护；关注更新 |
| 补丁更新后数据缺失 | 显示旧数据 | 爬虫自动检测新版本；数据过期提示 |
| Vanguard 误判 | 封号风险 | 开源代码透明实现；不碰内存 |

## Errors Encountered
| Error | Attempt | Resolution |
|-------|---------|------------|
|       | 1       |            |

## Notes
- Wails v2 API 与 v3 不同，需参考 v2 文档调整代码
- 技术规格文档中的代码片段基于 v3，实现时需适配 v2
- 每个 Phase 完成后更新状态并记录 progress.md
- 数据源验证是 Phase 2 的关键里程碑，若爬虫失败需及时调整方案
