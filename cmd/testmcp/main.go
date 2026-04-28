package main

import (
	"fmt"
	"haxPlugins/internal/scraper"
)

func main() {
	client := scraper.NewMCPClient()
	for _, pos := range []string{"adc", "support", "mid", "top", "jungle", "mage", "tank"} {
		_, err := client.FetchChampionAnalysis("ASHE", pos, "zh_CN")
		if err != nil {
			fmt.Printf("%s: %v\n", pos, err)
		} else {
			fmt.Printf("%s: OK\n", pos)
		}
	}
}
