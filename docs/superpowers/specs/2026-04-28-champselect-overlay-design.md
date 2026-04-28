# 选人阶段嵌入式提示面板 — 设计文档

## Context

用户需要在 LOL 海克斯大乱斗的选人阶段获得实时数据提示：英雄梯度评级、胜率、推荐海克斯。现有前端是一个全尺寸窗口，不符合"嵌入式提示"的轻量定位。本次完全重建选人阶段 UI，改为紧凑浮动面板，仅显示我方队伍数据。

## 窗口配置

- **类型**: Wails 无边框置顶窗口，340×520px 固定尺寸
- **外观**: 半透明深色背景 (`rgba(6,14,26,0.92)`)，圆角 12px，玻璃拟态 `backdrop-filter: blur(16px)`
- **定位**: 始终置顶 (`AlwaysOnTop: true`)，用户可拖拽标题栏移动
- **背景色**: `#00000000` 支持透明

## 布局结构

```
┌──────────────────────┐
│  海克斯大乱斗 · 我方  │  标题栏 (36px, 可拖拽)
├──────────────────────┤
│                      │
│  ┌────┬────┬────┐   │
│  │ S  │ A  │ B  │   │  英雄网格 (2行, 自适应)
│  │ 🧙 │ ⚔️ │ 🏹 │   │
│  └────┴────┴────┘   │
│  ┌────┬────┬────┐   │
│  │ A  │ S  │     │   │
│  │ 🛡️ │ 💀 │     │   │
│  └────┴────┴────┘   │
│                      │
├──────────────────────┤
│  ┌────────────────┐  │
│  │ 🧙 阿狸  S级   │  │  详情区 (选中后展开)
│  │ 胜率 52.3%     │  │
│  │ 选取率 18.7%   │  │
│  │                │  │
│  │ ⭐ 推荐海克斯   │  │
│  │ 召唤专家 61.2% │  │
│  │ 法术之刃 58.7% │  │
│  │ 启迪之光 56.1% │  │
│  │                │  │
│  │ [海克斯][出装] │  │  底部标签切换
│  └────────────────┘  │
└──────────────────────┘
```

## 组件设计

### ChampSelectPanel（主面板容器）
- **路径**: `frontend/src/panels/ChampSelectPanel.tsx`
- **职责**: 监听 `game:champselect` 事件，调用 `GetMyTeamStats()`，管理选中英雄状态
- **状态**: `heroes: TeamMember[]`, `selectedHero: TeamMember | null`, `loading: boolean`
- **事件**: `EventsOn("game:champselect", handleUpdate)`
- **无数据时**: 显示"等待进入选人阶段..."占位

### HeroGrid（英雄网格）
- **路径**: `frontend/src/panels/HeroGrid.tsx`
- **职责**: 渲染英雄卡片网格，处理选中/取消选中
- **Props**: `heroes: TeamMember[]`, `selectedId: number | null`, `onSelect: (id: number) => void`
- **布局**: `grid grid-cols-3 gap-2`（3列），自动换行

### HeroCard（单英雄卡片）
- **路径**: `frontend/src/panels/HeroCard.tsx`
- **职责**: 显示英雄头像 + 梯度角标 + 名称 + 胜率
- **Props**: `hero: TeamMember`, `selected: boolean`, `onClick: () => void`
- **选中态**: 金色发光边框 (`shadow-glow-gold`)
- **头像**: DDragon CDN 加载 (`ddragon.championIconURL(nameEn)`)，圆角 8px
- **梯度角标**: 右下角覆盖，颜色映射 S=金, A=紫, B=蓝, C=绿, D=灰

### HeroDetail（选中英雄详情）
- **路径**: `frontend/src/panels/HeroDetail.tsx`
- **职责**: 展示选中英雄的完整数据：胜率、选取率、海克斯/出装/协同推荐
- **Props**: `hero: TeamMember | null`
- **标签切换**: `augments | build | synergies` 三标签
- **动画**: 内容区 `transition-all duration-300`，展开时 `max-height` 过渡
- **空态**: `hero` 为 null 时不渲染

## 数据流

```
LCU → Go 后端 (app.go) → EventsEmit("game:champselect")
    → ChampSelectPanel (useEffect 监听)
        → GetMyTeamStats() 返回 TeamMemberStats[]
            → HeroGrid (heroes prop)
                → HeroCard × N
            → HeroDetail (selectedHero prop)
```

**复用现有 API**: `GetMyTeamStats()` 已返回所有需要数据（梯度、胜率、海克斯、出装、协同），无需新增后端代码。

## 梯度颜色映射

| 梯度 | 颜色 | Tailwind Class |
|------|------|---------------|
| S | 金色 `#c8aa6e` | `bg-tier-gold text-black` |
| A | 紫色 `#b48ef0` | `bg-tier-a text-white` |
| B | 蓝色 `#0397ab` | `bg-tier-b text-white` |
| C | 绿色 `#2dd4a0` | `bg-tier-c text-black` |
| D | 灰色 `#5a6a7e` | `bg-tier-d text-white` |

## 调整现有 App.tsx

`App.tsx` 阶段路由简化：
- `ChampSelect` 阶段 → 渲染 `<ChampSelectPanel />`（替代旧的 `<ChampSelectView />`）
- 其他阶段不变（保留现有占位）

## 不做的事

- 不修改任何 Go 后端代码
- 不删除现有前端组件文件
- 不新增数据库表或迁移
- 不实现游戏内海克斯识别（场景2，后续迭代）

## 文件清单

| 文件 | 操作 |
|------|------|
| `frontend/src/App.tsx` | 修改：ChampSelect 阶段路由指向新面板 |
| `frontend/src/panels/ChampSelectPanel.tsx` | 新增 |
| `frontend/src/panels/HeroGrid.tsx` | 新增 |
| `frontend/src/panels/HeroCard.tsx` | 新增 |
| `frontend/src/panels/HeroDetail.tsx` | 新增 |
| `frontend/src/shared/types.ts` | 新增 |
| `frontend/src/style.css` | 修改：清理旧样式，添加新面板所需样式 |
| `main.go` | 修改：窗口配置（无边框、置顶、尺寸） |

## 验证方法

1. **无客户端测试**: 使用 Mock 模式启动应用，确认面板渲染正常
2. **功能验证**: 点击不同英雄卡片，确认详情区切换正确
3. **标签切换**: 在详情区切换海克斯/出装/协同标签
4. **窗口行为**: 确认无边框窗口可拖拽移动、始终置顶
5. **Go 编译**: `go build -o /dev/null ./...` 确保无编译错误
6. **TypeScript 检查**: `cd frontend && npx tsc --noEmit` 确保无类型错误