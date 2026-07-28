# NOTES — experience register (append-only, newest at the bottom)

## 2026-07-24 — workstream born from discussion (session "experience register")

Owner brief: a directory of small reusable experiences — carousel behaviour is approved but
what a click on a card does is undocumented and re-invented each use; wanted: a register
recording e.g. "read more" expands in place, card click → article page → onward links to
info/tools. Several thousand eventually, used as base plans, forked per site, testable end to
end; links get "probably should go" instead of guesses.

### Round 1 — four parallel doc searches (experience loop / travelling docs+tool-improver /
### concept register+feature-builder / UX+link-integrity)

Findings that held: no experience-register-like construct exists anywhere (no table, no stub);
`content_components` has no behaviour/destination fields; bug 023 is the exact complaint
("nothing pairs label and URL"); EXPERIENCE_PLAN has the right vocabulary (journeys with
page/control/action/observable-outcome, promise ledger, criteria) but is per-site, authored
from a blank page, one exists; LNK-011 reserved an agent boundary for intent-matching, never
built; concept register = proven harvest→adversarial-verify→register pipeline (1,633 entries).

Positions I took in round 1 that were later corrected:
- Proposed a concept-register-style flat-markdown register seeded into a table.
- Leaned approval-by-use over per-entry approval.

### Owner corrections (round-1 reply)

- **Travelling docs was created for this** — documents a tool's provenance and direction;
  closer ancestor than I stated. Correct: I had under-weighted it.
- **DB-based storage preferred.**
- Taxonomy: our own, loosely based on the UX industry playbook.
- **Approval per experience**; formalise the acceptance-test side.
- vonc creates the first viable product; question: does that include T4, or does owner
  initiate?

### Round 2 — three parallel searches (site-plan machinery / travelling-docs mechanics /
### vonc pilot state)

Load-bearing findings, each verified by file:line in the reports:
- `doc_plans` has **no site_id** — already library-level. subject_type CHECK today =
  tool|pipeline|experience|action (163, 184).
- **doc_plans is exact-key only** — no metadata column, no structured search; RAG indexing of
  plans is design-intent, not a verified pipe. Travelling docs give provenance/direction, NOT
  selection → the register table supplies selection, the travelling doc travels with each
  entry (the content_components + doc_plans 'tool' precedent).
- **184 split contract found** (verified first-hand this session, not just from the agent
  report): `docResolveSubject` (write_doc_plan_action.go:136-144) rejects 'action' and is
  shared by ALL THREE doc actions (write:59, append_doc_note:59, load_doc_context:56), so the
  184-seeded action rows are unreachable through the doc actions. Separately
  `persist_diagnosis_note_action.go:78` allows only tool/pipeline — it silently skips even
  'experience' (163 missed it). Filed as **bugs_open/064**.
- Criteria fence: read-time parse only; stale criteria ×3 sightings, `-EDIT` placeholders
  silently skipped, unclosed fence → silently "". Formalisation target confirmed.
- Site-plan: page set decided in build-site-planner `plan_site`; `roadmap_brief` is the
  authority-override precedent ("the roadmap is the authority") — the experience_brief hook
  copies it. Preservation set + owned-page guard are the re-plan-safety precedents.
- **Dormant `site_flows`/`flow_pages`** discovered: narrative-funnel schema
  (awareness/consideration/conversion), ZERO platform/ references. Design supersedes the
  tables, adopts the funnel-stage vocabulary (recorded in PLAN §4).
- vonc/T4: nothing auto-fires T4 (session-driven); run-12 plan REJECTED (must not build);
  re-fire 092 only after tools-api deployed+smoke-POSTed. tools-api: design approved corr
  278a37c3 (2026-07-24 09:59), implementer refused twice (max_tokens; map-not-string), round
  3 in flight, zero code landed, no PR; branch feat/278a37c3 carries no tools-api commits.
  Owner hard gates: PR merge + 4 bastion/tunnel infra tasks. T5.1 journey runner not started.

### Owner rulings (round-2 answers, recorded in PLAN §2)

Substrate = register table + travelling doc. Site-plan hook = experience_brief aspect.
Taxonomy = layered, seeded from harvest. vonc = **wait for tools-api** (my recommendation was
the feasibility veto's static cut; owner ruled wait — the full debate path is the first
viable product; harvest waits for it).

### Missteps this session

- Round 1: stated travelling docs was per-instance-tool-only prior art and proposed flat
  markdown storage — under-read; owner corrected; round-2 research confirmed the substrate
  is library-level by construction (no site_id).
- Round 1: recommended approval-by-use; owner ruled per-experience approval. Synthesis kept:
  'proven' remains an evidence upgrade after council approval.

Artifacts written this session: PLAN, RUNBOOK, this file, README_where_we_are, design/ (4
draft artifacts, nothing applied), bugs_open/064, 016b §9 pattern entry, memory pointer.

---

## 2026-07-26 — the vonc gate lifted; first harvest done

**Gate check, first-hand (not trusting the other session's log):**
- `GET https://vonc.com/provocations/index.html` → 200, 23,670 B.
- `GET https://vonc.com/data/provocations.json` → 200, 9,797 B, `diff -q` **byte-identical**
  to `gauntlet_dead_cta/p4_sources/provocations.json`.
- Live page carries `data-component="provocations-archive-list"` + `data-archive-template`.
- Live `/assets/js/snippets.js` (22,475 B) contains the loader by **discriminating** strings
  (ones the loader creates, not uses): `provocations-archive-detail-style` ×1,
  `__item--linked` ×1, `__item--static` ×2, `entriesBySlug` ×3.
- Journeys themselves: the owning session's browser run, 72/73 desktop+mobile
  (`p4_sources/verify_live_2026-07-26.txt`); the 1 failure is upstream 503s = `bugs_open/083`.

**Component grounding (live DB):** `provocations-archive-list` `70d6662a` (quality 100),
`gauntlet-interface` `5da50747` (100), `provocation-card` `6163ff14` (50), `lobby-grid`
`9304f14d` (50) — **all four `usage_count = 0`**. Bespoke components; the clauses repeat, the
components do not. That is the register's case in one number.

**Harvested:** four draft entries in `harvest/entries/` — `CC-001 feed-driven-teaser-list`,
`MJ-001 teaser-detail-deeplink`, `CC-002 feed-promised-cta`, `MJ-002
timed-remote-challenge-loop`. Full record + the ten design corrections they forced:
`harvest/HARVEST_01_2026-07-26_vonc_provocations.md`.

**Criteria vocabulary, read from source (not assumed):**
`check_tool_acceptance.go:343-380` → `selector_exists`, `selector_count`, `interaction`,
`asset_loads`, `page_status_ok`. `run_checks_action.go:374-467` adds `no_horizontal_overflow`
at Tier 4; steps are `fill|click|select` only (`:772-776`); `runInteraction` (`:493`) asserts
straight after the last step with `stepDelay = 300ms` (`:199`), `settleDelay = 2s`,
`runDeadline = 120s`; `expect` is `{selector, text_matches}`.

> **CORRECTED 2026-07-26 (my own artifact, caught by harvest):**
> `design/criteria_template_schema_v1.md` listed schema v0's types and said "no new check-type
> invention beyond `journey`". Both parts were wrong. It **missed
> `no_horizontal_overflow`** (Tier-4 only), and four separate extensions turn out to be
> load-bearing, not optional: attribute assertions, a navigation step + cross-page status,
> an empty-region assertion, and waits/ordering/retries. Caught by trying to write the
> criteria for two real, live patterns and finding the clauses that matter unassertable.

> **CORRECTED 2026-07-26:** `design/taxonomy_seed.md` named pattern #1
> `teaser-detail-related` with a third "related links and tools" leg. **The live
> implementation has no such leg** — the onward links live in the feed's `today`/`arena` CTAs
> (harvested separately as `CC-002`). Renamed `teaser-detail-deeplink`. First time the
> bottom-up rule paid for itself: authored top-down, the register's first entry would have
> carried a leg no implementation has.

**064 — CLOSED, and my first reading of it was stale.** I pod-grepped the running chassis
(`agent-chassis-5785dd5c85-jff28`, image **v1.0.1167**, started 2026-07-26T17:11:30Z):
`unsupported subject_type` ×1 (the string the fix CREATED), `validDocSubjectTypes` ×1,
`no explicit subject` ×1 (the *other*, deliberately distinct skip reason — both present is
the correct post-fix state), `experience-pattern` ×0 (correctly not added — that is P2's),
positive control `write_doc_plan` ×7. I then wrote that the failing branch still needed a
live run and folded it into P2.

> **CORRECTED, same session, before committing:** the file is not in `bugs_open/` at all —
> `bugs_closed/064…`, commit `eb81de7b5`, **closed 2026-07-25 on image v1.0.1156**, a day
> before the pod I grepped. The closing session proved BOTH branches live: a scratch
> one-step `load_doc_context` with `subject_type='action'`/`diagnose_build_gate` returned the
> 184-seeded PLAN (orch `91d550a2`), and the negative branch with `subject_type='site'`
> failed with the new error verbatim (orch `f74d2442`), then cleaned up every scratch row.
> **What caught it:** the Edit tool refused the path. **The cheap check I skipped:** the bug
> file was in my context from a summary, and I re-read the *content* without re-checking the
> *location* — `ls bugs_open/ bugs_closed/ | grep 064` costs nothing and is the standing
> "grep BOTH directories" rule I had already been told. Nothing downstream was wrong (my
> P2-shrink conclusion holds and is if anything stronger), but the claim "not closed" would
> have shipped in a commit. Logged in WRONG_CALLS.

**Missteps / near-misses this session**
- Nearly wrote the harvest's openable/inert clause into BOTH the component contract and the
  journey. Caught while writing `MJ-001`: two accounts of one clause is the drift class this
  workstream keeps filing bugs about. Resolved by the composition rule (HARVEST_01 §3.6) —
  render-time properties live in the component contract; the journey references them.
- First draft of `MJ-001` carried `ai-never-funny-on-purpose` as a literal in the criteria.
  That is the `bugs_open/045` static-fallback mistake in a new place: the openable set changes
  with the feed. It is now a binding (`sample_item_key`), resolved at bind time.

## 2026-07-26 (later) — HARVEST 02, the brochure component library

Owner ruled the brochure set be harvested BEFORE building P2 (my recommendation, on the
grounds that harvesting 4 entries had already changed the schema five ways). Justified: it
changed it again, structurally.

**Live verification — all five components, one per page, fetched this session:**

| component | page | HTTP | proof |
|---|---|---|---|
| hero-card-carousel | /capabilities.html | 200 | 4 slides, `data-hcc-autoplay="false"`, `[data-hcc-live]` present, **zero** `[data-hcc-pause]` |
| image-hover-card-grid | /model-fine-tuning.html | 200 | 4 `:focus-visible` rules in the shipped CSS |
| swipeable-insight-carousel | /multi-agent-review-council.html | 200 | `scroll-snap-type` present; links on 2 cards only |
| stat-band | /index.html | 200 | 4 `[data-countup]`, each with the authored value on `aria-label` |
| people-feature-block | /about.html | 200 | exactly one anchor |

Served bundle `/assets/js/snippets.js` (7,614 B) carries both behaviours by created strings:
`data-hcc-track` ×1, `__statBandInit` ×2, the literal `Card " + (current + 1)` ×1,
`data-countup` ×1.

**The finding: one invariant, six sightings.** `no-inert-control` — see HARVEST_02 §2 for the
table. Five different authors, five different mechanisms, one rule. P2 gains an
`experience_invariants` table + `requires_invariant`; second invariant already visible
(`pointer-behaviour-has-a-keyboard-equal`).

**A harvested clause found a live defect.** Probing `CC-003`'s destination clause:
`/capabilities` → **404** (all four carousel cards link there), and none of the four fragments
exists on `/capabilities.html` either; the hover grid's cards point at absent fragments on a
live page. Checked `bugs_open/` first — it is `071`'s class (extension-less + fragment blind
spot), already owned and actively worked, with a hand-over from 049 the same day. **Evidence
appended to 071, no new bug filed**, per the grep-before-filing and who-owns rules.

**Missteps this session (second batch)**
- Grepped the live page for `:focus-within` and concluded the hover reveal was pointer-only —
  wrong pseudo-class. The component uses `:focus-visible`, correctly paired with `:hover` on
  every rule. Caught by reading the component source before writing the entry, which is the
  standing "read the function before changing it" rule applied to CSS. **Had I written the
  entry from the grep, the register's first accessibility clause would have been a fabricated
  defect report about someone else's component.**
- Nearly recorded `stat-band` as interaction-free by carrying my own 07-24 taxonomy line
  forward. It has the most interesting behaviour of the five. The line was written from the
  component list, not from the component.

## 2026-07-26 (evening) — P2a BUILT: the substrate + the validator

Owner's sequence (harvest brochure → build P2) reached the build. Commit **`2f220f261`**,
4 files, 973 insertions. **Inert**: Go needs an image roll, and migration 218 is deliberately
**NOT applied**.

**What shipped**
- `docs/agent_docs/sql_for_agents/218_experience_register_substrate.sql` — `experience_patterns`,
  `site_experiences`, `experience_invariants`, and `'experience-pattern'` added to both doc
  subject CHECKs. Idempotent throughout. Seeds the two invariants (with their sightings as
  evidence); seeds **no entries** — the first rows must be ones the validator accepted, or the
  register has no proof its own contract holds.
- `platform/orchestration/actions/doc_subjects_common.go` — `+"experience-pattern"`. One string.
- `platform/orchestration/actions/experience_criteria.go` — `ValidateExperienceCriteria`: parse,
  placeholder closure (template AND contract), no absolute URL / no unbound site value, no
  `-EDIT` ids, and the tier-capability check. A check beyond the platform is **deferred with its
  reason**, never dropped and never counted as a pass.
- `platform/orchestration/actions/experience_criteria_test.go` — behaviour tests pinned to the
  harvested templates + the capability lockstep test.

**Falsification probes run BEFORE committing** (the standing rule: break the thing the check
checks and watch it fail):
- Migration lockstep: removed `experience-pattern` from the Go list → FAIL naming
  `218_experience_register_substrate.sql` and listing both sets. Restored → pass.
- Capability lockstep, **both directions**: added a type the checkers lack → FAIL "table says
  \"selector_visible\" is Tier 2 but …check_tool_acceptance.go does not implement it"; removed
  `no_console_errors` from the table → FAIL "the browser runner implements
  \"no_console_errors\" but the validator's table does not know it". Restored → pass.

**A third thing my own criteria doc had wrong**, found while writing the capability table from
source: Tier 4 also implements **`no_console_errors`** (7 types in total, not 6). And Tier 2's
`expect` struct carries only `Selector` — **`text_matches` is silently ignored at Tier 2**, so a
text assertion is anchor-confirmed there and only really asserted in a browser.

**Council gate**: submitted corr **`f4610451-6bff-45d0-8d18-6f25d26640cd`** (4 edits, 5
byte-verified `grounded_in` quotes — each checked to appear verbatim in its file before
submission, per the quote-fidelity lesson). Verdict pending; queue depth was 8 at submission.
**No `Council-Reviewed:` trailer is possible on `2f220f261`** — the verdict post-dates the
commit, so 098 will list it as unreviewed by design. Recording it here is the substitute.

**Deliberately NOT done, and why**
- No write path or bind path yet — the next slice. The validator has to exist first, since the
  register's first rows are meant to be ones it accepted.
- The migration is not applied. Image before migration, or the widened CHECK recreates 184's
  split exactly.

---

## 2026-07-27 — the substrate is LIVE; and a council run that returned no verdict

**The gate lifted on its own.** Chassis **v1.0.1172** (pod `agent-chassis-7f88c4bd7f-bhhbf`,
started `2026-07-27T10:55:44Z`) carries the Go half. Pod-grep, with controls, because a bare
count proves nothing:

```
strings /app/agent-chassis | grep -c "experience-pattern"      → 1   (my change)
strings /app/agent-chassis | grep -c "docResolveSubject"       → 2   (positive control)
strings /app/agent-chassis | grep -c "experience-nonsense-xyz" → 0   (negative control)
```
`ValidateExperienceCriteria` greps **0** — expected, and worth writing down so nobody reads it
as a failed deploy: the function has no caller yet, so the linker drops it. **A dead-code symbol
is not a deployment check.** The discriminating string here is the one in `validDocSubjectTypes`,
which *is* reachable, via `docResolveSubject`. Confirmed the string has exactly one Go source:
`grep -rn '"experience-pattern"' --include=*.go .` → `doc_subjects_common.go:37` and the test file.

**Migration 218 applied — by hand, deliberately.** `run-migrations.sh` listed **20 pending
files**, 19 of them other threads'. `--apply` applies *all* of them in order. Several are parked
on purpose: `229_vonc_about_swapped_stat_values.sql`'s own probe already says *"the exact swapped
state … was not found — re-survey before forcing this"*. So: `psql -f` the one file, then the
runner's `--record-only` path to register it in the ledger (a hand-applied file that is never
recorded stays "pending" for ever and eventually gets replayed — `bugs_open/007`, which cost ~3
days in July).

> **MISSTEP (mine, caught by re-reading the runner's README before applying):** the migration had
> **no guard block**. The README requires one — `DO $$ … RAISE EXCEPTION … $$` inside the same
> `BEGIN/COMMIT`, so a partial apply rolls itself back — and mine ended with bare `SELECT`
> verifies *after* `COMMIT`, which report but cannot prevent. I had probed the Go side by induced
> fault and simply not applied the same standard to the SQL. Added the guard (and the rollback
> recipe) before applying.
>
> Then **proved the guard bites**, which is the part that matters: run standalone against the
> un-migrated database it raised
> `ERROR: 218: subject_type CHECK not widened on: doc_plans_subject_type_check, doc_notes_subject_type_check`
> — precisely the half-made split contract of `bugs_closed/064`. A guard that has never been seen
> to fail is a comment.

**Post-apply verification, in a transaction that rolled back** (positive *and* negative, because
a positive alone cannot tell a widened CHECK from a dropped one):

| probe | result |
|---|---|
| `INSERT doc_plans … subject_type='experience-pattern'` | **ACCEPTED** |
| `INSERT doc_plans … subject_type='site'` | **REJECTED** — `violates check constraint "doc_plans_subject_type_check"` |
| `INSERT doc_notes … subject_type='experience-pattern'` | **ACCEPTED** |
| rows surviving the probe | **0** |

So 064's contract now agrees in both directions on the live database. `experience_patterns`,
`site_experiences`, `experience_invariants` all exist; two invariants seeded (6 and 2 sightings);
**zero patterns**, by design.

### The first council submission returned no verdict at all

`f4610451-6bff-45d0-8d18-6f25d26640cd` did **not** come back REVISE or REJECTED. It wedged:

```
current_step = review_editquality | status = FAILED
error = 'reaper: stale EXECUTING_STEP for >4h; step=review_editquality'
created 18:34Z → killed 22:36Z
```

**This is a third outcome I had not accounted for, and it is invisible if you only look for a
verdict.** Polling `current_step`/`status` the way the runbook says would show a run "in
progress" for four hours and then a `FAILED` row that carries no objections, no reviewers, and
nothing to act on. It is not a REVISE and must never be recorded as one.

It was not unique to me. Eight runs were reaped that day and **none on any other day in the
previous seven**:

```sql
SELECT date_trunc('day', created_at) AS day, count(*) FROM orchestration_states
WHERE error LIKE 'reaper: stale EXECUTING_STEP%' AND created_at > now() - interval '7 days'
GROUP BY 1 ORDER BY 1;                          -- 2026-07-26: 8. Every other day: 0.
```
Spread across 14:44→21:02 and six different steps (`council_decide`, `review_constitution`,
`review_reuse_agent`, `review_editquality` ×2, `review_debug_historian`, `extract_claims`), each
killed at exactly +4h.

`[UNDIAGNOSED]` — I am **not** asserting a cause. The window overlaps several chassis rolls that
evening (v1.0.1169→1170→1171), and an in-flight step whose response returns to a pod that no
longer owns it is a known shape here (`coordinator.go:271` discards non-owner responses), but
that is a **hypothesis I have not tested** and 8 wedges over 6 hours is more than the roll count.
Not filed: a one-day spike with no recurrence is not yet a bug, and `bugs_open/029`
(hung spawns) and `bugs_open/075` (orphaned ownership) are adjacent and owned. **The trigger to
file is recurrence** — if reaped runs appear on a second day, that query is the evidence.

**Resubmitted** as `bbdd2c5e-1b9d-4179-a31c-a8a5c3c3bf32`, with the submission saved *in the repo*
this time (`council_submission_p2a_substrate_and_criteria.json`) rather than a session scratchpad
— the first one was lost when the scratchpad went, and rebuilding it cost an hour of re-deriving
quotes. The rationale now states plainly that the change has already landed and the migration is
already applied, because that changes what is worth objecting to: **objections to the schema are
the valuable ones now, since a live table is far more expensive to change than unlanded code.**
Eight `grounded_in` quotes, each re-verified byte-exact with `grep -rF` before submitting.

---

## 2026-07-27b — the verdict, and the objection worth more than the approval

**P2a: APPROVED.** Corr `bbdd2c5e-1b9d-4179-a31c-a8a5c3c3bf32`, round 1,
`{"decision":"approved","reviewers":11,"abstained":5,"unreadable":0}` — **`unreadable: 0` checked
first**, per the standing trap: on the 16-seat gate a filtered seat counts as abstained, so an
approval with a non-zero `unreadable` is seats that could not read the submission, not seats that
agreed with it. Five advisory objections, none high-severity. Turnaround was ~10 minutes, not the
~30 the runbook budgets.

**The most useful objection was one I had already made about myself, and had answered badly.**
`bug_historian` and `guardian` both raised the deferral escape hatch: *"an entry can defer every
hard check and still reach 'approved' with criteria that assert nothing substantive"*. It was
risk #5 of my own submission. What matters is not that they found it — I had pointed at it — but
that my stated answer was **"a minimum-executable-checks rule at the APPROVAL step; it is not
built"**, and `bug_historian` named exactly why that is not an answer:

> *"Until that gate exists, an all-deferred pattern is indistinguishable from a real one at every
> status this schema can report."*

A rule a future step must remember is a schema defect wearing a documentation costume. So it is a
constraint now — migration **230**, `status = 'draft' OR executable_checks > 0`. The bad state is
unrepresentable rather than discouraged, and it holds no matter which path later writes the
status. **Candour in a risk list does not close a gap**; naming a risk and shipping it are the
same act unless the naming changes the code.

> **MISSTEP, and it cost two medium objections:** my sketch for the migration wrote the doc_notes
> half as a comment — `-- same pair for doc_notes` — instead of the DDL. `editquality` and
> `guardian` both objected that they could not confirm it had landed, and `guardian` said so
> exactly right: *"'in principle' is doing the work here because I cannot verify it."* The real
> file does carry it and I had probed it live before they ever saw the submission. That is not
> the point. **The submission was ABOUT two halves that must move together, and I elided one of
> them for brevity.** Reviewers cannot open the file; an abbreviated quote is a different claim,
> and an abbreviated *diff* is a different change. Never elide the half a change is about.

**Two objections I answered with a query rather than an argument**, because both were of the
asserted-absence shape and both were mine to prove:

- `reuse_agent` + `prior_art_librarian`: was there already a write-time criteria validator to
  extend? **No.** `grep -n "func .*[Vv]alidat\|func .*[Cc]riteria"` over the tool-acceptance path
  returns `evaluateStaticCriteria` (evaluates against fetched HTML), `loadCurrentCriteria`,
  `extractCriteriaFence` (both parse at read time) and `criteriaOrNull` (a null helper). Every one
  runs at **read** time. That is the failure class the new validator addresses, not a duplicate of
  it. Worth stating plainly: the seat was right to demand the check even though the answer was the
  one I expected — I had asserted the absence without running it.
- `tooling_provenance`: does splitting `experience` from `experience-pattern` fragment
  travelling-doc history? **No.**
  ```
  subject_type | count | distinct subjects
  experience   |    59 |        1
  ```
  59 rows, **one** subject_key — the vonc EXPERIENCE_PLAN superseded 59 times. `experience` is a
  per-site level-4 plan; `experience-pattern` is a library-level entry. Different subjects, no
  shared history to fragment.

**One objection I am NOT acting on, with the reason.** `debug_historian` asked for migration 218
to be renamed off its number collision with `218_evidence_facts_for_043_sites.sql` *"before this
becomes load-bearing"*. It already is: the ledger keys on **filename**, and 218 is recorded as
applied. Renaming it now would make it look pending again and invite a replay — the precise trap
(`bugs_open/007`) the record-only path exists to avoid. The shop already carries duplicates at
217, 219, 220, 222, 226 and 227 and resolves by full name. Recorded here so the decision is
visible rather than silently ignored.

**No `Council-Reviewed:` trailer anywhere.** The verdict approved P2a, whose commit (`2f220f261`)
predates it and cannot be amended; and it did not review P2b. A trailer on either would be a
permanent false claim. The 098 report will list both as unreviewed, which is accurate.

## P2b — the write path

`write_experience_pattern` (`36bb6c992`) is the only door into the register. Three rules, each
closing a door rather than asking anyone to remember something: **status is not writable**
(supplying one is an error, not a silent no-op — silently dropping it lets a caller believe it
approved something); **a contract change demotes** an approved entry back to draft, on canonical-
JSON comparison so key order cannot cause a phantom demotion, with cosmetic fields excluded by an
explicit list; and **validation refuses the write**, reporting every problem at once because each
refusal round trip is an LLM call. `executable_checks`/`deferred_checks` are refused from the
caller for the same reason as status: an entry that scores itself walks through the constraint
that reads the score.

Falsification before commit, as standing practice: adding `site-journey` to `experiencePatternKinds`
fails the new lockstep test naming `218_experience_register_substrate.sql` — which also proves the
regex matched the intended file rather than passing vacuously. Migration 230's guard proves its own
constraint **in both directions inside the transaction** before committing itself: it refuses
`status='approved', executable_checks=0`, and it accepts `executable_checks=1`.

Submitted as corr **`2e71f640-06b9-48c8-a674-a121f74be22c`**.

**Still true and worth repeating:** the Go for P2b is committed but **not in a rolled image**, so
the write path is inert in production. Migration 230 is applied, and I argue it is safe ahead of
the image (defaults on both columns, tightening-only constraint, constrained states unreachable
because nothing can write a non-draft status yet). That argument is mine and is in the submission's
risk list precisely so a reviewer can test it rather than accept it.

---

## 2026-07-27c — the roll, the writer pipeline, and the shape I invented

**v1.0.1175 is live** (pod `agent-chassis-566bf56b78-jtjnj`, started 18:00:40Z) and carries the
write path. Pod-grep with controls: `write_experience_pattern` → **16**; my derived-field refusal
string `is derived from validation` → **1**; `docResolveSubject` → 2 (positive control);
`experience-nonsense-xyz` → 0 (negative). Note `experience_patterns_approved_needs_executable_check`
greps **0** and that is correct — a constraint name exists only in SQL. Reading that as a failed
deploy would be the same vacuous-marker trap in reverse.

**P2b: APPROVED** (corr `2e71f640-06b9-48c8-a674-a121f74be22c`, 10 reviewers, `unreadable: 0`,
5 advisory objections). Three seats — editquality, guardian, debug_historian — converged
independently on my own risk #1: `experiencePatternContractFields` decides whether an approved
entry loses its approval, and it was a hand-maintained literal with no structural check.

The fix worth recording is the *shape* of the fix. A derived list is impossible — which fields
are clause-bearing is a judgement, not a fact about the schema. So the guard forces the judgement
to be **made**: every column of `experience_patterns` must be classified as contract, selection,
cosmetic or system, and a column named by none of them fails the build. **A silent omission
becomes a required decision**, which is the most a mechanical check can do for a question that is
not mechanical. Proved by induced fault; the same parser also closes bug_historian's separate
objection (a jsonb column absent from the marshalling list is silently never written).

Two absence claims the reviewers refused to take on my word, both of which I had asserted without
running:
- *Does any other writer touch `experience_patterns`?* No — `grep -rn` over `platform/ internal/
  cmd/` returns 2 hits, both in my own action. So migration 230's CHECK could break nothing.
- *Does any workflow call `write_experience_pattern`?*
  `SELECT count(*) FROM agent_definitions WHERE default_config::text LIKE '%write_experience_pattern%'`
  → **0**. Genuinely inert, as claimed.

**Migration 238 seeds `experience-register-writer`** — the owning pipeline guardian asked to see
named rather than smuggled into an existing agent. Its one design decision: the entry write and
its travelling doc are **one workflow**, closing risk #3 (an entry with no provenance) by shape
rather than by a rule a caller must remember. The guard fails if a later edit points
`write_pattern` straight at `complete`.

### Then I tried to load a real entry, and the build fell over

> **CORRECTED 2026-07-27 — the `contract` shape in migration 218, in
> `write_experience_pattern_action.go`, in the P2b council submission and in every test was
> INVENTED.** It expected an object carrying a `triggers` array with `when`/`then`. **All nine
> harvested entries use an ARRAY of clauses keyed `control_role` / `primitive` / `outcome`.** The
> live write path would have refused **9 of 9** of the only entries that exist.
>
> Two more from the same look: **CC-006** has a deliberately EMPTY contract — nothing a visitor
> does drives a count-up stat band, its behaviour is in `automatic_triggers` and
> `honesty_clauses` — so "contract must be non-empty" refused a legitimate entry; and
> **`honesty_clauses` / `latency_envelope`** are fields the harvest produces for which the table
> had no columns, so they would have been silently dropped (migration **239**).
>
> **What caught it:** trying to load the real files. Nothing else could have — seven behaviour
> tests, a migration lockstep, and induced-fault probes in both directions were all green,
> because the fixture was an inline literal I wrote to match the code and *called*
> `validHarvestedEntry()`. The name asserted a provenance the value did not have. The council
> approved it too, and could not have caught it: reviewers see the submission, not the repo.
>
> **Why this one stings:** this workstream's founding finding, written by me the day before in
> this very directory, is that harvesting bottom-up catches shapes invented top-down. I built the
> validator from the harvest **NOTES** rather than the harvest **ENTRIES**. Writing a lesson down
> is not applying it. Full entry in `WRONG_CALLS.md`.

**The fix that generalises is not the corrected shape — it is that the fixtures are now the real
files.** `validHarvestedEntry()` reads `CC-001` off disk; two new tests load all nine, one
asserting the validator accepts every one and the other that every field they carry has a column.
There is no longer a second copy of the truth to drift from. Proved by induced fault: restoring
the non-empty-contract rule fails naming `CC-006_count-up-stat-band.json` exactly.

**Live state:** migrations 218, 230, 238, 239 applied and in the ledger. `experience_patterns`
still **0 rows** — and it must stay that way for now, because **v1.0.1175 carries the INVENTED
shape**. The register cannot be populated until the next roll ships `799c0c97e`. That is the
whole remaining gap between "built" and "real".

---

## 2026-07-28 — the register is POPULATED, and the first live run refused most of it

**v1.0.1180** (pod `agent-chassis-5987b8749b-4p4bl`, started 2026-07-27T22:06Z) carries the
corrected contract shape. The trigger's two-direction precheck passed — `new=1, old=0` — and
that same precheck had been **refusing** against v1.0.1175 minutes earlier, which is the
cleanest confirmation available that the invented shape really would have refused everything.

**All nine entries are in.** 9 rows, 9 travelling docs, **0 non-draft**, 29 executable checks,
23 deferrals. Every row written through `write_experience_pattern` — none by raw INSERT — so the
register's contents are, as designed, exactly the set its own contract accepted.

### The first live run refused 6 of 9, and every refusal was correct

This is the validator doing the job it was built for, on its first contact with real data. In
order found:

1. **`binding "X" is declared but never used`** — five entries. Reading them, **every unused
   binding named a part of the experience the entry's own clauses describe and our checkers
   cannot assert**: `empty_state_selector` (no empty-region assertion), `item_openable_selector`
   / `item_static_selector` / `countup_attribute` / `prev_control` / `item_template` /
   `marker_count` (no attribute assertion), `tool_page` (no navigation), `timer_selector` /
   `progress_pct_selector` / `final_input` / `final_submit` / `outcome_reasons_selector` (no
   waits — 300 ms vs a measured 8–23 s), `detail_close_selector` (no ordering).
   **So an unused binding is the SHADOW of a deferred check.** Blanket tolerance would hide a
   real mismatch; a hard error would delete the record of what the clause needs the day the
   harness can run it. It is now a **deferral** on the same terms as a deferred check —
   declared, carried, reported, never a pass — and `unused_ok` now **requires a reason**. It
   previously accepted a bare `true` while its own error message promised a reason; an escape
   hatch that does not demand its justification is a slower way of ignoring the rule.

2. **A genuine defect: MJ-002 used `{{binding.sample_turn_text}}` and never declared it.**
   Unbindable, so the interaction would have typed the literal placeholder into the box — the
   `-EDIT` failure mode in different clothes: it runs, it looks green, it asserts nothing.

3. **`min` is read by NOBODY.** Verified in the checkers' source, not from my own table: Tier 2
   does `anchorPresent`, Tier 4 does `n > 0`, and **neither check struct has a `Min` field**. So
   `selector_count` is a misnomer — it asserts existence, never a threshold. Five checks carried
   `min`; three were `min: 1` (which *is* existence, so the inert field is dropped) and two were
   `min: 2` — including **`at_least_two_cards`, whose entire clause is that the carousel's arrows
   exist only when there are two or more cards.** A check named `at_least_two_cards` that asserts
   "at least one" is precisely what this validator exists to catch.

4. **MINE, again the hand-maintained-list class.** The action passed an *enumerated* list of
   documents to scan for placeholders (`contract`, `states`, `data_contract`, `degraded_states`,
   `entry_points`) and omitted `section_types` and `destination_roles` — where the harvested
   entries put theirs. Bindings were reported unused when they were used all along. Fixed the
   same way as the column lists: **pass every field except the schema and the template.** There
   is no reason to keep a list of documents-that-might-contain-placeholders.

5. **The entries' own design error, which the refusal exposed: `section_types` held a
   PLACEHOLDER**, in all nine. It is a selection axis with a GIN index — you cannot match
   `{{binding.section_type}}` against a real page's section type, so the axis would index a
   literal placeholder and **the pattern could never be selected at all**. Replaced with the real
   `section_type` of the component each entry was harvested from, read from `content_components`:
   `provocations-archive-list`, `provocation-card` + `lobby-grid`, `hero-carousel`,
   `image-hover-cards`, `insight-carousel`, `stat-band`, `people-feature`, `gauntlet-interface`.
   A **starting point, not a claim** about every section the pattern suits — the axis widens on a
   new sighting, the same rule the invariants' `sightings` follow.

6. **CC-002 promised a role it never named.** Its outcome prose says the destination "is live and
   of the promised role", but the clause carried no `destination_role` — unlike the four
   analogous entries, which all do. Under-specified against its own words; naming it makes the
   promise checkable.

### The payoff: the harness gap is now MEASURED, not asserted

Since 07-26 this workstream has carried a note that "four register-worthy clauses are
unassertable by our own criteria". That was a hand-count from reading. The register now answers
it from data — 23 deferrals, grouped by the capability that blocks them:

| blocked on | clauses | entries |
|---|---|---|
| **attribute assertion (has/lacks an attribute)** | **7** | **7 of 9** |
| waits + retries (the 300 ms vs 8–23 s gap) | 3 | 2 |
| cross-page status (a url on `page_status_ok`) | 2 | 2 |
| focus step + visibility expectation | 2 | 1 |
| zero-count / scoped negation | 2 | 2 |
| count threshold (`min`) | 2 | 2 |
| navigation as a journey step | 1 | 1 |
| ordering / conditional checks | 1 | 1 |
| inducing a degraded state | 1 | 1 |
| empty-region assertion | 1 | 1 |

**Attribute assertion is the single biggest win available, and it is not close**: 7 clauses
across 7 of the 9 entries. It is also the *anti-dead-control* clause — `no-inert-control`, the
invariant found in six independent implementations, the reason this workstream exists. We cannot
currently check the rule we built the register to enforce. That is a prioritised harness roadmap
derived from evidence, and it is the first thing the register has produced that we did not
already know.

`[LIMITATION]` The **binding-level** deferrals are not in that table. The tightened P4 that
records them is committed but **not deployed** — v1.0.1180 silences a justified unused binding
without recording it. So the true figure is 23 **plus** the fifteen marked bindings; the table
above undercounts, and will correct itself at the next roll with no action needed.

---

## 2026-07-28b — the consumer, and the corrected harness figures

**The register now checks its own promises.** `verify_site_experience` runs a bound fork's
criteria against the page actually being served and moves the fork — and, on a first green run of
an approved entry, the entry — along their lifecycles.

**The door that matters is the vacuous pass.** A criteria document can produce ZERO executed
checks: everything deferred, or everything skipped (Tier 2 could not anchor the selector, or the
id ends in `-EDIT`). Such a run has **no failures**. Read as "no failures ⇒ pass", it verifies a
fork that asserted nothing — the register committing the exact defect it exists to end. So the
rule is **at least one PASSED and zero FAILED**, never an absence of failures; zero executed is
`INCONCLUSIVE` and changes no status. The decision is one testable function
(`experienceVerdict`) because it is the rule most likely to be "simplified" later by someone
reading `len(failed) == 0` as sufficient. Proved by induced fault — making that read succeed fails
four cases, naming the empty and all-skipped documents.

**Reuse rather than a copy:** `discovery_checks.EvaluateStaticCriteriaJSON` is a new exported door
onto the *same* Tier 2 evaluator tool acceptance uses. Reimplementing the anchor rule would create
a second definition of "this check passed" — and a divergent *judgement* is worse than a divergent
value, because it produces confident disagreement rather than an error.

> **CORRECTED — the deferral figures published this morning were an UNDERCOUNT, as flagged, and
> the correction changes the conclusion's strength.** v1.0.1189 carries the tightened P4, so
> binding-level deferrals are now recorded. Re-seeded all nine; the register holds **38 deferrals,
> not 23** — 23 check-level plus exactly the **15** marked bindings predicted. Re-ranked:
>
> | blocked on | clauses | entries |
> |---|---|---|
> | **attribute assertion** | **13** | **9 of 9** |
> | waits + retries | 8 | 2 |
> | cross-page status | 3 | 3 |
> | empty-region assertion | 3 | 2 |
> | zero-count / scoped negation | 2 | 2 |
> | count threshold | 2 | 2 |
> | focus step + visibility | 2 | 1 |
> | ordering / conditional | 2 | 2 |
> | navigation as a step · degraded state · uncategorised | 3 | 3 |
>
> This morning's table said attribute assertion blocked 7 clauses across 7 of 9. It is **13 across
> all nine** — a third of every deferral in the register, and there is no entry it does not
> affect. The ranking did not change; its margin roughly doubled. **The undercount made the right
> answer look like a closer call than it is**, which is the practical cost of publishing a figure
> with a known gap in it — worth remembering that I did label it `[UNDERCOUNT]`, and that the
> label is what made this a scheduled correction rather than a discovery.

**Branch note for whoever reads `git log`:** another session moved the working tree to
`087_towards_multiple_domains` mid-task, so the consumer commit landed there rather than on
`086_experience_loop`. Verified before continuing: `087` branched from `086` at `b82b3d8b4`, which
is *after* every commit in this workstream, so the whole chain is reachable and nothing was lost.
Forward-only holds — no reset was needed and none was made.

---

## 2026-07-28c — the first run against a real page, and what it corrected

Ran the whole chain locally against **vonc.com/provocations/index.html** — the page CC-001 was
harvested from — before the actions are deployed: read the criteria **as stored in the register**,
substitute real selectors read off the served page, fetch, evaluate with the same Tier 2 evaluator
tool acceptance uses. It found four things, three of them mine.

**The honest final result**, after the corrections below:

```
HELD BACK (4, each with its reason)
  feed_loads                  deferred — asset_loads only matches the path as TEXT in the page HTML
  rows_rendered               tier 4  — rows are cloned client-side; Tier 2 sees only the template
  no_dead_row_hrefs           deferred — no attribute assertion exists
  template_row_not_a_control  deferred — same gap
EXECUTED (1)
  list_exists                 PASS — anchor .provocations-archive__list present
VERDICT: verified
```

**Of this entry's five clauses, exactly one is checkable today against that page.** It passes.
That is a thin verification and the result says so, which is the point: `verified` now arrives
with its coverage attached rather than as a bare word.

### 1. Deferral did not bind the consumer — the sharpest one

`feed_loads` is recorded **deferred**, and the consumer **ran it anyway and reported a FAILURE**.
The first verdict of the register's life was `broken`, on a page that is fine.

> **A correct page called broken by a check we had already written down as impossible.** This
> workstream's standing rule is *fix the harness, never the page* — and here the harness was the
> one lying. Deferral that binds only the validator and not the runner is a comment.

Deferred checks are now held back **with their reason**, so the report says what is not being
checked and why.

### 2. Tier was ignored, in the dangerous direction

Everything went to the Tier 2 evaluator, including checks declaring `tier: 4`. A `selector_count`
evaluated statically is **vacuously green** — it confirms the container's anchor and never counts
anything. So a `verified` could have rested on an assertion that cannot fail. Now split before
evaluation, on three independent signals: the register says deferred, the author says `tier: 4`,
or the platform says the type only exists at Tier 4.

### 3. `dry_run` was a dead config key in my own action

Declared twice in the input spec, read nowhere — exactly the failure the standing note *"grep the
config key before calling it a win"* describes, committed by me, in a file I wrote yesterday.
Implemented now, and it earns its place rather than being a convenience: **per-experience approval
is designed and NOT BUILT**, so every entry is `draft`, every fork is `proposed`, and a proposed
fork cannot be verified. Without a report-only mode the entire path would be unexercisable — which
is how a mechanism rots before it is ever used. It writes **nothing**, not even
`last_check_result`: a dry run that leaves a trace is one somebody later reads as a real result.

### 4. The entry was wrong too, in two ways worth keeping

- `feed_loads` cannot succeed for this component. Verified live: `/data/provocations.json` serves
  **200**, and the page references **no json at all** — the loader is `/assets/js/snippets.js`.
  `asset_loads` matches the path as *text in the page HTML*, so for any component whose loader is
  bundled externally it can only ever fail. Deferred, with that reason.
- `rows_rendered` said nothing about rows. The rows are cloned client-side, so at Tier 2
  `selector_count` collapses to the container's anchor — precisely what `list_exists` already
  asserts. **Two checks, one fact, and the one named for rows was the empty one.** Marked `tier: 4`.

### The general finding: IMPLEMENTED is not SATISFIABLE

I have been describing the write-time validator as catching "checks the platform cannot execute".
That is too strong, and this run is the counter-example. The capability table answers *does this
check TYPE exist* — a real question, and `asset_loads` passes it. Whether a check can **succeed**
depends on how the page is built, and **write-time validation sees the template, never the page.**

So there are two different unassertables, and only the first is caught before a run:

| | caught by | example |
|---|---|---|
| the type does not exist | the validator, at write time | `attribute_absent` |
| the type exists but cannot match here | **only a real run** | `asset_loads` against a bundled loader |

That is the argument for the dry run being a first-class mode rather than a debugging aid, and it
is a correction to the claim in `SUMMARY_2026-07-28` that write-time validation covers this class.

`[LIMITATION, stated]` This was a **local** run — the bind and verify actions are committed and
registered but **not deployed and not called by any workflow**, so nothing has yet gone through
the real orchestration path. What is proven is the machinery and the entry, not the wiring.
