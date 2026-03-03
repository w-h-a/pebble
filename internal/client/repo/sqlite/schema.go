package sqlite

const schemaVersion = 1

const ddl = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS issues (
	id            TEXT PRIMARY KEY,
	title         TEXT NOT NULL,
	description   TEXT DEFAULT '',
	status        TEXT NOT NULL DEFAULT 'open'
		CHECK(status IN ('open','in_progress','approved','rejected','closed')),
	type          TEXT NOT NULL DEFAULT 'task'
		CHECK(type IN ('task','bug','feature','chore','decision','epic')),
	priority      INTEGER NOT NULL DEFAULT 2 CHECK(priority BETWEEN 0 AND 4),
	assignee      TEXT DEFAULT '',
	estimate_mins INTEGER DEFAULT 0,
	defer_until   DATETIME,
	due_at        DATETIME,
	created_at    DATETIME NOT NULL,
	updated_at    DATETIME NOT NULL,
	closed_at     DATETIME,
	parent_id     TEXT REFERENCES issues(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS dependencies (
	issue_id      TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
	depends_on_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
	created_at    DATETIME NOT NULL,
	PRIMARY KEY (issue_id, depends_on_id)
);

CREATE TABLE IF NOT EXISTS labels (
	issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
	label    TEXT NOT NULL,
	PRIMARY KEY (issue_id, label)
);

CREATE TABLE IF NOT EXISTS comments (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	issue_id   TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
	author     TEXT DEFAULT '',
	body       TEXT NOT NULL,
	created_at DATETIME NOT NULL
);

CREATE VIEW IF NOT EXISTS ready_issues AS
SELECT i.*
FROM issues i
WHERE i.status = 'open'
AND NOT EXISTS (
	SELECT 1 FROM dependencies d
	JOIN issues blocker ON blocker.id = d.depends_on_id
	WHERE d.issue_id = i.id
	AND blocker.status IN ('open', 'in_progress')
)
AND (i.defer_until IS NULL OR i.defer_until <= datetime('now'));
`
