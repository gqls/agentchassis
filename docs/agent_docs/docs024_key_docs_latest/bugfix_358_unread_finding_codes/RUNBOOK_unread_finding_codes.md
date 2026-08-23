# RUNBOOK — bugs_open/358 (unread finding codes)

Commands that were hard to get right, with the gotcha attached. Update HERE when one changes.

## The census, re-runnable

```sql
-- totals + resolved usage + retention liveness (oldest row sits on the 30d boundary)
SELECT count(*) AS total, count(*) FILTER (WHERE resolved) AS resolved,
       min(occurred_at) AS oldest FROM agent_error_log;

-- per-code counts, newest first by volume
SELECT error_code, count(*) AS n, min(occurred_at)::date AS first,
       max(occurred_at)::date AS last
FROM agent_error_log GROUP BY 1 ORDER BY 2 DESC LIMIT 45;

-- who is using the resolved workflow (first user ever: content-loss-check, 2026-08-22)
SELECT error_code, count(*), array_agg(DISTINCT resolved_by)
FROM agent_error_log WHERE resolved GROUP BY 1;
```

GOTCHA — **a zero over this table proves 30 days of nothing, never "never"** (358 §8):
retention (mig 466 pre_query) truncates history. All-history claims need git or tests.

GOTCHA — **grep the CONSTANT, not just the literal** (358 §3.2): the one real per-code
reader (`page_build_failure_guard.go:131`) binds a Go const to `$1`; a literal-only grep
verdicts the code unread.

GOTCHA — **`error_code` is free text**: uppercase and lowercase families coexist, and
`create_tool_cross_link_items.go` emits colon-suffixed variants
(`tool_crosslink_not_emitted:*`). Any registry or GROUP BY must state its normalisation
or a family double-counts as compliance.

## Reader census greps

```bash
# all writer literals (struct-field form)
grep -rn 'ErrorCode:' platform/ --include='*.go' | grep -v _test.go
# all direct readers of the table (then check each WHERE for a code filter)
grep -rn 'FROM agent_error_log\|from agent_error_log' platform/ cmd/ scripts/ docs/agent_docs/sql_for_agents/
# for each code carried by a const: grep the const NAME too
grep -rn '<ConstName>' platform/ cmd/ --include='*.go'
```

## Ownership / queue checks (before routing work)

```bash
python3 scripts/who-owns.py 358
```
```sql
SELECT item_type, item_key, status FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected')
  AND (summary ILIKE '%agent_error_log%' OR item_key ILIKE '%error_log%');
```

---

## The shipped check (added 2026-08-22)

```bash
./scripts/audit-finding-codes.sh            # exit 0 = every observed code declared
./scripts/audit-finding-codes.sh --json     # machine-readable report
```

Direct, if you want to control the input:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -At -c "SELECT DISTINCT error_code FROM agent_error_log
          WHERE error_code IS NOT NULL AND error_code <> ''" > /tmp/live_codes.txt
go run ./cmd/config-key-audit --finding-codes < /tmp/live_codes.txt
```

GOTCHA — **`go run` collapses the child's exit status.** The vacuity refusal is exit **2** from
the compiled binary and **1** under `go run`. Branch on **empty stdout**, never on exit code 2 —
that branch would be dead code. (Same discipline `audit-optional-key-budget.sh` records.)

GOTCHA — **`--report` needs `PG_CLIENTS_HOST`**, which the CronJob supplies from pod env. By hand,
omit `--report` and pipe the code list on stdin instead. The report path exists because the
service account has no `pods/exec` RBAC, so a scheduled job cannot shell to `kubectl`
(`cmd/config-key-audit/fleetdb.go:8-14`).

## The controls — how to prove the check can still fail

⚠ **Mutate a COPY, never the shipped registry.** Several sessions have this tree open;
`WRONG_CALLS.md` 2026-08-22 records a session mutating a shared file in place to prove a guard and
another session committing it mid-window.

```bash
SP=/tmp                                       # or your scratchpad
# (A) the real list must be clean
go run ./cmd/config-key-audit --finding-codes < /tmp/live_codes.txt >/dev/null; echo "want 0: $?"

# (B) one undeclared code must be caught — THE RATCHET
{ cat /tmp/live_codes.txt; echo "TEST_UNREGISTERED_X"; } > $SP/mutated.txt
go run ./cmd/config-key-audit --finding-codes < $SP/mutated.txt >/dev/null 2>&1; echo "want 1: $?"

# (C) a colon variant must NOT read as undeclared (normalisation)
{ cat /tmp/live_codes.txt; echo "tool_crosslink_not_emitted:a_brand_new_reason"; } \
  | go run ./cmd/config-key-audit --finding-codes >/dev/null 2>&1; echo "want 0: $?"

# (D) a 'consumed' reader pointed at the WRONG file must be rejected
python3 - <<'PY'
import json
p="docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json"
d=json.load(open(p))
d["CONTENT_VALIDATION_BLOCKER_DETAIL"]["reader"]="platform/orchestration/actions/page_build_failure_guard.go:131"
json.dump(d,open("/tmp/reg_wrong_reader.json","w"),indent=2)
PY
go run ./cmd/config-key-audit --finding-codes --registry /tmp/reg_wrong_reader.json \
  < /tmp/live_codes.txt >/dev/null 2>&1; echo "want 1: $?"

# (E) THE CONTROL for (D) — the true reader must pass, or (D) proved only that any change fails
go run ./cmd/config-key-audit --finding-codes < /tmp/live_codes.txt >/dev/null 2>&1; echo "want 0: $?"

# (F) vacuity guard — empty input must REFUSE. Read STDOUT, not the exit code (see gotcha above)
out=$(printf '' | go run ./cmd/config-key-audit --finding-codes 2>/dev/null); echo "want 0 bytes: ${#out}"
```

## Retention — the clock, and how to see it move

```sql
-- the whole window in one query: what dies next
SELECT error_code, count(*), min(occurred_at)::date AS first, max(occurred_at)::date AS last,
       (min(occurred_at)::date + 30) AS oldest_row_dies
FROM agent_error_log GROUP BY 1 ORDER BY min(occurred_at) LIMIT 10;
```

The retention job is the `database-cleanup` row in `scheduled_tasks` (enabled, hourly), whose
`pre_query` embeds the DELETE — `466_orchestration_status_vocabulary.sql:184-189`. **Marking a row
`resolved` HALVES its remaining life** (14 days, not 30), so any triage flow must extract what it
needs *before* resolving. `cmd/content-loss-check` gets this right and is the exemplar.

GOTCHA — **this is not theoretical and the numbers move within a session.** On 2026-08-22 the
observed-code count went 43 → 42 in three hours: `REVIEW_SUPERSEDED_BY_PASSING_SAVE` (25 rows, all
2026-07-23) was deleted between two runs of the check. Never quote a count from these docs without
re-running.

## Committing on this tree (learned the hard way, 2026-08-22)

```bash
# AFTER committing any file another session is also in — build the COMMITTED tree, not yours
scripts/verify-head-builds.sh

# package won't build locally because of someone's transient edit? test YOUR file against HEAD:
cp <your-file> /tmp/h/<same-path> && (cd /tmp/h && go test ./<pkg>/ -run '<YourTests>')
```

GOTCHA — a pathspec commit takes the **working tree**, so a same-file passenger can be
**half-written**: a mid-typed comment, or a call into a file that is still untracked. Both happened
here in one commit and broke HEAD twice; a green working-tree build is evidence of nothing.

GOTCHA — **use a quoted heredoc for every commit message carrying prose**:
`-m "$(cat <<'EOF' … EOF)"`. Backticks inside `-m "…"` execute; this lane lost two words from a
commit message that way, and forward-only means they cannot be amended back.
