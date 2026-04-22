package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const mcpEndpoint = "https://mcp-api.op.gg/mcp"

// MCPClient MCP API client
type MCPClient struct {
	client *http.Client
}

// NewMCPClient creates an MCP client
func NewMCPClient() *MCPClient {
	return &MCPClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// MCPClass represents a parsed class node from MCP response
type MCPClass struct {
	Name   string
	Fields []any
}

// StringField gets a string field by index
func (c *MCPClass) StringField(idx int) string {
	if idx >= len(c.Fields) {
		return ""
	}
	if s, ok := c.Fields[idx].(string); ok {
		return s
	}
	return ""
}

// FloatField gets a float64 field by index
func (c *MCPClass) FloatField(idx int) float64 {
	if idx >= len(c.Fields) {
		return 0
	}
	if f, ok := c.Fields[idx].(float64); ok {
		return f
	}
	return 0
}

// IntField gets an int field by index
func (c *MCPClass) IntField(idx int) int {
	return int(c.FloatField(idx))
}

// ArrayField gets an array field by index
func (c *MCPClass) ArrayField(idx int) []any {
	if idx >= len(c.Fields) {
		return nil
	}
	if a, ok := c.Fields[idx].([]any); ok {
		return a
	}
	return nil
}

// ClassField gets a nested class field by index
func (c *MCPClass) ClassField(idx int) *MCPClass {
	if idx >= len(c.Fields) {
		return nil
	}
	if cl, ok := c.Fields[idx].(*MCPClass); ok {
		return cl
	}
	return nil
}

// CallTool calls an MCP tool and returns the text content
func (c *MCPClient) CallTool(toolName string, arguments map[string]any) (string, error) {
	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.client.Post(mcpEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
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
		return "", fmt.Errorf("decode response: %w", err)
	}

	if mcpResp.Error != nil {
		return "", fmt.Errorf("mcp error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	if mcpResp.Result == nil || len(mcpResp.Result.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return mcpResp.Result.Content[0].Text, nil
}

// AnalysisResult holds parsed champion analysis data
type AnalysisResult struct {
	ChampionID   int
	ChampionName string
	CoreItems    []BuildItem
	Boots        *BuildItem
	SkillOrder   []string
	Runes        []string
	Winrate      float64
	Pickrate     float64
	SampleSize   int
}

// ParseAnalysis parses lol_get_champion_analysis response into AnalysisResult
func ParseAnalysis(text string, championID int, championName string) (*AnalysisResult, error) {
	text = stripClassDefinitions(text)
	root, err := parseMCPClass(text)
	if err != nil {
		return nil, fmt.Errorf("parse mcp class: %w", err)
	}

	if root.Name != "LolGetChampionAnalysis" || len(root.Fields) == 0 {
		return nil, fmt.Errorf("unexpected root class: %s", root.Name)
	}

	data := root.ClassField(2)
	if data == nil || data.Name != "Data" {
		return nil, fmt.Errorf("no Data field found")
	}

	result := &AnalysisResult{
		ChampionID:   championID,
		ChampionName: championName,
	}

	// Data field order (from class schema):
	// 0:summary, 1:damage_type, 2:strong_counters, 3:weak_counters, 4:synergies,
	// 5:core_items, 6:mythic_items, 7:boots, 8:starter_items, 9:last_items,
	// 10:fourth_items, 11:fifth_items, 12:sixth_items, 13:summoner_spells,
	// 14:runes, 15:skills, 16:skill_combos, 17:skill_masteries, 18:trends

	var coreItems, bootsItems, skillsClass, runesClass, summaryClass *MCPClass

	if len(data.Fields) > 0 {
		summaryClass, _ = data.Fields[0].(*MCPClass)
	}
	if len(data.Fields) > 5 {
		coreItems, _ = data.Fields[5].(*MCPClass)
	}
	if len(data.Fields) > 7 {
		bootsItems, _ = data.Fields[7].(*MCPClass)
	}
	if len(data.Fields) > 14 {
		runesClass, _ = data.Fields[14].(*MCPClass)
	}
	if len(data.Fields) > 15 {
		skillsClass, _ = data.Fields[15].(*MCPClass)
	}

	// Parse summary for winrate
	if summaryClass != nil {
		avgStats := summaryClass.ClassField(0)
		if avgStats != nil && avgStats.Name == "AverageStats" {
			result.Winrate = avgStats.FloatField(1) * 100
			result.Pickrate = avgStats.FloatField(2) * 100
			result.SampleSize = avgStats.IntField(0)
		}
	}

	// Parse core items (first CoreItems after summary)
	if coreItems != nil {
		itemIDs := intArrayFromInterface(coreItems.ArrayField(0))
		itemNames := stringArrayFromInterface(coreItems.ArrayField(1))
		pickRate := coreItems.FloatField(4)
		for i := 0; i < len(itemIDs) && i < len(itemNames); i++ {
			result.CoreItems = append(result.CoreItems, BuildItem{
				ItemID:  itemIDs[i],
				NameCN:  itemNames[i],
				Slot:    i + 1,
				Winrate: pickRate * 100,
			})
		}
	}

	// Parse boots (second CoreItems, typically single item)
	if bootsItems != nil {
		itemIDs := intArrayFromInterface(bootsItems.ArrayField(0))
		itemNames := stringArrayFromInterface(bootsItems.ArrayField(1))
		pickRate := bootsItems.FloatField(4)
		if len(itemIDs) > 0 && len(itemNames) > 0 {
			result.Boots = &BuildItem{
				ItemID:  itemIDs[0],
				NameCN:  itemNames[0],
				Slot:    0,
				Winrate: pickRate * 100,
			}
		}
	}

	// Parse skill order (first Skills with array order)
	if skillsClass != nil {
		order := stringArrayFromInterface(skillsClass.ArrayField(0))
		result.SkillOrder = order
	}

	// Parse runes (primary names at idx 4, secondary names at idx 8, stat mods at idx 10)
	if runesClass != nil {
		primaryNames := stringArrayFromInterface(runesClass.ArrayField(4))
		secondaryNames := stringArrayFromInterface(runesClass.ArrayField(8))
		statMods := stringArrayFromInterface(runesClass.ArrayField(10))
		result.Runes = append(primaryNames, secondaryNames...)
		result.Runes = append(result.Runes, statMods...)
	}

	return result, nil
}

// FetchChampionAnalysis fetches build analysis for a champion via MCP
func (c *MCPClient) FetchChampionAnalysis(championNameEN, position, lang string) (*AnalysisResult, error) {
	text, err := c.CallTool("lol_get_champion_analysis", map[string]any{
		"champion":   championNameEN,
		"game_mode":  "aram",
		"position":   position,
		"lang":       lang,
	})
	if err != nil {
		return nil, err
	}
	return ParseAnalysis(text, 0, championNameEN)
}

// FetchChampionSynergies fetches synergy recommendations for a champion
func (c *MCPClient) FetchChampionSynergies(championNameEN, myPosition, synergyPosition, lang string) (string, error) {
	return c.CallTool("lol_get_champion_synergies", map[string]any{
		"champion":         championNameEN,
		"my_position":      myPosition,
		"synergy_position": synergyPosition,
		"lang":             lang,
	})
}

// SynergyResult represents a single synergy recommendation
type SynergyResult struct {
	SynergyChampionID int
	SynergyName       string
	ScoreRank         int
	Score             float64
	Play              int
	Win               int
	WinRate           float64
	Tier              int
}

// ParseSynergies parses lol_get_champion_synergies response
func ParseSynergies(text string, championID int, championName string) ([]SynergyResult, error) {
	text = stripClassDefinitions(text)
	root, err := parseMCPClass(text)
	if err != nil {
		return nil, fmt.Errorf("parse mcp class: %w", err)
	}

	if root.Name != "LolGetChampionSynergies" || len(root.Fields) == 0 {
		return nil, fmt.Errorf("unexpected root class: %s", root.Name)
	}

	data := root.ClassField(4)
	if data == nil || data.Name != "Data" {
		return nil, fmt.Errorf("no Data field found")
	}

	var synergies []SynergyResult
	for _, field := range data.Fields {
		arr, ok := field.([]any)
		if !ok {
			continue
		}
		for _, elem := range arr {
			syn, ok := elem.(*MCPClass)
			if !ok || syn.Name != "Synergie" {
				continue
			}
			// Synergie(champion_id, champion_name, position, synergy_champion_id,
			//   synergy_champion_name, synergy_position, score_rank, score, play, win, win_rate, synergy_tier_data)
			tier := 0
			if tierData, ok := syn.Fields[11].(*MCPClass); ok {
				tier = tierData.IntField(0)
			}
			synergies = append(synergies, SynergyResult{
				SynergyChampionID: syn.IntField(3),
				SynergyName:       syn.StringField(4),
				ScoreRank:         syn.IntField(6),
				Score:             syn.FloatField(7),
				Play:              syn.IntField(8),
				Win:               syn.IntField(9),
				WinRate:           syn.FloatField(10),
				Tier:              tier,
			})
		}
	}

	return synergies, nil
}

// --- Parser internals ---

func parseMCPClass(text string) (*MCPClass, error) {
	text = strings.TrimSpace(text)
	parenIdx := strings.Index(text, "(")
	if parenIdx == -1 {
		return nil, fmt.Errorf("no opening paren in: %.50s", text)
	}

	name := strings.TrimSpace(text[:parenIdx])
	endIdx := findMatchingParen(text, parenIdx)
	if endIdx == -1 {
		return nil, fmt.Errorf("no matching paren for %s", name)
	}

	inner := text[parenIdx+1 : endIdx]
	fields := splitFields(inner)

	var parsedFields []any
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		parsedFields = append(parsedFields, parseValue(f))
	}

	return &MCPClass{Name: name, Fields: parsedFields}, nil
}

func parseValue(text string) any {
	text = strings.TrimSpace(text)

	// String literal
	if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
		return unescapeString(text[1 : len(text)-1])
	}

	// Array
	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		return parseArray(text)
	}

	// Nested class
	if strings.Contains(text, "(") {
		if nested, err := parseMCPClass(text); err == nil {
			return nested
		}
	}

	// Number
	if num, err := strconv.ParseFloat(text, 64); err == nil {
		return num
	}

	// Fallback: return as string
	return text
}

func parseArray(text string) []any {
	inner := text[1 : len(text)-1]
	fields := splitFields(inner)
	var result []any
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		result = append(result, parseValue(f))
	}
	return result
}

func findMatchingParen(text string, start int) int {
	depth := 1
	inString := false
	for i := start + 1; i < len(text); i++ {
		c := text[i]
		if c == '"' && (i == 0 || text[i-1] != '\\') {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitFields(text string) []string {
	var fields []string
	var current strings.Builder
	depth := 0
	inString := false

	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '"' && (i == 0 || text[i-1] != '\\') {
			inString = !inString
			current.WriteByte(c)
			continue
		}
		if inString {
			current.WriteByte(c)
			continue
		}
		if c == '(' || c == '[' {
			depth++
			current.WriteByte(c)
		} else if c == ')' || c == ']' {
			depth--
			current.WriteByte(c)
		} else if c == ',' && depth == 0 {
			fields = append(fields, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}
	return fields
}

func unescapeString(s string) string {
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}

func stringArrayFromInterface(arr []any) []string {
	var result []string
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func intArrayFromInterface(arr []any) []int {
	var result []int
	for _, v := range arr {
		switch n := v.(type) {
		case float64:
			result = append(result, int(n))
		case int:
			result = append(result, n)
		}
	}
	return result
}

// ToUpperSnakeCase converts TitleCase champion name to UPPER_SNAKE_CASE
func ToUpperSnakeCase(s string) string {
	if s == "" {
		return ""
	}
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			if prev >= 'a' && prev <= 'z' {
				result.WriteRune('_')
			} else if i+1 < len(runes) {
				next := runes[i+1]
				if next >= 'a' && next <= 'z' {
					result.WriteRune('_')
				}
			}
		}
		result.WriteRune(r)
	}
	return strings.ToUpper(result.String())
}

// stripClassDefinitions removes class definition lines from MCP text responses.
// Some MCP tools return class schemas before the actual serialized data.
func stripClassDefinitions(text string) string {
	lines := strings.Split(text, "\n")
	var dataLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "class ") {
			continue
		}
		dataLines = append(dataLines, line)
	}
	return strings.Join(dataLines, "\n")
}
