package data

import (
	"database/sql"
	"fmt"
)

// Item represents a game item
type Item struct {
	ItemID int    `json:"item_id"`
	NameEN string `json:"name_en"`
	NameCN string `json:"name_cn"`
	Tags   string `json:"tags"`
	Stats  string `json:"stats"`
}

// SaveItems batch saves item data
func (d *DB) SaveItems(items []Item) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO items (item_id, name_en, name_cn, tags, stats)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(item_id) DO UPDATE SET
			name_en = excluded.name_en,
			name_cn = excluded.name_cn,
			tags = excluded.tags,
			stats = excluded.stats
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, it := range items {
		if _, err := stmt.Exec(it.ItemID, it.NameEN, it.NameCN, it.Tags, it.Stats); err != nil {
			return fmt.Errorf("save item %d: %w", it.ItemID, err)
		}
	}

	return tx.Commit()
}

// GetItemByID fetches an item by ID
func (d *DB) GetItemByID(itemID int) (*Item, error) {
	var it Item
	err := d.conn.QueryRow(`
		SELECT item_id, name_en, name_cn, tags, stats
		FROM items WHERE item_id = ?
	`, itemID).Scan(&it.ItemID, &it.NameEN, &it.NameCN, &it.Tags, &it.Stats)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &it, nil
}
