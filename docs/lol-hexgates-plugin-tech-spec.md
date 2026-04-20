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
| **桌面框架** | Wails v3 | Go 后端 + React/Vue 前端，透明窗口、置顶、现代 UI |
| **LCU 连接** | lcu-gopher | 自动检测 LOL 客户端，HTTP + WebSocket |
| **数据爬取** | Go net/http + goquery | 爬取 OP.GG / U.GG / Lolalytics 海克斯数据 |
| **本地缓存** | SQLite + mattn/go-sqlite3 | 英雄、海克斯、出装、胜率数据 |
| **游戏内截图** | kbinani/screenshot | 跨平台截图，获取海克斯选项区域 |
| **OCR 识别** | gosseract (Tesseract) | 识别海克斯文字，中文需训练数据 |
| **全局热键** | robotn/gohook | F1 呼出/隐藏 Overlay |
| **窗口管理** | Wails 窗口 API | 置顶、透明、无边框、定位到 LOL 窗口上方 |
| **配置管理** | Viper | 热键、数据源偏好、显示设置 |
| **日志** | Zap | 结构化日志 |

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
    role         TEXT,             -- 可选，如 'AP', 'AD', 'Tank'
    items        TEXT NOT NULL,     -- JSON [{"item_id": 3157, "slot": 1, "winrate": 0.55}, ...]
    boots        TEXT,              -- JSON {"item_id": 3020, "winrate": 0.52}
    skill_order  TEXT,              -- JSON ["Q", "Q", "W", "Q", ...]
    patch        TEXT NOT NULL,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, game_mode, role, patch)
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

### 5.3 OP.GG 爬虫实现

```go
package scraper

import (
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "github.com/PuerkitoBio/goquery"
)

// OPGGSource OP.GG 数据源（中文，有海克斯大乱斗独立页面）
type OPGGSource struct {
    client  *http.Client
    baseURL string
}

func NewOPGGSource() *OPGGSource {
    return &OPGGSource{
        client:  &http.Client{Timeout: 30 * time.Second},
        baseURL: "https://op.gg/zh-cn/lol/modes/aram-mayhem",
    }
}

func (s *OPGGSource) Name() string { return "opgg" }

// GetCurrentPatch 从 OP.GG 页面获取当前游戏版本
func (s *OPGGSource) GetCurrentPatch() string {
    // OP.GG 页面有版本信息，如 "16.8.1"
    // 或通过 Riot API 获取当前版本，统一数据源
    resp, err := s.client.Get(s.baseURL)
    if err != nil {
        return ""
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return ""
    }

    // 从页面或 meta 标签提取版本号
    version := doc.Find("meta[name='version']").AttrOr("content", "")
    return version
}

// ScrapeChampionStats 爬取英雄胜率排行
// OP.GG 页面: https://op.gg/zh-cn/lol/modes/aram-mayhem
// 页面结构：表格，列：排名、英雄、段位、胜率、登场率
func (s *OPGGSource) ScrapeChampionStats(mode, patch string) []ChampionStat {
    url := fmt.Sprintf("%s?patch=%s", s.baseURL, patch)

    resp, err := s.client.Get(url)
    if err != nil {
        return nil
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil
    }

    var stats []ChampionStat

    // OP.GG 英雄列表表格
    // 选择器需要根据实际页面结构调整
    doc.Find("table.aram-champions-table tbody tr").Each(func(i int, s *goquery.Selection) {
        // 提取英雄名称
        nameCN := s.Find("td.champion-cell .name").Text()
        nameEN := s.Find("td.champion-cell .english-name").Text()

        // 提取胜率
        winrateStr := s.Find("td.winrate-cell").Text()
        winrateStr = strings.TrimSuffix(winrateStr, "%")
        winrate, _ := strconv.ParseFloat(winrateStr, 64)

        // 提取登场率
        pickrateStr := s.Find("td.pickrate-cell").Text()
        pickrateStr = strings.TrimSuffix(pickrateStr, "%")
        pickrate, _ := strconv.ParseFloat(pickrateStr, 64)

        // 提取段位
        tier := s.Find("td.tier-cell img").AttrOr("alt", "")

        stats = append(stats, ChampionStat{
            NameCN:   strings.TrimSpace(nameCN),
            NameEN:   strings.TrimSpace(nameEN),
            Winrate:  winrate / 100,
            Pickrate: pickrate / 100,
            Tier:     tier,
            Source:   s.Name(),
            Patch:    patch,
        })
    })

    return stats
}

// ScrapeAugmentStats 爬取海克斯数据
// OP.GG 页面有 augment 筛选区：
// - 按等级筛选：白银 / 黄金 / 棱彩
// - 每个海克斯显示：图标、名称、描述、推荐英雄 Top 5
func (s *OPGGSource) ScrapeAugmentStats(mode, patch string) []HeroAugmentStat {
    url := fmt.Sprintf("%s?patch=%s", s.baseURL, patch)

    resp, err := s.client.Get(url)
    if err != nil {
        return nil
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil
    }

    var stats []HeroAugmentStat

    // 爬取海克斯列表
    // OP.GG 的 augment 区域通常是独立区块
    doc.Find(".augment-list .augment-item").Each(func(i int, sel *goquery.Selection) {
        augmentName := sel.Find("strong").Text()
        augmentDesc := sel.Find("p").Text()

        // 提取推荐英雄（Top 5）
        var topChampions []string
        sel.Find(".recommended-champions img").Each(func(j int, img *goquery.Selection) {
            champName := img.AttrOr("alt", "")
            topChampions = append(topChampions, champName)
        })

        // 注意：OP.GG 的 augment 页面可能没有按英雄+海克斯的精确胜率
        // 但提供了"该海克斯最适合的英雄"推荐
        // 这部分数据可以作为补充，精确胜率仍需 U.GG / Lolalytics

        stats = append(stats, HeroAugmentStat{
            AugmentName:    strings.TrimSpace(augmentName),
            Description:    strings.TrimSpace(augmentDesc),
            TopChampions:   topChampions,
            Source:         s.Name(),
            Patch:          patch,
        })
    })

    return stats
}

// ScrapeBuilds 爬取英雄出装
// OP.GG 每个英雄有独立的 aram-mayhem build 页面
// 如：https://op.gg/zh-cn/lol/modes/aram-mayhem/ahri/build
func (s *OPGGSource) ScrapeBuilds(mode, patch string) []BuildRecommendation {
    // 1. 先获取英雄列表
    champions := s.getChampionList()

    var builds []BuildRecommendation

    for _, champ := range champions {
        buildURL := fmt.Sprintf("%s/%s/build?patch=%s", s.baseURL, champ.Key, patch)

        resp, err := s.client.Get(buildURL)
        if err != nil {
            continue
        }

        doc, err := goquery.NewDocumentFromReader(resp.Body)
        resp.Body.Close()
        if err != nil {
            continue
        }

        // 爬取出装顺序
        var items []BuildItem
        doc.Find(".build-items .item-slot").Each(func(i int, sel *goquery.Selection) {
            itemIDStr := sel.AttrOr("data-item-id", "0")
            itemID, _ := strconv.Atoi(itemIDStr)
            itemName := sel.Find(".item-name").Text()

            items = append(items, BuildItem{
                ItemID: itemID,
                NameCN: strings.TrimSpace(itemName),
                Slot:   i + 1,
            })
        })

        // 爬取符文
        var runes []string
        doc.Find(".rune-path .rune").Each(func(i int, sel *goquery.Selection) {
            runeName := sel.AttrOr("alt", "")
            runes = append(runes, runeName)
        })

        builds = append(builds, BuildRecommendation{
            ChampionID:   champ.ID,
            ChampionName: champ.NameCN,
            Items:        items,
            Runes:        runes,
            Source:       s.Name(),
            Patch:        patch,
        })
    }

    return builds
}

// OP.GG 数据特点说明
//
// 优势：
// - 中文界面，海克斯名称和描述都是中文，无需翻译
// - 有独立的 aram-mayhem 页面，数据针对性强
// - 有海克斯筛选（白银/黄金/棱彩）
// - 每个海克斯有推荐英雄 Top 5
// - 每个英雄有独立的出装/符文页面
//
// 局限：
// - 页面有 Cloudflare 反爬，需要配置 User-Agent、延时、可能需绕过
// - 部分数据是动态渲染（Next.js），可能需要爬取 API 接口而非 HTML
// - augment 按英雄+海克斯的精确胜率数据可能不如 U.GG 详细
// - 建议作为中文数据源和补充验证，精确胜率仍以 U.GG 为主
//
// 反爬策略：
// - User-Agent 轮换（模拟真实浏览器）
// - 请求间隔 1-3 秒随机延时
// - 失败重试（指数退避）
// - 考虑使用 headless browser（如 rod/playwright）作为备选
```

### 5.4 数据源优先级策略

| 数据类型 | 主数据源 | 备用/验证 | 说明 |
|----------|----------|-----------|------|
| 英雄胜率 | **U.GG** | OP.GG, Lolalytics | U.GG 数据最全面 |
| 海克斯名称（中文）| **OP.GG** | 本地翻译表 | OP.GG 原生中文 |
| 英雄+海克斯胜率 | **U.GG** | Lolalytics | U.GG 有精确组合数据 |
| 出装推荐 | **OP.GG** | U.GG | OP.GG 中文装备名 |
| 符文推荐 | **OP.GG** | U.GG | OP.GG 中文符文名 |

---

## 6. 项目目录结构

```
lol-hexgates-helper/
├── main.go                          # 入口
├── go.mod
├── wails.json                       # Wails 配置
│
├── cmd/
│   └── app/
│       └── main.go                  # 应用启动逻辑
│
├── internal/
│   ├── app/                         # Wails 应用生命周期
│   │   ├── app.go                   # 主 App 结构
│   │   ├── events.go                # 事件定义
│   │   └── windows.go               # 窗口管理
│   │
│   ├── lcu/                         # LCU 客户端封装
│   │   ├── client.go                # lcu-gopher 封装
│   │   ├── champ_select.go          # 选人阶段事件
│   │   ├── events.go                # LCU 事件类型
│   │   └── game_phase.go            # 游戏阶段监听
│   │
│   ├── overlay/                     # 游戏内 Overlay
│   │   ├── window.go                # 透明窗口管理
│   │   ├── hotkey.go                # 全局热键
│   │   ├── screenshot.go            # 截图逻辑
│   │   ├── ocr.go                   # OCR 识别
│   │   └── position.go              # 窗口定位计算
│   │
│   ├── data/                        # 数据层
│   │   ├── db.go                    # SQLite 连接
│   │   ├── migrations/              # 数据库迁移
│   │   ├── champion.go              # 英雄数据操作
│   │   ├── augment.go               # 海克斯数据操作
│   │   ├── build.go                 # 出装数据操作
│   │   └── queries.go               # SQL 查询
│   │
│   ├── scraper/                     # 数据爬取
│   │   ├── scraper.go               # 爬取引擎
│   │   ├── source.go                # 数据源接口
│   │   ├── opgg.go                  # OP.GG 爬取实现
│   │   ├── ugg.go                   # U.GG 爬取实现
│   │   ├── lolalytics.go            # Lolalytics 爬取实现
│   │   └── http.go                  # HTTP 客户端配置
│   │
│   ├── service/                     # 业务逻辑层
│   │   ├── champion.go              # 英雄服务
│   │   ├── augment.go               # 海克斯服务
│   │   ├── build.go                 # 出装服务
│   │   └── overlay.go               # Overlay 服务
│   │
│   ├── config/                      # 配置
│   │   ├── config.go                # 配置结构
│   │   └── viper.go                 # Viper 初始化
│   │
│   └── logger/                      # 日志
│       └── zap.go                   # Zap 初始化
│
├── frontend/                        # Wails 前端 (React + TypeScript)
│   ├── src/
│   │   ├── App.tsx                  # 主应用
│   │   ├── main.tsx                 # 入口
│   │   ├── components/
│   │   │   ├── ChampionCard.tsx     # 英雄卡片
│   │   │   ├── AugmentList.tsx      # 海克斯列表
│   │   │   ├── AugmentCard.tsx      # 海克斯卡片
│   │   │   ├── BuildTree.tsx        # 出装树
│   │   │   ├── WinrateBadge.tsx     # 胜率徽章
│   │   │   └── Overlay.tsx          # 游戏内悬浮窗
│   │   ├── hooks/
│   │   │   ├── useChampion.ts       # 英雄数据 Hook
│   │   │   ├── useAugment.ts        # 海克斯数据 Hook
│   │   │   └── useOverlay.ts        # Overlay 状态 Hook
│   │   ├── services/
│   │   │   └── api.ts               # Go 后端 API 封装
│   │   └── types/
│   │       └── index.ts             # TypeScript 类型定义
│   ├── package.json
│   └── vite.config.ts
│
├── assets/                          # 静态资源
│   ├── icons/                       # 应用图标
│   └── images/                      # 占位图
│
└── docs/                            # 文档
    ├── ARCHITECTURE.md              # 架构文档
    ├── DATA_SOURCES.md              # 数据源说明
    └── API.md                       # 接口文档
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

### Phase 2：客户端选人阶段（2-3 周）

| 任务 | 输出 |
|------|------|
| 监听选人事件 | 能获取队友英雄列表 |
| 英雄胜率查询 | 选人界面显示队友英雄胜率排行 |
| 海克斯推荐 | 选中英雄后显示该英雄的海克斯胜率 |
| 出装建议 | 显示推荐出装和技能加点 |
| 数据爬虫 MVP | 手动触发一次爬取，填充测试数据 |

### Phase 3：数据服务完善（1-2 周）

| 任务 | 输出 |
|------|------|
| OP.GG 爬虫 | 自动爬取英雄胜率、海克斯数据（中文数据源）|
| U.GG 爬虫 | 交叉验证 |
| Lolalytics 爬虫 | 交叉验证 |
| 定时更新 | 每 6 小时自动检查更新 |
| 数据版本管理 | 按补丁版本存储，支持历史查询 |

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

## 11. 下一步行动

1. **确认方案**：是否有需要调整或补充的部分？
2. **初始化项目**：创建 Wails + Go 项目骨架
3. **MVP 实现**：先做客户端选人阶段的英雄胜率展示
4. **数据验证**：验证 OP.GG / U.GG / Lolalytics 的数据是否可以稳定爬取
