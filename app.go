package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"haxPlugins/internal/data"
	"haxPlugins/internal/lcu"
	"haxPlugins/internal/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	gameModeHexgates = "hexgates"
	currentPatch     = "16.8.1"
	ddragonBaseURL  = "https://ddragon.leagueoflegends.com/cdn"
)

// App struct
type App struct {
	ctx       context.Context
	db        *data.DB
	lcuClient *lcu.Client
	mockMode  bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := logger.Init(); err != nil {
		runtime.LogErrorf(ctx, "初始化日志失败: %v", err)
	}

	// 初始化数据库
	dbPath := filepath.Join(".", "data", "haxplugins.db")
	db, err := data.New(dbPath)
	if err != nil {
		runtime.LogErrorf(ctx, "初始化数据库失败: %v", err)
		return
	}
	a.db = db

	// 初始化 LCU 客户端
	lcuClient, err := lcu.NewClient()
	if err != nil {
		runtime.LogWarningf(ctx, "创建 LCU 客户端失败，启用 Mock 模式: %v", err)
		a.startMockMode()
	} else {
		a.lcuClient = lcuClient
		// 连接 LOL 客户端（非阻塞，因为 LOL 可能还未启动）
		go a.connectLCU()
	}
}

// connectLCU 连接 LOL 客户端并启动事件监听
func (a *App) connectLCU() {
	if err := a.lcuClient.Connect(); err != nil {
		runtime.LogWarningf(a.ctx, "连接 LOL 客户端失败，自动切换到 Mock 模式: %v", err)
		a.startMockMode()
		return
	}

	runtime.LogInfo(a.ctx, "LCU 连接成功")

	// 订阅游戏阶段变化事件并推送到前端
	a.lcuClient.Subscribe(lcu.EventGamePhaseChanged, func(data interface{}) {
		phase, _ := data.(string)
		runtime.EventsEmit(a.ctx, "game:phase", phase)
	})

	// 订阅选人会话更新事件并推送到前端
	a.lcuClient.Subscribe(lcu.EventChampSelectUpdate, func(data interface{}) {
		runtime.EventsEmit(a.ctx, "game:champselect", data)
	})

	if err := a.lcuClient.StartListening(); err != nil {
		runtime.LogErrorf(a.ctx, "启动 LCU 事件监听失败: %v", err)
	}
}

// mockTeamChampions 预设 Mock 队友英雄 ID
var mockTeamChampions = []int{86, 22, 17, 222, 157} // 盖伦, 艾希, 提莫, 金克丝, 亚索

// mockEnemyChampions 预设 Mock 敌方英雄 ID
var mockEnemyChampions = []int{238, 103, 84, 11, 1} // 劫, 阿狸, 赵信, 易, 安妮

// startMockMode 启动 Mock 模式（无 LOL 客户端时用于 UI 测试）
func (a *App) startMockMode() {
	a.mockMode = true
	runtime.LogInfo(a.ctx, "Mock 模式已启动")

	// 推送模拟游戏阶段
	runtime.EventsEmit(a.ctx, "game:phase", "ChampSelect")

	// 每 10 秒推送一次选人会话更新，模拟数据刷新
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				runtime.EventsEmit(a.ctx, "game:champselect", map[string]interface{}{"mock": true})
			}
		}
	}()
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	if a.lcuClient != nil {
		_ = a.lcuClient.Disconnect()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
	logger.Sync()
}

// GetCurrentPhase 获取当前游戏阶段
func (a *App) GetCurrentPhase() (string, error) {
	if a.mockMode {
		return "ChampSelect", nil
	}
	if a.lcuClient == nil {
		return "", fmt.Errorf("LCU 客户端未初始化")
	}
	phase, err := a.lcuClient.GetGamePhase()
	if err != nil {
		return "", err
	}
	return string(phase), nil
}

// TeamMemberStats 队友统计信息
type TeamMemberStats struct {
	ChampionID     int                    `json:"champion_id"`
	ChampionName   string                 `json:"champion_name"`
	ChampionNameEN string                 `json:"champion_name_en"`
	CellID         int                    `json:"cell_id"`
	Winrate        float64                `json:"winrate"`
	Pickrate       float64                `json:"pickrate"`
	Tier           string                 `json:"tier"`
	Augments       []data.HeroAugmentStat `json:"augments"`
	Build          *data.Build            `json:"build"`
	Synergies      []data.ChampionSynergy `json:"synergies"`
}

// GetMyTeamStats 获取队友列表及统计数据
func (a *App) GetMyTeamStats() ([]TeamMemberStats, error) {
	return a.getTeamStats(func() ([]int, error) {
		if a.mockMode {
			return mockTeamChampions, nil
		}
		if a.lcuClient == nil {
			return nil, fmt.Errorf("LCU 客户端未初始化")
		}
		members, err := a.lcuClient.GetMyTeam()
		if err != nil {
			return nil, fmt.Errorf("获取队友列表失败: %w", err)
		}
		var IDs []int
		for _, m := range members {
			IDs = append(IDs, m.ChampionID)
		}
		return IDs, nil
	})
}

// GetEnemyTeamStats 获取敌方队伍列表及统计数据
func (a *App) GetEnemyTeamStats() ([]TeamMemberStats, error) {
	return a.getTeamStats(func() ([]int, error) {
		if a.mockMode {
			return mockEnemyChampions, nil
		}
		if a.lcuClient == nil {
			return nil, fmt.Errorf("LCU 客户端未初始化")
		}
		members, err := a.lcuClient.GetEnemyTeam()
		if err != nil {
			return nil, fmt.Errorf("获取敌方列表失败: %w", err)
		}
		var IDs []int
		for _, m := range members {
			IDs = append(IDs, m.ChampionID)
		}
		return IDs, nil
	})
}

func (a *App) getTeamStats(getChampionIDs func() ([]int, error)) ([]TeamMemberStats, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	championIDs, err := getChampionIDs()
	if err != nil {
		return nil, err
	}

	var result []TeamMemberStats
	for i, cid := range championIDs {
		champion, err := a.db.GetChampionByID(cid)
		if err != nil {
			runtime.LogErrorf(a.ctx, "获取英雄信息失败: %v", err)
		}

		stats := TeamMemberStats{
			ChampionID:     cid,
			ChampionName:   "",
			ChampionNameEN: "",
			CellID:         i,
		}

		if champion != nil {
			stats.ChampionName = champion.NameCN
			stats.ChampionNameEN = champion.NameEN
		}

		championStats, err := a.db.GetChampionStats([]int{cid}, gameModeHexgates, currentPatch)
		if err != nil {
			runtime.LogErrorf(a.ctx, "获取英雄胜率失败: %v", err)
		} else if len(championStats) > 0 {
			stats.Winrate = championStats[0].Winrate
			stats.Pickrate = championStats[0].Pickrate
			stats.Tier = championStats[0].Tier
		}

		augments, err := a.db.GetAugmentsForChampion(cid, gameModeHexgates, currentPatch)
		if err != nil {
			runtime.LogErrorf(a.ctx, "获取海克斯推荐失败: %v", err)
		} else if len(augments) > 5 {
			stats.Augments = augments[:5]
		} else {
			stats.Augments = augments
		}

		build, err := a.db.GetBuildForChampion(cid, gameModeHexgates, "", currentPatch)
		if err != nil {
			runtime.LogErrorf(a.ctx, "获取出装推荐失败: %v", err)
		} else {
			stats.Build = build
		}

		synergies, err := a.db.GetSynergiesForChampion(cid, gameModeHexgates, currentPatch)
		if err != nil {
			runtime.LogErrorf(a.ctx, "获取协同推荐失败: %v", err)
		} else {
			stats.Synergies = synergies
		}

		result = append(result, stats)
	}

	return result, nil
}

// GetAugmentRecommendations 获取指定英雄的海克斯推荐
func (a *App) GetAugmentRecommendations(championID int) ([]data.HeroAugmentStat, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	augments, err := a.db.GetAugmentsForChampion(championID, gameModeHexgates, currentPatch)
	if err != nil {
		return nil, fmt.Errorf("获取海克斯推荐失败: %w", err)
	}

	return augments, nil
}

// ChampionImage 英雄图标信息
type ChampionImage struct {
	ImageURL   string `json:"image_url"`
	NameEN     string `json:"name_en"`
	ChampionID int    `json:"champion_id"`
}

// GetChampionImages 批量获取指定英雄的图标 URL
func (a *App) GetChampionImages(championIDs []int) ([]ChampionImage, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var images []ChampionImage
	for _, cid := range championIDs {
		champion, err := a.db.GetChampionByID(cid)
		if err != nil || champion == nil {
			images = append(images, ChampionImage{ChampionID: cid, ImageURL: "", NameEN: ""})
			continue
		}
		key := champion.NameEN
		imageURL := fmt.Sprintf("%s/%s/img/champion/%s.png", ddragonBaseURL, currentPatch, key)
		images = append(images, ChampionImage{
			ImageURL:   imageURL,
			NameEN:     key,
			ChampionID: cid,
		})
	}

	return images, nil
}

// GetItemImageURL 获取装备图标 URL
func (a *App) GetItemImageURL(itemID int) string {
	return fmt.Sprintf("%s/%s/img/item/%d.png", ddragonBaseURL, currentPatch, itemID)
}

// GetBuildRecommendation 获取指定英雄的出装推荐
func (a *App) GetBuildRecommendation(championID int) (*data.Build, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	build, err := a.db.GetBuildForChampion(championID, gameModeHexgates, "", currentPatch)
	if err != nil {
		return nil, fmt.Errorf("获取出装推荐失败: %w", err)
	}

	return build, nil
}
