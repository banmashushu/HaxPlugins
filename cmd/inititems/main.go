package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"haxPlugins/internal/data"
	"haxPlugins/internal/logger"
	"haxPlugins/internal/scraper"
)

func main() {
	fmt.Println("=== 初始化物品数据 ===")
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

	client := scraper.NewMCPClient()

	fmt.Println("获取嚎哭深渊物品数据...")
	text, err := client.CallTool("lol_list_items", map[string]any{
		"map":  "HOWLING_ABYSS",
		"lang": "zh_CN",
	})
	if err != nil {
		log.Fatalf("获取物品数据失败: %v", err)
	}

	var resp struct {
		Lang string `json:"lang"`
		Map  string `json:"map"`
		Data struct {
			Items []struct {
				ItemID         int    `json:"item_id"`
				Name           string `json:"name"`
				IntoItems      []int  `json:"into_items"`
				FromItems      []int  `json:"from_items"`
				GoldSell       int    `json:"gold_sell"`
				GoldTotal      int    `json:"gold_total"`
				GoldPurchasable bool  `json:"gold_purchasable"`
				Description    string `json:"description"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		log.Fatalf("解析物品数据失败: %v", err)
	}

	fmt.Printf("获取到 %d 个物品\n", len(resp.Data.Items))

	var items []data.Item
	for _, it := range resp.Data.Items {
		tagsJSON, _ := json.Marshal(it.IntoItems)
		statsJSON, _ := json.Marshal(map[string]any{
			"gold_sell":  it.GoldSell,
			"gold_total": it.GoldTotal,
			"purchasable": it.GoldPurchasable,
		})

		items = append(items, data.Item{
			ItemID:   it.ItemID,
			NameEN:   "", // API only returns localized name
			NameCN:   it.Name,
			Tags:     string(tagsJSON),
			Stats:    string(statsJSON),
		})
	}

	fmt.Println("保存物品数据...")
	if err := db.SaveItems(items); err != nil {
		log.Fatalf("保存失败: %v", err)
	}
	fmt.Printf("已保存 %d 个物品\n", len(items))

	fmt.Println("\n=== 初始化完成 ===")
}
