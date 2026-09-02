# HANDOFF — after 694: the empty-index rule, a bug in my own detector, and the routing still owed

**Opened 2026-09-02 (evening)** by the experience_loop lane, closing the session that worked
`HANDOFF_2026-09-02_content_quality_auditor_into_the_new_build_path.md`.
Supersedes nothing — read that file's **§ GATE RESULT** block first for why the original task
turned out to be three jobs, then this file for what is done and what is left.

**Full path of this file:**
`docs/agent_docs/docs024_key_docs_latest/experience_loop/HANDOFF_2026-09-02b_after_694_the_empty_index_rule_and_the_routing_still_owed.md`

---

## 0. DO THIS FIRST — two things, and neither is optional

### 0a. The kubeconfig token is EXPIRED. You cannot verify anything until the owner refreshes it.

At 21:09Z every cluster call returned `You must be logged in to the server (Unauthorized)`,
including `kubectl get --raw /version`. That is the fleet-wide expiry signature (the token lapses
about every 3 days; **the owner refreshes it**, you cannot). Context is
`personae-uk001-prod-agent-chassis-cluster`.

**Do not read any "0 rows" or empty output as a finding until this is cleared** — an expired token
and a genuine absence look identical at the shell.

### 0b. A fresh chassis was rolled at ~21:0xZ and migration 694's survival is UNVERIFIED

694 is **DB config**, so a chassis roll should not touch it — but "should not" is an assumption and
this estate's own rule is to prove it at the artefact. I could not, because of 0a. **Run this
first**, and expect the row shown:

```sql
SELECT
  (default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query' LIKE '%p.name IN (%') AS allowlist_back,      -- expect f
  (default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query' LIKE '%<style[^>]*?>%')  AS nongreedy,        -- expect t
  (default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query' LIKE '%ORDER BY pc.position%') AS ordered,    -- expect t
  (default_config->'workflow'->'steps'->'run_content_llm_audit'->'config'->>'prompt' LIKE '%9. GUIDE PROMINENCE:%') AS dims,  -- expect t
  (default_config->'workflow'->'steps'->'write_findings'->'config'->>'filing_mode') AS filing_mode,                           -- expect record
  updated_at
FROM agent_definitions WHERE type='content-quality-auditor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Then the behavioural check, which is the one that actually matters — **the config reading right is
not the same as the seat sampling right**:

```sql
SELECT o.created_at,
       jsonb_array_length(o.collected_data->'page_samples'->'rows') AS pages,
       (SELECT count(DISTINCT r->>'page_type') FROM jsonb_array_elements(o.collected_data->'page_samples'->'rows') r) AS page_types,
       (o.collected_data->'page_samples'->'rows'->0 ? 'page_type') AS has_new_columns
FROM orchestration_states o
WHERE o.collected_data ? 'page_samples' AND o.created_at > '2026-09-02 21:00:00+00'
ORDER BY o.created_at DESC LIMIT 5;
```
**Pre-694 the answer was 3 pages and no `page_type` key. Post-694 and pre-roll it was 14–18 pages
across 4–7 types with `has_new_columns = t`.** If post-roll runs read 3 again, 694 has been
reverted by something and that is your whole afternoon — start at `agent_definitions_backup`
(`snapshot_taken_at DESC`, **not** `created_at` — that is a landmine) and at
`agent_definitions_bak_694`, which holds the PRE-694 config and is deliberately left in place.

⚠ **No orchestration dispatch within ~300s of a chassis pod (re)start** — the spawn is silently
dropped. Check `startTime` before firing anything.

---

## 1. What is DONE, and how it was proven

### Migration 694 — `content-quality-auditor` can see the site

| | |
|---|---|
| files | `docs/agent_docs/sql_for_agents/694_content_quality_auditor_can_see_the_site.sql` (+ `_ROLLBACK`) |
| council | **APPROVED at round 2**, `SUBMISSION_CORR=d52a0e45-5c64-4d32-a1ab-f73532684d37` |
| commits | `5ff171327` (round 1), `0fa679a28` (round 1 REVISE acted on) |
| applied | **2026-09-02 14:36:08Z**, by hand, then `run-migrations.sh --record-only` so no other session's `--apply` re-runs it |
| owner ruling carried | `filing_mode` stays **`record`** — findings are verdict rows for human approval, nothing auto-regenerates |

**What it changed.** `load_page_content` hardcoded `p.name IN ('index','about','services','contact')`
— 3 of 22 pages on boxingonline, **92 of 1,196 fleet-wide (7.7%)** across 36 sites. It now samples
every `page_type`, 4 per type, 1,200 chars each, ordered by `pc.position`, with `<style>`/`<script>`
stripped **before** truncation, and `page_type` + `url` exposed. The prompt gained `audience` in its
enum (10 of 210 stored findings were already emitting it out-of-vocabulary), TOP 5 → TOP 8, and four
promise-keeping dimensions: **6 PROMISE · 7 EMPTY INDEX · 8 TOOL DATA · 9 GUIDE PROMINENCE.**

**Proven, not asserted:** every audit between the apply and the roll sampled **14–18 pages across
4–7 page_types** carrying the new column, against 3 and no column before.

> ⚠ **Boxing Online itself has still never been audited post-694.** Its latest run is **06:23Z,
> pre-694, 3 pages**. The improvement sweep takes one site per 15-min tick over ~54 sites and had
> not returned; a manual dispatch published at ~14:37Z produced **no orchestration row at all**
> (unexplained, not retried — CLAUDE.md's "a missing row is almost always latency"). So the change
> is proven **fleet-wide and not on the motivating site**. Do not close that loop by quoting a
> different site's run.

### The council round 1 REVISE found a real bug in my own fix — worth reading before you write SQL here

I wrote `regexp_replace(html, '<style[^>]*>.*?</style>', '', 'gs')`. **PostgreSQL takes the
greediness of the FIRST quantifier**, so `[^>]*` makes the whole expression greedy and the `.*?`
does not save it — on a component with two style blocks it matches first-open to LAST-close and
deletes the prose between. Demonstrated:

```
regexp_replace('<style>.a{}</style>KEEP THIS PROSE<style>.b{}</style>AND THIS',
               '<style[^>]*>.*?</style>','','gs')   ->   'AND THIS'
```

`[MEASURED 2026-09-02]` **7 of 2,871 components** carry 2+ style blocks and the greedy form
destroyed content on **all seven** — avg 2,528 chars, worst 9,076. Fixed by adopting migration
**601**'s proven pattern `<style[^>]*?>.*?</style>` with `'gi'` (the lazy FIRST quantifier is what
makes it non-greedy) and a **space** replacement, not `''` — `''` joins the words either side
(`ALPHA`+`BETA` → `ALPHABETA`), a second defect no seat caught.
**This is PostgreSQL-only. It does NOT transfer to the Go call sites** (`rerender_single_page_action.go:27`
and four others) — Go/RE2 is per-quantifier lazy, proven by running it. Fleet census of live
`agent_definitions` for the greedy pattern: **0**.

Landmine filed: `LANDMINES.md`, *"Sampling a page into an LLM prompt with `LEFT(rendered_html, N)`
feeds it mostly CSS…"*. Two of my own missteps in `WRONG_CALLS.md`.

---

## 2. WHAT IS LEFT — in the order I would do it

### 2a. The empty-index rule (SQ-005 rule D, or a widening of C) — **the funded one, and I promised it to a peer**

**The gap, measured.** `designblog.co.uk` serves four listing pages with **zero** items, each
carrying prose about its own brief instead (`/glossary.html`, `/inspiration/`, `/the-design-feed/`,
`/uk-studios-directory/`). Owner's words via that lane: *"the glossary has text about the brief and
is not a glossary… the directory is empty."* **Both my detectors report ZERO on it, run by hand.**

Rule C's corpus, in `scripts/audit-experience-promises.py` (`INDEX_SQL`, ~line 253):
```sql
AND jsonb_typeof(COALESCE(pc.content_data->'articles', pc.content_data->'items')) = 'array'
), sized AS MATERIALIZED (SELECT * FROM inst WHERE jsonb_array_length(arr) > 0)
```
Two independent filters drop an empty index **before the rule runs**. Measured on designblog: all
**11** component rows under those four pages have `arr_type` **NULL** — no `articles`/`items` key
at all — so they die at the FIRST filter, never reaching the `> 0` one. **Rule C asks "does this
index list things from OUTSIDE its own directory?", which presupposes it lists something. An index
that lists NOTHING is invisible to it.**

Design constraints you must not lose:
- **`/glossary.html` is `page_type='content'`** — outside Rule C's selector entirely. So the fix
  needs a **content-class trigger, not merely a page_type widening**. This is the constraint that
  makes it more than a one-line change.
- **It will fire on pages that CANNOT fill themselves, and that is correct.** Per `bugs_open/444`
  (portfolio positioning): the four remakes' feeds have **0 `content_sources` rows**, glossary and
  inspiration have **no item producer anywhere in the estate**, and the directory kind does not
  exist. 444's fix candidate — plan validation refusing or degrading a listing page whose item
  source resolves to zero — is the upstream door-closer this detector then holds shut. **Say this
  in the docstring** so nobody "fixes" the detector for over-reporting.
- This is the CONTRIB's open **ask 2 listing half**. Second independent instance (boxingonline,
  then designblog) is what funded it.
- **I told the `designblog.co.uk` session I would build it and re-run over their four pages.**
  That promise is outstanding — reply to them when it lands.

### 2b. SQ-004's `--site` control bug — cheap, and it poisons every scoped run until fixed

`scripts/audit-listing-class-promise.py --site <anything-but-leopardess>` prints
`positive_leopardess: FAIL — classifier drifted or site changed`. **It is not a regression and not
a finding.** The positive control is a leopardess page; `--site` filters the control's own case out
of the corpus, so it reports FAIL where it means **N/A**. Proof: `--site leopardessconsulting.co.uk`
→ control **PASS**, and it still finds the real mismatch (7/13 off-class on `/blog.html`).

Why it matters more than it sounds: that line is the only one telling a reader whether a zero is
trustworthy, and it currently cries wolf on every scoped run. Fix = report `N/A (control case not
in scope)` when `--site` excludes the control, or exempt the control row from the site filter.
**Until then, every `--site` clean result is UNTESTED, not clean** — I said exactly that to two
peer lanes, so the caveat is on the record in their notes too.

### 2c. The ORIGINAL task, still owed: route the seat into the new-build path

Unchanged from the previous handoff's §3, and **deliberately held**:

- Edit **`site-work-orchestrator`**. Live tail was
  `fix_items_loop -> apply_site_design -> update_site_status -> complete`;
  **insert between `apply_site_design` and `update_site_status`** (site fully built and designed,
  nothing marked delivered).
- **Use `call_agent`, not `spawn_agent`** — the spawn→call handshake fails ~half the time
  fleet-wide, and `apply_site_design` in that very workflow is already a `call_agent` to copy
  verbatim (`action`, `config.agent_type`, `config.input_mapping`, `timeout_seconds: 300`,
  `output_field`).
- The auditor needs **no storage client** (DB queries + one LLM call), so the inline-chassis
  `params.StorageClient` nil landmine does not bite. Say so — a reviewer will ask.
- Not redundant with `content-reviewer`: that is **per-page**, inside `build_items_loop`; this is a
  **group/site** auditor. Boxingonline is the argument — every page passed individually and the
  SITE was wrong.
- **Hold until you have watched one real post-694 audit on a freshly built site** (see §0b). Wiring
  in a checker before seeing it work is how the original task went wrong.
- Council gate applies (a migration under `sql_for_agents/` is in scope). `DRY_RUN=1` tests
  admission free; `Council-Submitted:` lets you commit before the verdict.

### 2d. CONTRIB ask 3, planner half — still open

Refusing to SELECT a tool at planning time when we hold no data to put in it. The mechanical half
shipped as SQ-005 rule B; this is the planner-side door-closer.

---

## 3. Live commitments to other sessions — these are promises, not notes

| to | what I owe | state |
|---|---|---|
| `designblog.co.uk` | build the empty-index rule, then **re-run over their four pages and report** | outstanding (§2a) |
| `vetcomparison` | **re-run both detectors after migration 701 applies**, and fire the widened `content-quality-auditor` (record mode) at vetcomparison post-701 | they will ping when 701's remainder batch lands |

Both lanes have my two caveats on their record (the leopardess-control N/A-as-FAIL, and the
empty-index blindness), so their quoted zeros carry the right weight. Do not let those go stale.

**Findings already given to vetcomparison, for continuity:** promise ledger clean; one finding —
`tool-compliance-deadline-calculator` is `status='active'`, `build_status='planned'`, **0
components**, serves 404, and is **`in_header=false` with 0 references on the served homepage**, but
carries `nav_label='Deadline Calculator'`. Half-wired, not a live dead link: it becomes one the
moment anyone turns that nav on. Their lane says it is standing owner item `e30dc7b9` (07-17),
recommendation BUILD with relative periods only. They also correct the vigilant-designer's claim
that the parked contrast items concern an unrendered colour — **the three grey ones do render**,
inside the tool markup. Relevant if the auditor's dimensions ever quote the parked-item list.

**Design directive (owner, relayed via designblog):** sites should not all share the same top/bottom
nav + big-hero composition; *"We are trying to make these sites best in class."* **NOT this lane's
work** — routed to and ACKed by `components`, `theme kits`, `site design planner`,
`vigilant-designer` and `editorial design uplift`. This lane is **context-holder only**; do not
silently adopt it.

---

## 4. Things that will mislead you — read before touching anything here

- **`orchestration_states` is REAPED.** The previous handoff recorded 44 auditor runs; the same
  query returned 42 hours later. Every "N runs since" figure about this seat has a shelf life of
  days. Date it or do not quote it.
- **Count auditor runs by `collected_data ? 'content_audit'`**, never by text-matching the agent
  name in `workflow_plan` — that matches the PARENT's spawn record.
- **`content-quality-auditor`'s `default_config.workflow.steps` is an OBJECT keyed by step name,
  not an array.** `jsonb_path_query_array(default_config,'$.workflow.steps[*] ? (...)')` returns
  `[]` and reads as "no such step". Use `jsonb_object_keys` or `$.**`.
- **The seat's lifetime 65% LLM failure rate is NOT a code defect and is closed.** 5,013 of 5,114
  failures are `"You have reached your specified API usage limits"` plus 99 credit-balance 400s,
  spanning 2026-04-08 → 2026-08-31. Since 2026-09-01: **74 calls, 0 failures.** This answers the
  CONTRIB's parting "34% failure rate is its own question".
- **The prompt's category enum is TEXT NO GO CODE READS.** `grep -rn 'tone|gap|cta' --include=*.go`
  returns nothing. The router keys on the category string; `audience` matches no rule and lands in
  the `capability_gap` fallback identically whether the prompt declares it or not. And the older
  belief that the fallback mints `item_type='audit_finding_'+category` is **stale** — that was
  removed 2026-08-15 (`write_audit_findings_action.go:724`).
- **The record-mode reader already exists and needs no building** — I over-promised to the owner
  that I would build one, then found it: admin dashboard `recordVerdictsOnly` filter
  (`frontends/admin-dashboard/src/App.tsx:474` → `GET /admin/work-items?filing_mode=record`), the
  per-finding Release button (`App.tsx:737` → `POST /admin/work-items/:item_id/release`), route at
  `internal/core-manager/api/server.go:455`, handler `site_admin_handlers.go:1049`.
- **`experience-promise-check` has NEVER fired on schedule.** `lastScheduleTime` was empty at 19:15Z
  — the CronJob was created 2026-09-02 10:30:27Z, after its own `40 7 * * *` slot, so its first
  scheduled run is the following morning. Every dated receipt before then is a manual trigger.
  **Check `lastScheduleTime` before reading a clean cron history as coverage.**
- **A migration's verify block of bare `IF position(...) = 0` FAILS OPEN on NULL** — `position(x in
  NULL)` is NULL, `NULL = 0` is NULL, and plpgsql treats `IF NULL` as false. 694 now refuses
  explicitly on NULL first; copy that shape.
- **Migration numbers collide.** I took 692, found it claimed mid-work, and landed on 694; 691 is
  double-claimed by two lanes today. Re-check `ls docs/agent_docs/sql_for_agents/` immediately
  before you write the file, not when you start thinking about it.

---

## 5. Commands worth keeping (all cluster ones blocked until §0a clears)

```bash
# both detectors, scoped — NB the SQ-004 control reads FAIL when scoped away (§2b)
python3 scripts/audit-experience-promises.py --site <domain>
python3 scripts/audit-listing-class-promise.py --site <domain>
python3 scripts/audit-experience-promises.py --self-test     # fixtures, no cluster — works today

# migration: dry-run WITHOUT persisting (this is how 694 was proven)
sed -e "s/^SELECT snapshot_agent(.*$/-- dry/" -e "s/^COMMIT;/ROLLBACK;/" <file>.sql \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

# council: free admission test, then submit, then resubmit on the SAME correlation
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json
RESUBMIT_CORR=<corr> ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json

# out-of-band apply, then tell the ledger so nobody re-runs it
./scripts/migration/run-migrations.sh --record-only <file>.sql --note "<why>"
```

**Council lesson worth more than the commands: reviewers judge the SKETCH, and it is the only view
of your code they get.** Three of round 1's objections were about a `CREATE TABLE` and a
`snapshot_agent()` call that were *always in the file* and that I had trimmed out of the sketch.
Round 2 sent the full executable bodies and those objections vanished. Send the whole body.

---

## 6. The five standing docs for this lane

| doc | path |
|---|---|
| PLAN | `docs/agent_docs/docs024_key_docs_latest/experience_loop/PLAN_experience_loop.md` |
| RUNBOOK | `docs/agent_docs/docs024_key_docs_latest/experience_loop/RUNBOOK_experience_loop.md` |
| NOTES | `docs/agent_docs/docs024_key_docs_latest/experience_loop/RUNNING_NOTES_experience_loop.md` — two entries dated 2026-09-02, the second covering the apply and the designblog exchange |
| README (owner's) | `docs/agent_docs/docs024_key_docs_latest/experience_loop/README_where_we_are.md` — **append only, never rewrite; it is the owner's document** |
| SUMMARY | `docs/agent_docs/docs024_key_docs_latest/experience_loop/SUMMARY_where_experience_loop_is_now.md` — **not written this session, deliberately**: the milestone is the routing going live, and "where we are now" would repeat the last one. Write it when 2c lands. |

Register entries for this lane: **SQ-004** `listing-class-promise-check` (`25 7 * * *`),
**SQ-005** `experience-promise-check` (`40 7 * * *`).

**Refused deliberately, do not re-attempt as a regex:** *"does this page contain the thing its title
asserts?"* — the disproof is in SQ-005's docstring. The empty-index rule (§2a) is decidable because
it counts items; the title-assertion one is not.
