# LOL 海克斯大乱斗插件 — 技术实现方案

## 1. 项目概述

### 1.1 目标
开发一个 LOL 海克斯大乱斗（Hexgates ARAM）辅助插件，在以下两个阶段提供数据支持：

| 阶段 | 场景 | 功能 |
|------|------|------|
| 阶段一 | **选人界面**（客户端） | 显示队友英雄胜率排行，选中英雄后推荐最优海克斯 |
| 阶段二 | **游戏内选海克斯** | Overlay 悬浮窗展示当前 3 个海克斯选项的胜率、选取率、出装建议 |

### 1.2 核心约束
- **主语言：Go**
- **不触碰游戏内存**（Vanguard 反作弊）
- **不注入 DLL / 不 Hook 渲染**
- 仅通过**截图 OCR** + **LCU API** + **窗口叠加**实现

---

## 2. 技术选型

### 2.1 技术栈总览

| 层级 | 技术 | 版本/说明 |
|------|------|----------|
| **桌面框架** | Wails v2 | Go 后端 + React/TypeScript 前端 |
| **LCU 连接** | lcu-gopher (vendored) | 自动检测 LOL 客户端，HTTP + WebSocket |
| **数据爬取** | OP.GG MCP API + 内部 REST API | `mcp-api.op.gg/mcp` JSON-RPC 2.0 |
| **本地缓存** | SQLite + mattn/go-sqlite3 | WAL 模式，自动迁移 |
| **图标资源** | DDragon CDN | 英雄头像、装备图标 |
| **游戏内截图** | kbinani/screenshot | 跨平台截图（规划中）|
| **OCR 识别** | gosseract (Tesseract) | 识别海克斯文字（规划中）|
| **全局热键** | robotn/gohook | F1 呼出/隐藏 Overlay（规划中）|
| **窗口管理** | Wails 窗口 API | 置顶、透明、无边框（规划中）|
| **配置管理** | Viper | 热键、数据源偏好、显示设置 |
| **日志** | Zap | 结构化日志 |
| **样式** | Tailwind CSS | 深色主题（LOL 风格）|

### 2.2 技术选型理由

**为什么选 Wails 而不是 Fyne？**
- 需要展示复杂表格（英雄胜率排行）+ 卡片式海克斯对比 + 出装树
- Wails 支持任意前端技术栈，React + Tailwind 做数据可视化更方便
- 透明背景 + 无边框窗口更适合游戏 Overlay

**为什么选截图 OCR 而不是读内存？**
- Vanguard 会扫描所有进程，读内存 = 封号风险
- 截图 + OCR 是视觉分析，与玩家自己看屏幕没有本质区别，风险极低
- 类似 Discord Overlay、NVIDIA GeForce Experience 浮窗的实现原理

**为什么选 SQLite 而不是远程 API？**
- 第三方数据站（OP.GG、U.GG、Lolalytics）没有公开 API
- 海克斯数据变化频率不高（每补丁一次），本地缓存 + 定时更新足够
- 离线可用，不依赖网络延迟

---

## 3. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        用户桌面层                             │
│  ┌──────────────┐        ┌─────────────────────────────┐   │
│  │  LOL 客户端   │◄──────►│      Go 后端服务 (Wails)     │   │
│  │              │  LCU   │                             │   │
│  │  选人界面     │  API   │  ┌─────────┐  ┌─────────┐  │   │
│  └──────────────┘        │  │ LCU连接  │  │ 数据爬取 │  │   │
│         │                │  │  模块   │  │  服务   │  │   │
│         │ 进入游戏        │  └────┬────┘  └────┬────┘  │   │
│         ▼                │       │            │       │   │
│  ┌──────────────┐        │  ┌────┴────────────┴────┐  │   │
│  │  LOL 游戏窗口 │        │  │     SQLite 缓存       │  │   │
│  │              │◄──────►│  │  (英雄/海克斯/出装)   │  │   │
│  │  海克斯选择   │ 截图   │  └──────────────────────┘  │   │
│  └──────────────┘ OCR    │                             │   │
│         ▲                │  ┌─────────────────────────┐│   │
│         │ 叠加显示        │  │   透明 Overlay 窗口     ││   │
│  ┌──────────────┐        │  │  (React + Wails 前端)   ││   │
│  │  悬浮窗 (Wails)│◄──────►│  │  胜率/出装/热键响应     ││   │
│  └──────────────┘        │  └─────────────────────────┘│   │
│                          └─────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. 详细实现方案

### 4.1 阶段一：客户端选人阶段

#### 4.1.1 连接 LCU 并监听选人事件

```go
package lcu

import (
    "github.com/Its-Haze/lcu-gopher"
)

type Client struct {
    lcuClient *lcu.Client
    eventBus  *EventBus
}

func NewClient() (*Client, error) {
    config := lcu.DefaultConfig()
    config.AwaitConnection = true  // 等待 LOL 启动

    client, err := lcu.NewClient(config)
    if err != nil {
        return nil, err
    }

    return &Client{
        lcuClient: client,
        eventBus:  NewEventBus(),
    }, nil
}

func (c *Client) StartListening() {
    // 监听游戏阶段变化
    c.lcuClient.SubscribeToGamePhase(func(phase lcu.GamePhase) {
        switch phase {
        case lcu.GamePhaseChampSelect:
            c.eventBus.Publish(EventEnterChampSelect, nil)
        case lcu.GamePhaseInProgress:
            c.eventBus.Publish(EventGameStart, nil)
        case lcu.GamePhaseLobby:
            c.eventBus.Publish(EventReturnToLobby, nil)
        }
    })

    // 监听选人会话变化（获取队友英雄）
    c.lcuClient.Subscribe("/lol-champ-select/v1/session", func(event *lcu.Event) {
        var session ChampSelectSession
        if err := json.Unmarshal(event.Data, &session); err != nil {
            return
        }
        c.eventBus.Publish(EventChampSelectUpdate, session)
    }, "Update", "Create")
}
```

#### 4.1.2 获取队友英雄列表

```go
// /lol-champ-select/v1/session 返回的结构
// 解析 myTeam 数组，提取每个队友的 championId

type ChampSelectSession struct {
    MyTeam []struct {
        ChampionID   int    `json:"championId"`
        CellID       int    `json:"cellId"`
        SummonerID   int64  `json:"summonerId"`
        AssignedRole string `json:"assignedPosition"`
    } `json:"myTeam"`
}
```

#### 4.1.3 显示英雄胜率排行

前端 React 组件接收 Go 后端推送的数据：

```typescript
// 前端：Wails 事件绑定
EventsOn("champ-select:team-heroes", (heroes: TeamHero[]) => {
  // 查询本地 SQLite 获取每个英雄的 ARAM 胜率
  // 按胜率排序展示
})
```

```go
// 后端：查询英雄胜率
type HeroWinrate struct {
    ChampionID   int
    Name         string
    NameCN       string
    Winrate      float64
    Pickrate     float64
    Tier         string  // S/A/B/C
    BestAugments []AugmentWinrate
}

func (s *HeroService) GetTeamHeroWinrates(championIDs []int) ([]HeroWinrate, error) {
    query := `
        SELECT champion_id, name_cn, winrate, pickrate, tier
        FROM champion_stats
        WHERE game_mode = 'hexgates'
        AND champion_id IN (?)`

    // 使用 sqlx.In 处理 IN 查询
    // 按胜率降序返回
}
```

#### 4.1.4 选中英雄后显示海克斯推荐

当玩家锁定英雄后，前端显示该英雄的海克斯胜率排行：

```go
type AugmentWinrate struct {
    AugmentID    string
    NameCN       string
    NameEN       string
    Winrate      float64
    Pickrate     float64
    Tier         string
    Description  string
}

func (s *AugmentService) GetAugmentsForHero(championID int) ([]AugmentWinrate, error) {
    query := `
        SELECT a.augment_id, a.name_cn, a.name_en,
               ha.winrate, ha.pickrate, ha.tier, a.description
        FROM hero_augment_stats ha
        JOIN augments a ON ha.augment_id = a.augment_id
        WHERE ha.champion_id = ?
        AND ha.game_mode = 'hexgates'
        ORDER BY ha.winrate DESC`
}
```

#### 4.1.5 出装建议

```go
type BuildRecommendation struct {
    ChampionID int
    Role       string  // 根据英雄定位
    Items      []struct {
        ItemID   int
        NameCN   string
        Winrate  float64
        Slot     int  // 第几件出
    }
    Boots struct {
        ItemID  int
        NameCN  string
        Winrate float64
    }
    SkillOrder []string  // 技能加点顺序
}
```

---

### 4.2 阶段二：游戏内 Overlay

#### 4.2.1 核心挑战

进入游戏后：
- LCU 客户端关闭，**LCU API 不可用**
- Vanguard 反作弊运行，**禁止读内存**
- 需要在 LOL 游戏窗口上方显示信息

#### 4.2.2 解决方案：截图 OCR + 手动查询双模式

##### 模式 A：手动查询（MVP 优先实现）

1. 游戏内按 **F1** 呼出悬浮窗
2. 悬浮窗显示搜索框 + 该英雄推荐的海克斯列表
3. 玩家手动对比当前出现的 3 个海克斯

```go
// 全局热键监听
package hotkey

import "github.com/robotn/gohook"

func StartListener(toggleOverlay func()) {
    hook.Register(hook.KeyDown, []string{"f1"}, func(e hook.Event) {
        toggleOverlay()
    })
    s := hook.Start()
    <-hook.Process(s)
}
```

##### 模式 B：截图 OCR 自动识别（进阶）

```go
package ocr

import (
    "github.com/kbinani/screenshot"
    "github.com/otiai10/gosseract/v2"
)

type AugmentOCR struct {
    client *gosseract.Client
}

func NewAugmentOCR() *AugmentOCR {
    client := gosseract.NewClient()
    // 加载中文训练数据
    client.SetLanguage("chi_sim", "eng")
    return &AugmentOCR{client: client}
}

// CaptureAugmentArea 截取海克斯选择区域的屏幕
func (o *AugmentOCR) CaptureAugmentArea() ([]string, error) {
    // 1. 获取 LOL 游戏窗口位置
    bounds := getLOLWindowBounds()

    // 2. 根据 16:9 / 16:10 分辨率计算海克斯选项区域
    // 海克斯选择 UI 通常出现在屏幕中央偏下
    // 3 个选项水平排列
    augmentArea := calculateAugmentArea(bounds)

    // 3. 截图
    img, err := screenshot.CaptureRect(augmentArea)
    if err != nil {
        return nil, err
    }

    // 4. 图像预处理（灰度、二值化、锐化）提高 OCR 准确率
    processed := preprocessImage(img)

    // 5. OCR 识别
    o.client.SetImageFromBytes(processed)
    text, err := o.client.Text()
    if err != nil {
        return nil, err
    }

    // 6. 分割 3 个海克斯名称
    return splitAugmentNames(text), nil
}
```

**海克斯选择区域定位（基于 1920x1080）：**

```go
func calculateAugmentArea(bounds image.Rectangle) image.Rectangle {
    // LOL 海克斯选择 UI 在 1920x1080 下的大致位置
    // 中央偏下，3 个选项水平排列
    centerX := bounds.Dx() / 2
    centerY := bounds.Dy() * 3 / 4  // 屏幕下方 75% 处

    width := bounds.Dx() * 2 / 3    // 占据屏幕 2/3 宽度
    height := 200                    // 约 200px 高度

    return image.Rect(
        centerX - width/2,
        centerY - height/2,
        centerX + width/2,
        centerY + height/2,
    )
}
```

#### 4.2.3 透明 Overlay 窗口

```go
// Wails 窗口配置
// app.go

func (a *App) createOverlayWindow() {
    overlay := application.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
        Title:  "Hexgates Overlay",
        Width:  400,
        Height: 300,
        X:      calculateOverlayX(),  // 定位到 LOL 窗口中央
        Y:      calculateOverlayY(),
        Windows: application.WindowsWindow{
            AlwaysOnTop: true,
            Transparent: true,
            Frameless:   true,
            BackdropType: application.Mica,
        },
        Mac: application.MacWindow{
            TitleBar:              application.MacTitleBarHidden,
            Appearance:            application.NSAppearanceNameVibrantDark,
            WebviewIsTransparent:  true,
            BackgroundColour:      application.NewRGB(0, 0, 0, 0), // 全透明
            AlwaysOnTop:           true,
        },
    })

    overlay.SetHTML(`
        <html style="background: transparent;">
        <body style="background: rgba(0,0,0,0.7); color: white; font-family: sans-serif;">
            <!-- React App 挂载点 -->
            <div id="root"></div>
        </body>
        </html>
    `)
}
```

#### 4.2.4 Overlay 内容设计

```typescript
// 前端：Overlay 组件
interface OverlayProps {
  currentHero: HeroInfo       // 当前英雄（进游戏前缓存）
  augments: AugmentOption[]   // 当前出现的 3 个海克斯选项
}

interface AugmentOption {
  name: string
  winrate: number
  pickrate: number
  tier: 'S' | 'A' | 'B' | 'C'
  description: string
  isRecommended: boolean
}

// 显示效果：
// ┌─────────────────────────────────┐
// │  ⚡ 闪电打击          53.2% ✅ │  <- S级，推荐
// │     选取率: 24.1%               │
// ├─────────────────────────────────┤
// │  🔥 火焰增幅          48.7%    │  <- A级
// │     选取率: 18.3%               │
// ├─────────────────────────────────┤
// │  ❄️ 冰霜之触          42.1%    │  <- B级
// │     选取率: 12.5%               │
// └─────────────────────────────────┘
```

---

## 5. 数据层设计

### 5.1 SQLite Schema

```sql
-- 英雄基础信息
CREATE TABLE champions (
    champion_id INTEGER PRIMARY KEY,
    name_en     TEXT NOT NULL,
    name_cn     TEXT NOT NULL,
    title       TEXT,
    tags        TEXT  -- JSON ["Mage", "Support"]
);

-- 英雄在各模式下的统计数据
CREATE TABLE champion_stats (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id  INTEGER NOT NULL,
    game_mode    TEXT NOT NULL,  -- 'hexgates', 'aram', 'classic'
    winrate      REAL,
    pickrate     REAL,
    banrate      REAL,
    tier         TEXT,           -- 'S', 'A', 'B', 'C', 'D'
    sample_size  INTEGER,        -- 统计样本量
    patch        TEXT NOT NULL,  -- 版本号，如 "14.7"
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, game_mode, patch)
);

-- 海克斯（Augment）基础信息
CREATE TABLE augments (
    augment_id  TEXT PRIMARY KEY,  -- OP.GG / U.GG / Lolalytics 的 ID
    name_en     TEXT NOT NULL,
    name_cn     TEXT NOT NULL,
    description TEXT,
    tier        TEXT,              -- 海克斯等级：Silver/Gold/Prismatic
    icon_url    TEXT
);

-- 英雄 + 海克斯 组合胜率
CREATE TABLE hero_augment_stats (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id INTEGER NOT NULL,
    augment_id  TEXT NOT NULL,
    game_mode   TEXT NOT NULL,
    winrate     REAL,
    pickrate    REAL,
    tier        TEXT,
    patch       TEXT NOT NULL,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, augment_id, game_mode, patch),
    FOREIGN KEY (champion_id) REFERENCES champions(champion_id),
    FOREIGN KEY (augment_id) REFERENCES augments(augment_id)
);

-- 出装推荐
CREATE TABLE build_recommendations (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id  INTEGER NOT NULL,
    game_mode    TEXT NOT NULL,
    role         TEXT,             -- 如 'adc', 'mid', 'top', 'support'
    items        TEXT NOT NULL,     -- JSON [{"item_id": 3157, "slot": 1, "winrate": 55.2}, ...]
    boots        TEXT,              -- JSON {"item_id": 3020, "winrate": 52.1}
    skill_order  TEXT,              -- JSON ["Q", "Q", "W", "Q", ...]
    runes        TEXT,              -- JSON ["Press the Attack", "Triumph", ...]
    patch        TEXT NOT NULL,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, game_mode, role, patch)
);

-- 英雄协同推荐
CREATE TABLE champion_synergies (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id           INTEGER NOT NULL,
    champion_name         TEXT NOT NULL,
    synergy_champion_id   INTEGER NOT NULL,
    synergy_name          TEXT NOT NULL,
    score_rank            INTEGER,        -- 排名
    score                 REAL,           -- 协同评分
    play                  INTEGER,        -- 场次
    win                   INTEGER,        -- 胜场
    win_rate              REAL,           -- 胜率 (%)
    tier                  INTEGER,        -- 协同 tier
    game_mode             TEXT NOT NULL,
    patch                 TEXT NOT NULL,
    updated_at            DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, synergy_champion_id, game_mode, patch)
);

-- 装备基础信息
CREATE TABLE items (
    item_id  INTEGER PRIMARY KEY,
    name_en  TEXT NOT NULL,
    name_cn  TEXT NOT NULL,
    tags     TEXT,  -- JSON ["Magic", "Defense"]
    stats    TEXT   -- JSON {"ap": 80, "cd": 20}
);

-- 版本更新追踪
CREATE TABLE patch_tracker (
    patch       TEXT PRIMARY KEY,
    released_at DATETIME,
    scraped_at  DATETIME,
    status      TEXT  -- 'pending', 'scraping', 'completed', 'failed'
);
```

### 5.2 数据更新策略

```go
package scraper

// Scraper 定时爬取第三方数据站

type Scraper struct {
    sources     []DataSource  // OPGGSource, UGGSource, LolalyticsSource
    db          *sql.DB
    interval    time.Duration // 默认 6 小时
}

func (s *Scraper) Start() {
    ticker := time.NewTicker(s.interval)
    go func() {
        for range ticker.C {
            s.scrapeAll()
        }
    }()
}

func (s *Scraper) scrapeAll() {
    for _, source := range s.sources {
        // 1. 检查当前游戏版本
        patch := source.GetCurrentPatch()

        // 2. 检查是否已爬取
        if s.isPatchScraped(patch, source.Name()) {
            continue
        }

        // 3. 爬取英雄胜率
        heroStats := source.ScrapeChampionStats("hexgates", patch)
        s.saveChampionStats(heroStats)

        // 4. 爬取海克斯数据
        augmentStats := source.ScrapeAugmentStats("hexgates", patch)
        s.saveAugmentStats(augmentStats)

        // 5. 爬取出装
        builds := source.ScrapeBuilds("hexgates", patch)
        s.saveBuilds(builds)
    }
}
```

### 5.3 数据获取实现

#### 5.3.1 OP.GG MCP API（主要数据源）

OP.GG 提供 MCP (Model Context Protocol) API，通过 JSON-RPC 2.0 调用获取结构化 ARAM 数据。

```go
const mcpEndpoint = "https://mcp-api.op.gg/mcp"

// MCPClient MCP API 客户端
type MCPClient struct {
    client *http.Client
}

// CallTool 调用 MCP 工具
func (c *MCPClient) CallTool(toolName string, arguments map[string]any) (string, error) {
    reqBody := map[string]any{
        "jsonrpc": "2.0",
        "id":      1,
        "method":  "tools/call",
        "params": map[string]any{
            "name":      toolName,
            "arguments": arguments,
        },
    }
    // POST 到 mcpEndpoint，返回 text 内容（Python-like 类序列化）
}
```

**可用的 MCP 工具：**

| 工具名 | 功能 | 返回格式 |
|--------|------|----------|
| `lol_get_champion_analysis` | ARAM 出装、符文、技能加点 | Python-like 类序列化 |
| `lol_get_champion_synergies` | 英雄协同推荐 | Python-like 类序列化 |
| `lol_list_items` | 嚎哭深渊装备列表 | JSON |
| `lol_list_aram_augments` | ARAM 海克斯列表 | JSON |

**MCP 响应解析：**

MCP 工具返回 Python-like 的类序列化文本（非 JSON），需要自定义解析器：

```go
// 示例响应：
// class LolGetChampionAnalysis: champion,position,data
// LolGetChampionAnalysis("ASHE", "adc", Data(...))

// 解析步骤：
// 1. stripClassDefinitions() — 去除 "class Xxx:" 定义行
// 2. parseMCPClass() — 递归解析类名和字段
// 3. 按索引访问字段（Data 在 index 2，synergies 的 Data 在 index 4）
```

#### 5.3.2 OP.GG 内部 REST API（英雄胜率）

英雄胜率数据来自 OP.GG 内部 REST API `https://lol-api-champion.op.gg/api`：

```go
package scraper

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"

    "github.com/PuerkitoBio/goquery"
)

const (
    opggBaseURL     = "https://op.gg"
    opggChampionAPI = "https://lol-api-champion.op.gg/api"
)

// OPGGSource OP.GG 数据源（通过内部 REST API）
type OPGGSource struct {
    client *http.Client
}

func NewOPGGSource() *OPGGSource {
    return &OPGGSource{
        client: NewHTTPClient(),
    }
}

func (s *OPGGSource) Name() string { return "opgg" }

// GetCurrentPatch 从 OP.GG 页面静态资源路径提取版本号
// 例如从 https://opgg-static.akamaized.net/meta/images/lol/16.8.1/champion/Neeko.png 中提取 "16.8.1"
func (s *OPGGSource) GetCurrentPatch() (string, error) {
    req, err := newRequest("GET", opggBaseURL+"/zh-cn/lol/modes/aram-mayhem")
    if err != nil {
        return "", err
    }

    randomDelay()
    resp, err := retryRequest(s.client, req, 3)
    if err != nil {
        return "", fmt.Errorf("fetch patch: %w", err)
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("read body: %w", err)
    }
    body := string(bodyBytes)

    // 查找 /lol/ 后面的版本号
    for {
        idx := strings.Index(body, "/lol/")
        if idx == -1 {
            break
        }
        body = body[idx+5:]
        slashIdx := strings.Index(body, "/")
        if slashIdx == -1 {
            break
        }
        candidate := body[:slashIdx]
        segments := strings.Split(candidate, ".")
        if len(segments) >= 2 {
            _, err1 := strconv.Atoi(segments[0])
            _, err2 := strconv.Atoi(segments[1])
            if err1 == nil && err2 == nil {
                return candidate, nil
            }
        }
        body = body[slashIdx+1:]
    }

    return "", fmt.Errorf("could not detect patch version from OP.GG")
}

// === API 数据模型 ===

// 英雄元数据响应：/api/meta/champions?hl=zh_CN
type opggChampionMeta struct {
    ID   int    `json:"id"`   // 英雄数字 ID
    Key  string `json:"key"`  // 英文 key，如 "Garen"
    Name string `json:"name"` // 中文名，如 "德玛西亚之力"
}

// 英雄统计响应：/api/{region}/champions/{mode}?tier=all&hl=zh_CN
type opggChampionStats struct {
    ID   int    `json:"id"`
    Tier string `json:"tier"` // tier 等级（如 "3"）
    Rank int    `json:"rank"`
    AverageStats struct {
        WinRate  float64 `json:"win_rate"`  // 0.0 ~ 1.0
        PickRate float64 `json:"pick_rate"` // 0.0 ~ 1.0
        BanRate  float64 `json:"ban_rate"`  // 可能为 null
        Play     int     `json:"play"`      // 样本量
    } `json:"average_stats"`
}

// === 核心爬取逻辑 ===

// fetchChampionMeta 获取英雄元数据（ID → 中文名/英文 key 映射）
func (s *OPGGSource) fetchChampionMeta() (map[int]opggChampionMeta, error) {
    url := fmt.Sprintf("%s/meta/champions?hl=zh_CN", opggChampionAPI)
    req, err := newRequest("GET", url)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/json")

    randomDelay()
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch champion meta: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("fetch champion meta: status %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read champion meta: %w", err)
    }

    var result struct {
        Data []opggChampionMeta `json:"data"`
    }
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("decode champion meta: %w", err)
    }

    metaMap := make(map[int]opggChampionMeta, len(result.Data))
    for _, champ := range result.Data {
        metaMap[champ.ID] = champ
    }
    return metaMap, nil
}

// mapMode 将内部模式名称映射为 OP.GG API 有效模式
// OP.GG 的 "aram-mayhem" mode 在 API 中无效，必须用 "aram"
func mapMode(mode string) string {
    switch mode {
    case "hexgates":
        return "aram"
    case "aram-mayhem":
        return "aram"
    default:
        return mode
    }
}

// ScrapeChampionStats 爬取英雄胜率排行（通过内部 API）
func (s *OPGGSource) ScrapeChampionStats(mode, patch string) ([]ChampionStat, error) {
    // 1. 获取中文名映射
    metaMap, err := s.fetchChampionMeta()
    if err != nil {
        return nil, fmt.Errorf("fetch meta: %w", err)
    }

    // 2. 获取统计数据
    apiMode := mapMode(mode)
    url := fmt.Sprintf("%s/NA/champions/%s?tier=all&hl=zh_CN", opggChampionAPI, apiMode)
    // NOTE: 不要传 version 参数，会导致 422

    req, err := newRequest("GET", url)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Referer", "https://op.gg/")

    randomDelay()
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch champion stats: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("fetch champion stats: status %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read champion stats: %w", err)
    }

    var apiResp struct {
        Data []opggChampionStats `json:"data"`
    }
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return nil, fmt.Errorf("decode champion stats: %w", err)
    }

    if len(apiResp.Data) == 0 {
        return nil, fmt.Errorf("no champion stats returned from API")
    }

    // 3. 合并元数据和统计数据
    var stats []ChampionStat
    for _, cs := range apiResp.Data {
        meta, ok := metaMap[cs.ID]
        if !ok {
            continue
        }

        stats = append(stats, ChampionStat{
            ChampionID: cs.ID,
            NameCN:     meta.Name,
            NameEN:     meta.Key,
            Winrate:    cs.AverageStats.WinRate,
            Pickrate:   cs.AverageStats.PickRate,
            Banrate:    cs.AverageStats.BanRate,
            Tier:       cs.Tier,
            SampleSize: cs.AverageStats.Play,
            Patch:      patch,
            Source:     s.Name(),
        })
    }

    if len(stats) == 0 {
        return nil, fmt.Errorf("no valid champion stats after merging with meta")
    }

    return stats, nil
}

// ScrapeAugmentStats 爬取海克斯数据
func (s *OPGGSource) ScrapeAugmentStats(mode, patch string) ([]HeroAugmentStat, error) {
    // TODO: OP.GG 内部 API 暂未发现 augment 统计 endpoint
    return nil, fmt.Errorf("augment scraping not yet implemented for OP.GG")
}

// ScrapeBuilds 爬取英雄出装
func (s *OPGGSource) ScrapeBuilds(mode, patch string) ([]BuildRecommendation, error) {
    // TODO: 需要遍历每个英雄的 build 页面
    return nil, fmt.Errorf("build scraping not yet implemented for OP.GG")
}

// === 辅助方法（HTML 爬取，用于 augment/build 等非 API 数据）===

// ScrapeAugmentList 爬取海克斯列表（基础信息，无胜率）
func (s *OPGGSource) ScrapeAugmentList() ([]struct {
    NameCN string
    NameEN string
    Tier   string
}, error) {
    url := opggBaseURL + "/zh-cn/lol/modes/aram-mayhem"

    req, err := newRequest("GET", url)
    if err != nil {
        return nil, err
    }

    randomDelay()
    resp, err := retryRequest(s.client, req, 3)
    if err != nil {
        return nil, fmt.Errorf("fetch augment list: %w", err)
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("parse augment list: %w", err)
    }

    var augments []struct {
        NameCN string
        NameEN string
        Tier   string
    }

    doc.Find("[data-type='augment']").Each(func(i int, sel *goquery.Selection) {
        name := strings.TrimSpace(sel.Text())
        tier, _ := sel.Attr("data-tier")
        if name != "" {
            augments = append(augments, struct {
                NameCN string
                NameEN string
                Tier   string
            }{
                NameCN: name,
                Tier:   tier,
            })
        }
    })

    return augments, nil
}

// ScrapeChampionBuild 爬取单个英雄的出装
func (s *OPGGSource) ScrapeChampionBuild(championKey, patch string) (*BuildRecommendation, error) {
    url := fmt.Sprintf("%s/zh-cn/lol/modes/aram-mayhem/%s/build", opggBaseURL, championKey)
    if patch != "" {
        url += "?patch=" + patch
    }

    req, err := newRequest("GET", url)
    if err != nil {
        return nil, err
    }

    randomDelay()
    resp, err := retryRequest(s.client, req, 3)
    if err != nil {
        return nil, fmt.Errorf("fetch build for %s: %w", championKey, err)
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("parse build for %s: %w", championKey, err)
    }

    build := &BuildRecommendation{
        ChampionName: championKey,
        Patch:        patch,
        Source:       s.Name(),
    }

    doc.Find(".build-items .item").Each(func(i int, sel *goquery.Selection) {
        itemIDStr, _ := sel.Attr("data-item-id")
        itemID, _ := strconv.Atoi(itemIDStr)
        itemName := strings.TrimSpace(sel.Find(".item-name").Text())

        if itemID > 0 {
            build.Items = append(build.Items, BuildItem{
                ItemID: itemID,
                NameCN: itemName,
                Slot:   i + 1,
            })
        }
    })

    doc.Find(".rune-page .rune").Each(func(i int, sel *goquery.Selection) {
        runeName, _ := sel.Attr("alt")
        if runeName == "" {
            runeName = strings.TrimSpace(sel.Text())
        }
        if runeName != "" {
            build.Runes = append(build.Runes, runeName)
        }
    })

    return build, nil
}

// FetchChampionKeys 获取英雄英文 key 列表
func (s *OPGGSource) FetchChampionKeys() ([]string, error) {
    // 通过 meta API 获取，比 HTML 解析更可靠
    metaMap, err := s.fetchChampionMeta()
    if err != nil {
        return nil, err
    }

    keys := make([]string, 0, len(metaMap))
    for _, meta := range metaMap {
        keys = append(keys, meta.Key)
    }
    return keys, nil
}

// OP.GG 数据特点说明
//
// 架构演进：
// - 原计划：HTML 爬取（goquery）
// - 问题：Next.js streaming SSR，初始 HTML 表格为空
// - 解决：改用内部 REST API（参考 OPGG.py 项目）
//
// API 端点：
// - Base: https://lol-api-champion.op.gg/api
// - Meta:  /meta/champions?hl=zh_CN          → 英雄元数据
// - Stats: /{region}/champions/{mode}?tier=all → 胜率统计
//
// 优势：
// - 中文名直接来自 API，无需翻译
// - JSON 结构化数据，解析稳定
// - 无需 headless browser，直接 HTTP 即可
// - 172 英雄数据完整，与 DDragon 数量一致
//
// 局限：
// - 内部 API，OP.GG 可能随时变更
// - augment/build 暂无对应 API endpoint，仍需 HTML 爬取
// - version 参数会导致 422，不能传版本号
// - "aram-mayhem" mode 无效，必须用 "aram"
//
// 请求要求：
// - Accept: application/json
// - Referer: https://op.gg/ (stats endpoint 需要)
// - User-Agent: 浏览器 UA（通过 newRequest 自动设置）
//
// 反爬策略：
// - User-Agent 轮换
// - 请求间隔 1-3 秒随机延时
// - 失败重试（指数退避）
```

### 5.4 数据源优先级策略

| 数据类型 | 主数据源 | 备用/验证 | 说明 |
|----------|----------|-----------|------|
| 英雄胜率 | **OP.GG** | U.GG, Lolalytics | OP.GG 内部 API 稳定，直接返回结构化 JSON |
| 海克斯名称（中文）| **OP.GG** | 本地翻译表 | OP.GG 原生中文 |
| 英雄+海克斯胜率 | **U.GG** | Lolalytics | U.GG 有精确组合数据 |
| 出装推荐 | **OP.GG** | U.GG | OP.GG 中文装备名（HTML 爬取）|
| 符文推荐 | **OP.GG** | U.GG | OP.GG 中文符文名（HTML 爬取）|

---

## 6. 项目目录结构

```
lol-hexgates-helper/
├── main.go                          # 入口
├── go.mod
├── wails.json                       # Wails 配置
│
├── cmd/                             # 数据初始化工具
│   ├── initdata/main.go             # DDragon 英雄/装备基础数据
│   ├── initaugments/main.go         # OP.GG MCP 海克斯数据
│   ├── initbuilds/main.go           # OP.GG MCP 出装+符文+技能
│   ├── initsynergies/main.go        # OP.GG MCP 协同数据
│   ├── inititems/main.go            # OP.GG MCP 装备数据
│   └── lcudemo/main.go              # LCU 连接测试
│
├── internal/
│   ├── lcu/                         # LCU 客户端封装
│   │   ├── client.go                # lcu-gopher 封装（vendored + patched）
│   │   └── events.go                # 事件总线
│   │
│   ├── data/                        # 数据层
│   │   ├── db.go                    # SQLite 连接 + 迁移
│   │   ├── migrations/              # 数据库迁移文件
│   │   ├── champion.go              # 英雄数据操作
│   │   ├── augment.go               # 海克斯数据操作
│   │   ├── build.go                 # 出装数据操作（含符文）
│   │   ├── synergy.go               # 协同数据操作
│   │   ├── item.go                  # 装备数据操作
│   │   └── queries.go               # SQL 查询
│   │
│   ├── scraper/                     # 数据爬取
│   │   ├── ddragon.go               # DDragon API 客户端
│   │   ├── opgg.go                  # OP.GG 内部 REST API
│   │   ├── mcp.go                   # OP.GG MCP API 客户端 + 解析器
│   │   ├── http.go                  # HTTP 客户端配置
│   │   └── source.go                # 数据源接口
│   │
│   └── logger/                      # 日志
│       └── zap.go                   # Zap 初始化
│
├── frontend/                        # Wails 前端 (React + TypeScript + Tailwind)
│   ├── src/
│   │   ├── App.tsx                  # 主应用（阶段路由）
│   │   ├── main.tsx                 # 入口
│   │   ├── style.css                # 全局样式（LOL 深色主题）
│   │   ├── components/
│   │   │   ├── ChampSelectView.tsx  # 选人阶段主容器
│   │   │   ├── TeamMemberCard.tsx   # 队友卡片（胜率/海克斯/出装/协同）
│   │   │   ├── AugmentList.tsx      # 海克斯推荐列表
│   │   │   └── BuildPanel.tsx       # 出装面板（装备+符文+技能）
│   │   └── utils/
│   │       └── ddragon.ts           # DDragon CDN URL 工具
│   ├── package.json
│   ├── tailwind.config.js           # Tailwind 配置（LOL 配色）
│   └── vite.config.ts
│
├── data/                            # SQLite 数据库（运行时生成）
│   └── haxplugins.db
│
└── docs/                            # 文档
    ├── lol-hexgates-plugin-tech-spec.md  # 技术实现方案
    └── superpowers/specs/                # 设计文档
```

---

## 7. 开发计划（分阶段）

### Phase 1：基础骨架（1-2 周）

| 任务 | 输出 |
|------|------|
| Wails 项目初始化 | 可运行的空窗口 |
| SQLite 数据库 + 迁移 | 空表结构就绪 |
| lcu-gopher 集成 | 能连接 LOL 客户端，打印游戏阶段 |
| 前端基础组件 | 英雄卡片、海克斯列表的 UI 骨架 |

### Phase 2：数据层与爬虫 MVP（2-3 周）

| 任务 | 输出 |
|------|------|
| SQLite Schema 设计 | 英雄/海克斯/出装/协同/装备 表结构 |
| DDragon 数据导入 | 172 英雄 + 695 装备基础数据 |
| OP.GG MCP API 集成 | builds / augments / synergies / items 数据获取 |
| 数据初始化工具 | `cmd/initdata`、`cmd/initaugments`、`cmd/initbuilds`、`cmd/initsynergies`、`cmd/inititems` |

### Phase 3：客户端选人阶段（2 周）

| 任务 | 输出 |
|------|------|
| LCU 连接 + 事件监听 | 游戏阶段推送、选人会话更新 |
| Mock 模式 | 无 LOL 客户端时自动启用，支持 UI 测试 |
| 前端选人界面 | 队友卡片：胜率、Tier、头像 |
| 海克斯推荐面板 | 强度评分、选取率、胜率 |
| 出装推荐面板 | 核心装备、鞋子、符文、技能加点 |
| 协同推荐面板 | 最佳搭档英雄、胜率、评分 |
| 数据绑定 | Go ↔ React 实时推送 |

### Phase 4：游戏内 Overlay（2-3 周）

| 任务 | 输出 |
|------|------|
| 透明窗口 | 无边框、置顶、透明背景的悬浮窗 |
| 全局热键 | F1 呼出/隐藏 |
| 手动查询模式 | 搜索框 + 海克斯列表（不依赖 OCR）|
| 截图 + OCR | 自动识别当前 3 个海克斯选项 |
| 英雄信息缓存 | 进游戏前缓存当前英雄，Overlay 直接用 |

### Phase 5：优化与发布（1-2 周）

| 任务 | 输出 |
|------|------|
| OCR 准确率优化 | 图像预处理、中文训练数据调优 |
| UI 美化 | 参考 LOL 游戏风格设计 |
| 配置面板 | 热键修改、数据源切换、显示设置 |
| 错误处理 | 连接失败、数据缺失的降级方案 |
| 打包发布 | Windows / macOS 安装包 |

---

## 8. 关键 Go 代码片段

### 8.1 主应用启动

```go
package main

import (
    "context"
    "embed"

    "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

type App struct {
    lcuClient     *lcu.Client
    overlay       *overlay.Manager
    championSvc   *service.ChampionService
    augmentSvc    *service.AugmentService
    buildSvc      *service.BuildService
    scraper       *scraper.Scraper
}

func main() {
    app := application.New(application.Options{
        Name: "LOL Hexgates Helper",
        Assets: application.AssetOptions{
            Handler: application.AutoDiscoveryFS(assets, "frontend/dist"),
        },
    })

    myApp := &App{}

    // 绑定 Go 方法到前端
    app.Bind(myApp.GetTeamWinrates)
    app.Bind(myApp.GetAugmentRecommendations)
    app.Bind(myApp.GetBuildRecommendation)
    app.Bind(myApp.ToggleOverlay)

    // 创建主窗口
    mainWindow := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
        Title:  "海克斯大乱斗助手",
        Width:  1200,
        Height: 800,
    })

    // 创建 Overlay 窗口（初始隐藏）
    myApp.overlay = overlay.NewManager(app)

    // 启动 LCU 监听
    myApp.lcuClient = lcu.NewClient()
    myApp.lcuClient.StartListening()

    // 启动数据爬取
    myApp.scraper = scraper.New(myApp.db)
    myApp.scraper.Start()

    // 运行
    err := app.Run()
    if err != nil {
        logger.Fatal("app run failed", zap.Error(err))
    }
}
```

### 8.2 前端调用 Go 方法

```typescript
// frontend/src/services/api.ts
import { GetTeamWinrates, GetAugmentRecommendations } from "../../wailsjs/go/main/App";

export async function fetchTeamWinrates(): Promise<TeamHero[]> {
    return await GetTeamWinrates();
}

export async function fetchAugments(championId: number): Promise<Augment[]> {
    return await GetAugmentRecommendations(championId);
}
```

### 8.3 窗口定位到 LOL 上方

```go
package overlay

import (
    "github.com/wailsapp/wails/v3/pkg/application"
    "github.com/robotn/gohook"
)

type Manager struct {
    app           *application.Application
    overlayWindow application.WebviewWindow
    visible       bool
}

func NewManager(app *application.Application) *Manager {
    m := &Manager{app: app}

    // 创建透明 Overlay 窗口（初始隐藏）
    m.overlayWindow = app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
        Title:     "Overlay",
        Width:     400,
        Height:    350,
        Hidden:    true,
        Frameless: true,
        AlwaysOnTop: true,
        BackgroundColour: application.NewRGB(0, 0, 0, 0),
    })

    // 注册热键
    hook.Register(hook.KeyDown, []string{"f1"}, func(e hook.Event) {
        m.Toggle()
    })
    go hook.Process(hook.Start())

    return m
}

func (m *Manager) Toggle() {
    if m.visible {
        m.overlayWindow.Hide()
        m.visible = false
    } else {
        // 定位到 LOL 窗口中央
        pos := getLOLWindowCenter()
        m.overlayWindow.SetPosition(pos.X-200, pos.Y-175)
        m.overlayWindow.Show()
        m.visible = true
    }
}
```

---

## 9. 风险与注意事项

### 9.1 安全风险（Critical）

| 风险 | 后果 | 规避方案 |
|------|------|----------|
| 读游戏内存 | **封号** | 绝对禁止，仅用截图 OCR |
| 注入 DLL / Hook | **封号** | 绝对禁止，独立进程运行 |
| 修改游戏文件 | **封号** | 不触碰游戏目录任何文件 |
| 网络请求频率过高 | IP 被封 | 爬虫加延时、User-Agent 轮换、失败重试 |

### 9.2 技术风险

| 风险 | 影响 | 应对方案 |
|------|------|----------|
| OP.GG / U.GG / Lolalytics 改版 | 爬虫失效 | 多数据源冗余，监控爬取失败告警 |
| OCR 中文识别率低 | 游戏内体验差 | 先用 MVP 手动查询模式，OCR 作为加分项 |
| 不同分辨率 UI 位置不同 | Overlay 错位 | 支持常见分辨率配置，或让玩家手动校准 |
| LCU API 变更 | 客户端功能失效 | lcu-gopher 社区维护，关注更新 |
| 补丁更新后数据缺失 | 显示旧数据 | 爬虫自动检测新版本，数据过期提示 |

### 9.3 法律风险

- Riot 的 Vanguard 反作弊可能将任何第三方工具视为违规
- 虽然截图 + Overlay 类似 OBS / Discord Overlay 的原理，但仍存在被误判的风险
- **建议**：
  - 开源代码，透明实现
  - 不在游戏中修改任何内存
  - 社区使用时自行承担风险

---

## 10. 参考资源

| 资源 | 链接 | 用途 |
|------|------|------|
| lcu-gopher | https://github.com/Its-Haze/lcu-gopher | Go LCU 客户端 |
| Wails | https://wails.io/ | Go 桌面应用框架 |
| Pengu Loader | https://pengu.lol/ | LOL 客户端插件参考 |
| Yimikami Plugins | https://github.com/Yimikami/pengu-plugins | 现有插件参考（Rune Plugin 爬取逻辑）|
| Riot API Docs | https://developer.riotgames.com/docs/lol | LCU API 官方文档 |
| OP.GG | https://op.gg/zh-cn/lol/modes/aram-mayhem | 数据源（中文、有海克斯筛选和推荐）|
| U.GG | https://u.gg/ | 数据源 |
| Lolalytics | https://lolalytics.com/ | 数据源 |

---

## 11. 数据初始化命令

```bash
# 首次使用或补丁更新后运行：
go run ./cmd/initdata        # DDragon 英雄/装备基础数据
go run ./cmd/initaugments    # OP.GG MCP 海克斯数据
go run ./cmd/initbuilds      # OP.GG MCP 出装+符文+技能数据
go run ./cmd/initsynergies   # OP.GG MCP 协同数据
go run ./cmd/inititems       # OP.GG MCP 嚎哭深渊装备数据
```

## 12. 下一步行动

1. **Phase 4：游戏内 Overlay** — 透明窗口 + 热键 + 手动查询模式
2. **数据自动更新** — 检测补丁版本变化，自动触发数据爬取
3. **Windows 支持** — 修复 lcu-gopher 跨平台编译问题
4. **OCR 自动识别** — 截图识别海克斯选项（进阶功能）
