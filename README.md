<div align="center">

# 🏏 Auto-Pick Draft Fantasy

**A turn-based fantasy cricket draft engine: users join a league, take turns picking players, and a background worker auto-picks when someone runs out of time.**

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![migrate](https://img.shields.io/badge/golang--migrate-v4.18-7B42BC?style=for-the-badge)
![Postman](https://img.shields.io/badge/Postman-Collection-FF6C37?style=for-the-badge&logo=postman&logoColor=white)

![status](https://img.shields.io/badge/status-learning_project-yellow?style=flat-square)
![deps](https://img.shields.io/badge/HTTP-net%2Fhttp_stdlib-blue?style=flat-square)
![driver](https://img.shields.io/badge/driver-pgx%2Fv5-blueviolet?style=flat-square)
![players](https://img.shields.io/badge/seeded_players-115-success?style=flat-square)

[Quick start](#-quick-start) •
[Simulation](#-run-the-simulation) •
[How a draft works](#-how-a-draft-works) •
[API](#-api-reference) •
[Design notes](#-design-notes)

</div>

---

> [!NOTE]
> This is a **learning / practice project**. The goal is to explore how an industry-style draft system is built (transactions, row locks, background workers) while keeping the code small and readable. There is no auth on purpose.

## ✨ Features

| | Feature | What it does |
|---|---|---|
| 👤 | **Users** | Full CRUD |
| 🏆 | **Leagues** | `season_long` or `draft_pick`, with a fixed `size` and a per-pick `timeout` (1–300 s) |
| 🎟️ | **Entries** | Users join a league and get the next `pick_order` automatically. A user can join a league only once, and only until it's full |
| 🚦 | **Start league** | Only when the league is full. Puts the first entry on the clock |
| 🤖 | **Auto-pick worker** | Every 2 s it finds turns whose time has run out and gives that entry the first available player |
| 📋 | **Rosters** | See every team in a league, or one entry's team |
| 🌱 | **Seed data** | 115 real international cricketers (Bat / Bowl / AR / Keeper) |
| 🧪 | **Simulation** | `make simulate` runs a full colored demo in your terminal |

---

## 🧱 Architecture

```mermaid
flowchart LR
    C([🧑‍💻 Client<br/>Postman / curl]):::client

    subgraph APP["⚙️ Go app (cmd/main.go)"]
        direction LR
        H[🌐 Handler<br/><i>HTTP only</i>]:::handler --> S[🧠 Service<br/><i>validation, rules</i>]:::service --> R[🗄️ Repository<br/><i>SQL only</i>]:::repo
        W[🤖 Auto-pick worker<br/><i>ticks every 2s</i>]:::worker
    end

    DB[(🐘 PostgreSQL 17)]:::db

    C -- JSON --> H
    R -- pgx pool --> DB
    W -- pgx pool --> DB

    classDef client  fill:#fde68a,stroke:#b45309,color:#1f2937
    classDef handler fill:#bfdbfe,stroke:#1d4ed8,color:#1e3a8a
    classDef service fill:#c7d2fe,stroke:#4338ca,color:#1e1b4b
    classDef repo    fill:#ddd6fe,stroke:#6d28d9,color:#2e1065
    classDef worker  fill:#fecaca,stroke:#b91c1c,color:#450a0a
    classDef db      fill:#bbf7d0,stroke:#15803d,color:#052e16
```

Each layer only talks to the one to its right:

- **Handler**: reads the request, calls the service and writes JSON. It doesn't know SQL.
- **Service**: validation and business rules. It doesn't know HTTP.
- **Repository**: SQL and transactions. It returns simple errors (`ErrNotFound`, `ErrConflict`, …).
- **DTOs** (`internal/dto`) keep the API shape separate from the **DB models** (`internal/models`).

---

## 🗂️ Data model

```mermaid
erDiagram
    USERS ||--o{ ENTRIES : "joins as"
    LEAGUES ||--o{ ENTRIES : "has"
    LEAGUES ||--o| LEAGUE_PICKS : "current turn"
    ENTRIES ||--o{ LEAGUE_PICKS : "on the clock"
    ENTRIES ||--o{ ROSTER_OWNERS : "owns"
    PLAYERS ||--o{ ROSTER_OWNERS : "picked as"
    LEAGUES ||--o{ ROSTER_OWNERS : "scoped to"

    USERS {
        bigint id PK
        varchar first_name
        varchar last_name
        int age
    }
    LEAGUES {
        bigint id PK
        varchar name
        text description
        varchar type "season_long | draft_pick"
        varchar series_name
        varchar status "created | started | completed"
        int size "max members, >= 2"
        int timeout "seconds per pick, 1..300"
    }
    ENTRIES {
        bigint id PK
        bigint league_id FK
        bigint user_id FK
        int pick_order "1, 2, 3 ..."
        timestamptz joined_at
    }
    PLAYERS {
        bigint id PK
        varchar name
        varchar position "Bat | Bowl | AR | Keeper"
    }
    ROSTER_OWNERS {
        bigint id PK
        bigint league_id FK
        bigint entry_id FK
        bigint player_id FK
    }
    LEAGUE_PICKS {
        bigint id PK
        bigint league_id FK "UNIQUE"
        bigint entry_id FK
        timestamptz expires_at "deadline -> auto-pick"
    }
```

**Rules the database enforces:**

| Constraint | Guarantees |
|---|---|
| `entries UNIQUE (league_id, user_id)` | a user joins a league only once |
| `entries UNIQUE (league_id, pick_order)` | no two entries share a turn number |
| `roster_owners UNIQUE (league_id, player_id)` | a player belongs to **one** entry per league |
| `league_picks UNIQUE (league_id)` | one "current turn" per league |
| `CHECK` on `type`, `status`, `position`, `size`, `timeout` | only valid values are stored |

---

## 🔄 League lifecycle

```mermaid
stateDiagram-v2
    direction LR
    [*] --> created : POST /leagues
    created --> created : POST /leagues/{id}/entries<br/>(until full)
    created --> started : POST /leagues/{id}/start<br/>(entries == size)
    started --> completed : planned
    completed --> [*]

    classDef open fill:#bbf7d0,stroke:#15803d,color:#052e16
    classDef live fill:#fde68a,stroke:#b45309,color:#451a03
    classDef todo fill:#e5e7eb,stroke:#6b7280,color:#374151,stroke-dasharray: 5 5
    class created open
    class started live
    class completed todo
```

| Status | Join? | Start? |
|---|:---:|:---:|
| 🟢 `created` | ✅ until full | ✅ when full |
| 🟡 `started` | ❌ `409` | ❌ `409` |
| ⚪ `completed` | ❌ | ❌ |

---

## 🎬 How a draft works

A full run with a league of size 3 and a 5-second timeout:

```mermaid
sequenceDiagram
    autonumber
    actor U as 👤 Users
    participant API as 🌐 API
    participant DB as 🐘 Postgres
    participant W as 🤖 Worker

    rect rgba(59,130,246,0.12)
    Note over U,DB: 1 · Setup
    U->>API: POST /leagues {size: 3, timeout: 5}
    API->>DB: INSERT leagues (status = created)
    API-->>U: 201 league #1
    end

    rect rgba(34,197,94,0.12)
    Note over U,DB: 2 · Join (the league row is locked for each join)
    loop 3 users
        U->>API: POST /leagues/1/entries {user_id}
        API->>DB: BEGIN · SELECT … FOR UPDATE<br/>status = created and not full yet?<br/>INSERT pick_order = MAX + 1 · COMMIT
        API-->>U: 201 pick_order 1, 2, 3
    end
    U->>API: POST /leagues/1/entries (4th user)
    API-->>U: 409 league is full
    end

    rect rgba(234,179,8,0.15)
    Note over U,DB: 3 · Start (one transaction)
    U->>API: POST /leagues/1/start
    API->>DB: lock league · entries == size?<br/>first entry = ORDER BY pick_order LIMIT 1<br/>status = started<br/>INSERT league_picks (expires_at = now + 5s)
    API-->>U: 200 current_pick: entry #1
    end

    rect rgba(239,68,68,0.12)
    Note over U,W: 4 · Clock runs out ⏳
    loop every 2s
        W->>DB: SELECT league_picks WHERE expires_at ≤ NOW()
    end
    W->>DB: load players not in roster_owners for league 1
    W->>DB: BEGIN · INSERT roster_owners (first player)<br/>DELETE league_picks · COMMIT
    Note right of W: auto-pick: entry #1 got player #1
    end

    U->>API: GET /entries/1/roster
    API-->>U: 200 players: [Ruturaj Gaikwad (Bat)]
```

---

## 🚀 Quick start

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) with Compose
- `jq`, only for `make simulate` (preinstalled on recent macOS; otherwise `brew install jq` or `apt install jq`)

### 1 · Configure

```bash
cp .env.example .env
```

```ini
# Postgres container
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password_dev
POSTGRES_DB=fantasy-db

# Go app (runs on the host)
DATABASE_URL=postgres://postgres:password_dev@localhost:5433/fantasy-db?sslmode=disable
PORT=8080
```

> [!TIP]
> Postgres is exposed on host port **5433**, not 5432, so it won't clash with a local Postgres install.

### 2 · Start the database and apply migrations

```bash
docker compose up -d      # starts Postgres; the `migrate` service applies migrations, then exits
make db-seed              # inserts 115 cricket players (safe to run again)
```

### 3 · Run the API and the worker

```bash
make run
```

```text
auto-pick worker started (every 2s)
listening on :8080
```

Press <kbd>Ctrl</kbd> + <kbd>C</kbd> to stop. The worker stops and in-flight requests get 5 s to finish.

### 🛠️ Make targets

| Command | What it does |
|---|---|
| `make run` | Start the API and the auto-pick worker |
| `make db-migrate` | Apply pending migrations (`docker compose run --rm migrate`) |
| `make db-seed` | Insert the seeded players |
| `make simulate` | Run the end-to-end draft simulation 👇 |

> [!WARNING]
> To reset everything: `docker compose down -v && docker compose up -d && make db-seed`. The `-v` deletes the database volume.

---

## 🧪 Run the simulation

With `make run` going in one terminal, run this in another:

```bash
make simulate                     # defaults: SIZE=3 TIMEOUT=5
make simulate SIZE=4 TIMEOUT=10   # bigger league, longer clock
make simulate KEEP=1              # keep the demo data afterwards
```

The script ([`scripts/simulate.sh`](scripts/simulate.sh)) creates users, fills a league, starts it, shows a countdown bar, waits for the worker to auto-pick and prints the rosters. It deletes its own data when it finishes. Here's a real run (the colors approximate the terminal output):

```diff
   ____             __ _     ____  _      _
  |  _ \ _ __ __ _ / _| |_  |  _ \(_) ___| | __
  | | | | '__/ _` | |_| __| | |_) | |/ __| |/ /
  | |_| | | | (_| |  _| |_  |  __/| | (__|   <
  |____/|_|  \__,_|_|  \__| |_|   |_|\___|_|\_\   auto-pick simulation

@@ ▶ Creating 4 users (one extra to show a full league) @@
+ ✔ user #8  Aarav Mehta
+ ✔ user #9  Diya Kapoor
+ ✔ user #10  Kabir Singh
+ ✔ user #11  Meera Iyer

@@ ▶ Creating a draft league @@
+ ✔ league #4  Simulation Cup · draft_pick · size 3 · 3s per pick · status created

@@ ▶ Users join the league (pick_order is assigned automatically) @@
+ ✔ entry #7  user #8  → pick_order 1
+ ✔ entry #8  user #9  → pick_order 2
+ ✔ entry #9  user #10  → pick_order 3
- ✘ user #11 tried to join → 409 conflict: league is full  (expected)

@@ ▶ Starting the draft @@
+ ✔ status → started
+ ✔ entry #7 is on the clock until 2026-11-12T18:32:20+05:30
- ✘ start again → 409 conflict: league can only be started from 'created' status  (expected)

@@ ▶ Entry #7 does nothing… ⏳ @@
  [██████████████████████████████]  time's up!

@@ ▶ Waiting for the auto-pick worker @@
+ ✔ 🤖 auto-picked: Ruturaj Gaikwad (Bat) for entry #7

@@ ▶ League rosters @@
  ENTRY    USER     ORDER  PLAYERS
  #7       #8       1      Ruturaj Gaikwad (Bat)
  #8       #9       2      —
  #9       #10      3      —

+ ✔ Simulation complete
```

At the same time, the server terminal logs:

```text
auto-pick: league 4, entry 7 got player 1 (Ruturaj Gaikwad, Bat)
```

---

## 📡 API reference

Base URL is `http://localhost:8080`. All bodies are JSON.

> [!TIP]
> Import [`postman/fantasy-draft.postman_collection.json`](postman/fantasy-draft.postman_collection.json) into Postman. Create requests save `userId`, `leagueId` and `entryId` automatically, so the other requests just work.

### 👤 Users

| Method | Path | Body | Success | Errors |
|---|---|---|---|---|
| ![POST](https://img.shields.io/badge/POST-49cc90?style=flat-square) | `/users` | `{first_name, last_name, age}` | `201` | `400` |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/users?limit=20&offset=0` | | `200` | |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/users/{id}` | | `200` | `400` `404` |
| ![PUT](https://img.shields.io/badge/PUT-fca130?style=flat-square) | `/users/{id}` | `{first_name, last_name, age}` | `200` | `400` `404` |
| ![DELETE](https://img.shields.io/badge/DELETE-f93e3e?style=flat-square) | `/users/{id}` | | `204` | `404` |

### 🏆 Leagues

| Method | Path | Body | Success | Errors |
|---|---|---|---|---|
| ![POST](https://img.shields.io/badge/POST-49cc90?style=flat-square) | `/leagues` | `{name, description?, type, series_name, size, timeout}` | `201` | `400` |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/leagues?limit=20&offset=0` | | `200` | |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/leagues/{id}` | | `200` | `400` `404` |
| ![PUT](https://img.shields.io/badge/PUT-fca130?style=flat-square) | `/leagues/{id}` | same as create (status is not editable) | `200` | `400` `404` `409`¹ |
| ![DELETE](https://img.shields.io/badge/DELETE-f93e3e?style=flat-square) | `/leagues/{id}` | | `204` | `404` |
| ![POST](https://img.shields.io/badge/POST-49cc90?style=flat-square) | `/leagues/{id}/start` | | `200` | `404` `409`² |

¹ `size` is smaller than the number of entries that already joined. ² The league is not full, or is already started.

### 🎟️ Entries

| Method | Path | Body | Success | Errors |
|---|---|---|---|---|
| ![POST](https://img.shields.io/badge/POST-49cc90?style=flat-square) | `/leagues/{id}/entries` | `{user_id}` | `201` | `400` `404` `409`³ |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/leagues/{id}/entries` | | `200` (in draft order) | `404` |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/users/{id}/entries` | | `200` | `404` |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/entries/{id}` | | `200` | `404` |
| ![PUT](https://img.shields.io/badge/PUT-fca130?style=flat-square) | `/entries/{id}` | `{pick_order}` | `200` | `400` `404` `409` |
| ![DELETE](https://img.shields.io/badge/DELETE-f93e3e?style=flat-square) | `/entries/{id}` | | `204` (leave league) | `404` |

³ Already joined, the league is full, or the league has started.

### 📋 Rosters

| Method | Path | Success | Errors |
|---|---|---|---|
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/leagues/{id}/rosters` | `200`: every entry with its players | `404` |
| ![GET](https://img.shields.io/badge/GET-61affe?style=flat-square) | `/entries/{id}/roster` | `200`: one entry with its players | `404` |

<details>
<summary><b>📦 Example requests and responses</b></summary>

**Create a league**

```bash
curl -X POST localhost:8080/leagues -d '{
  "name": "Friends Draft", "description": "Weekend league",
  "type": "draft_pick", "series_name": "IPL 2026",
  "size": 4, "timeout": 60
}'
```

```json
{
  "id": 1, "name": "Friends Draft", "description": "Weekend league",
  "type": "draft_pick", "series_name": "IPL 2026",
  "status": "created", "size": 4, "timeout": 60,
  "created_at": "2026-11-12T17:12:05+05:30"
}
```

**Start a league**

```bash
curl -X POST localhost:8080/leagues/1/start
```

```json
{
  "league": { "id": 1, "status": "started", "size": 4, "timeout": 60, "...": "..." },
  "current_pick": { "entry_id": 1, "expires_at": "2026-11-12T17:13:05+05:30" }
}
```

**An entry's roster**

```bash
curl localhost:8080/entries/1/roster
```

```json
{
  "entry_id": 1, "user_id": 1, "pick_order": 1,
  "players": [
    { "id": 1, "name": "Ruturaj Gaikwad", "position": "Bat", "picked_at": "2026-11-12T17:13:06+05:30" }
  ]
}
```

</details>

### ❗ Errors

Every error has the same shape:

```json
{ "error": "conflict: league is full" }
```

| Status | When |
|---|---|
| `400` | invalid JSON, invalid id, or a field fails validation |
| `404` | the user, league or entry doesn't exist (the message says which) |
| `409` | a business rule was violated (duplicate join, full league, already started, …) |
| `500` | unexpected error; logged on the server, the message to the client is generic |

---

## 🧠 Design notes

These are the parts worth reading in the code.

<details open>
<summary><b>🔒 Joining a league without race conditions</b>: <code>entry_repository.go → Create</code></summary>

If two users join at the same moment, both could read `MAX(pick_order) = 2` and both try to insert `3`. To prevent that, the join runs in a transaction:

```sql
BEGIN;
SELECT status, size FROM leagues WHERE id = $1 FOR UPDATE;  -- 🔒 others wait here
--  status must be 'created'; COUNT(entries) must be < size
INSERT INTO entries (league_id, user_id, pick_order)
SELECT $1, $2, COALESCE(MAX(pick_order), 0) + 1 FROM entries WHERE league_id = $1;
COMMIT;
```

The lock only queues up joins for the **same** league. In testing, 6 joins fired at a size-3 league at the same moment gave exactly 3 × `201` and 3 × `409`.

</details>

<details>
<summary><b>🚦 Starting a league atomically</b>: <code>league_repository.go → Start</code></summary>

All in one transaction, so a league is never half-started:

1. `SELECT … FOR UPDATE`: lock the league (this also blocks joins that happen at the same time)
2. status must be `created`
3. `COUNT(entries)` must equal `size`
4. first entry = `ORDER BY pick_order LIMIT 1`
5. `UPDATE status = 'started'`
6. `INSERT league_picks (expires_at = NOW() + timeout)`

</details>

<details>
<summary><b>🤖 The auto-pick worker</b>: <code>internal/workers/auto_draft_pick_worker.go</code></summary>

```mermaid
flowchart TD
    T([⏰ tick every 2s]):::tick --> Q["SELECT league_picks<br/>WHERE expires_at ≤ NOW()"]:::step
    Q --> E{any?}:::decide
    E -- no --> T
    E -- yes --> G[group by league]:::step
    G --> P["load players NOT IN roster_owners<br/>for that league"]:::step
    P --> A["BEGIN<br/>INSERT roster_owners (first player)<br/>DELETE league_picks<br/>COMMIT"]:::tx
    A --> L[📝 log: auto-pick ...]:::step
    L --> T

    classDef tick   fill:#fde68a,stroke:#b45309,color:#451a03
    classDef step   fill:#bfdbfe,stroke:#1d4ed8,color:#1e3a8a
    classDef decide fill:#e9d5ff,stroke:#7e22ce,color:#3b0764
    classDef tx     fill:#bbf7d0,stroke:#15803d,color:#052e16
```

- **Strategy:** give the first available player (lowest id).
- **Assign and remove in one transaction:** a crash can't leave the expired pick behind, which would make the worker assign a second player.
- **Errors in one league are logged** and don't stop the other leagues.
- **Stops cleanly** when the context is cancelled (<kbd>Ctrl</kbd> + <kbd>C</kbd>).

</details>

---

## 📁 Project structure

```text
.
├── cmd/
│   └── main.go                 # wiring: db → repos → services → handlers, worker, graceful shutdown
├── internal/
│   ├── db/                     # pgx connection pool
│   ├── models/                 # DB models (one per table)
│   ├── dto/                    # request / response shapes
│   ├── repository/             # SQL + transactions
│   ├── service/                # validation + business rules
│   ├── handler/                # HTTP routes (net/http, Go 1.22+ patterns)
│   └── workers/                # auto-pick background worker
├── migrations/                 # golang-migrate SQL (up/down per table)
├── seeds/
│   └── 001_players.sql         # 115 cricketers (safe to run again)
├── scripts/
│   └── simulate.sh             # colored end-to-end demo
├── postman/                    # importable Postman collection
├── docker-compose.yaml         # Postgres 17 + migrate
├── Makefile
└── .env.example
```

---

## 🗺️ Roadmap

- [x] Users, leagues, entries CRUD
- [x] League `size`, per-pick `timeout`, `status` lifecycle
- [x] Race-free join and start (row locks + transactions)
- [x] Auto-pick worker for timed-out turns
- [x] Roster views (per league, per entry)
- [ ] Advance the turn to the next entry after a pick (round-robin or snake)
- [ ] Manual pick endpoint for the entry on the clock
- [ ] Move the league to `completed` when every roster is full
- [ ] Block leaving or reordering entries once a league has started
- [ ] Re-check the deadline under a row lock in the worker (`FOR UPDATE`) so it can't collide with a manual pick
- [ ] Tests

> [!IMPORTANT]
> **Current behaviour:** after an auto-pick, the worker deletes the `league_picks` row and doesn't create the next turn yet. Each draft therefore runs **one** auto-pick and then waits. Advancing the turn is the next step on the roadmap.

---

<div align="center">

Built for learning 🏏 · Go · PostgreSQL · golang-migrate

</div>
