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
	fmt.Println("=== 初始化英雄协同数据 ===")
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
	totalSynergies := 0
	failCount := 0

	for i, champ := range champions {
		myPos := selectRole(champ.Tags)
		synergyPos := "support"
		if myPos == "support" {
			synergyPos = "adc"
		}
		nameEN := scraper.ToUpperSnakeCase(champ.NameEN)

		fmt.Printf("  [%3d/%d] %s (%s→%s)... ", i+1, len(champions), champ.NameCN, myPos, synergyPos)

		text, err := client.FetchChampionSynergies(nameEN, myPos, synergyPos, "zh_CN")
		if err != nil {
			fmt.Printf("失败: %v\n", err)
			failCount++
			continue
		}

		synergies, err := scraper.ParseSynergies(text, champ.ChampionID, champ.NameCN)
		if err != nil {
			fmt.Printf("解析失败: %v\n", err)
			failCount++
			continue
		}

		if len(synergies) == 0 {
			fmt.Printf("无数据\n")
			continue
		}

		var records []data.ChampionSynergy
		for _, s := range synergies {
			records = append(records, data.ChampionSynergy{
				ChampionID:        champ.ChampionID,
				ChampionName:      champ.NameCN,
				SynergyChampionID: s.SynergyChampionID,
				SynergyName:       s.SynergyName,
				ScoreRank:         s.ScoreRank,
				Score:             s.Score,
				Play:              s.Play,
				Win:               s.Win,
				WinRate:           s.WinRate * 100,
				Tier:              s.Tier,
				GameMode:          "hexgates",
				Patch:             patch,
			})
		}

		if err := db.SaveSynergies(records, "hexgates", patch); err != nil {
			fmt.Printf("保存失败: %v\n", err)
			failCount++
			continue
		}

		totalSynergies += len(records)
		fmt.Printf("%d 条协同\n", len(records))

		if i < len(champions)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	fmt.Printf("\n成功获取 %d 个英雄的协同数据，共 %d 条记录 (失败 %d)\n",
		len(champions)-failCount, totalSynergies, failCount)
	fmt.Println("\n=== 初始化完成 ===")
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
