# HANDOFF — bug-sweep thread, 2026-07-27 → 2026-07-28. Start here.

**What this thread is.** A triage-and-closure sweep of `/bugs_open/`, picked up
from a session that hit a usage limit mid-flight (`a6973bb7`). It has since
closed five bugs, filed two, applied two config fixes, shipped one Go fix through
the council gate, and drained 57 items from the human-review queue.

**Its remit is the whole backlog**, so it is not "owned" in the workstream sense
and does not hold a lock on anything. Check `scripts/who-owns.py <n>` before
picking a bug up, and re-run `git log` at implementation start — that script
reads commits and is blind to a session mid-fix.

---

## 1. Fleet state, checked 2026-07-28 10:01 UTC

| fact | value |
|---|---|
| chassis | **v1.0.1182**, pod started `09:55:02Z`, binary built `09:37` |
| open bugs | **48** (was 36 at 21:00 yesterday — other threads are filing fast) |
| `needs_human_review` queue | **327** |
| `orchestration_states` retention floor | 2026-07-13 (≈15 days, 1,913 rows) |

**Clock trap:** this box is BST (UTC+1). `git log` prints BST, `kubectl` prints
UTC. Compare in UTC (`TZ=UTC git log --date=iso-local`) or live fixes look inert.

---

## 2. Closed this sweep — with the evidence, so nothing is taken on trust

| bug | evidence |
|---|---|
| **010** (oldest open, filed 07-17) | discriminating pod-grep on v1.0.1174, 4 positives = 1, negative control = 0. File is explicit that defect 3's *behaviour* was never induced |
| **070** | live row read: `stale-work-item-reaper.pre_query` now keys on `updated_at`, not row age. Migration 237. DB-config, live on apply |
| **034**, **086** | closed by the prior session's agents; commits `fb6b4f69d`, `1f6c70f49` |
| **112** | **both acceptance tests passed on the real path.** Key reaches a spawned pod (`len=53`, with pre-roll pods at 0 as a live control); and `page-content-writer` made 2 successful `gemini-pro-latest` calls. Council round 2 **APPROVED**, trailer earned |

## 3. Filed this sweep

- **`bugs_open/119`** — one council seat's malformed JSON voids a whole round.
  Measured: **2 of the last 18 rounds** were decided by an unreadable seat rather
  than a judgement, on two *different* seats. Not truncation (`bugs_closed/019`
  cannot reach it — every call finished inside its ceiling). Fix candidate 1 is
  "retry a seat once on parse failure", which keeps 019's downgrade contract intact.
- **`bugs_open/125`** — `page-rebuild` deploys to a path from the page **name**,
  ignoring `pages.url`, publishing an orphaned duplicate. **65% of pages (280/431)**
  would deploy wrong. **LATENT** — see §5, this is the important one.

---

## 4. OPEN LOOPS — what is actually owed

### 4a. `bugs_open/125` — pre-work is DONE, the fix is not written

Candidate 1 is ~3 lines in `determinePageFilename`
(`platform/orchestration/actions/git_deployer_actions.go:368-400`): try
`p["url"]` before `slug`/`name`. **Its stated pre-checks are already run:**

- **`pages.url` is safe to use verbatim** — 431/431 leading slash, 430/431
  `.html`, 0 empty, 0 suspicious. **One exception:** `/tools.html#audience-check`
  carries a **fragment**, which used as a path makes a file literally named
  `tools.html#audience-check`. **The fix must strip `#…`.**
- **Three callers** share the buggy path, so the change hits all of them:
  `pageflow-builder`, `page-rebuild`, `site-work-orchestrator` (all
  `…_loop/deploy_page`). None has run in the retention window.

**Go change ⇒ needs a build + roll, and a council submission** (`platform/`).

### 4b. `bugs_open/087` — fix applied and verified; the case stays OPEN

Migration 246 applied and **survived four rolls** (1177→1179→1180→1182), so it is
not re-seed-fragile. Verified live: the writer child received
`section_plan.sections_ready = 4` and completed the loop that used to kill it.

It stays open because its own acceptance bar ("the page must deploy with its
components rewritten") could not be met on the chosen target:
`save_page_sections` refused with `rebuild_policy=owned`. **That guard is
correct.** To finish: re-test on a page that is **not** `rebuild_policy=owned`,
**after 125 is fixed** — otherwise the test publishes another orphan.

### 4c. `bugs_open/033` — the drain is LIVE and is a loaded gun

First live run closed **57** items (382 → 325); `resolution_path='auto:revalidated'`
went 0 → 57, all `status=complete` with `completed_at` set. Reversible by
construction — verified `idx_swi_dedup` excludes `complete`, so a wrong close
releases the dedup key and the check re-raises.

**⚠️ `dry_run` is now `false` with `max_items: 500` and there is no cron.** Nothing
fires it on a clock, but `TRIGGER_revalidate_review_queue_v1.sh` now **closes
items** for whoever runs it next. Restore simulation by setting the **scalar at
the leaf path** (a literal-object `jsonb_set` at `{…,sweep,config}` drops the
`max_items` sibling and silently reverts the cap to 50):

```sql
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,sweep,config,dry_run}', 'true'::jsonb, false)
WHERE type='diagnosis-review-queue-revalidator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Untouched and larger: the owner's *"the queue should not fill"* reframing. The
drain treats the symptom. 223 of the remaining items are **uncovered types**, and
72 are covered-but-unjudgeable (`no content_data` 32, `no deployed component
matches` 25, `spec.missing names no fields` 15) — that is where a v2 belongs.

### 4d. Small, unstarted

- **`bugs_open/101`** — do **not** delete the four "inert" keys by key name. `max_pages`
  is **live** for `select_representative_content` and `validate_site_plan`
  (`build-site-planner` at 80, `site-planner` at 20). Delete by **(action, key)**.
- **`[UNVERIFIED]` in 125** — 16 `git_commit` steps appear to default to
  `index.html`, and `page-rerender` alone has 49 runs. Almost certainly fine (a
  separate multi-file `filesMap` path). **Read that path before anyone acts on it.**
- **Memory index nag** — the tooling asks to compact `MEMORY.md` below 17KB.
  **Decline it.** There is an owner ruling dated 2026-07-28 inside the file: ~20KB
  is accepted, and the only lever that reaches 17KB is moving the auto-loaded
  durable practices out, which is explicitly forbidden.

---

## 5. The finding worth carrying forward

**Fixing 087 armed 125.** `page-rebuild` was dormant *because it was broken*. The
first run that got past 087's crash reached `deploy_page` for the first time ever
and published a live orphaned page. The resolver had been wrong since it was
written; 087 was the only reason nobody had met it.

So: **"this path never runs, so its bugs are low priority" is often backwards** —
it may never run because something upstream is broken, and your fix is what
starts the traffic. A step's execution count is evidence about its *predecessors*,
not its own quality. Written up in `016b` §9.

Practical consequence: **087 and 125 should ship together**, or 087 gets no
traffic until 125 lands.

---

## 6. Landmines this thread paid for

- **A check answers the question you ENCODED, not the one you MEANT** — and a
  non-empty result *feels* verified. Four instances in 40 minutes, after writing
  the weaker version of this rule the night before. Never type a value the DB
  stores (I `curl`ed a guessed URL and told the owner a live 35KB page was a 404,
  with `pages.url` already printed on my screen). Never use `ORDER BY created_at
  DESC LIMIT 1` as a proxy for `parent_orchestration_id IS NULL` — in a spawning
  workflow that is a *grandchild*, and it reported a live run as finished.
- **A zero is not a finding until the instrument is shown able to return
  non-zero.** `strings | grep -c "^SYMBOL$"` returns 0 for everything, because Go
  constants sit inside a shared string table and `^…$` can never match. The
  positive control is what catches it.
- **`make deploy-agents` / `push-backend` are FLEET-WIDE** and assume every
  service was built at `IMAGE_TAG`. After building one service they point the
  other thirteen at an image that does not exist. Deploy one service by applying
  its overlay alone.
- **A payload lives at its step's `output_field`**, not under the terminal step.
  A trigger script's own printed query pointed at the wrong path and returned
  empty — which reads exactly like "the sweep did nothing".
- **A council `REVISE` may be the harness, not you.** Read `decided_by` and
  `unreadable` (**not** `abstained`) before rewriting anything.

All logged in `WRONG_CALLS.md` (2026-07-27, 2026-07-28) and `016b` §9.

---

## 7. Suggested order for the next session

1. **`bugs_open/125` candidate 1** — pre-work done (§4a), needs the fragment
   strip, a council submission, a build and a roll. Highest value: it unblocks
   087 and removes an estate-wide landmine.
2. **`bugs_open/087` re-test** on a non-owned page, once 125 is live. That closes
   087 and confirms 125 in one run — assert the **path**
   (`page_deployed…file_path` must equal `pages.url`), not the success flag.
3. **`bugs_open/119`** — retry-a-seat-once. Small, and it stops ~11% of council
   rounds being wasted.
4. **Pick fresh from `/bugs_open/`** — 48 open. Prefer ones with no commit in
   24h; almost everything else has an active session on it.

**Full per-thread map of the estate:** `docs024_key_docs_latest/OPEN_THREADS_RESTART_LIST.md`
(written by this sweep, 07-27; half-life about a day on this tree).
