package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const ddragonBaseURL = "https://ddragon.leagueoflegends.com"

// DDragonClient Riot DDragon API 客户端
type DDragonClient struct {
	client *http.Client
}

// NewDDragonClient 创建 DDragon 客户端
func NewDDragonClient() *DDragonClient {
	return &DDragonClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetLatestVersion 获取最新游戏版本
func (c *DDragonClient) GetLatestVersion() (string, error) {
	resp, err := c.client.Get(ddragonBaseURL + "/api/versions.json")
	if err != nil {
		return "", fmt.Errorf("fetch versions: %w", err)
	}
	defer resp.Body.Close()

	var versions []string
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", fmt.Errorf("decode versions: %w", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found")
	}

	return versions[0], nil
}

// ChampionInfo DDragon 英雄数据
type ChampionInfo struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Title string `json:"title"`
	Tags  []string `json:"tags"`
}

// ChampionData DDragon 响应结构
type ChampionData struct {
	Data map[string]ChampionInfo `json:"data"`
}

// FetchChampions 获取所有英雄基础数据
func (c *DDragonClient) FetchChampions(version string) (map[int]ChampionInfo, error) {
	// 获取英文数据（用于 ID 和 key）
	enData, err := c.fetchChampionData(version, "en_US")
	if err != nil {
		return nil, err
	}

	// 获取中文数据（用于中文名称）
	zhData, err := c.fetchChampionData(version, "zh_CN")
	if err != nil {
		return nil, err
	}

	result := make(map[int]ChampionInfo)
	for key, enChamp := range enData.Data {
		var id int
		fmt.Sscanf(enChamp.Key, "%d", &id)

		zhChamp := zhData.Data[key]
		result[id] = ChampionInfo{
			ID:    enChamp.ID,
			Key:   enChamp.Key,
			Name:  zhChamp.Name,
			Title: zhChamp.Title,
			Tags:  enChamp.Tags,
		}
	}

	return result, nil
}

func (c *DDragonClient) fetchChampionData(version, lang string) (*ChampionData, error) {
	url := fmt.Sprintf("%s/cdn/%s/data/%s/champion.json", ddragonBaseURL, version, lang)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch champion data (%s): %w", lang, err)
	}
	defer resp.Body.Close()

	var data ChampionData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode champion data: %w", err)
	}

	return &data, nil
}

// ItemInfo DDragon 装备数据
type ItemInfo struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// ItemData DDragon 装备响应
type ItemData struct {
	Data map[string]ItemInfo `json:"data"`
}

// FetchItems 获取所有装备基础数据
func (c *DDragonClient) FetchItems(version string) (map[int]ItemInfo, error) {
	enData, err := c.fetchItemData(version, "en_US")
	if err != nil {
		return nil, err
	}

	zhData, err := c.fetchItemData(version, "zh_CN")
	if err != nil {
		return nil, err
	}

	result := make(map[int]ItemInfo)
	for idStr, enItem := range enData.Data {
		var id int
		fmt.Sscanf(idStr, "%d", &id)

		zhItem := zhData.Data[idStr]
		name := zhItem.Name
		if name == "" {
			name = enItem.Name
		}

		result[id] = ItemInfo{
			Name: name,
			Tags: enItem.Tags,
		}
	}

	return result, nil
}

func (c *DDragonClient) fetchItemData(version, lang string) (*ItemData, error) {
	url := fmt.Sprintf("%s/cdn/%s/data/%s/item.json", ddragonBaseURL, version, lang)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch item data (%s): %w", lang, err)
	}
	defer resp.Body.Close()

	var data ItemData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode item data: %w", err)
	}

	return &data, nil
}
