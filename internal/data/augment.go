package data

import (
	"database/sql"
	"fmt"
)

// Augment 海克斯
type Augment struct {
	AugmentID   string `json:"augment_id"`
	NameEN      string `json:"name_en"`
	NameCN      string `json:"name_cn"`
	Description string `json:"description"`
	Tier        string `json:"tier"`
	IconURL     string `json:"icon_url"`
}

// HeroAugmentStat 英雄+海克斯组合统计
type HeroAugmentStat struct {
	ChampionID    int     `json:"champion_id"`
	ChampionName  string  `json:"champion_name"`
	AugmentID     string  `json:"augment_id"`
	AugmentName   string  `json:"augment_name"`
	AugmentNameCN string  `json:"augment_name_cn"`
	Winrate       float64 `json:"winrate"`
	Pickrate      float64 `json:"pickrate"`
	Tier          string  `json:"tier"`
	Patch         string  `json:"patch"`
}

// SaveAugments 批量保存海克斯基础数据
func (d *DB) SaveAugments(augments []Augment) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO augments (augment_id, name_en, name_cn, description, tier, icon_url)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(augment_id) DO UPDATE SET
			name_en = excluded.name_en,
			name_cn = excluded.name_cn,
			description = excluded.description,
			tier = excluded.tier,
			icon_url = excluded.icon_url
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, a := range augments {
		if _, err := stmt.Exec(a.AugmentID, a.NameEN, a.NameCN, a.Description, a.Tier, a.IconURL); err != nil {
			return fmt.Errorf("save augment %s: %w", a.AugmentID, err)
		}
	}

	return tx.Commit()
}

// GetAllAugments 获取所有海克斯
func (d *DB) GetAllAugments() ([]Augment, error) {
	rows, err := d.conn.Query(`SELECT augment_id, name_en, name_cn, description, tier, icon_url FROM augments ORDER BY name_cn`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var augments []Augment
	for rows.Next() {
		var a Augment
		if err := rows.Scan(&a.AugmentID, &a.NameEN, &a.NameCN, &a.Description, &a.Tier, &a.IconURL); err != nil {
			return nil, err
		}
		augments = append(augments, a)
	}

	return augments, rows.Err()
}

// SaveHeroAugmentStats 批量保存英雄+海克斯组合统计
func (d *DB) SaveHeroAugmentStats(stats []HeroAugmentStat, gameMode, patch string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO hero_augment_stats (champion_id, augment_id, game_mode, winrate, pickrate, tier, patch)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(champion_id, augment_id, game_mode, patch) DO UPDATE SET
			winrate = excluded.winrate,
			pickrate = excluded.pickrate,
			tier = excluded.tier,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range stats {
		if _, err := stmt.Exec(s.ChampionID, s.AugmentID, gameMode, s.Winrate, s.Pickrate, s.Tier, patch); err != nil {
			return fmt.Errorf("save hero augment stat %d-%s: %w", s.ChampionID, s.AugmentID, err)
		}
	}

	return tx.Commit()
}

// GetAugmentsForChampion 获取指定英雄的海克斯推荐
func (d *DB) GetAugmentsForChampion(championID int, gameMode, patch string) ([]HeroAugmentStat, error) {
	query := `
		SELECT ha.champion_id, c.name_cn, ha.augment_id, a.name_en, a.name_cn,
		       ha.winrate, ha.pickrate, ha.tier, ha.patch
		FROM hero_augment_stats ha
		JOIN champions c ON ha.champion_id = c.champion_id
		JOIN augments a ON ha.augment_id = a.augment_id
		WHERE ha.champion_id = ? AND ha.game_mode = ? AND ha.patch = ?
		ORDER BY ha.winrate DESC`

	rows, err := d.conn.Query(query, championID, gameMode, patch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []HeroAugmentStat
	for rows.Next() {
		var s HeroAugmentStat
		if err := rows.Scan(&s.ChampionID, &s.ChampionName, &s.AugmentID, &s.AugmentName,
			&s.AugmentNameCN, &s.Winrate, &s.Pickrate, &s.Tier, &s.Patch); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

// GetAugmentByID 按ID获取海克斯
func (d *DB) GetAugmentByID(id string) (*Augment, error) {
	var a Augment
	err := d.conn.QueryRow(`
		SELECT augment_id, name_en, name_cn, description, tier, icon_url
		FROM augments WHERE augment_id = ?
	`, id).Scan(&a.AugmentID, &a.NameEN, &a.NameCN, &a.Description, &a.Tier, &a.IconURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}
