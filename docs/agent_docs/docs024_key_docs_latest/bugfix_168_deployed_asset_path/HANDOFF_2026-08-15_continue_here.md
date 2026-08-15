# HANDOFF — 2026-08-15 — FOUR OWNER INSTRUCTIONS, NONE STARTED. Read this file only

**Supersedes `HANDOFF_2026-08-14b_continue_here.md` for state.** That file's §3.1 said *"WAIT FOR IT
— do not manufacture one"* about the first gate refusal; **the owner reversed that today** (§C below).
Its LLM-capability row is corrected in place. Its §4 traps still hold except where superseded here.

Working record: `NOTES_deployed_asset_path.md` (newest at the bottom). Owner's log:
`README_where_we_are.md`. Runbook: `RUNBOOK_deployed_asset_path.md`.

**Nothing is in flight. Nothing is half-applied. Everything below is committed.**
**Nothing is blocked. Three pieces of work are queued and none has been started.**

---

## 1. State — verified 2026-08-15 09:13Z

| thing | state |
|---|---|
| Daily sweep | **RAN 2026-08-15 08:45:17.441082Z**, 20 rows restamped. Anchor now `2026-08-15 08:45:17.441082Z`; **next due ~2026-08-16 08:45:17Z** (`interval_seconds` 86400) |
| The measurement | `refused_at_a_gate 0 \| passed_all_gates 1 \| uninstrumented 0 \| total 30`. Arms: `scan_still_trips 17 · <vintage> 9 · page_absent 2 · evidence_base_absent 1 · resolved_all_gates_passed 1` |
| **The first refusal** | **STILL UNOBSERVED.** All refusal arms 0 and unexercised since the instrument shipped |
| Fleet | **`v1.0.1300`**, both chassis pods, started 2026-08-14 20:36Z. Gate arms probed in this binary with controls |
| ⚠ Fleet LLM capability | **LIVE — the 08-14 handoff's "DOWN until 2026-09-01" was WRONG.** Measured 08:52Z: 31 ok / 0 failed in the 08:00Z hour, unbroken 30 h. **The council gate is AVAILABLE.** Verify on the SUCCESS side of `llm_call_log`, never the failure side |
| Code index | **STALE — indexed commit `a85ad401`, 2026-08-12 16:01Z**, and `.go` only (7,358 symbols, no SQL). The landmine verifier says so itself. Do not read "not found in index" as "does not exist" |

**Today's prediction experiment CLOSED and HELD on all four points** — see NOTES *2026-08-15
08:45–08:50Z*. The one number that moved (`a355d78b` 14 → 19) was moved by **the register, not the
page**; that is now a `LANDMINES.md` entry, synced, verifier verdict `NEEDS_HUMAN_REVIEW` (it could
confirm the Go half; the SQL literals are outside its `.go`-only index — not a refutation).

---

## 2. The four owner instructions (chat, 2026-08-15)

> *"yes go ahead with the evidence base for webdesign.co.uk and every component needs to be
> decomposed so it becomes editable and managed by the framework. 2. scan level exclusion.
> Manufacture a flagged item so it gets a refusal. Please create a handoff so we can start this in a
> fresh chat"*

Instruction 4 is this file. A, B and C below are **not started**.

---

## A. webdesign.co.uk — give it an evidence base, and decompose every component

### A.1 Why this is the biggest of the three

**Plain terms first: an "evidence base" is the per-site list of facts the site is allowed to assert,
plus phrasings it must not use.** It is the yardstick every claim check measures pages against.

**The rule:** a site with **no current** `evidence_base` row is **opted out** of claim checking.
`check_unverified_claims.go:310` loads `WHERE site_id = $1 AND aspect = 'evidence_base' AND
is_current = true`; no row → `ParseEvidenceBase` returns nil → the numeric/stat scan does not run.
Only the **fleet-wide** banned set still applies (`claims_global.go` is deliberately nil-safe, so a
new site is not born completely unarmed).

**This case, measured 2026-08-15 09:00–09:13Z:**

| | `webdesign.uk` | `webdesign.co.uk` |
|---|---|---|
| site id | `1fcfa4f3-ec80-4010-878b-b971cd46711f` | **`6b49db8e-d447-4467-8277-4f3018af9897`** |
| deployed pages | 7 | **103** |
| `evidence_base` rows, any version | 10 (1 current) | **0 — none, ever** |
| serves visitors? | **no — redirects to `.co.uk`** | **yes** |

⚠ **`webdesign.uk` redirects to `webdesign.co.uk`** (measured: `curl -sL` on `webdesign.uk/` ends at
`https://webdesign.co.uk/`, 200, 27,584 bytes). **So the register the owner re-attested yesterday
protects the 7-page site nobody lands on, and the 103-page site visitors actually reach has no
register at all.**

⚠ **Do not repeat my counting error.** `count(*)` over a `LEFT JOIN site_specs` returned **1** for
`webdesign.co.uk` — that counts the unmatched `sites` row, not a spec. The direct query returns
**0 rows**. Ask `count(ss.id)`, or query the child table directly.

### A.2 The decomposition — what the owner is asking for, and the measured gap

**Plain terms: a framework-managed page is stored as several components (slots) — hero, body,
CTA — each separately editable and re-renderable. A hand-built page is stored as ONE blob**, so
every edit is a whole-page rewrite and nothing can be managed, locked, or re-rendered granularly.

Measured 2026-08-15 09:13Z:

| site | deployed pages | components | per page |
|---|---|---|---|
| **webdesign.co.uk** | 103 | 113 | **1.10** |
| leopardessconsulting.co.uk | 46 | 120 | 2.61 |
| webdesign.uk | 7 | 19 | 2.71 |
| robot-hands.com | 35 | 97 | 2.77 |
| vonc.com | 22 | 61 | 2.77 |
| finetuning.uk | 50 | 155 | 3.10 |

**98 of the 103 deployed pages carry exactly ONE component** (2 pages have 2, 3 have 3, 1 has 4).
Zero locked. The framework-built sites cluster at 2.6–3.1; `webdesign.co.uk` is the outlier by a
wide margin, which is the shape the owner's instruction describes.

### A.3 ⚠ THE RULE THAT GOVERNS HOW THIS IS DONE

**CLAUDE.md, OWNER RULING 2026-08-04: `EVERY SITE GOES THROUGH THE FRAMEWORK. Never hand-build one.`**
No hand-authored HTML, however small or temporary. That ruling was *raised by the webdesign
shopfront itself* — a session hand-wrote it and shipped it to `portfolio-sites/`. **The
decomposition is the remedy for that original sin; doing it by hand would repeat it.** Seed the
specs and dispatch through the pipeline. If the framework cannot yet decompose an existing blob,
**that is a bug to file, not a licence to hand-roll.**

### A.4 Suggested order, and it is a judgement the next session may revisit

1. **Create the evidence base first.** It is cheap, it is read-only in effect, and it *measures the
   scale of the problem* before any content moves. `refresh_evidence_base` is the registered action
   (`refresh_evidence_base_action.go:104`). `webdesign.uk`'s current row was written by
   `created_by='ai-site-selling-automation-2026-08-14'`, `source='owner_ruling'` — read that row as
   the worked example of the shape.
2. **Then decompose**, through the framework.
3. **Then fix** whatever the register flags.

**Reason for that order:** fixing a flagged claim inside a single-blob page means rewriting the whole
page, and CLAUDE.md's standing warning applies — *an agent that rewrites a whole artifact can persist
a fragment and report success* (`output_tokens == max_tokens` means CUT, not finished; `bugs_open/012`
saved a 10,272-char component back as 1,253 chars). Decomposed slots make every later fix small.

⚠ **Expect a wave, and sequence for it.** Switching on a register for 103 previously-unchecked pages
will file a batch of `claims_unverified` items at once. That is the mechanism working, but do not
let it look like an incident: say in advance that it is expected, and check the first few by hand for
false positives before trusting the batch (see the `finetuning.uk/privacy-policy` case in §C.2).

---

## B. Scan-level exclusion — stop scanning archived, never-deployed pages

### B.1 Where it goes

`platform/orchestration/actions/discovery_checks/check_unverified_claims.go:371-381`, the
`pageQuery` in `ScanDeployedClaims`. Today it filters on **`p.site_id` and content presence only** —
there is no page-status and no deploy filter, which is exactly why a rejected draft is scanned:

```sql
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = $1
  AND ( (pc.rendered_html IS NOT NULL AND pc.rendered_html <> '')
     OR (pc.content_data IS NOT NULL AND pc.content_data::text <> '{}') )
```

### B.2 ⚠ THE PREDICATE — and getting this wrong is worse than not doing it

**Do NOT exclude on `p.status = 'archived'` alone.** `LANDMINES.md` carries the entry
**"an `archived` page can be SERVING"** — status is not a serving signal, and excluding on it would
silently stop checking live pages. The safe exclusion is the **CONJUNCTION**:

```sql
AND NOT (p.status = 'archived' AND p.deployed_at IS NULL)   -- archived AND never deployed
```

Both halves are load-bearing: archived-but-deployed may still be serving; deployed-but-not-archived
obviously stays.

### B.3 ⚠ Three consequences to state in the council submission

1. **It edits a DOCUMENTED property, deliberately.** `revalidate_unverified_claims.go` (~:358) says
   in terms: *"Note this check has never filtered on page status, so — unlike the voice scan — an
   archived page still comes back and is still judged."* That is a design statement, not an
   oversight. Say you are changing it and why; do not let a reviewer discover it.
2. **`ScanDeployedClaims` is SHARED between the emit side and the revalidator** — the "predicate
   parity" doc comment (council round 6, `reuse_agent`) exists precisely so both ends judge the same
   text. Changing page selection changes **filing AND revalidation** together. That is the correct
   direction, but it is a shared-seam change, so name the consumers rather than measuring them
   (owner ruling 2026-07-29 §3).
3. **⚠ THIS DOES NOT RETIRE `a355d78b`, AND A READER WILL ASSUME IT DOES.** An excluded page returns
   no scan row → `scan == nil` → `unverifiedClaimsVerdict` returns `revalidationUnknown` with arm
   **`page_absent`**, which is **not a closure**. So the item stays open, with a different arm, and
   the `page_absent` count goes 2 → 3. **Disposing of the item is a SEPARATE act** — cancel the row
   explicitly. The exclusion stops the *class* (and stops the count climbing); it does not clean up
   the instance.

### B.4 Gate

`platform/` scope → **submit to the council gate**, which is available (LLM capability verified live
**08:52Z**, see §1). It may also be **architecture-scope**: it narrows what a shared scanner covers,
and the 2026-07-29 ruling makes the test *"does it change what the shared mechanism GUARANTEES?"* —
narrowing coverage arguably does. **Flag it in the submission and let the seats rule**; do not decide
it silently either way.

---

## C. Manufacture a flagged item so a gate REFUSES

### C.1 The authorisation, and what it supersedes

**Owner, 2026-08-15: "Manufacture a flagged item so it gets a refusal."** This **reverses**
`HANDOFF_2026-08-14b` §3.1 (*"WAIT FOR IT — do not manufacture one"*) and the reasoning under it.
The wait has produced nothing: two sweeps since the instrument shipped, zero refusals, all refusal
arms still unexercised.

### C.2 The candidate, chosen against the ladder

**`20d5da84` — `leopardessconsulting.co.uk` / `for-engineering-teams`.**
`n_findings = 1` · `matched = "90,790"` (**6 chars, highly distinctive**) · `build_status='deployed'`
· `deployed_at` non-null · currently `scan_still_trips`.

**Rejected candidates, and the reasons are the trap:**

| item | page | token | why not |
|---|---|---|---|
| `15479f3c` | finetuning.uk/privacy-policy | `16` | ⚠ **NEVER TOUCH** — matches the age in *"…from anyone under the age of 16"*. Deleting it damages a legal notice |
| `b090e138` | gamesdesign.co.uk/about-index | `100%` | short needle |
| `f55f833c`, `3375653f`, `dc7a2041`, `08a42559` | various | `3` | short needle |
| `b561c826` | robot-hands.com/matchmatrix-methodology | `2` | short needle |
| `962da5c9`, `9a0e67d9` | various | `47`, `11` | short needle |

**Why short needles disqualify:** `claimStillOnPage` (`revalidate_unverified_claims.go:325-340`) is a
**case-insensitive substring scoped to the slot**, and its doc comment says the crudeness is
deliberate — *"a token that matches something unrelated in its own slot produces a refusal … never a
closure."* So a one-character token pins the item at `gate_claims_still_present` **for ever**, which
is a refusal, but the *wrong* refusal and an unclearable item.

### C.3 The recipe

1. **First, verify `90,790` is a genuine overclaim** — read the `snippet`, and check the register:
   ```sql
   SELECT f->>'id', f->>'value', f->>'tolerance', f->>'writer_line'
   FROM site_specs ss JOIN sites s ON s.id=ss.site_id, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE s.domain='leopardessconsulting.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;
   ```
   If it is a checker false positive, pick again — do not delete a legitimate figure to win an
   observation.
2. **Edit BOTH surfaces.** ⚠ `ScanDeployedClaims` reads `rendered_html` **and** `content_data`, and
   the claim-granular gate searches `html || contentJSON`. **A one-surface edit leaves the finding
   standing** — this cost the 08-14 session a near-miss (its planned single edit would have shipped a
   no-op behind a `COMPLETED` orchestration). Worked pair of scripts in this directory:
   `SQL_2026-08-14_clean_case_studies_content_data.sql` / `..._rendered_html.sql`.
3. **DO NOT REDEPLOY.** That is the whole point.
4. **Wait for the next daily** (~2026-08-16 08:45:17Z), condition-based — poll `last_triggered_at`
   until it CHANGES; never compute a wake time with `date -u -d '<literal>'`, which parses in LOCAL
   time. Reusable watcher pattern in NOTES *2026-08-15 08:0xZ* and the `LANDMINES.md` timing entry.
5. **Expected arm: `gate_published_correction_unpublished`.**

**Why it produces that arm:** the published gate compares `deployed_at` against
`newest_component_update`. Editing components without redeploying makes
`newest_component_update > deployed_at`, which is the `bugs_open/262` case exactly. The 08-14 session
had this window open for ~3 minutes by accident and the rerender closed it before the sweep landed.

### C.4 ⚠ Safety notes

- **The live page does not change.** We edit stored copy and do not deploy, so visitors keep seeing
  exactly what they see now. The overclaim stays live a little longer — that is the status quo, not
  new harm — and **once the refusal is observed, redeploy to close the item properly.** Do not leave
  it parked.
- **Guard both UPDATEs with `DO` / `RAISE`, and induce the guard before trusting it.** A verify block
  of bare `SELECT`s **cannot stop a `COMMIT`** (`ON_ERROR_STOP` ignores a non-empty result). Run it
  once with the wrong expected delta and confirm it aborts.
- **psql does not interpolate `:vars` inside a dollar-quoted body** — pass via
  `SET LOCAL app.expect_delta = :expect_delta;` and read `current_setting(...)::int`.
- **A manual sweep is FLEET-WIDE (500 items, oldest-first).** If you fire one rather than waiting,
  report everything it closes; and do **not** wind `scheduled_tasks.last_triggered_at` back — that
  moves the daily anchor permanently on a row this lane does not own.

---

## 3. Traps carried forward — still live

- ⚠ **`last_completed_at` is NOT a completion signal** on `review-queue-revalidate-daily`: it equals
  `last_triggered_at` exactly (both today and 08-14) because the scheduler stamps both at dispatch.
  **The evidence a sweep did work is the `revalidation.at` stamps on the rows.** (New today.)
- ⚠ **Never read a measurement of a scheduled run without the anchor and a clock stamp in the same
  output.** A pre-run and post-run reading of a last-write-wins field are structurally identical.
- ⚠ **`result.revalidation.arm` is LAST-WRITE-WINS** — a snapshot, never a history. Drop *ever*,
  *how often*, *rate*.
- ⚠ **`resolved_all_gates_passed` carries no `gate_` prefix** — a prefix-only reach query counts only
  refusals and misses every closure. Use `arm LIKE 'gate\_%' OR arm = 'resolved_all_gates_passed'`.
  `arm IS NULL` is a **vintage** marker, not a gap; the gap check is `arm LIKE 'unreported:%'`.
- ⚠ **A claims count is a COMPARISON — the register moves it with no page edit**, and the half that
  moves it (`banned_claims`) lives *inside* the `evidence_base` row's `data`, not in its own aspect.
  Full entry in `LANDMINES.md` (filed today). **Diff whole objects first, project second.**
- ⚠ **An `archived` page can be SERVING** — never conclude a page is dead from `pages.status`.
- ⚠ **The `page-rerender` pod restarts often** — check `.status.startTime`; CLAUDE.md's ~300s
  post-restart spawn-drop rule is real and bit the 08-14 session.
- ⚠ **`kcat -P` exits 0 having sent nothing** — after any trigger script, confirm the row landed
  (`orchestration_states` by correlation) rather than trusting the exit code. Verified working today.
- ⚠ **`LANDMINES.md` takes same-file passengers** — another session had 17 uncommitted lines in it
  at 09:05Z today. Gate on `git diff --numstat` and commit by pathspec.

---

## 4. Commits from the 2026-08-15 session

| sha | what |
|---|---|
| `56160db8a` | pre-run addendum (prediction 2's "no third outcome" is incomplete) |
| `5b8f04831` | sweep result + the new `LANDMINES.md` entry |
| `2772ef074` | owner log + the LLM-capability correction to the 08-14b handoff |

Landmine verifier: corr `095e988a-bc90-4a5e-847e-6bf23d8aa815`, verdict in `doc_notes`
(`categories ? 'landmine-verification'`), `NEEDS_HUMAN_REVIEW` — Go half confirmed, SQL literals
outside its index.

**No platform code changed this session, so no council submission is owed.** B will owe one.
