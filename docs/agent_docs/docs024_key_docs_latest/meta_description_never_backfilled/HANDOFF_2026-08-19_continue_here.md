# HANDOFF 2026-08-19 (final) — meta descriptions, `bugs_open/320`: DONE, and it unblocked `309`

**Read this first if you are picking this up cold.** Everything below was verified at the
artefact; where a claim is unproven it says so.

> **Rewritten 21:10Z.** An earlier version of this file said "355 pages remain" in §1 while
> a banner added later said 50. Two numbers for one fact in one document is the drift this
> lane spent the day fixing elsewhere, so it was rewritten rather than patched.

---

## 1. The one-paragraph version

`pages.meta_description` is the sentence a search engine prints under a page title, and on
this estate it is also each blog card's excerpt. **407 of 731 live pages had none.** Two
live mechanisms: the planner was never asked for one, and an unguarded upsert clause
blanked the ones that existed. Both are fixed and live. A backfiller was built, proven, and
run across the fleet: **47 of 737 pages are empty now (6.4%), and 43 of those have no
content at all to describe.** The five pages that were blocking `bugs_open/309` were filled,
the rerender was dispatched, and **309 is fixed and verified at the served page.**

## 2. State, every line proven

| thing | state | proof |
|---|---|---|
| Chassis | **`v1.0.1316`**, revision `07eeba4a1…` | pod `imageID` == local `RepoDigests` (a real build, not a cached same-tag rebuild); commits are ancestors; **control: `HEAD` correctly NOT an ancestor** |
| M2 guard, **four** write paths | **LIVE** | `grep -rn 'meta_description = EXCLUDED' --include='*.go' . \| grep -v COALESCE` → 0 live sites |
| M1 — planner asked (mig `485`) | **LIVE** | chain verified: `write_site_plan_action.go:535` reads the exact key `485` adds, `:631` inserts it |
| `save_page_meta_description` (SEO-004) | **LIVE** | binary probe PRESENT with positive **and** negative controls |
| Copy gates (the owner's condition) | **LIVE, in the action** | 3 tests; mutations RUN: delete the gate call → 3 fail; make the unreadable branch pass → 1 fails |
| `meta-description-backfiller` (`488` + `493`) | **LIVE, PROVEN, fleet-run** | ~571 pages written; idempotence proven at row level |
| `fail_on_non_numeric` tripwire | **ARMED and PROVEN TO FIRE** | two identical workflows: with the flag → step **FAILED**, neither branch taken; without → `took_else` COMPLETED, i.e. the silent skip reproduced |
| Council | **APPROVED** round 2 | corr `46734ae9-…`; round 1 REVISE found a real defect (I had guarded 1 of 4 paths) |
| **`bugs_open/309`** | **FIXED** | 6 cards/0 anchors → **8 cards/16 anchors**, all 8 targets HTTP 200, shrink guard **passed untouched** |

**Fleet: 407/731 empty (55.7%) → 47/737 (6.4%).** Mean description length **129** chars.
Of the 47, **43 have zero components** — the floor — leaving **4** reachable.

## 3. What is actually left

1. **4 reachable pages** may need one more pass: `./scripts/backfill-meta-descriptions.sh <domain>`.
   The workflow takes 25 pages per run, which is why large sites needed several.
2. **43 pages have ZERO components.** This is a **floor, not a backlog** — a page with no
   content cannot be described from its content, and the alternative is invention. They
   need content before they need a description. Do not "fix" this by lowering the floor.
3. **Nothing is scheduled.** The backfiller is hand-driven. Now that a fleet run has been
   read and is good, putting it on a cron is a reasonable next step and a small one — but
   it is a decision, not a leftover.
4. **`bugs_open/313`'s wider sweep** — the 313 lane has built `audit-array-producer-conditions.sh`
   (WFA-018) and reports it clean; nothing owed from here.

## 4. Traps this lane paid for — do not re-pay

- **A `COMPLETED` orchestration that wrote nothing.** `output_format: "array"` returns a
  BARE ARRAY, so a gate reading `X.count > 0` resolves to nothing and silently routes to
  else. `.count` exists **only** under `output_format: "object"`
  (`database_actions.go:129-145`). **This is `bugs_open/313`, and it arrived here because I
  modelled the workflow on `internal-linker` — the agent 313 was filed against. Copying a
  live agent copies its bugs.**
- **`check_voice_tells` cannot see this column.** It scans `page_components.rendered_html`;
  `pages.meta_description` is invisible to it and to every `rendered_html` census. Wiring
  it would have produced a confident pass over text it never examined. The reusable
  text-level entry points are `VoiceGate.ScanVoice([]string, longForm)` and
  `checkBannedClaims([]string, …)`.
- **`content_sample` built from raw markup is mostly CSS.** A model handed a stylesheet
  still writes you a fluent, wrong sentence, and no copy gate catches that.
- **`--record-only` REFUSES an uppercase-suffixed sidecar.** A `_HOLD` file that is
  hand-applied and left named `_HOLD` ends up applied with **no ledger row**. Rename it
  back the moment the hold is satisfied.
- **`display_name` and `category` on `agent_definitions` are NOT NULL with no default.**
- **The guard does NOT "match `nav_label`" — it is the mirror image.** `nav_label` is
  `COALESCE(NULLIF(pages.…,''), EXCLUDED.…)`, so the **existing** value wins;
  `meta_description` lets the **incoming** value win unless blank. Both deliberate,
  deliberately different. **Do not unify them.**
- **`000` from curl is the client failing to ask, not a 404.** One card target read `000`
  and was 200 on retry. Reading it as "missing" would have manufactured a defect.
- **When two of your own measurements disagree, neither is evidence** until you can say
  which is wrong and why (1 vs 314 visible chars on the same unchanged page).
- **A `[MEASURED]` figure can be honest, dated and still unrepresentative.** I recorded mean
  description length as 102 from a 26-page two-site sample and flagged it as a shortfall
  bearing on 309's shrink guard. On the full population it is 129. **The marker says how a
  figure was obtained, not how far it generalises.**

## 5. Where everything lives

- `bugs_open/320` — the case; **§11 the owner ruling**, **§12 the live results**.
- `bugs_open/309` — **§13 is the close**, verified at the served page. Its §10 carries my
  correction of its own explanation of the blocker.
- This directory: `PLAN`, `NOTES` (append-only, newest last — the missteps are the point),
  `RUNBOOK`, `README_where_we_are` (the owner's prose log), `SUMMARY_2026-08-19`, both
  council submissions.
- Migrations `485` (planner), `488` (the agent), `493` (the canary fixes) — applied,
  ledger-recorded, each with a ROLLBACK sidecar.
- Code: `save_page_meta_description_action.go` + two test files; the four guards in
  `site_db_actions.go`, `apply_adoption_plan_action.go`, `adopt_verbatim.go`,
  `cmd/webdesignport/import.go`.
- Register **SEO-004** (`register/seo.md`) and its `000_concept_index.md` row.
- `scripts/backfill-meta-descriptions.sh`.

## 6. The owner's standing instructions on this lane

1. **Backfill authorised FLEET-WIDE, review pass WAIVED** (`320` §11, 2026-08-19).
2. **Condition of that waiver:** the summaries go through the copy guidance and checks so
   they don't sound like AI. Guidance is in the prompt (v3's rules 1-6 — the ones that
   govern a single sentence; 7-13 are about tables and paragraph rhythm and would have
   invited a table into a 155-character field). Checks are in the **action**, before the
   write, where a workflow author cannot forget to wire them.
3. **The framework writes the content, not a session** (2026-08-06). If a generator does not
   exist, that is the finding to report — which is why this lane exists at all.
4. **`309`: wait for the writer/replan** — honoured. The writer ran, then the rerender.
   No article was regenerated and `section_shrink_floor` was never touched.
