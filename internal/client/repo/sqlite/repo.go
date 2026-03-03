package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/w-h-a/pebble/internal/client/repo"
	"github.com/w-h-a/pebble/internal/domain"
	_ "modernc.org/sqlite"
)

type sqliteRepo struct {
	option repo.Options
	db     *sql.DB
}

func (r *sqliteRepo) CreateIssue(ctx context.Context, issue *domain.Issue) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO issues (id, title, description, status, type, priority, assignee, estimate_mins, defer_until, due_at, created_at, updated_at, parent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		issue.ID,
		issue.Title,
		issue.Description,
		string(issue.Status),
		string(issue.Type),
		issue.Priority,
		issue.Assignee,
		issue.EstimateMins,
		issue.DeferUntil,
		issue.DueAt,
		issue.CreatedAt,
		issue.UpdatedAt,
		issue.ParentID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert issue: %w", err)
	}

	return nil
}

func (r *sqliteRepo) IssueExists(ctx context.Context, id string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM issues WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check issue exists: %w", err)
	}

	return exists, nil
}

func (r *sqliteRepo) AddLabels(ctx context.Context, issueID string, labels []string) error {
	for _, label := range labels {
		_, err := r.db.ExecContext(
			ctx,
			"INSERT OR IGNORE INTO labels (issue_id, label) VALUES (?, ?)",
			issueID,
			label,
		)
		if err != nil {
			return fmt.Errorf("failed to insert label %q: %w", label, err)
		}
	}

	return nil
}

func (r *sqliteRepo) GetIssue(ctx context.Context, id string) (*domain.Issue, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, title, description, status, type, priority, assignee, estimate_mins, defer_until, due_at, created_at, updated_at, closed_at, parent_id
		FROM issues WHERE id = ?`,
		id,
	)

	var issue domain.Issue

	err := row.Scan(
		&issue.ID,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.Type,
		&issue.Priority,
		&issue.Assignee,
		&issue.EstimateMins,
		&issue.DeferUntil,
		&issue.DueAt,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.ClosedAt,
		&issue.ParentID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("issue not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan issue: %w", err)
	}

	return &issue, nil
}

func (r *sqliteRepo) GetLabels(ctx context.Context, issueID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT label FROM labels WHERE issue_id = ? ORDER BY label", issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels: %w", err)
	}
	defer rows.Close()

	var labels []string

	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	return labels, rows.Err()
}

func (r *sqliteRepo) GetDependencies(ctx context.Context, issueID string) ([]domain.Dependency, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT issue_id, depends_on_id, created_at FROM dependencies WHERE issue_id = ? ORDER BY created_at",
		issueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	var deps []domain.Dependency

	for rows.Next() {
		var d domain.Dependency
		if err := rows.Scan(&d.IssueID, &d.DependsOnID, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		deps = append(deps, d)
	}

	return deps, rows.Err()
}

func (r *sqliteRepo) GetComments(ctx context.Context, issueID string) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, issue_id, author, body, created_at FROM comments WHERE issue_id = ? ORDER BY created_at",
		issueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []domain.Comment

	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.Author, &c.Body, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	return comments, rows.Err()
}

func (r *sqliteRepo) ResolveID(ctx context.Context, partial string) (string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id FROM issues WHERE id LIKE ?", partial+"%")
	if err != nil {
		return "", fmt.Errorf("failed to resolve id: %w", err)
	}
	defer rows.Close()

	var matches []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", fmt.Errorf("failed to scan id: %w", err)
		}
		matches = append(matches, id)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("failed to resolve id rows: %w", err)
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no issue found matching %q", partial)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous id %q matches %d issues", partial, len(matches))
	}
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
