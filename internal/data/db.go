package data

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB 数据库连接
type DB struct {
	conn *sql.DB
}

// New 创建数据库连接
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	return d.conn.Close()
}

// Conn 获取原始连接
func (d *DB) Conn() *sql.DB {
	return d.conn
}

// runMigrations 执行数据库迁移
func (d *DB) runMigrations() error {
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file.Name(), err)
		}

		if _, err := d.conn.Exec(string(content)); err != nil {
			// 忽略重复列错误，允许迁移安全地重复执行
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return fmt.Errorf("execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}
