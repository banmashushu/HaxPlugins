package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"haxPlugins/internal/data"
	"haxPlugins/internal/logger"
	"haxPlugins/internal/scraper"
)

func main() {
	fmt.Println("=== 初始化游戏基础数据 ===")
	fmt.Println()

	if err := logger.Init(); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	// 创建数据目录
	dataDir := "./data"
	_ = os.MkdirAll(dataDir, 0755)

	// 连接数据库
	db, err := data.New(filepath.Join(dataDir, "haxplugins.db"))
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()
	fmt.Println("数据库已连接")
	fmt.Println()

	// 获取 DDragon 数据
	client := scraper.NewDDragonClient()

	version, err := client.GetLatestVersion()
	if err != nil {
		log.Fatalf("获取版本失败: %v", err)
	}
	fmt.Printf("当前游戏版本: %s\n", version)
	fmt.Println()

	// 初始化英雄数据
	fmt.Println("[1] 初始化英雄数据...")
	if err := initChampions(db, client, version); err != nil {
		log.Printf("  英雄数据初始化失败: %v", err)
	} else {
		fmt.Println("  英雄数据初始化完成")
	}

	// 初始化装备数据
	fmt.Println("[2] 初始化装备数据...")
	if err := initItems(db, client, version); err != nil {
		log.Printf("  装备数据初始化失败: %v", err)
	} else {
		fmt.Println("  装备数据初始化完成")
	}

	// 初始化模拟胜率数据（MVP 测试用）
	fmt.Println("[3] 初始化模拟胜率数据...")
	if err := initMockStats(db, version); err != nil {
		log.Printf("  模拟胜率数据初始化失败: %v", err)
	} else {
		fmt.Println("  模拟胜率数据初始化完成")
	}

	// 打印统计
	fmt.Println()
	fmt.Println("=== 初始化完成 ===")
	champions, _ := db.GetAllChampions()
	fmt.Printf("英雄数量: %d\n", len(champions))
}

func initChampions(db *data.DB, client *scraper.DDragonClient, version string) error {
	ddChampions, err := client.FetchChampions(version)
	if err != nil {
		return fmt.Errorf("获取英雄数据: %w", err)
	}

	var champions []data.Champion
	for id, info := range ddChampions {
		champions = append(champions, data.Champion{
			ChampionID: id,
			NameEN:     info.ID,
			NameCN:     info.Name,
			Title:      info.Title,
			Tags:       info.Tags,
		})
	}

	if err := db.SaveChampions(champions); err != nil {
		return fmt.Errorf("保存英雄数据: %w", err)
	}

	fmt.Printf("  已导入 %d 个英雄\n", len(champions))
	return nil
}

func initItems(db *data.DB, client *scraper.DDragonClient, version string) error {
	ddItems, err := client.FetchItems(version)
	if err != nil {
		return fmt.Errorf("获取装备数据: %w", err)
	}

	tx, err := db.Conn().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO items (item_id, name_en, name_cn, tags)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(item_id) DO UPDATE SET
			name_en = excluded.name_en,
			name_cn = excluded.name_cn,
			tags = excluded.tags
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	count := 0
	for id, info := range ddItems {
		tagsJSON, _ := json.Marshal(info.Tags)
		if _, err := stmt.Exec(id, info.NameEN, info.NameCN, string(tagsJSON)); err != nil {
			continue
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("  已导入 %d 件装备\n", count)
	return nil
}

// initMockStats 初始化模拟胜率数据（用于 MVP 测试）
func initMockStats(db *data.DB, patch string) error {
	champions, err := db.GetAllChampions()
	if err != nil {
		return err
	}

	var stats []data.ChampionStat
	for _, c := range champions {
		// 生成随机胜率（45%-55%之间）
		winrate := 0.45 + float64(c.ChampionID%100)/1000.0
		if winrate > 0.55 {
			winrate = 0.55
		}
		pickrate := 0.05 + float64(c.ChampionID%50)/1000.0

		stats = append(stats, data.ChampionStat{
			ChampionID: c.ChampionID,
			NameCN:     c.NameCN,
			Winrate:    winrate,
			Pickrate:   pickrate,
			Tier:       calculateTier(winrate),
			Patch:      patch,
		})
	}

	return db.SaveChampionStats(stats, "hexgates", patch)
}

func calculateTier(winrate float64) string {
	switch {
	case winrate >= 0.54:
		return "S"
	case winrate >= 0.51:
		return "A"
	case winrate >= 0.48:
		return "B"
	case winrate >= 0.45:
		return "C"
	default:
		return "D"
	}
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
