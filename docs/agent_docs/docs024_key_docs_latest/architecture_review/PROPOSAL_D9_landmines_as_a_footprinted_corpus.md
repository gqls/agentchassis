# PROPOSAL D9 — landmines as a footprinted corpus, not prose in an auto-loaded file

**Status: DRAFT — proposal only. Nothing built, no config or code changed.**

**Provenance.** Written 2026-07-27 by session *"bugfix 61"* (the `bugs_closed/061`
med-scrape thread) at the owner's direction: *"with regard the landmines, we have the
council and the missteps list and we could also add landmines, see the architecture
council thread."* Handed to this thread rather than applied, because **D8 is this
thread's finding and this extends it**, and because `DECISIONS_open_for_owner_…` was
being actively edited (14:00 today) while this was written. Fold it in as D9, reject it,
or bin it — it is yours. A one-line pointer has been inserted at the end of §5 of that
document; nothing else there was touched.

---

## 1. The finding: D8's defect has a second symptom, and nobody had connected them

**D8 established** that the council structurally cannot read the misstep record —
`code_checks` indexes Go only (4,535 symbols, **0 markdown**; `WRONG_CALLS.md` has 0 rows
in `code_symbols`), and the SQL `checks` reach ten tables, none a document store. So
`WRONG_CALLS.md`, `/bugs_open/`, `/bugs_closed/` and every working doc are invisible to
every seat.

**The second symptom, measured 2026-07-26/27:** the auto-memory index grows without
bound for the *same underlying reason*. `MEMORY.md` auto-loads in full into every thread,
so a landmine that must be seen has to live there — because prose that cannot be queried
can only be delivered by broadcasting it. From the memory repo's own git history:

| time (07-26) | `MEMORY.md` bytes | |
|---|---|---|
| 22:11 | 20,613 | |
| 22:15 | 17,621 | ← compaction #1 (~3 KB cut) |
| 23:01 | 20,575 | **re-inflated past its pre-compaction size in 46 minutes** |
| 23:05 | 17,573 | ← compaction #2, by a different session |
| 07-27 11:51 | 17,978 | creeping again |

Two compactions inside one hour, by two sessions, the first fully undone before the
second. Composition today: **76 entries, mean 223 chars, only 2 over 400** — so this is
*uniform* bloat, and trimming the worst offenders provably cannot fix it. `MEMORY_closed.md`,
the archive created by the 2026-07-24 lifecycle split, is now **24 KB — larger than the
index it drains**, and near the same ~24.4 KB read cap.

**One defect, stated once:** *hard-won knowledge lives in prose; prose cannot be queried;
what cannot be queried must be broadcast; broadcast has no upper bound.* The memory index
is the broadcast channel, and it is now the binding constraint on both problems.

## 2. There is a ladder, and its middle rung has no mechanism

- **`WRONG_CALLS.md`** — raw incidents, append-only, human-read. One row is an anecdote;
  the file's own thesis is that the value is *"the tally of the cheap check that would
  have caught it — because a check that keeps appearing is one worth automating."*
- **A landmine** — the distilled check, attached to the thing it guards, delivered when
  that thing is touched. **No mechanism exists for this rung.** It is improvised as prose
  in `MEMORY.md`, and that improvisation *is* the bloat in §1.
- **A lint or a council seat** — full automation, once the tally justifies it. Exactly how
  `check_append_only_docs` earned its place in `scripts/pattern-check.py`.

`WRONG_CALLS.md` already states the promotion rule for rung 1 → rung 3. Naming rung 2
gives landmines a home that is not an auto-loaded file, and — the point for this thread —
gives the council something it can actually read.

## 3. Reuse before building: the store already exists and is already used this way

`public.doc_notes` — **370 rows today**:

| column | why it fits |
|---|---|
| `subject_type` + `subject_key` | indexed together (`idx_doc_notes_subject`). `subject_key` is the natural **footprint**: the path, table, or symbol the landmine guards |
| `categories` jsonb | GIN-indexed (`idx_doc_notes_categories`) — cross-cutting tags |
| `body` | the landmine text |
| `source_agent`, `created_by`, `created_at` | provenance and staleness, which `MEMORY.md` lines do not carry |

**And the categories in use are already landmines in all but name:**
`do-not-lock-derived`, `envelope-vs-payload`, `derived-fields`, `bug-020`, `bugfix-056`.
This is not a new idea being proposed — it is an existing practice that never got a name,
a convention, or a delivery path.

**Relevance-gating is likewise already proven here:** the gate fires two seats always and
the rest only when the submission's edited paths match their footprint (CLAUDE.md; a real
submission drew 10 of 16). That is precisely the delivery model a landmine needs — a
session touching `vet_med_price_scrape_action.go` should receive that file's three
landmines, not all seventy-six.

**Distinct from `bugs_open/108`** — do not merge them. 108 fixes the **code** index
(`code_symbols`: stale-but-FRESH, and no function bodies). D9 concerns the **prose**
corpus, which `code_checks` will never cover: D8 established it is Go-only *by design*,
not by defect. 108 makes the code index tell the truth; D9 gives the seats a second
corpus that has never existed. They are complements.

## 4. Proposed decisions

**D9(a) — Name the ladder** (incident → landmine → automation) in `WRONG_CALLS.md`'s
header and in `016b` §9, so the middle rung stops being improvised. Costs nothing;
prevents the next person doing what we have all been doing.

**D9(b) — Landmines become `doc_notes` rows**, `subject_type='landmine'`, `subject_key`
= the guarded path/table/symbol, `categories` for cross-cutting tags. No migration
needed; the table, both indexes and the practice already exist.

**D9(c) — Delivery to the council: add `doc_notes` to the schema hint.** This is D8a's
original proposal, which **D8a′ superseded only for the *minutes* case** (`council_report`
in `diagnosis_artifacts`, applied and live today). The landmine case is still open, and
it is a config change — live immediately, no image, no roll.

**D9(d) — Delivery to a *session*: OPEN, and the weak link. Do not let (a)–(c) be
approved as though this were solved.** Nothing currently queries `doc_notes` at session
start. Three candidates, none costed:
  1. a hook that queries landmines for paths in the session's opening diff;
  2. a standing discipline ("query landmines for what you are about to touch");
  3. piggyback the existing memory-recall mechanism, which surfaces topic files by
     `description:` frontmatter — but `doc_notes` rows are not memory files, so this may
     not reach them at all.

  **(2) is the tempting one and I recommend against it standing alone.** "Detail does not
  live in the index" is *already the written rule* in `memory-index-how-it-works.md`, and
  §1's measurements are what a discipline achieves against a standing incentive. Authors
  put landmines in the auto-loaded file because it is the only thing guaranteed to be read;
  that incentive survives any amount of exhortation.

## 5. Risks and what I could not verify

- **Staging risk, the serious one.** Moving landmines out of `MEMORY.md` before (d) is
  solved would *remove* protection sessions have today in exchange for protection they
  might get. Landmines in the index have genuinely prevented cross-workstream mistakes.
  **Sequence must be: build delivery, run both in parallel, measure, then drain the index
  — never drain first.**
- **[UNVERIFIED] the council's footprint map location.** I could not find it under
  `council_decide.config` (keys there are `error_step`, `max_rounds`, `review_fields`,
  `hard_veto_from`, `fix_correlation_id`), so it presumably lives per-`review_field` or in
  the mirrored `fix-proposer` row that `099_SYNC_gate_roster.py` copies. §3's claim that
  the gating mechanism is reusable rests on CLAUDE.md's description of the behaviour, not
  on my having read the config.
- **[UNVERIFIED] whether `doc_notes`' existing 370 rows need partitioning or curation**
  before they are exposed to seats, and whether `subject_key` granularity (file? function?
  table?) actually matches how landmines are phrased.
- **[UNMEASURED] migration cost.** The 76 index entries are prose; converting them to
  footprinted rows is manual curation, not a script. Nobody has sized it.
- **The same test D8a′ set for itself applies here:** the honest measure is not that the
  rows exist, but whether a seat's `checks` array ever cites them and whether a session's
  behaviour changes. Present text is not use.
