package data

import (
	"fmt"
)

// ChampionSynergy represents a synergy recommendation between two champions
type ChampionSynergy struct {
	ChampionID        int     `json:"champion_id"`
	ChampionName      string  `json:"champion_name"`
	SynergyChampionID int     `json:"synergy_champion_id"`
	SynergyName       string  `json:"synergy_name"`
	ScoreRank         int     `json:"score_rank"`
	Score             float64 `json:"score"`
	Play              int     `json:"play"`
	Win               int     `json:"win"`
	WinRate           float64 `json:"win_rate"`
	Tier              int     `json:"tier"`
	GameMode          string  `json:"game_mode"`
	Patch             string  `json:"patch"`
}

// SaveSynergies batch saves synergy data
func (d *DB) SaveSynergies(synergies []ChampionSynergy, gameMode, patch string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO champion_synergies (champion_id, champion_name, synergy_champion_id, synergy_name,
			score_rank, score, play, win, win_rate, tier, game_mode, patch)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(champion_id, synergy_champion_id, game_mode, patch) DO UPDATE SET
			champion_name = excluded.champion_name,
			synergy_name = excluded.synergy_name,
			score_rank = excluded.score_rank,
			score = excluded.score,
			play = excluded.play,
			win = excluded.win,
			win_rate = excluded.win_rate,
			tier = excluded.tier,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range synergies {
		if _, err := stmt.Exec(s.ChampionID, s.ChampionName, s.SynergyChampionID, s.SynergyName,
			s.ScoreRank, s.Score, s.Play, s.Win, s.WinRate, s.Tier, gameMode, patch); err != nil {
			return fmt.Errorf("save synergy %d-%d: %w", s.ChampionID, s.SynergyChampionID, err)
		}
	}

	return tx.Commit()
}

// GetSynergiesForChampion fetches top synergies for a champion
func (d *DB) GetSynergiesForChampion(championID int, gameMode, patch string) ([]ChampionSynergy, error) {
	query := `
		SELECT champion_id, champion_name, synergy_champion_id, synergy_name,
			score_rank, score, play, win, win_rate, tier, game_mode, patch
		FROM champion_synergies
		WHERE champion_id = ? AND game_mode = ? AND patch = ?
		ORDER BY score_rank ASC
		LIMIT 5`

	rows, err := d.conn.Query(query, championID, gameMode, patch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ChampionSynergy
	for rows.Next() {
		var s ChampionSynergy
		if err := rows.Scan(&s.ChampionID, &s.ChampionName, &s.SynergyChampionID, &s.SynergyName,
			&s.ScoreRank, &s.Score, &s.Play, &s.Win, &s.WinRate, &s.Tier, &s.GameMode, &s.Patch); err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, rows.Err()
}
