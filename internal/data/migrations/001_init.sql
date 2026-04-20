-- 英雄基础信息
CREATE TABLE IF NOT EXISTS champions (
    champion_id INTEGER PRIMARY KEY,
    name_en     TEXT NOT NULL,
    name_cn     TEXT NOT NULL,
    title       TEXT,
    tags        TEXT
);

-- 英雄在各模式下的统计数据
CREATE TABLE IF NOT EXISTS champion_stats (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id  INTEGER NOT NULL,
    game_mode    TEXT NOT NULL,
    winrate      REAL,
    pickrate     REAL,
    banrate      REAL,
    tier         TEXT,
    sample_size  INTEGER,
    patch        TEXT NOT NULL,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, game_mode, patch)
);

-- 海克斯（Augment）基础信息
CREATE TABLE IF NOT EXISTS augments (
    augment_id  TEXT PRIMARY KEY,
    name_en     TEXT NOT NULL,
    name_cn     TEXT NOT NULL,
    description TEXT,
    tier        TEXT,
    icon_url    TEXT
);

-- 英雄 + 海克斯 组合胜率
CREATE TABLE IF NOT EXISTS hero_augment_stats (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id INTEGER NOT NULL,
    augment_id  TEXT NOT NULL,
    game_mode   TEXT NOT NULL,
    winrate     REAL,
    pickrate    REAL,
    tier        TEXT,
    patch       TEXT NOT NULL,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, augment_id, game_mode, patch),
    FOREIGN KEY (champion_id) REFERENCES champions(champion_id),
    FOREIGN KEY (augment_id) REFERENCES augments(augment_id)
);

-- 出装推荐
CREATE TABLE IF NOT EXISTS build_recommendations (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    champion_id  INTEGER NOT NULL,
    game_mode    TEXT NOT NULL,
    role         TEXT,
    items        TEXT NOT NULL,
    boots        TEXT,
    skill_order  TEXT,
    patch        TEXT NOT NULL,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(champion_id, game_mode, role, patch)
);

-- 装备基础信息
CREATE TABLE IF NOT EXISTS items (
    item_id  INTEGER PRIMARY KEY,
    name_en  TEXT NOT NULL,
    name_cn  TEXT NOT NULL,
    tags     TEXT,
    stats    TEXT
);

-- 版本更新追踪
CREATE TABLE IF NOT EXISTS patch_tracker (
    patch       TEXT PRIMARY KEY,
    released_at DATETIME,
    scraped_at  DATETIME,
    status      TEXT
);
