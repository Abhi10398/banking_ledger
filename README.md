# Banking Ledger

A production-style banking ledger service built in Go. Supports account management, fund transfers, reversals, deposits/withdrawals, and a full audit trail — all backed by PostgreSQL with strict consistency guarantees.

---

## What I built

A RESTful HTTP service that acts as the core ledger for a single-currency banking system. It handles the mechanics that matter most in financial software: atomic double-entry bookkeeping, deadlock-safe concurrent transfers, idempotent reversals, and an append-only audit log written inside the same database transaction as every mutation.

A lightweight SPA (`static/index.html`) ships alongside the API and provides a live dashboard for accounts, transfers, and the audit log — no build step required.

---

## Architecture

```
┌─────────────────────────────────────────────┐
│  Client (browser / Postman / curl)          │
└──────────────────┬──────────────────────────┘
                   │ HTTP
┌──────────────────▼──────────────────────────┐
│  Go + Fiber v2                              │
│  ├── Correlation-ID middleware              │
│  ├── Structured logging (Zap / Logrus)      │
│  ├── Panic recovery                         │
│  ├── /api/accounts  — AccountAPI            │
│  └── /api/transfers — TransferAPI           │
└──────────────────┬──────────────────────────┘
                   │ GORM (pgx driver)
┌──────────────────▼──────────────────────────┐
│  PostgreSQL 16                              │
│  ├── accounts   (BIGINT balance, CHECK ≥ 0) │
│  ├── transfers  (ENUM status, UNIQUE idx)   │
│  └── audit_log  (append-only, indexed)      │
└─────────────────────────────────────────────┘
```

**Why this stack**

| Choice | Reason |
|---|---|
| Go + Fiber v2 | Low-latency, low-allocation HTTP layer; straightforward concurrency model |
| PostgreSQL | ACID transactions, row-level locking (`SELECT … FOR UPDATE`), native UUID, ENUMs, and `CHECK` constraints enforce invariants at the DB layer |
| GORM (pgx driver) | Ergonomic ORM without sacrificing raw-SQL escape hatches; pgx gives native PostgreSQL protocol support |
| Google Wire | Compile-time DI — dependency graph is explicit and verified at build time, no runtime reflection |
| Liquibase | Version-controlled schema migrations with rollback support; same tool works locally and in CI/CD |

---

## How to run

**Prerequisites:** [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) installed. Nothing else needed.

**Step 1 — Start the stack**

```bash
make run
```

This starts PostgreSQL, runs all database migrations, then starts the app. Wait until you see the app logs settle and see Fiber v2.52.9 http://127.0.0.1:8080 in the logs

**Step 2 — Open the dashboard**

```
http://localhost:8080
```

**Step 3 — Seed sample data**

Creates five accounts with starting balances. Copy the printed UUIDs — you'll need two of them for the concurrency test.

```bash
make seed
```

Output:

```
Alice     id=aaaaaaaa-…  balance=500000
Bob       id=bbbbbbbb-…  balance=300000
Carol     id=cccccccc-…  balance=750000
Dave      id=dddddddd-…  balance=200000
Eve       id=eeeeeeee-…  balance=1000000
```

**Step 4 — Run the concurrency test** (optional)

Fires 50 concurrent transfers between two accounts and verifies no money is created or destroyed.

Copy two account IDs printed in Step 3 and paste them in:

```bash
make test-concurrency A=aaaaaaaa-0000-0000-0000-000000000001 B=bbbbbbbb-0000-0000-0000-000000000002
```

Expected output:

```
Initial state:
  A (aaaaaaaa…)  balance=500000
  B (bbbbbbbb…)  balance=300000
  Total=800000

Firing 50 concurrent transfers of 100 each…

Results: 47 succeeded, 3 failed/rejected

Final state:
  A (aaaaaaaa…)  balance=495300
  B (bbbbbbbb…)  balance=304700
  Total=800000

Invariant checks:
  A balance non-negative : PASS
  B balance non-negative : PASS
  Total conserved        : PASS

✅ All invariants hold.
```

Some transfers will be rejected when the sender's balance would go negative — this is correct behaviour. The invariant is that the total is always conserved.

---

### Tear down

```bash
make docker-down        # stop containers, keep database volume
make docker-clean       # stop containers AND delete all data
```

---

## Development & testing via Postman

Use this flow when iterating on the API locally — no Docker required.

**Step 1 — Install and start PostgreSQL**

```bash
brew install postgresql@18
brew services start postgresql@18
```

**Step 2 — Create the database** (first time only)

```bash
createdb banking_ledger
```

**Step 3 — Run migrations**

```bash
make pg-migrate
```

**Step 4 — Start the server**

```bash
go run main.go api static
```

The API is available at `http://localhost:3000/api`.

**Step 5 — Import the Postman collection**

Open Postman → Import → select `postman/banking_ledger.postman_collection.json` from this repo.

The collection covers every endpoint. Requests that create resources automatically set the `accountId` and `transferId` collection variables so subsequent requests are pre-filled.

**Step 6 — Stop everything when done**

```bash
# Ctrl+C to stop the server, then:
brew services stop postgresql@18
```

---

## Design decisions & trade-offs

### Balance as `BIGINT` (paise) — no float drift

All monetary amounts are stored and transmitted as integers in the smallest currency unit (paise for INR, cents for USD). `BIGINT` gives 9,223,372,036,854,775,807 paise — roughly ₹92 trillion — which is enough for any realistic account balance.

**Why not `DECIMAL`/`NUMERIC`?** Floating-point types (`FLOAT`, `DOUBLE`) cannot represent 0.1 exactly in binary. Even `NUMERIC` introduces rounding surface area when the application layer does arithmetic in Go `float64`. Integers are exact and arithmetic on them is deterministic across every layer of the stack.

The API accepts and returns amounts in the same integer unit. The frontend formats them for display only — the canonical value never passes through a float.

### `FOR UPDATE` with ordered locking — deadlock prevention

Every transfer that touches two accounts acquires row-level locks with `SELECT … FOR UPDATE`. If two concurrent transfers involve the same pair of accounts in opposite directions, each will try to lock the rows in a different order — the classic deadlock.

The fix: **always lock accounts in ascending UUID order**, regardless of which is the sender.

```sql
-- Both goroutines lock the same row first because UUIDs are ordered
SELECT * FROM accounts WHERE id IN ($1, $2) ORDER BY id FOR UPDATE
```

This eliminates the deadlock entirely without needing serializable isolation or retry loops. The cost is a slightly wider lock window (both rows locked simultaneously rather than sequentially), which is negligible at the scale this service targets.

### READ COMMITTED vs SERIALIZABLE

The service uses PostgreSQL's default `READ COMMITTED` isolation level, not `SERIALIZABLE`.

`SERIALIZABLE` prevents every anomaly but serializes all conflicting transactions — under high concurrency (e.g., the concurrency test fires 50 threads) this causes frequent `serialization_failure` errors that the application must catch and retry. The retry logic adds latency and complexity.

`READ COMMITTED` with explicit `FOR UPDATE` locking gives the same safety for the transfer use-case:
- Non-repeatable reads are irrelevant because we re-read inside the transaction after acquiring the lock.
- Phantom reads cannot occur because we lock by primary key, not range.
- The `CHECK (balance >= 0)` constraint is evaluated at commit time and will reject any transaction that would produce a negative balance, acting as a final backstop.

The net result: stronger throughput, simpler error handling, same correctness.

### Idempotent reversal — two-layer enforcement

A transfer reversal must never execute twice. Two independent layers prevent double-reversal:

1. **Application layer**: Before executing the reversal, the service fetches the original transfer and checks `status = 'completed'`. If it is `'reversed'`, it returns a `409 Conflict` immediately.

2. **Database layer**: The `transfers` table has a `UNIQUE` constraint on `reversal_of`:
   ```sql
   reversal_of UUID REFERENCES transfers(id) UNIQUE
   ```
   Even if two concurrent requests both pass the application-layer check (a narrow race), only one can insert a row with `reversal_of = <original_id>`. The second gets a unique-constraint violation, which the service maps to `409 Conflict`.

Neither layer alone is sufficient — the app check is vulnerable to a race, the DB constraint only helps at insert time. Together they are airtight.

### Audit log in the same transaction — consistency guarantee

Every audit entry is written inside the same database transaction as the balance change it describes. If the transaction rolls back (e.g. insufficient funds, constraint violation), the audit entry is also rolled back — there are no phantom audit records for operations that never completed.

This is stricter than common patterns that write audit events to a queue or a separate table after the fact. The trade-off is that a slow audit insert blocks the transfer commit, but given that audit rows are small and the table is append-only with a partial index on `created_at`, this is negligible in practice.

Failed operations (e.g. transfers that are rejected for insufficient funds) are still recorded — the service catches the error, writes an audit entry with `outcome = 'FAILURE'` in a fresh standalone transaction, and then returns the error to the caller.

---

## Assumptions

| Assumption | Rationale |
|---|---|
| **Minimum balance floor is zero** | No overdraft facility. `CHECK (balance >= 0)` is a DB-level hard constraint. If a transfer would take the sender below zero it is rejected with `402 Insufficient Funds`. |
| **Single currency per account** | Currency is stored on the account and validated at transfer time — both parties must hold the same currency. No FX conversion is performed. |
| **Reversal reverses the full amount** | Partial reversals are not supported. A reversal creates a new transfer for the exact original amount in the opposite direction. |
| **Transfers are immediate** | There is no scheduled, pending, or batched transfer state. A transfer either succeeds (status `completed`) or fails (status `failed`) synchronously within the HTTP request. |

---

## API Reference

All endpoints are prefixed with `/api`. Amounts are always in paise (integer). The server returns `application/json`.

Request bodies that fail struct validation return `400 Bad Request` with:
```json
{ "error": "validation_error", "message": "<details>" }
```

Every response includes an `X-Correlation-ID` header. Pass one in your request to trace it end-to-end in logs.

---

### Accounts

| Method | Path | Description | Body | Success |
|---|---|---|---|---|
| `GET` | `/api/accounts` | List all accounts | — | `200` array |
| `POST` | `/api/accounts` | Create an account | `{"name":"Alice","currency":"INR"}` | `201` account |
| `GET` | `/api/accounts/:id` | Get account by ID | — | `200` account |
| `POST` | `/api/accounts/:id/deposit` | Deposit funds | `{"amount":50000}` | `200` account |
| `POST` | `/api/accounts/:id/withdraw` | Withdraw funds | `{"amount":50000}` | `200` account |
| `GET` | `/api/accounts/:id/audit` | Audit log for account | `?limit=20&offset=0` | `200` array |

**Account object**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Alice",
  "balance": 500000,
  "currency": "INR",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:05:00Z"
}
```

---

### Transfers

| Method | Path | Description | Body | Success |
|---|---|---|---|---|
| `POST` | `/api/transfers` | Execute a transfer | `{"from_account_id":"…","to_account_id":"…","amount":10000}` | `200` result |
| `GET` | `/api/transfers` | List transfers | `?limit=20&offset=0` | `200` array |
| `POST` | `/api/transfers/:id/reverse` | Reverse a transfer | — | `200` result |

**Transfer result**
```json
{
  "transfer_id": "…",
  "new_from_balance": 490000,
  "new_to_balance": 310000
}
```

**Reversal result**
```json
{
  "reversal_id": "…",
  "original_transfer_id": "…"
}
```

**Transfer object** (from list)
```json
{
  "id": "…",
  "from_account_id": "…",
  "to_account_id": "…",
  "amount": 10000,
  "status": "completed",
  "reversed_by": null,
  "reversal_of": null,
  "created_at": "…",
  "updated_at": "…"
}
```

---

### Audit Log

| Method | Path | Description | Success |
|---|---|---|---|
| `GET` | `/api/audit` | List all audit entries | `200` array |
| `GET` | `/api/accounts/:id/audit` | Audit entries for one account | `200` array |

**Audit entry object**
```json
{
  "id": "…",
  "operation": "TRANSFER",
  "from_account_id": "…",
  "to_account_id": "…",
  "amount": 10000,
  "outcome": "SUCCESS",
  "failure_reason": null,
  "transfer_id": "…",
  "created_at": "…",
  "updated_at": "…"
}
```

`operation` ∈ `TRANSFER | REVERSAL | DEPOSIT | WITHDRAWAL`  
`outcome` ∈ `SUCCESS | FAILURE`

---

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/ping` | Liveness probe — returns `200 {"message":"pong"}` |

---

## Error responses

| HTTP status | `error` code | Meaning |
|---|---|---|
| `400` | `validation_error` | Request body failed struct validation |
| `400` | `bad_request` | Malformed input (e.g. invalid UUID) |
| `402` | `insufficient_funds` | Sender balance would go negative |
| `404` | `account_not_found` | No account with that ID |
| `404` | `transfer_not_found` | No transfer with that ID |
| `409` | `already_reversed` | Transfer has already been reversed |
| `422` | `invalid_reversal` | Transfer cannot be reversed (status ≠ completed) |
| `500` | `internal_error` | Unexpected server error |

---

## Makefile reference

```bash
make run                  # docker-compose up --build (full stack)
make dev                  # go run main.go api static (local, no Docker)
make migrate              # run Liquibase migrations locally
make seed                 # POST 5 sample accounts with balances
make test-concurrency A=<id> B=<id>  # fire 50 concurrent transfers

make docker-up            # docker compose up --build -d (detached)
make docker-down          # docker compose down
make docker-logs          # docker compose logs -f app
make docker-migrate       # docker compose run --rm migrate
make docker-clean         # docker compose down -v --remove-orphans

make build                # go mod tidy && go build -o myapp .
make pg-migrate           # liquibase update (postgres)
make pg-migrate-rollback  # rollback last changeset
make pg-migrate-history   # show applied changesets
```
