package scraper

import "time"

// DataSource 数据源接口
type DataSource interface {
	Name() string
	GetCurrentPatch() (string, error)
	ScrapeChampionStats(mode, patch string) ([]ChampionStat, error)
	ScrapeAugmentStats(mode, patch string) ([]HeroAugmentStat, error)
	ScrapeBuilds(mode, patch string) ([]BuildRecommendation, error)
}

// ChampionStat 英雄统计
type ChampionStat struct {
	ChampionID int
	NameCN     string
	NameEN     string
	Winrate    float64
	Pickrate   float64
	Banrate    float64
	Tier       string
	SampleSize int
	Patch      string
	Source     string
}

// HeroAugmentStat 英雄+海克斯统计
type HeroAugmentStat struct {
	ChampionID    int
	ChampionName  string
	AugmentID     string
	AugmentName   string
	AugmentNameCN string
	Winrate       float64
	Pickrate      float64
	Tier          string
	Patch         string
	Source        string
}

// BuildRecommendation 出装推荐
type BuildRecommendation struct {
	ChampionID   int
	ChampionName string
	Role         string
	Items        []BuildItem
	Boots        *BuildItem
	SkillOrder   []string
	Runes        []string
	Patch        string
	Source       string
}

// BuildItem 出装物品
type BuildItem struct {
	ItemID  int
	NameCN  string
	Slot    int
	Winrate float64
}

// Scraper 爬取引擎
type Scraper struct {
	sources  []DataSource
	interval time.Duration
}

// NewScraper 创建爬虫引擎
func NewScraper(sources []DataSource) *Scraper {
	return &Scraper{
		sources:  sources,
		interval: 6 * time.Hour,
	}
}

// ScrapeAll 爬取所有数据源
func (s *Scraper) ScrapeAll(mode, patch string) (champions []ChampionStat, augments []HeroAugmentStat, builds []BuildRecommendation, err error) {
	for _, source := range s.sources {
		c, err := source.ScrapeChampionStats(mode, patch)
		if err == nil {
			champions = append(champions, c...)
		}

		a, err := source.ScrapeAugmentStats(mode, patch)
		if err == nil {
			augments = append(augments, a...)
		}

		b, err := source.ScrapeBuilds(mode, patch)
		if err == nil {
			builds = append(builds, b...)
		}
	}

	return champions, augments, builds, nil
}
