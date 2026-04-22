-- Champion synergy recommendations
CREATE TABLE IF NOT EXISTS champion_synergies (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id         INTEGER NOT NULL,
    champion_name       TEXT NOT NULL,
    synergy_champion_id INTEGER NOT NULL,
    synergy_name        TEXT NOT NULL,
    score_rank          INTEGER,
    score               REAL,
    play                INTEGER,
    win                 INTEGER,
    win_rate            REAL,
    tier                INTEGER,
    game_mode           TEXT NOT NULL,
    patch               TEXT NOT NULL,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, synergy_champion_id, game_mode, patch)
);
