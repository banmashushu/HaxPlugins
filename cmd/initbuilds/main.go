package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"haxPlugins/internal/data"
	"haxPlugins/internal/logger"
	"haxPlugins/internal/scraper"
)

const patch = "16.8.1"

func main() {
	fmt.Println("=== 初始化 ARAM 出装数据 ===")
	fmt.Println()

	if err := logger.Init(); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	dataDir := "./data"
	_ = os.MkdirAll(dataDir, 0755)

	db, err := data.New(filepath.Join(dataDir, "haxplugins.db"))
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()
	fmt.Println("数据库已连接")
	fmt.Println()

	champions, err := db.GetAllChampions()
	if err != nil {
		log.Fatalf("获取英雄列表失败: %v", err)
	}
	fmt.Printf("英雄数量: %d\n", len(champions))
	fmt.Println()

	client := scraper.NewMCPClient()
	var builds []data.Build
	failCount := 0

	for i, champ := range champions {
		role := selectRole(champ.Tags)
		nameEN := scraper.ToUpperSnakeCase(champ.NameEN)

		fmt.Printf("  [%3d/%d] %s (role=%s)... ", i+1, len(champions), champ.NameCN, role)

		analysis, err := client.FetchChampionAnalysis(nameEN, role, "zh_CN")
		if err != nil {
			fmt.Printf("失败: %v\n", err)
			failCount++
			continue
		}

		if analysis == nil || (len(analysis.CoreItems) == 0 && analysis.Boots == nil) {
			fmt.Printf("无数据\n")
			continue
		}

		build := data.Build{
			ChampionID:   champ.ChampionID,
			ChampionName: champ.NameCN,
			GameMode:     "hexgates",
			Role:         role,
			Items:        toDataBuildItems(analysis.CoreItems),
			Boots:        toDataBuildItem(analysis.Boots),
			SkillOrder:   analysis.SkillOrder,
			Runes:        analysis.Runes,
			Patch:        patch,
		}
		builds = append(builds, build)
		fmt.Printf("items=%d boots=%s skills=%d runes=%d\n",
			len(analysis.CoreItems), boolStr(analysis.Boots != nil), len(analysis.SkillOrder), len(analysis.Runes))

		if i < len(champions)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	fmt.Printf("\n成功获取 %d/%d 个英雄的出装数据 (失败 %d)\n", len(builds), len(champions), failCount)

	if len(builds) > 0 {
		fmt.Println("\n保存出装数据...")
		if err := db.SaveBuilds(builds); err != nil {
			log.Fatalf("保存失败: %v", err)
		}
		fmt.Printf("已保存 %d 条出装数据\n", len(builds))
	}

	fmt.Println("\n=== 初始化完成 ===")
}

func toDataBuildItems(items []scraper.BuildItem) []data.BuildItem {
	var result []data.BuildItem
	for _, it := range items {
		result = append(result, data.BuildItem{
			ItemID:  it.ItemID,
			NameCN:  it.NameCN,
			Slot:    it.Slot,
			Winrate: it.Winrate,
		})
	}
	return result
}

func toDataBuildItem(item *scraper.BuildItem) *data.BuildItem {
	if item == nil {
		return nil
	}
	return &data.BuildItem{
		ItemID:  item.ItemID,
		NameCN:  item.NameCN,
		Slot:    item.Slot,
		Winrate: item.Winrate,
	}
}

func selectRole(tags []string) string {
	priority := map[string]string{
		"Marksman": "adc",
		"Support":  "support",
		"Mage":     "mid",
		"Assassin": "mid",
		"Fighter":  "top",
		"Tank":     "top",
	}
	for _, tag := range tags {
		if role, ok := priority[tag]; ok {
			return role
		}
	}
	return "adc"
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
