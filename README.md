# Pebble (pb)

## Problem

[Beads](https://github.com/steveyegge/beads) is a powerful issue tracker that grew to serve multi-agent orchestration. For a solo developer that likes to keep their hands on the wheel while pairing with an agentic navigator, ~80% of it is dead weight.

## Solution

Pebble keeps the features of beads that I use everyday and drops everything else.

## Architecture

### Flow Chart

```mermaid
graph TD
  subgraph CLI ["pb CLI (cobra)"]
    CMD[Command Layer]
  end

  subgraph Service ["Service Layer"]
    SVC[Service]
  end

  subgraph Client ["Client Layer"]
    REPO_IF[Repo Interface]
  end

  subgraph Infra ["Infrastructure"]
    SQLITE[SQLite via modernc.org/sqlite]
    DB[(pebble.db)]
  end

  subgraph Domain ["Domain Layer"]
    ISSUE[Issue]
    DEP[Dependency]
    COMMENT[Comment]
  end

  CMD --> SVC
  SVC --> Domain
  SVC --> REPO_IF
  REPO_IF -.-> SQLITE
  SQLITE --> DB
```

### ER Diagram

```mermaid
erDiagram
  issues {
    text id PK
    text title
    text description
    text status "open|in_progress|approved|rejected|closed"
    text type "task|bug|feature|chore|decision|epic"
    int priority "0-4"
    int estimate_mins
    text parent_id FK "self-ref"
  }

  dependencies {
    text issue_id PK "FK → issues"
    text depends_on_id PK "FK → issues"
  }

  labels {
    text issue_id PK "FK → issues"
    text label PK
  }

  comments {
    int id PK
    text issue_id FK
    text body
  }

  issues ||--o{ dependencies : "blocked by"
  issues ||--o{ labels : has
  issues ||--o{ comments : has
```

## Usage

```text
pb init [--stealth] [--prefix]      pb ready [--sort --limit]
pb create "title" [flags]           pb upcoming [--days --assignee]
pb show <id>                        pb search <query>
pb update <id> [flags]              pb dep add <id> --blocks <id>
pb close <id>                       pb dep remove <id> <id>
pb reopen <id>                      pb comment <id> "text"
pb delete <id>                      pb stale [--days]
pb list [--status --type ...]       pb config set|get|list
pb export [-o file.jsonl]           pb version
pb import <file.jsonl>
```