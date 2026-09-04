// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"testing"
	"time"

	"github.com/asciimoo/hister/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrationDropsRequiredWebSessionLastSeenColumn(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	cfg.App.Directory = t.TempDir()
	cfg.Server.Database = "legacy.sqlite3"
	_, dsn := cfg.DatabaseConnection()

	legacyDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec("CREATE TABLE databases (id integer PRIMARY KEY AUTOINCREMENT, version integer)").Error; err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec("INSERT INTO databases (id, version) VALUES (1, 1)").Error; err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec(`CREATE TABLE web_sessions (
		id integer PRIMARY KEY AUTOINCREMENT,
		token_hash text NOT NULL,
		data blob NOT NULL,
		created_at datetime,
		updated_at datetime,
		last_seen_at datetime NOT NULL,
		expires_at datetime NOT NULL
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec("CREATE INDEX idx_web_sessions_last_seen_at ON web_sessions(last_seen_at)").Error; err != nil {
		t.Fatal(err)
	}
	legacySQLDB, err := legacyDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := legacySQLDB.Close(); err != nil {
		t.Fatal(err)
	}

	if err := Init(cfg); err != nil {
		t.Fatal(err)
	}
	currentSQLDB, err := DB.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = currentSQLDB.Close() })

	if DB.Migrator().HasColumn("web_sessions", "last_seen_at") {
		t.Fatal("web_sessions.last_seen_at still exists after migration")
	}
	if err := CreateWebSession(&WebSession{
		TokenHash: "2b7847b6ff9b65017f19fbdc4019f3a287b0a16669b4370d03ab8d4a899a0f47",
		Data:      []byte("session data"),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("create web session after migration: %v", err)
	}
	dbVersion, initialized := migrationVersion()
	if !initialized || dbVersion != int64(len(migrations)) {
		t.Fatalf("database version = %d, initialized = %v, want %d and true", dbVersion, initialized, len(migrations))
	}
}
