package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/w-h-a/pebble/internal/client/repo"
	_ "modernc.org/sqlite"
)

type sqliteRepo struct {
	option repo.Options
	db     *sql.DB
}

func (r *sqliteRepo) Close() error {
	return r.db.Close()
}

func (r *sqliteRepo) configure() error {
	var journalMode string
	if err := r.db.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		return fmt.Errorf("failed to set journal mode: %w", err)
	}

	if _, err := r.db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	slog.Debug("pragmas configured", "journal_mode", journalMode, "foreign_keys", true)

	var count int
	err := r.db.QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_version'",
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check schema_version: %w", err)
	}

	var current int
	if count > 0 {
		err = r.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&current)
		if err != nil {
			return fmt.Errorf("read schema version: %w", err)
		}
	}

	if current >= schemaVersion {
		slog.Debug("schema up to date", "version", current)
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(ddl); err != nil {
		return fmt.Errorf("exec ddl: %w", err)
	}

	if count == 0 {
		_, err = tx.Exec("INSERT INTO schema_version (version) VALUES (?)", schemaVersion)
	} else {
		_, err = tx.Exec("UPDATE schema_version SET version = ?", schemaVersion)
	}
	if err != nil {
		return fmt.Errorf("update schema version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Debug("schema migrated", "from", current, "to", schemaVersion)

	return nil
}

func NewRepo(opts ...repo.Option) (repo.Repo, error) {
	options := repo.NewOptions(opts...)

	db, err := sql.Open("sqlite", options.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	r := &sqliteRepo{
		option: options,
		db:     db,
	}

	if err := r.configure(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure database: %w", err)
	}

	return r, nil
}
