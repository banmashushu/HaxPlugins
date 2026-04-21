package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"haxPlugins/internal/data"
	"haxPlugins/internal/logger"
)

const mcpEndpoint = "https://mcp-api.op.gg/mcp"

// MCP 文本响应中的 Augment 条目正则
// 格式: Augment(id,"name","desc",tier,performance,popular)
// 注意: name/desc 可能为 null
var augmentRe = regexp.MustCompile(`Augment\((\d+),("(?:[^"\\]|\\.)*"|null),("(?:[^"\\]|\\.)*"|null),(\d+),([\d.]+),([\d.]+)\)`)

// MCPAugment 解析后的海克斯数据
type MCPAugment struct {
	ID          int
	Name        string
	Desc        string
	Tier        int
	Performance float64
	Popular     float64
}

func main() {
	fmt.Println("=== 初始化 ARAM 海克斯数据 ===")
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

	// 获取所有英雄
	champions, err := db.GetAllChampions()
	if err != nil {
		log.Fatalf("获取英雄列表失败: %v", err)
	}
	fmt.Printf("英雄数量: %d\n", len(champions))
	fmt.Println()

	// 获取当前版本（从数据库已有数据推断，或固定值）
	patch := "16.8.1"

	// 收集所有 augment 数据
	augmentMap := make(map[int]*data.Augment)        // 去重: id -> Augment
	heroStats := make([]data.HeroAugmentStat, 0, 5000) // 英雄+海克斯组合

	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Println("开始批量获取海克斯数据...")
	for i, champ := range champions {
		fmt.Printf("  [%3d/%d] %s (id=%d)... ", i+1, len(champions), champ.NameCN, champ.ChampionID)

		augments, err := fetchAugmentsForChampion(client, champ.ChampionID)
		if err != nil {
			fmt.Printf("失败: %v\n", err)
			continue
		}
		fmt.Printf("%d 个海克斯\n", len(augments))

		for _, a := range augments {
			// 收集基础数据（去重）
			if _, exists := augmentMap[a.ID]; !exists {
				augmentMap[a.ID] = &data.Augment{
					AugmentID:   strconv.Itoa(a.ID),
					NameCN:      a.Name,
					Description: a.Desc,
					Tier:        tierToString(a.Tier),
				}
			}

			// 收集英雄+海克斯统计
			heroStats = append(heroStats, data.HeroAugmentStat{
				ChampionID:    champ.ChampionID,
				ChampionName:  champ.NameCN,
				AugmentID:     strconv.Itoa(a.ID),
				AugmentNameCN: a.Name,
				Winrate:       a.Performance, // performance 为原始强度评分（非胜率百分比）
				Pickrate:      a.Popular,
				Tier:          tierToString(a.Tier),
				Patch:         patch,
			})
		}

		// 限速: 每次请求间隔 500ms
		if i < len(champions)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	fmt.Println()
	fmt.Printf("共收集 %d 个独立海克斯，%d 条英雄+海克斯组合\n", len(augmentMap), len(heroStats))
	fmt.Println()

	// 保存海克斯基础数据
	fmt.Println("[1] 保存海克斯基础数据...")
	var augmentList []data.Augment
	for _, a := range augmentMap {
		augmentList = append(augmentList, *a)
	}
	if err := db.SaveAugments(augmentList); err != nil {
		log.Printf("  保存失败: %v", err)
	} else {
		fmt.Printf("  已保存 %d 个海克斯\n", len(augmentList))
	}

	// 保存英雄+海克斯组合数据
	fmt.Println("[2] 保存英雄+海克斯组合数据...")
	if err := db.SaveHeroAugmentStats(heroStats, "hexgates", patch); err != nil {
		log.Printf("  保存失败: %v", err)
	} else {
		fmt.Printf("  已保存 %d 条组合数据\n", len(heroStats))
	}

	// 验证
	fmt.Println()
	fmt.Println("=== 初始化完成 ===")
	allAugments, _ := db.GetAllAugments()
	fmt.Printf("海克斯总数: %d\n", len(allAugments))
}

// fetchAugmentsForChampion 调用 MCP API 获取指定英雄的海克斯数据
func fetchAugmentsForChampion(client *http.Client, championID int) ([]MCPAugment, error) {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "lol_list_aram_augments",
			"arguments": map[string]interface{}{
				"champion_id": championID,
				"lang":        "zh_CN",
				"desired_output_fields": []string{
					"data.augments[].{id,name,desc,tier,performance,popular}",
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := client.Post(mcpEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	var mcpResp struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  *struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("mcp error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	if mcpResp.Result == nil || len(mcpResp.Result.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	text := mcpResp.Result.Content[0].Text
	return parseAugmentText(text), nil
}

// parseAugmentText 解析 MCP 文本响应中的 Augment 数据
func parseAugmentText(text string) []MCPAugment {
	var augments []MCPAugment

	matches := augmentRe.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if len(m) < 7 {
			continue
		}

		id, _ := strconv.Atoi(m[1])
		name := parseMCPString(m[2])
		desc := parseMCPString(m[3])
		tier, _ := strconv.Atoi(m[4])
		perf, _ := strconv.ParseFloat(m[5], 64)
		pop, _ := strconv.ParseFloat(m[6], 64)

		augments = append(augments, MCPAugment{
			ID:          id,
			Name:        name,
			Desc:        desc,
			Tier:        tier,
			Performance: perf,
			Popular:     pop,
		})
	}

	return augments
}

// parseMCPString 解析 MCP 响应中的字符串值（处理 null 和转义）
func parseMCPString(s string) string {
	s = strings.TrimSpace(s)
	if s == "null" {
		return ""
	}
	// 去掉首尾双引号
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	// 处理转义
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}

// tierToString 将数字 tier 转为字符串等级
func tierToString(tier int) string {
	switch tier {
	case 3:
		return "silver"
	case 4:
		return "gold"
	case 5:
		return "prismatic"
	default:
		return fmt.Sprintf("tier%d", tier)
	}
}
