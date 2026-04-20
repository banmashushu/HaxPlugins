package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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
		fmt.Println("  提示: OP.GG 可能有 Cloudflare 反爬保护。")
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
	// 先手动测试 HTTP 请求
	fmt.Println("  [调试] 测试直接 HTTP 请求...")
	req, _ := http.NewRequest("GET", "https://op.gg/zh-cn/lol/modes/aram-mayhem", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  HTTP 请求失败: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("  HTTP 状态码: %d\n", resp.StatusCode)
	fmt.Printf("  Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("  Content-Encoding: %s\n", resp.Header.Get("Content-Encoding"))

	// 读取更多响应体
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("  响应体大小: %d bytes\n", len(bodyBytes))
	body := string(bodyBytes)
	if idx := strings.Index(body, "/lol/"); idx != -1 {
		fmt.Printf("  找到 '/lol/' 在位置 %d\n", idx)
	} else {
		fmt.Println("  未在响应体中找到 '/lol/'")
		// 打印前 500 字符看看是什么内容
		if len(body) > 500 {
			fmt.Printf("  前500字符: %s...\n", body[:500])
		}
	}

	source := scraper.NewOPGGSource()

	// 跳过版本检测，直接用已知版本测试
	patch := "16.8.1"
	fmt.Printf("  OP.GG 版本: %s (跳过检测)\n", patch)

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
