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
