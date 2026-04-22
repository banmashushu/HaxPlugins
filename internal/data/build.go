package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// BuildItem 出装物品
type BuildItem struct {
	ItemID  int     `json:"item_id"`
	NameCN  string  `json:"name_cn"`
	Slot    int     `json:"slot"`
	Winrate float64 `json:"winrate"`
}

// Build 出装推荐
type Build struct {
	ChampionID int         `json:"champion_id"`
	ChampionName string    `json:"champion_name"`
	GameMode   string      `json:"game_mode"`
	Role       string      `json:"role"`
	Items      []BuildItem `json:"items"`
	Boots      *BuildItem  `json:"boots"`
	SkillOrder []string    `json:"skill_order"`
	Runes      []string    `json:"runes"`
	Patch      string      `json:"patch"`
}

// SaveBuilds 批量保存出装推荐
func (d *DB) SaveBuilds(builds []Build) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO build_recommendations (champion_id, game_mode, role, items, boots, skill_order, runes, patch)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(champion_id, game_mode, role, patch) DO UPDATE SET
			items = excluded.items,
			boots = excluded.boots,
			skill_order = excluded.skill_order,
			runes = excluded.runes,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range builds {
		itemsJSON, _ := json.Marshal(b.Items)
		bootsJSON, _ := json.Marshal(b.Boots)
		skillJSON, _ := json.Marshal(b.SkillOrder)
		runesJSON, _ := json.Marshal(b.Runes)

		if _, err := stmt.Exec(b.ChampionID, b.GameMode, b.Role, string(itemsJSON), string(bootsJSON), string(skillJSON), string(runesJSON), b.Patch); err != nil {
			return fmt.Errorf("save build for champion %d: %w", b.ChampionID, err)
		}
	}

	return tx.Commit()
}

// GetBuildForChampion 获取指定英雄的出装推荐
func (d *DB) GetBuildForChampion(championID int, gameMode, role, patch string) (*Build, error) {
	var b Build
	var itemsJSON, bootsJSON, skillJSON string

	var runesJSON string
	err := d.conn.QueryRow(`
		SELECT br.champion_id, c.name_cn, br.game_mode, br.role,
		       br.items, br.boots, br.skill_order, br.runes, br.patch
		FROM build_recommendations br
		JOIN champions c ON br.champion_id = c.champion_id
		WHERE br.champion_id = ? AND br.game_mode = ? AND br.patch = ?
		AND (br.role = ? OR ? = '')
		ORDER BY br.role
		LIMIT 1
	`, championID, gameMode, patch, role, role).Scan(
		&b.ChampionID, &b.ChampionName, &b.GameMode, &b.Role,
		&itemsJSON, &bootsJSON, &skillJSON, &runesJSON, &b.Patch)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal([]byte(itemsJSON), &b.Items)
	_ = json.Unmarshal([]byte(bootsJSON), &b.Boots)
	_ = json.Unmarshal([]byte(skillJSON), &b.SkillOrder)
	_ = json.Unmarshal([]byte(runesJSON), &b.Runes)

	return &b, nil
}
