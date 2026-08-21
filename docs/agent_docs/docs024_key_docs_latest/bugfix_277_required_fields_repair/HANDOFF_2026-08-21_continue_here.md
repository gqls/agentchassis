# HANDOFF — 2026-08-21. `277`'s clause 1 is MET (7 pages repaired, proven at the served bytes), the owed `diagnosis_guardian` fix is APPLIED, and `083`'s soak is at day 4 of ~7

**Supersedes `HANDOFF_2026-08-20b_continue_here.md`.** That file's §2 sequence is now HISTORY — all
four steps ran today. Read this from disk; then `NOTES_required_fields_repair.md` **from the
bottom** (two new entries today, §§1–7 the canary and §8 the seat fix).

> **Written ~15:00Z. Deploy facts have a shelf life of hours — re-read, do not quote.**
> Chassis was **`0483e7f4e`** (`sha256:3ed50651…`, pods up 2026-08-20 19:51Z) all day; `af0f00bb5`
> and `6011f9657` are both aboard, proven by ancestry against the pod's own
> `buildinfo.GitCommit`. **Nothing this session needs another roll.**
> Live md5s after today's config changes: `fix-proposer.review_diagnosis_guardian`
> **`99bf2e45…`**, `council-gate.review_diagnosis_guardian` **`347a20cf…`**.

---

## 0. STATE TABLE

| bug | state | what blocks the close |
|---|---|---|
| **`bugs_open/277`** | **clause 1 MET** — 7 pages repaired, verified at the served bytes (§7 of the bug file) | the **`no_content_data` half**, untouched by all of this: 27 of 30 parked rows, a different agent, and `473`'s deterministic route does not cover it |
| **`bugs_open/083`** | fix live + artefact-proven; **door soak at day 4 of ~7** | ~**08-24/25**: `444`/`458` sat their week. Close then, with the two statements in §3 below |
| `bugs_open/333` | theirs | **CONTRIB filed today** — the two-strike rule reaches their false *"tried twice"* down a second road, with no refusal loop involved |
| `530`/`531` (council) | **APPLIED + APPROVED r1** (corr `c00fbfd8…`, 14:03:49Z, `point_fix`, no truncation) | **nothing** — verdict read, both medium advisories answered with the queries they asked for (NOTES §10), trailer written. Done |

---

## 1. WHAT HAPPENED TODAY, in the order it mattered

1. **The roll had already happened** (08-20 19:51Z). Proven per service with both controls — see the
   banner. The `build provenance` **startup line was NOT readable**: a chassis pod's log history is
   ~90 seconds under load, and the three "hits" a grep found were the phrase quoted inside a council
   submission's `collected_data`. Use the `buildinfo.GitCommit=` probe of `/proc/1/exe`.
2. **The sweep the canary needed was FOUR DAYS away, and nothing said so.** `literal_markdown` is a
   *quality*-discovery check; only `site-discovery-rotation-quality` runs it, `LIMIT 1` site older
   than 7 days. Every one of the 25 stamped sites was inside its floor (oldest: 5d 01h), so the
   rotation was **idle** while its `last_triggered_at` advanced every 3h. Owner approved forcing it.
   A one-shot task (**no `pre_query`, so the rotation stamp was not consumed** — the natural 08-25
   slot survives) fired 13:19:01Z; the run was confirmed at `orchestration_states`, not at the stamp.
   Recipe + the "when would the rotation reach my site" query: **RUNBOOK**, "Force a discovery sweep
   for ONE site". Also filed as a LANDMINE.
3. **The router filed 8 rows and every one carried the new shape.** No other check filed anything.
4. **One canary promoted by hand 13:21:42Z → `complete` 13:25:03Z → the promoter released the other
   six on its very next tick (13:27) → 7 of 7 `complete`, all `verified`, by 13:37Z.**
5. **Proof at the served bytes on all seven**, with the control that carried the risk: prose
   backticks to **zero** everywhere while `tool-head-architect` kept all **44** of its in-script
   template-literal backticks. Per-page table in NOTES §7; the worked before/after in `277` §7.
6. **The owed `diagnosis_guardian` message became migrations `530`/`531`** — see §2.

---

## 2. `530`/`531` — APPLIED, and what a reviewer of this needs to know

The seat's prompt asserted the coordinator *"reads ONLY `step.config.error_step`"* and that a
step-level one is *"silently inert"*. `routeToErrorStepOrFail` (`coordinator.go:3666-3679`) checks
**step-level FIRST** — its own comment calls that the preferred location — with config-level as the
backward-compatibility fallback. The seat was objecting to authors who did the right thing.

While reading both rosters side by side: council-gate's copy carried
`## The author's stated rationale loop's load-bearing disciplines`, produced by
`099_SYNC_gate_roster.py:85`'s **unanchored** `replace("## The diagnosis", …)`. `531` repairs the
live text and the script's substitution is now line-anchored so it cannot recur. **`099`'s OTHER
defect is untouched: its transform predates `377`, so `--apply` stays SUSPENDED.**

Exercised three ways and round-tripped byte-identically before applying — the table is in NOTES §8.
⚠ **`531`'s marker-offset guard is structurally unfirable for this anchor set and its header says
so**; the guard that can fail is the fleet-wide `17 seats / 1 prefix` check, induced and confirmed.

---

## 3. `083` — the close, ~08-24/25, and the two things it MUST say

The mechanism is proven twice over now (today's arc is a second independent instance — bug file,
2026-08-21 entry §1). When the doors have sat their week:

1. **Move the file with BOTH paths on the commit** (`git commit bugs_open/083_… bugs_closed/083_… -m …`)
   and verify **at HEAD**, not at the tree:
   `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 083` → exactly one line
   ⚠ **two unrelated cases share the number 083** — resolve by slug.
2. **State that `479`'s reclaim arm has never fired** — re-measured today, **0 live / 0 archive, all
   history**, while the escalated population **grew to 7** (today's 12:59Z tick added
   `dead_fragment_link` and `missing_conversion_path`). The door opens on schedule and nothing has
   ever come back through it. Do not let the close imply otherwise.
3. And carry §3 of that entry: the **third stranding shape** (two-strike strikes inherited across a
   re-route) is invisible to all three of 083's own instruments, because the row is born terminal.

---

## 4. STILL OWED

- ~~Read the `c00fbfd8` verdict~~ **DONE — APPROVED round 1, advisories answered, NOTES §10. Nothing is owed on the council trail.**
- **`083` close, ~08-24/25** (§3).
- **`277`'s `no_content_data` half** — the only thing holding that file open. Different agent;
  `473`'s deterministic route does not cover it.
- **`copy_edit_proposed` exclusion in the promoter's `pre_query`** (owner decision D2, 2026-08-12).
  Still deliberately not done by a session: it changes which rows are dispatched.
- **A disconfirmable prediction to check on ~08-25:** `learn-index`'s next filing should be born
  `detected`, not `unresolved`, because both of its strikes age out of the rolling 7-day window
  first. **Do not hand-flip the row** — the prediction is worth more than the one page.

## 5. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py` **by slug** for `277`,
`083`, `333` · the chassis stamp + `git merge-base --is-ancestor` for anything you think shipped ·
`SELECT status, count(*) FROM site_work_items WHERE item_type='literal_markdown' AND
handler_agent='section-editor' GROUP BY 1;` (expect 7 `complete`, 1 `unresolved` until the 08-25
sweep) · then §4.
