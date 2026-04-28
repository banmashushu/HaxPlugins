# 选人阶段嵌入式提示面板 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将选人阶段 UI 从全尺寸页面重建为紧凑浮动面板（340×520px），只显示我方队伍数据。

**Architecture:** 4 个新组件（ChampSelectPanel → HeroGrid → HeroCard + HeroDetail）通过 App.tsx 阶段路由挂载，复用现有 `GetMyTeamStats()` API 和 `ddragon.ts` 工具。

**Tech Stack:** React 18 + TypeScript + Tailwind CSS 3, Wails v2.11.0 (Go), DDragon CDN

---

### Task 1: Update Wails window configuration

**Files:**
- Modify: `main.go:19-21`

- [ ] **Step 1: Update main.go window options**

Change the window from 720×360 framed to 340×520 frameless, always-on-top with transparent background:

```go
// main.go — replace lines 19-31
err := wails.Run(&options.App{
    Title:          "haxPlugins",
    Width:          340,
    Height:         520,
    DisableResize:  true,
    Frameless:      true,
    AlwaysOnTop:    true,
    AssetServer: &assetserver.Options{
        Assets: assets,
    },
    BackgroundColour: &options.RGBA{R: 6, G: 14, B: 26, A: 255},
    OnStartup:        app.startup,
    OnShutdown:       app.shutdown,
    Bind: []interface{}{
        app,
    },
})
```

- [ ] **Step 2: Verify Go build**

Run: `cd /Users/zhangchuan/GolandProjects/HaxPlugins && go build -o /dev/null ./...`
Expected: PASS (no errors)

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: configure frameless always-on-top window for overlay panel"
```

---

### Task 2: Update Tailwind tier colors to match new spec

**Files:**
- Modify: `frontend/tailwind.config.js:30-39`

- [ ] **Step 1: Replace tier color definitions**

```js
// frontend/tailwind.config.js — replace tier block (lines 30-39)
tier: {
  s: '#c8aa6e',      // gold for S tier
  a: '#b48ef0',      // purple for A tier
  b: '#0397ab',      // blue for B tier
  c: '#2dd4a0',      // green for C tier
  d: '#5a6a7e',      // gray for D tier
  prismatic: '#b48ef0',
  gold: '#f0d878',
  silver: '#a8b4c2',
},
```

- [ ] **Step 2: Commit**

```bash
git add frontend/tailwind.config.js
git commit -m "refactor: update tier colors to match new overlay panel spec"
```

---

### Task 3: Create shared types

**Files:**
- Create: `frontend/src/shared/types.ts`

- [ ] **Step 1: Create types file**

```typescript
// frontend/src/shared/types.ts
// Shared type re-exports — single source of truth for all panels.
// Wraps the auto-generated Wails models for convenience.

import { data, main } from "../wailsjs/go/models";

export type TeamMember = main.TeamMemberStats;
export type HeroAugmentStat = data.HeroAugmentStat;
export type Build = data.Build;
export type ChampionSynergy = data.ChampionSynergy;
export type BuildItem = data.BuildItem;

export type TierLabel = 'S' | 'A' | 'B' | 'C' | 'D' | '';

export const TIER_COLORS: Record<string, string> = {
  S: 'bg-tier-s text-black',
  A: 'bg-tier-a text-white',
  B: 'bg-tier-b text-white',
  C: 'bg-tier-c text-black',
  D: 'bg-tier-d text-white',
};

export const TIER_BORDER_COLORS: Record<string, string> = {
  S: 'border-tier-s shadow-[0_0_12px_rgba(200,170,110,0.3)]',
  A: 'border-tier-a shadow-[0_0_12px_rgba(180,142,240,0.3)]',
  B: 'border-tier-b shadow-[0_0_12px_rgba(3,151,171,0.3)]',
  C: 'border-tier-c shadow-[0_0_12px_rgba(45,212,160,0.3)]',
  D: 'border-tier-d shadow-[0_0_12px_rgba(90,106,126,0.3)]',
};

export function formatPercent(n: number): string {
  if (n == null) return '--';
  return (n * 100).toFixed(1) + '%';
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd /Users/zhangchuan/GolandProjects/HaxPlugins/frontend && npx tsc --noEmit src/shared/types.ts 2>&1 || true`
Expected: No errors from this file (may have unrelated errors)

- [ ] **Step 3: Commit**

```bash
git add frontend/src/shared/types.ts
git commit -m "feat: add shared types and tier color map for overlay panel"
```

---

### Task 4: Create HeroCard component

**Files:**
- Create: `frontend/src/panels/HeroCard.tsx`

- [ ] **Step 1: Create HeroCard**

```typescript
// frontend/src/panels/HeroCard.tsx
import { TeamMember, TIER_COLORS, TIER_BORDER_COLORS, formatPercent } from "../shared/types";
import { championIconURL } from "../utils/ddragon";

interface HeroCardProps {
  hero: TeamMember;
  selected: boolean;
  onClick: () => void;
}

export default function HeroCard({ hero, selected, onClick }: HeroCardProps) {
  const tierClass = TIER_COLORS[hero.tier] || 'bg-lol-muted text-white';
  const borderClass = selected
    ? TIER_BORDER_COLORS[hero.tier] || 'border-lol-border-glow'
    : 'border-transparent';

  return (
    <button
      onClick={onClick}
      className={`
        relative flex flex-col items-center p-1.5 rounded-lg
        bg-lol-card/80 border-2 ${borderClass}
        transition-all duration-200 hover:bg-lol-card-hover
        focus:outline-none
        ${selected ? 'scale-105' : 'hover:scale-[1.02]'}
      `}
    >
      {/* Hero avatar */}
      <div className="relative w-14 h-14 rounded-md overflow-hidden">
        <img
          src={championIconURL(hero.champion_name_en)}
          alt={hero.champion_name}
          className="w-full h-full object-cover"
          loading="lazy"
        />
        {/* Tier badge bottom-right */}
        <span
          className={`
            absolute -bottom-0.5 -right-0.5
            text-[10px] font-bold px-1 py-px rounded
            leading-none ${tierClass}
          `}
        >
          {hero.tier || '?'}
        </span>
      </div>

      {/* Name + winrate */}
      <span className="text-[11px] text-lol-text font-semibold mt-1 leading-tight truncate max-w-full">
        {hero.champion_name}
      </span>
      <span className="text-[10px] text-lol-green font-medium">
        {formatPercent(hero.winrate)}
      </span>
    </button>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/panels/HeroCard.tsx
git commit -m "feat: add HeroCard component with tier badge and selection state"
```

---

### Task 5: Create HeroGrid component

**Files:**
- Create: `frontend/src/panels/HeroGrid.tsx`

- [ ] **Step 1: Create HeroGrid**

```typescript
// frontend/src/panels/HeroGrid.tsx
import { TeamMember } from "../shared/types";
import HeroCard from "./HeroCard";

interface HeroGridProps {
  heroes: TeamMember[];
  selectedId: number | null;
  onSelect: (hero: TeamMember) => void;
}

export default function HeroGrid({ heroes, selectedId, onSelect }: HeroGridProps) {
  if (heroes.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-xs text-lol-muted">
        等待英雄数据...
      </div>
    );
  }

  return (
    <div className="grid grid-cols-3 gap-2 px-3">
      {heroes.map((hero) => (
        <HeroCard
          key={hero.cell_id || hero.champion_id}
          hero={hero}
          selected={hero.champion_id === selectedId}
          onClick={() => onSelect(hero)}
        />
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/panels/HeroGrid.tsx
git commit -m "feat: add HeroGrid component for team champion grid layout"
```

---

### Task 6: Create HeroDetail component

**Files:**
- Create: `frontend/src/panels/HeroDetail.tsx`

- [ ] **Step 1: Create HeroDetail**

```typescript
// frontend/src/panels/HeroDetail.tsx
import { useState } from "react";
import {
  TeamMember, HeroAugmentStat, Build, ChampionSynergy,
  TIER_COLORS, formatPercent,
} from "../shared/types";
import { championIconURL, itemIconURL } from "../utils/ddragon";

type TabId = 'augments' | 'build' | 'synergies';

interface HeroDetailProps {
  hero: TeamMember | null;
}

export default function HeroDetail({ hero }: HeroDetailProps) {
  const [tab, setTab] = useState<TabId>('augments');

  if (!hero) {
    return (
      <div className="flex items-center justify-center py-6 text-xs text-lol-muted">
        点击英雄查看详情
      </div>
    );
  }

  const tierClass = TIER_COLORS[hero.tier] || '';

  return (
    <div className="px-3 pb-3 animate-[slideUp_0.25s_ease-out]">
      {/* Hero header */}
      <div className="flex items-center gap-3 mb-3">
        <img
          src={championIconURL(hero.champion_name_en)}
          alt={hero.champion_name}
          className="w-12 h-12 rounded-lg border border-lol-border/50"
        />
        <div>
          <div className="flex items-center gap-2">
            <span className="text-sm font-bold text-lol-text-bright">
              {hero.champion_name}
            </span>
            <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded leading-none ${tierClass}`}>
              {hero.tier || '?'}级
            </span>
          </div>
          <div className="flex gap-3 mt-0.5">
            <span className="text-xs text-lol-green font-semibold">
              胜率 {formatPercent(hero.winrate)}
            </span>
            <span className="text-xs text-lol-muted">
              选取率 {formatPercent(hero.pickrate)}
            </span>
          </div>
        </div>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 mb-2 border-b border-lol-border/30">
        {([
          ['augments', '海克斯'],
          ['build', '出装'],
          ['synergies', '协同'],
        ] as [TabId, string][]).map(([id, label]) => (
          <button
            key={id}
            onClick={() => setTab(id)}
            className={`
              text-xs font-semibold px-2.5 py-1.5 rounded-t
              transition-colors duration-150
              ${tab === id
                ? 'bg-lol-card text-lol-gold border-b-2 border-lol-gold'
                : 'text-lol-muted hover:text-lol-text'}
            `}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="min-h-[60px]">
        {tab === 'augments' && <AugmentTab augments={hero.augments || []} />}
        {tab === 'build' && <BuildTab build={hero.build} />}
        {tab === 'synergies' && <SynergyTab synergies={hero.synergies || []} />}
      </div>
    </div>
  );
}

function AugmentTab({ augments }: { augments: HeroAugmentStat[] }) {
  if (augments.length === 0) {
    return <div className="text-xs text-lol-muted py-2">暂无海克斯推荐数据</div>;
  }

  const top3 = augments.slice(0, 3);

  return (
    <div className="space-y-1.5">
      {top3.map((a, i) => {
        const tierClass = TIER_COLORS[a.tier] || 'bg-lol-muted text-white';
        return (
          <div
            key={a.augment_id}
            className="flex items-center justify-between px-2 py-1.5 rounded bg-lol-card/50 border border-lol-border/20"
          >
            <div className="flex items-center gap-2">
              <span className="text-[10px] font-bold text-lol-gold w-4">
                #{i + 1}
              </span>
              <span className="text-xs text-lol-text">{a.augment_name_cn || a.augment_name}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className={`text-[10px] font-bold px-1 rounded leading-none ${tierClass}`}>
                {a.tier || '?'}
              </span>
              <span className="text-xs text-lol-green font-semibold w-12 text-right">
                {formatPercent(a.winrate)}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function BuildTab({ build }: { build?: Build }) {
  if (!build) {
    return <div className="text-xs text-lol-muted py-2">暂无出装推荐数据</div>;
  }

  const items = [...(build.items || [])];
  if (build.boots) items.push(build.boots);

  return (
    <div className="space-y-2">
      {/* Items grid */}
      <div className="flex gap-1.5 flex-wrap">
        {items.map((item) => (
          <div
            key={item.item_id}
            className="relative w-8 h-8 rounded bg-lol-card/60 border border-lol-border/30 overflow-hidden"
            title={`${item.name_cn} (${formatPercent(item.winrate)})`}
          >
            <img
              src={itemIconURL(item.item_id)}
              alt={item.name_cn}
              className="w-full h-full object-cover"
              loading="lazy"
            />
          </div>
        ))}
      </div>
      {/* Skill order */}
      {build.skill_order && build.skill_order.length > 0 && (
        <div className="flex items-center gap-1">
          <span className="text-[10px] text-lol-muted">技能:</span>
          {build.skill_order.map((s, i) => (
            <span key={i} className="text-[11px] font-bold text-lol-text-bright bg-lol-card/60 px-1 rounded">
              {s}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function SynergyTab({ synergies }: { synergies: ChampionSynergy[] }) {
  if (synergies.length === 0) {
    return <div className="text-xs text-lol-muted py-2">暂无协同推荐数据</div>;
  }

  return (
    <div className="space-y-1.5">
      {synergies.map((s) => (
        <div
          key={s.synergy_champion_id}
          className="flex items-center justify-between px-2 py-1.5 rounded bg-lol-card/50 border border-lol-border/20"
        >
          <div className="flex items-center gap-2">
            <img
              src={championIconURL(s.synergy_name)}
              alt={s.synergy_name}
              className="w-6 h-6 rounded"
              loading="lazy"
            />
            <span className="text-xs text-lol-text">{s.synergy_name}</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-[10px] text-lol-muted">
              {s.play}场
            </span>
            <span className="text-xs text-lol-green font-semibold w-12 text-right">
              {formatPercent(s.win_rate)}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/panels/HeroDetail.tsx
git commit -m "feat: add HeroDetail component with augments/build/synergies tabs"
```

---

### Task 7: Create ChampSelectPanel container

**Files:**
- Create: `frontend/src/panels/ChampSelectPanel.tsx`

- [ ] **Step 1: Create ChampSelectPanel**

```typescript
// frontend/src/panels/ChampSelectPanel.tsx
import { useEffect, useState, useCallback } from "react";
import { EventsOn } from "../wailsjs/runtime";
import { GetMyTeamStats } from "../wailsjs/go/main/App";
import { TeamMember } from "../shared/types";
import HeroGrid from "./HeroGrid";
import HeroDetail from "./HeroDetail";

export default function ChampSelectPanel() {
  const [heroes, setHeroes] = useState<TeamMember[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await GetMyTeamStats();
      setHeroes((data || []).map((d) => typeof d === 'string' ? JSON.parse(d) : d));
      if (data && data.length > 0 && selectedId == null) {
        setSelectedId(data[0].champion_id);
      }
    } catch {
      setError('无法获取队伍数据');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
    const unsub = EventsOn("game:champselect", () => loadData());
    return () => unsub();
  }, [loadData]);

  const selectedHero = heroes.find((h) => h.champion_id === selectedId) || null;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="flex flex-col items-center gap-2">
          <div className="w-6 h-6 border-2 border-lol-gold/30 border-t-lol-gold rounded-full animate-spin" />
          <span className="text-xs text-lol-muted">加载队伍数据...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-xs text-lol-red">{error}</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Hero grid section */}
      <div className="flex-shrink-0 py-3">
        <HeroGrid
          heroes={heroes}
          selectedId={selectedId}
          onSelect={(h) => setSelectedId(h.champion_id)}
        />
      </div>

      {/* Divider */}
      <div className="mx-3 border-t border-lol-border/30" />

      {/* Detail section */}
      <div className="flex-1 overflow-y-auto py-2">
        <HeroDetail hero={selectedHero} />
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/panels/ChampSelectPanel.tsx
git commit -m "feat: add ChampSelectPanel container with LCU event handling"
```

---

### Task 8: Update App.tsx to route to new panel

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Replace ChampSelectView import with ChampSelectPanel**

```typescript
// frontend/src/App.tsx — line 5, replace:
import ChampSelectView from "./components/ChampSelectView";
// with:
import ChampSelectPanel from "./panels/ChampSelectPanel";
```

- [ ] **Step 2: Replace ChampSelect route render**

In `renderContent()`, replace line 76:
```typescript
// Replace:
return <ChampSelectView />;
// With:
return <ChampSelectPanel />;
```

- [ ] **Step 3: Update header title to reflect new design**

Replace the header span text from `HaxPlugins` to `海克斯大乱斗`:

```typescript
// Line 108, replace:
<span className="text-xs font-bold text-gold-shimmer tracking-widest uppercase">HaxPlugins</span>
// With:
<span className="text-xs font-bold text-gold-shimmer tracking-wider">海克斯大乱斗</span>
```

- [ ] **Step 4: Add drag region for frameless window**

On the header div (line 106), add `drag-region` class for frameless window dragging:

```typescript
// Line 106, change className from:
<div className="flex items-center gap-2.5 px-3 py-2 bg-header-gradient border-b border-lol-border/60 flex-shrink-0">
// To:
<div className="flex items-center gap-2.5 px-3 py-2 bg-header-gradient border-b border-lol-border/60 flex-shrink-0 drag-region">
```

- [ ] **Step 5: Verify TypeScript compiles**

Run: `cd /Users/zhangchuan/GolandProjects/HaxPlugins/frontend && npx tsc --noEmit 2>&1 | head -20`
Expected: No new errors from modified files.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat: route ChampSelect phase to new overlay panel"
```

---

### Task 9: Update style.css for frameless window and new panel

**Files:**
- Modify: `frontend/src/style.css`

- [ ] **Step 1: Add frameless window and panel styles**

Append to the end of `frontend/src/style.css`:

```css
/* --- Overlay Panel Styles --- */

/* Drag region for frameless window titlebar */
.drag-region {
  -webkit-app-region: drag;
}
.drag-region button,
.drag-region input,
.drag-region [role="button"] {
  -webkit-app-region: no-drag;
}

/* Slide-up animation for detail panel */
@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Panel body rounded corners (main.go sets transparent bg, CSS handles visual rounding) */
.panel-body {
  background: rgba(6, 14, 26, 0.94);
  border-radius: 12px;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(30, 52, 80, 0.5);
}

/* Subtle inset shadow for depth */
.panel-body::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 12px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
  pointer-events: none;
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/style.css
git commit -m "feat: add frameless window drag region and panel animation styles"
```

---

### Task 10: Verification — build and review

- [ ] **Step 1: Go build check**

Run: `cd /Users/zhangchuan/GolandProjects/HaxPlugins && go build -o /dev/null ./...`
Expected: PASS

- [ ] **Step 2: TypeScript check**

Run: `cd /Users/zhangchuan/GolandProjects/HaxPlugins/frontend && npx tsc --noEmit 2>&1`
Expected: No new errors

- [ ] **Step 3: Frontend build check**

Run: `cd /Users/zhangchuan/GolandProjects/HaxPlugins/frontend && npm run build 2>&1`
Expected: PASS, dist/ output generated

- [ ] **Step 4: Git status**

Run: `git status` — verify no unintended files changed

---

### Verification summary (manual testing with app running)

1. **Mock 模式启动**: 启动应用，确认无边框浮动窗口出现，340×520px
2. **窗口拖拽**: 拖拽标题栏确认窗口可移动
3. **窗口置顶**: 打开其他窗口确认本窗口始终在前
4. **英雄网格**: 点击不同英雄卡片确认选中态切换
5. **详情区**: 确认选中英雄后详情区展开，显示胜率 + 海克斯推荐
6. **标签切换**: 在详情区切换海克斯/出装/协同标签
7. **LCU 实时连接**: 连接 LCU 时确认自动加载队伍数据
8. **无数据状态**: 断开局时确认显示"等待进入选人阶段"
