package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Champion 英雄
type Champion struct {
	ChampionID int      `json:"champion_id"`
	NameEN     string   `json:"name_en"`
	NameCN     string   `json:"name_cn"`
	Title      string   `json:"title"`
	Tags       []string `json:"tags"`
}

// ChampionStat 英雄统计
type ChampionStat struct {
	ChampionID int     `json:"champion_id"`
	NameCN     string  `json:"name_cn"`
	Winrate    float64 `json:"winrate"`
	Pickrate   float64 `json:"pickrate"`
	Banrate    float64 `json:"banrate"`
	Tier       string  `json:"tier"`
	SampleSize int     `json:"sample_size"`
	Patch      string  `json:"patch"`
}

// SaveChampions 批量保存英雄基础数据
func (d *DB) SaveChampions(champions []Champion) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO champions (champion_id, name_en, name_cn, title, tags)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(champion_id) DO UPDATE SET
			name_en = excluded.name_en,
			name_cn = excluded.name_cn,
			title = excluded.title,
			tags = excluded.tags
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range champions {
		tagsJSON, _ := json.Marshal(c.Tags)
		if _, err := stmt.Exec(c.ChampionID, c.NameEN, c.NameCN, c.Title, string(tagsJSON)); err != nil {
			return fmt.Errorf("save champion %d: %w", c.ChampionID, err)
		}
	}

	return tx.Commit()
}

// GetAllChampions 获取所有英雄
func (d *DB) GetAllChampions() ([]Champion, error) {
	rows, err := d.conn.Query(`SELECT champion_id, name_en, name_cn, title, tags FROM champions ORDER BY name_cn`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var champions []Champion
	for rows.Next() {
		var c Champion
		var tagsJSON string
		if err := rows.Scan(&c.ChampionID, &c.NameEN, &c.NameCN, &c.Title, &tagsJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
		champions = append(champions, c)
	}

	return champions, rows.Err()
}

// GetChampionByID 按ID获取英雄
func (d *DB) GetChampionByID(id int) (*Champion, error) {
	var c Champion
	var tagsJSON string
	err := d.conn.QueryRow(`
		SELECT champion_id, name_en, name_cn, title, tags
		FROM champions WHERE champion_id = ?
	`, id).Scan(&c.ChampionID, &c.NameEN, &c.NameCN, &c.Title, &tagsJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
	return &c, nil
}

// SaveChampionStats 批量保存英雄统计数据
func (d *DB) SaveChampionStats(stats []ChampionStat, gameMode, patch string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO champion_stats (champion_id, game_mode, winrate, pickrate, banrate, tier, sample_size, patch)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(champion_id, game_mode, patch) DO UPDATE SET
			winrate = excluded.winrate,
			pickrate = excluded.pickrate,
			banrate = excluded.banrate,
			tier = excluded.tier,
			sample_size = excluded.sample_size,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range stats {
		if _, err := stmt.Exec(s.ChampionID, gameMode, s.Winrate, s.Pickrate, s.Banrate, s.Tier, s.SampleSize, patch); err != nil {
			return fmt.Errorf("save stats for champion %d: %w", s.ChampionID, err)
		}
	}

	return tx.Commit()
}

// GetChampionStats 获取英雄统计数据
func (d *DB) GetChampionStats(championIDs []int, gameMode, patch string) ([]ChampionStat, error) {
	if len(championIDs) == 0 {
		return d.getAllChampionStats(gameMode, patch)
	}

	query := `
		SELECT c.champion_id, c.name_cn, s.winrate, s.pickrate, s.banrate, s.tier, s.sample_size, s.patch
		FROM champion_stats s
		JOIN champions c ON s.champion_id = c.champion_id
		WHERE s.game_mode = ? AND s.patch = ? AND s.champion_id IN (`

	args := []interface{}{gameMode, patch}
	for i, id := range championIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += `)
		ORDER BY s.winrate DESC`

	return d.queryChampionStats(query, args...)
}

func (d *DB) getAllChampionStats(gameMode, patch string) ([]ChampionStat, error) {
	query := `
		SELECT c.champion_id, c.name_cn, s.winrate, s.pickrate, s.banrate, s.tier, s.sample_size, s.patch
		FROM champion_stats s
		JOIN champions c ON s.champion_id = c.champion_id
		WHERE s.game_mode = ? AND s.patch = ?
		ORDER BY s.winrate DESC`

	return d.queryChampionStats(query, gameMode, patch)
}

func (d *DB) queryChampionStats(query string, args ...interface{}) ([]ChampionStat, error) {
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ChampionStat
	for rows.Next() {
		var s ChampionStat
		if err := rows.Scan(&s.ChampionID, &s.NameCN, &s.Winrate, &s.Pickrate, &s.Banrate, &s.Tier, &s.SampleSize, &s.Patch); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}
