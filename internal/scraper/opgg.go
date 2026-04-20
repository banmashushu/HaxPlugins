package scraper

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const opggBaseURL = "https://op.gg"

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

	// 直接读取响应体，用字符串搜索提取版本号
	// 例如: https://opgg-static.akamaized.net/meta/images/lol/16.8.1/champion/Neeko.png
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	body := string(bodyBytes)

	// 查找 /lol/ 后面的版本号
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

// ScrapeChampionStats 爬取英雄胜率排行
func (s *OPGGSource) ScrapeChampionStats(mode, patch string) ([]ChampionStat, error) {
	url := fmt.Sprintf("%s/zh-cn/lol/modes/aram-mayhem", opggBaseURL)
	if patch != "" {
		url += "?patch=" + patch
	}

	req, err := newRequest("GET", url)
	if err != nil {
		return nil, err
	}

	randomDelay()
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch champion stats: %w", err)
	}

	// 检查是否是 Cloudflare 拦截页
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusServiceUnavailable {
		resp.Body.Close()
		return nil, fmt.Errorf("OP.GG blocked request (status %d), possibly Cloudflare protection", resp.StatusCode)
	}

	// 先读取 body 到字符串，再解析
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse champion stats: %w", err)
	}

	var stats []ChampionStat

	// OP.GG 英雄列表表格选择器（需要根据实际页面结构调整）
	doc.Find("table tbody tr").Each(func(i int, sel *goquery.Selection) {
		// 跳过表头
		if i == 0 && sel.Find("th").Length() > 0 {
			return
		}

		cells := sel.Find("td")
		if cells.Length() < 4 {
			return
		}

		// 提取英雄名称（通常是第一列）
		nameCell := cells.Eq(1) // 假设第2列是英雄名称
		nameCN := strings.TrimSpace(nameCell.Find("strong").Text())
		if nameCN == "" {
			nameCN = strings.TrimSpace(nameCell.Text())
		}

		// 提取胜率
		winrateText := strings.TrimSpace(cells.Eq(2).Text())
		winrateText = strings.TrimSuffix(winrateText, "%")
		winrate, _ := strconv.ParseFloat(winrateText, 64)

		// 提取登场率
		pickrateText := strings.TrimSpace(cells.Eq(3).Text())
		pickrateText = strings.TrimSuffix(pickrateText, "%")
		pickrate, _ := strconv.ParseFloat(pickrateText, 64)

		// 提取段位/评级
		tier := strings.TrimSpace(cells.Eq(0).Text())

		if nameCN != "" && winrate > 0 {
			stats = append(stats, ChampionStat{
				NameCN:   nameCN,
				Winrate:  winrate / 100,
				Pickrate: pickrate / 100,
				Tier:     tier,
				Patch:    patch,
				Source:   s.Name(),
			})
		}
	})

	if len(stats) == 0 {
		return nil, fmt.Errorf("no champion stats found, page structure may have changed")
	}

	return stats, nil
}

// ScrapeAugmentStats 爬取海克斯数据
func (s *OPGGSource) ScrapeAugmentStats(mode, patch string) ([]HeroAugmentStat, error) {
	// OP.GG 的 augment 数据通常在英雄详情页中
	// 这里返回空，需要通过其他方式获取
	// TODO: 实现 augment 数据爬取
	return nil, fmt.Errorf("augment scraping not yet implemented for OP.GG")
}

// ScrapeBuilds 爬取英雄出装
func (s *OPGGSource) ScrapeBuilds(mode, patch string) ([]BuildRecommendation, error) {
	// TODO: 需要遍历每个英雄的 build 页面
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

	// 根据 OP.GG 页面结构查找 augment 筛选区
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

	// 爬取出装顺序
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

	// 爬取符文
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
