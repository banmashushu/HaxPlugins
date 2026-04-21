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
	opggBaseURL       = "https://op.gg"
	opggChampionAPI   = "https://lol-api-champion.op.gg/api"
)

// OPGGSource OP.GG 数据源
type OPGGSource struct {
	client *http.Client
}

// NewOPGGSource 创建 OP.GG 数据源
func NewOPGGSource() *OPGGSource {
	return &OPGGSource{
		client: NewHTTPClient(),
	}
}

// Name 返回数据源名称
func (s *OPGGSource) Name() string {
	return "opgg"
}

// GetCurrentPatch 获取当前游戏版本
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

// opggChampionMeta OP.GG 英雄元数据
type opggChampionMeta struct {
	ID   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// opggChampionMetaResponse 英雄元数据响应
type opggChampionMetaResponse struct {
	Data []opggChampionMeta `json:"data"`
}

// fetchChampionMeta 获取英雄元数据（中文名映射）
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

	var result opggChampionMetaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode champion meta: %w", err)
	}

	metaMap := make(map[int]opggChampionMeta, len(result.Data))
	for _, champ := range result.Data {
		metaMap[champ.ID] = champ
	}
	return metaMap, nil
}

// opggChampionStats OP.GG API 中的单个英雄统计
type opggChampionStats struct {
	ID           int    `json:"id"`
	Tier         string `json:"tier"`
	Rank         int    `json:"rank"`
	AverageStats struct {
		WinRate  float64 `json:"win_rate"`
		PickRate float64 `json:"pick_rate"`
		BanRate  float64 `json:"ban_rate"`
		Play     int     `json:"play"`
	} `json:"average_stats"`
}

// opggChampionStatsResponse 英雄统计数据响应
type opggChampionStatsResponse struct {
	Data []opggChampionStats `json:"data"`
}

// mapMode 将内部模式名称映射为 OP.GG API 有效模式
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

// ScrapeChampionStats 爬取英雄胜率排行（使用 OP.GG 内部 API）
func (s *OPGGSource) ScrapeChampionStats(mode, patch string) ([]ChampionStat, error) {
	// 1. 获取中文名映射
	metaMap, err := s.fetchChampionMeta()
	if err != nil {
		return nil, fmt.Errorf("fetch meta: %w", err)
	}

	// 2. 获取统计数据
	apiMode := mapMode(mode)
	url := fmt.Sprintf("%s/NA/champions/%s?tier=all&hl=zh_CN", opggChampionAPI, apiMode)

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

	var apiResp opggChampionStatsResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode champion stats: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no champion stats returned from API")
	}

	// 3. 合并数据
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
	return nil, fmt.Errorf("augment scraping not yet implemented for OP.GG")
}

// ScrapeBuilds 爬取英雄出装
func (s *OPGGSource) ScrapeBuilds(mode, patch string) ([]BuildRecommendation, error) {
	return nil, fmt.Errorf("build scraping not yet implemented for OP.GG")
}

// ScrapeAugmentList 爬取海克斯列表（不含胜率，仅基础信息）
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

// FetchChampionKeys 获取英雄英文 key 列表（用于构建 build URL）
func (s *OPGGSource) FetchChampionKeys() ([]string, error) {
	req, err := newRequest("GET", opggBaseURL+"/zh-cn/lol/modes/aram-mayhem")
	if err != nil {
		return nil, err
	}

	randomDelay()
	resp, err := retryRequest(s.client, req, 3)
	if err != nil {
		return nil, fmt.Errorf("fetch champion keys: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse champion keys: %w", err)
	}

	var keys []string
	doc.Find("table tbody tr td a[href*='/aram-mayhem/']").Each(func(i int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		parts := strings.Split(href, "/")
		if len(parts) > 0 {
			key := parts[len(parts)-1]
			if key != "" && key != "aram-mayhem" {
				keys = append(keys, key)
			}
		}
	})

	return keys, nil
}
