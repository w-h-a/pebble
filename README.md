# Pebble (pb)

## Problem

[Beads](https://github.com/steveyegge/beads) is a powerful issue tracker that grew to serve multi-agent orchestration. For a solo developer that likes to keep their hands on the wheel while pairing with an agentic navigator, ~80% of it is dead weight.

## Solution

Pebble keeps the features of beads that I use everyday and drops everything else.

## Architecture

```mermaid
graph TD
  subgraph CLI ["pb CLI (cobra)"]
      CMD[Command Layer]
  end

  subgraph Service ["Service Layer"]
    SVC[Service]
  end

  subgraph Domain ["Domain Layer"]
      ISSUE[Issue]
      DEP[Dependency]
      COMMENT[Comment]
  end

  subgraph Client ["Client Layer"]
    REPO_IF[Repo Interface]
  end

  subgraph Infra ["Infrastructure"]
    IDGEN[ID Generator]
    SQLITE[SQLite via modernc.org/sqlite]
    DB[(pebble.db)]
  end

  subgraph IO ["I/O"]
      JSONL[JSONL Import/Export]
      JSON[--json formatter]
  end

  CMD --> SVC
  CMD --> JSONL
  CMD --> JSON
  SVC --> Domain
  SVC --> IDGEN
  SVC --> REPO_IF
  REPO_IF -.-> SQLITE
  SQLITE --> DB
  JSONL --> REPO_IF
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