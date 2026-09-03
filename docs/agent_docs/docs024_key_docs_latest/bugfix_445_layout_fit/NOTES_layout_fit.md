# NOTES — layout fit / bugs_open/445

Running record, append-only, newest at the bottom. Missteps are the point.

## 2026-09-03 (a) — picking it up

`who-owns.py` said OWNED/recently-active (designblog_couk 32 commits/14d). Read as
filed-and-parked; **asked the filing lane directly and they confirmed** ("stand up — 445 is
filed-and-parked, not in-flight"). who-owns reads commits, so it cannot see a session mid-fix —
the tree was also clean of every file I would touch.

## 2026-09-03 (b) — the measurement, and the control that makes it worth anything

Extracted every `resolved_composition.reasoning` (33 rows) and parsed the score out of the prose.
Then re-implemented the scorer in Python and checked it against those recorded scores:
**29 of 30 exact.** The one miss (gamedesign.uk) had its classification refreshed after
composition — a known event, not a replication fault. Everything downstream rests on that
agreement; without it the simulation would be an assertion.

Scores turned out nearly **discrete**: 8 sites at exactly 3.05 (tags 2.30), 9 at exactly 8.31
(tags 6.91). 2.3026 = `log(1+18/2)` — one tag present in exactly two layouts. Identical scores to
2dp across structurally different sites is not similarity; it is a single shared tag.

## 2026-09-03 (c) — MISSTEP: an all-time count that only queried the live table

Published **"exactly one `needs_new_layout_candidate` fleet-wide, ever"** to four lanes. It is
**two** — `site_work_items_archive` (33,350 rows) held a second. Two lanes had already written it
into their notes and a fleet memory on my word. Corrections sent to all three who carried it.
Full entry in `WRONG_CALLS.md`; the check is in RUNBOOK r3.

⚠ **The part that generalises, and it is not mine.** The `theme kits` lane had "independently
verified" the wrong figure by making the *same* omission — we each queried the table the question
named. Their formulation: *"independent corroboration is not protection when both parties inherit
the same framing; when a second check agrees, ask what BOTH checks assumed, not whether they
match."* They also had the rolling-window trap in their own auto-loaded memory index, with a note
saying it had already failed to fire once.

## 2026-09-03 (d) — MISSTEP: fired 090 blind

Fired `090` on a code-only symptom with no `SEED_SCOPE`. Failed at `assemble_bundle` after ~6
minutes and burned the item's only attempt (`max_attempts=1`). **The script had told me** —
`WARNING: nothing to key coverage on … dispatching blind` — and I had read as far as the
correlation id and stopped. Re-fired with `SEED_SCOPE`; completed.

## 2026-09-03 (e) — a peer's challenge, refuted, which then produced the real cause

`portfolio_positioning` challenged 445 §2 with a 12-site census: 8 sites carry the literal string
`magazine-grid` in `industry_tags`. Checked all three scoring paths — own tags, category, and the
**description** path §2 had not checked (magazine-grid's description opens *"Publication layout
with featured article…"*, so neither `" magazine-grid "` nor `" magazine grid "` is present).
Zero contribution. They accepted the refutation and corrected it in three places.

**And then they found what I had missed, which is the mechanical cause of my 87% figure.**
`layout_taxonomy` was fetched by `read_layout_taxonomy` and **dropped at the template boundary**
because `classify_and_extract`'s `input_fields` allow-list did not name it. **I verified it
myself** at `llm_call_log.prompt_rendered` before acting: the model was shown a literal `null`
where the tag list should be, and `<no value>` for the layout count, then told to coin a tag if
nothing fits. So my "87% of emitted terms match nothing" is not a vocabulary mismatch — it is the
arithmetic of an empty list.

**Two broken links in one loop.** Theirs produced the coined vocabulary; mine is why nobody
noticed, because 87% of tags vanishing is exactly what the silent signal existed to report.

## 2026-09-03 (f) — building against a tree someone else had broken

`go build ./platform/...` failed in `tool_acceptance_actions.go`, a file I never touched and which
was **dirty** in another session's tree. Used `verify-head-builds.sh --with` throughout (RUNBOOK
r6). It also surfaced `TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens` failing on **clean
HEAD** with zero changes of mine — pre-existing, someone else's; re-ran with no `--with` to prove
it rather than assume it.

## 2026-09-03 (g) — MISSTEP inside a test I wrote: a fixture that could not reach its own arm

`TestSchemeGapArmStillFires` failed. My fixture assumption was wrong and the reason is a real
finding: **the same-scheme bonus (0.50) alone lifts any same-scheme layout to `total > 0`**, so
`lmFirstEligible` takes it and the scheme arm is unreachable while any same-scheme layout exists.
That is the live mechanism behind `soft-editorial` scoring 0.50 as a permanent runner-up on 27 of
33 sites while matching none of their tags. Gave the test its own light-only fixture and wrote the
reason into the test.

## 2026-09-03 (h) — mutation results, including one prediction of mine that was wrong

Ran all three mutations rather than asserting them:
- (i) `lmMinTagCoverage = 0` → killed the two weak-fit tests only. Also printed
  `TagCoverage = 0.072` for the designblog fixture, reproducing that site's live 7% — an
  independent check that the Go formula matches the Python one validated against the fleet.
- (ii) denominator → `tagScore` → **killed ONE test, not the two I predicted.** The zero-overlap
  case has `tagScore == 0`, so the mutated expression is never evaluated. **A mutation can miss a
  test by never reaching it** — the surviving test is not evidence the denominator is unimportant
  there. Corrected in the test header in place rather than tidied away.
- (iii) predicate reverted to the old two arms → killed the two weak-fit cases; fallback and
  scheme cases still passed, which is what proves the change ADDED an arm rather than replacing one.

## 2026-09-03 (i) — shipped

Phase 1 `76db94fc7` (committed ahead of an announced chassis build; inert until it rolls).
Council `Council-Submitted: 34d57f60`. Phase 0 migration 735 applied and **verified at the live
row with controls**, including confirming 734's `layout_taxonomy` wiring survived my edit.

Deviation from the approved plan, deliberate: the `layoutmatch` package extraction was to land
first; it is deferred to Phase 2/3 because a package move under a build deadline on a shared tree
risks breaking HEAD for every other session.

**Simulated the archetype before recommending it, and it partly disconfirms 445's own fix
candidate 1** — four of seven sites improve, but designblog (the site that started this) and
apis.uk still win on a single tag at 6-8%. Recorded in 445 §8g so the bug cannot be closed on
"archetype drawn, problem solved".

## 2026-09-03 (j) — the archetype: tags by simulation, seeded behind a guard

**Grammar read first.** Dumped `magazine-grid` and `tool-portal-light` `css_template`s; the
renderer contract is in tool-portal-light's header (no `--section-*` defaults; five surface
classes must be surface-coloured; `var(--section-*, var(--color-*))`). Components mostly carry
inline `<style>` — `tool-list`, `guide-list`, `hero`, `article-body`, `header-with-categories`,
`call-to-action`, `features` do; **`faq`, `featured-content`, `category-listing` do NOT**, so
the layout frames those three fully. The sibling layouts style `.tool-card`/`.tools-grid` but
the live `tool-list` component emits `tl-card`/`tl-grid` — so the new template covers BOTH
vocabularies rather than guessing which one renders.

**Four candidate tag sets, simulated** (`scratchpad/simulate2.py`, scorer validated 29/30):
A (445's title words) rescued 4 of 7 and left designblog at 6%. **B** (A + the form words the
sites already emit: `editorial-guides, long-form-content, research-publication,
content-platform, guides`) rescued 6 of 7 — designblog 7→16%, apis 9→19% — and pulled in
gamedesign.uk and farmerinsurance.uk. C/D added nothing over B live. → **B.**

**Judging the pull-ins rather than waving them through.** gamedesign.uk: a design-practice
publication with `interactive-illustration`, `long-form-content` — the shape, not a steal.
farmerinsurance.uk: self-described `insurance-guidance, editorial-guides, content-platform,
interactive-calculators`, sitting on `industry-hub` at `tags 0.00` — moving from a bonus-only
pick to a two-tag match is what DES-086 exists to surface; its two industry-hub siblings
(garden-tools, vetcomparison) do NOT move. **oufe.com not rescued** under any candidate: its own
tags lead `interactive-platform`, so tool-portal wins it. Recorded rather than tuned away.

**Proxy for the 17 unbuilt remakes — `[ASSUMED]` tags, and the twins.** 7 of 14 proxies land on
the new layout; both Christmas twins land on it. That is correct: layout is about FORM, and
twin differentiation is `RFC_037`'s job (positioning), not the layout's. Written into the
register so nobody reads "both twins on one layout" as a defect.

**Live reachability, the guard's number:** 14 current classification specs emit at least one
candidate-B tag (raw). Per tag: `editorial` 5, `content-hub` 4, `editorial-guides` 3,
`research-publication` 2, `content-platform` 2, `interactive-tools` 2, `long-form-content` 2,
`editorial-publication` 1, **`long-form` 0, `guides` 0** — those two are in the set for the
vocabulary 735 now steers toward, not for today.

**Migration 736 applied** — `content-hub-tools`, category editorial, scheme light, 9 tags,
30,303 chars CSS. Verified at the live row with controls: 19 active layouts; the seven new
terms present in the DISTINCT list `read_layout_taxonomy` hands the classifier; **the seven
cluster sites verified UNMOVED**; 0 sites composed onto it. **Re-ran the simulation against the
real 19-layout dump: identical to the hypothetical** — 8 sites would move, oufe.com would not.

**Phase 4 in its minimal form.** The migration's own DO block refuses to seed a layout fewer
than 2 current specs can reach, and the verify block re-checks reachability against the seeded
row's own `industry_tags` (not a copy of the list). The reusable `assert_layout_reachable()` +
`pattern-check.py` rule is still owed.

**Not done, deliberately:** no site re-composed (owner: fix forward only); `layoutmatch`
extraction; `internal/cronchecks`; `cmd/layout-fit-check`.

## 2026-09-03 (k) — the Go fix MISSED the chassis build, and a pattern-check advisory answered

**`76db94fc7` is not in `v1.0.1358`.** The pods started 12:06:47Z; my commit landed later (I
was running the mutation proofs when the build was cut — the right trade: an unproven predicate
change fleet-wide is worse than one roll later). `portfolio_positioning` measured it first with a
NUL-split probe of `/proc/1/exe`, both controls; I confirmed: `enforceListingItemSources` 2,
`layout_match_score` 0, `weak_tag_fit` 0, absent control 0. **In HEAD, in no binary.** So
copyonline's `resolved_composition` row will have the OLD shape (no `layout_match_score`) even if
it composes onto the new layout — expected, not a regression. I had written "should ride it" in
the first draft of the handoff; corrected to MEASURED / NOT ROLLED.

**Pattern-check advisory on the 736 commit:** `new-capability-surface` — NOTES proposes
`cmd/layout-fit-check/`, which does not exist. Answer, for the record: no existing `cmd/`
scores layout fit or reads the `sites → style_collections → css_themes → layouts` join;
`cmd/verifier-remit-check` is the *template* (fleet-wide finding, `system.internal` shelf), not
a substitute. And per the owner's 2026-09-03 decision the new check is to be built on
`internal/cronchecks` (RFC_024 option 2) rather than as another copy — which is the concern the
advisory exists to raise.

## 2026-09-03 (l) — the roll landed; the verdict I was waiting for never existed

Picked up from `HANDOFF_2026-09-03_continue_here.md`. Ran its §1 liveness table first, as it
instructs. Two of the three rows changed.

**Phase 1 IS LIVE. `v1.0.1359`, pods started 13:28:18Z and 13:28:43Z.** The handoff's row 1 said
NOT ROLLED and that is now superseded.

⚠ **The instrument the handoff named did not work, and it failed in the direction that produces a
false answer rather than an empty one.** Two separate causes, stacked:

1. The pods were **3h05m old** by the time I looked (16:33Z), so the `build provenance` startup
   line had rotated out. `--tail=400` and `| head -200` both returned nothing on the structural
   grep. That is the documented time-limit, and an empty result there means *out of range*, not
   *unstamped*.
2. **The unstructured grep the handoff quotes returned a HIT anyway** — and the hit was
   `LANDMINES.md` prose *about* build provenance, synced into `doc_notes`, injected into an agent
   prompt, and logged by the chassis. Exactly the trap at `LANDMINES.md` §"`logs | grep 'build
   provenance'` now matches LANDMINE TEXT". It fired on me verbatim. **The text I got back was
   the landmine warning me not to trust the text I was getting back.**

So I used the structural match (`"caller":"…/main.go:NN","msg":"build provenance","git_commit":…`),
got nothing, and fell through to the binary probe, **with both controls, on BOTH replicas**:

| literal | nrqf7 | phgh2 | meaning |
|---|---|---|---|
| `weak_tag_fit` | 1 | 1 | Phase 1 present (was **0** on v1.0.1358) |
| `layout_match_score` | 1 | 1 | Phase 1 present (was **0**) |
| `enforceListingItemSources` | 2 | 2 | must-be-present control holds |
| `zzq_literal_that_cannot_exist_4417` | 0 | 0 | must-be-absent control holds |

`git merge-base --is-ancestor 76db94fc7 HEAD` → yes. **The fit evidence and the widened signal are
now in the running fleet.**

### The verdict I was told to read was never produced

The handoff says "the two council verdicts, neither read at handoff". Only one of them exists.

- **Phase 5 (migration 736, `39942a14`) — APPROVED**, round 1, "approved with 2 advisory
  objection(s) — none high-severity". Read in full from `diagnosis_artifacts kind='council_report'`.
  Eleven seats; nine approve, editquality and bug_historian object advisorily.
- **Phase 1 (`76db94fc7`, `34d57f60`) — NO VERDICT, and there never was one.** The run stalled at
  `council_decide` with `last_activity` 12:05:46Z and was swept to `FAILED` at **16:07:33Z**. There
  is a `fix_plan` artifact and **no `council_report` row at all**.

**Why it died, and it was not random.** Seven runs carry that identical `updated_at`
(`2026-09-03 16:07:33.448202+00`), and every one has `last_activity` between 12:05:35Z and
12:06:05Z. The previous chassis (`v1.0.1358`) started **12:06:47Z**. So the roll killed six
in-flight council runs plus one `generic-process`, they sat in `EXECUTING_STEP` for four hours,
and a stale sweep marked them FAILED. This is the known "a roll KILLS an in-flight council"
shape — what is new is what it leaves behind, below.

**Two of the seven belong to other lanes and are still orphaned** (checked their `fix_plan`
bodies): `45ae3ad3` is a `seed_analytics_default.go` submission, and `f0ad8366` is
`portfolio_positioning`'s **migration 734** round 2. Both had drawn a REVISE in round 1 and both
resubmissions were killed. Three of the seven (`8745ad9e`, `63be72d1`, `76288ff9`) do have later
completed runs, so those lanes are covered.

### MISSTEP AVOIDED, and the trap is worth more than the fix: AWAITING reads as *queued*, for ever

I nearly wrote this session off as "Phase 1 is submitted, `098` will credit it when the verdict
lands". It never would. `098`'s `db_decision` reads **only** `diagnosis_artifacts`; it never looks
at `orchestration_states`. A commit carrying `Council-Submitted:` for a **dead** run is bucketed
`AWAITING` and annotated **`no report yet (queued, or evidence cleared)`** — a string that says
*still coming*. There is no elapsed-time bound and no status join, so it says that for ever.

The discriminator is one query and nothing runs it automatically:
```sql
SELECT current_step, status, last_activity FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<corr>';
```
`FAILED` there, with no `council_report`, means **resubmit** — not wait. Written up in
`LANDMINES.md` (footprint `098_REPORT_unreviewed_commits_v1.sh`, `Council-Submitted`,
`orchestration_states`).

**Resubmitted Phase 1: new correlation `adfa4d03-67a8-419f-bc22-d0ef125f94ee`** (dry run first —
admission passed free). The submission JSON was unchanged and still accurate; the code it describes
is now live rather than pending, which is a fact for the reader, not a change to the plan.

### An advisory objection checked rather than accepted, and it does not hold

Phase 5's editquality seat objected (medium) that the reachability guard "treats the tag
`editorial` as canonicalising to `editorial-publication` for counting purposes, inflating the
reachable-site count (5 sites on `editorial` vs 1 on `editorial-publication`)… asserted, not
verified against the actual classifier/matcher behaviour".

**It is verified, in the matcher, and the objection is refuted:**

- `canonicalTag` maps it explicitly — `fork_theme_composition.go:146`,
  `case "editorial", "publication", "magazine", "editorial-content": return "editorial-publication"`.
- **Both sides go through it**: site terms at `:227` (`siteTerms := canonicalSet(append([]string{category}, industryTags...))`)
  and layout tags at `:278` (`layoutTags := canonicalSet(lr.tags)`).
- The seeded row carries `editorial-publication` (verified at the live row).

So a site emitting `editorial` genuinely does reach `content-hub-tools`, and the guard's count
stands. The seat said in its own notes that `layouts`/`site_specs` were outside its schema, so this
was a flagged blind spot, not a bad read — **the answer was one `grep` the seat could not run.**
The other objections (partial `--section-*` check; no positive assertion that the CSS covers
`faq-item` / `featured-article__*` / `category-listing`) are **real and unanswered**; they are
verify-block gaps, not defects in the seeded row, and they belong to the Phase 4 work.

### The peer retraction, verified rather than taken on trust

`portfolio_positioning` retracted migration 734: its register step's `query_database` used `$1`
with no `params` array, so every classifier run after 11:39Z failed with
`expected 1 arguments, got 0`. They rolled that step back and kept the `input_fields` half.

Checked both halves myself at the live row rather than accepting it:
- `input_fields` = `["input_data","search_results","scraped_data","site_specs","layout_taxonomy"]`
  — **`layout_taxonomy` retained**, so the fix for the 87%-unmatchable finding is intact.
- Step list has `read_layout_taxonomy` and **no register step** — the rollback landed.

⚠ **The consequence for this lane is a corrected boundary, and it invalidates a line in our own
README.** 11:39Z is NOT a before/after boundary: nothing classified successfully between 11:39Z and
16:01:39Z, because every attempt died at their broken step. **No classification has yet run with a
real tag list.** Our README said "that went live at 11:39" — corrected there in place.

### What the roll does NOT yet prove

Zero `resolved_composition` rows have been written since 13:28Z, and
`count(*) WHERE data->'lineage' ? 'layout_match_score'` is **0 fleet-wide**. `content-hub-tools`
still has **0** sites. So the handoff's §4 step 2 — *the first post-roll composition must carry
`layout_match_score`* — is **still owed and still the right test**. The roll is proven at the
binary; the fit evidence is unproven until a composition runs.

## 2026-09-03 (m) — the canary fired, and the first datum attacks my own metric

**The real before/after boundary is `2026-09-03 16:54:12Z`** — the first classification ever to run
with the layout tag library rendered. Not 11:39Z (retracted, see (l)). Verified independently at the
artefact before writing this down, on the one `classify_and_extract` row that exists today
(`created_at 16:54:07.139237+00`, `llm_call_log`):

| check | result |
|---|---|
| tag-list region contains `null` | **false** (was true on every prior run) |
| prompt contains `content-hub` | **true** — 736's new tags reached the classifier |
| prompt contains `editorial-guides` | **true** |
| 735's promise sentence removed | **true** |

So **734's surviving half, 735 and 736 are all proven at the artefact in a single prompt.** Before
today `llm_call_log` had zero `classify_and_extract` rows for 2026-09-03 — I checked that first,
because "the migration is verified at the live row" says nothing about whether a run has ever
exercised it. 735 had been applied-and-unexercised for five hours, exactly as 734 was.

### The result, from `portfolio_positioning`, and it is not the good news it looks like

copyonline emitted **10 tags, 10 of them matchable — 100%**, against the ~13% (28/216) baseline.
Coverage did exactly what my pre-registered prediction 1 said it would.

**And the tags describe the wrong site.** `marketplace, directory, community-platform, b2b,
professional-services, content-platform, creative-agency, interactive-platform,
practitioner-platform, industry-hub`, `category=hub`. Copyonline is positioned as an editorial
authority on writing commercial copy, with tools; its directory is one page and a marketplace is an
explicit must-not. **Not one tag says copywriting, editorial, guides or long-form.** So
`content-hub-tools` will not win it, and the classifier has typed the site the brief exists to
replace.

### ⚠ This is a disconfirmation of MY work, not only theirs, and I want it recorded as one

Their formulation is the sharp one: **"matchability is not accuracy, and coverage cannot
distinguish them — a site typed entirely in borrowed vocabulary scores 100%."**

That lands directly on the metric I shipped in Phase 1. `TagCoverage` measures *what fraction of a
site's emitted tags the chosen layout addresses*. It has no view on whether those tags describe the
site. So:

- **The weak-fit arm will stay SILENT on copyonline** — high coverage, no signal — even though this
  is precisely a site the library serves badly. My detector's first live encounter with a
  mis-typed site is a miss, and it is a miss BY CONSTRUCTION, not by threshold.
- **The 0.50 threshold cannot be re-derived on post-16:54 data.** Prediction 1 in the handoff says
  "coverage should rise fleet-wide; if new compositions land inside 38–62% the cut was a 33-site
  artefact". That test is now **contaminated**: the rise is caused by the classifier being steered
  toward library vocabulary, so re-deriving the cut on that population is fitting the metric to
  itself. **Prediction 1 is retired as a threshold test.** What it can still do is flag a *fall*.
- **A `[MEASURED]` coverage figure after 16:54Z is not comparable to one before it.** Any
  before/after on this metric must say which side of 16:54:12Z it sits.

**What I am NOT claiming.** n=1. The peer names an alternative I cannot exclude either: copyonline's
brief leads with its marketplace heritage and its directory, so the classifier may be reading the
brief correctly and the *brief* may be over-weighting. Testable on the next remake and on a
re-classification of this one. **One site is not a rate.**

**A risk in my own migration I should have flagged when I wrote it** `[UNMEASURED]`: 735's
form-over-industry sentence quotes four live library tags as examples (`long-form`, `tool-portal`,
`content-hub`, `comparison`). MEMORY carries the trap that *a quoted exemplar in a prompt ships
verbatim*. copyonline emitted none of the four, so there is no evidence of copying yet — but the
next few classifications are the place to look, and if exemplar tags start appearing at rate, the
sentence needs rewording to describe the shape without naming instances.

### An unrelated live defect found in the same prompt, contributed rather than filed

The same rendered prompt carries one `<no value>` — the **Pre-Defined Mission** block. The template
guards on the parent (`{{if .site_specs.specs.mission_brief}}`) and prints a child
(`{{.site_specs.specs.mission_brief.text}}`); copyonline's `mission_brief` is a rich structured
object with **no `text` key**. `[MEASURED]` **7 of 23** current `mission_brief` specs lack it. The
block is load-bearing — migration 464 licenses a regulated business model only when a Pre-Defined
Mission "is present above and explicitly asks for one" — so on those 7 the licence renders and the
constraint does not.

`who-owns.py 453` shows that lane six commits deep in exactly this class today, so it is written up
as a **CONTRIBUTION into `bugs_open/453`**, not a new bug number, and I have not touched the
template. Their `--template-input-fields` lint cannot catch it: the root `site_specs` **is** in
`input_fields`, so it is their shape 3 (root present, sub-field absent), decidable only at render.
