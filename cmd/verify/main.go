package main

import (
	"fmt"
	"log"

	"haxPlugins/internal/scraper"
)

func main() {
	fmt.Println("=== HaxPlugins 数据源验证工具 ===")
	fmt.Println()

	// 1. 验证 DDragon
	fmt.Println("[1] 验证 DDragon API...")
	if err := verifyDDragon(); err != nil {
		log.Printf("  DDragon 验证失败: %v", err)
	} else {
		fmt.Println("  DDragon 验证通过")
	}
	fmt.Println()

	// 2. 验证 OP.GG
	fmt.Println("[2] 验证 OP.GG...")
	if err := verifyOPGG(); err != nil {
		log.Printf("  OP.GG 验证失败: %v", err)
		fmt.Println()
		fmt.Println("  提示: OP.GG 有 Cloudflare 反爬保护和动态渲染。")
		fmt.Println("  备选方案:")
		fmt.Println("    - 使用 headless browser (rod/playwright)")
		fmt.Println("    - 手动导入数据")
		fmt.Println("    - 使用其他数据源 (U.GG / Lolalytics)")
	} else {
		fmt.Println("  OP.GG 验证通过")
	}

	fmt.Println()
	fmt.Println("验证完成")
}

func verifyDDragon() error {
	client := scraper.NewDDragonClient()

	version, err := client.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("获取版本失败: %w", err)
	}
	fmt.Printf("  最新版本: %s\n", version)

	champions, err := client.FetchChampions(version)
	if err != nil {
		return fmt.Errorf("获取英雄数据失败: %w", err)
	}
	fmt.Printf("  英雄数量: %d\n", len(champions))

	// 打印前 5 个英雄
	count := 0
	for id, champ := range champions {
		if count >= 5 {
			break
		}
		fmt.Printf("    - [%d] %s (%s)\n", id, champ.Name, champ.ID)
		count++
	}

	items, err := client.FetchItems(version)
	if err != nil {
		return fmt.Errorf("获取装备数据失败: %w", err)
	}
	fmt.Printf("  装备数量: %d\n", len(items))

	return nil
}

func verifyOPGG() error {
	source := scraper.NewOPGGSource()

	// 测试版本检测
	patch, err := source.GetCurrentPatch()
	if err != nil {
		return fmt.Errorf("获取版本: %w", err)
	}
	fmt.Printf("  OP.GG 版本: %s\n", patch)

	// 测试英雄胜率爬取
	stats, err := source.ScrapeChampionStats("hexgates", patch)
	if err != nil {
		return fmt.Errorf("爬取英雄胜率: %w", err)
	}
	fmt.Printf("  英雄胜率数据: %d 条\n", len(stats))

	if len(stats) > 0 {
		fmt.Println("  前 5 名英雄:")
		for i, s := range stats {
			if i >= 5 {
				break
			}
			fmt.Printf("    %d. %s - 胜率: %.1f%%\n", i+1, s.NameCN, s.Winrate*100)
		}
	}

	return nil
}
