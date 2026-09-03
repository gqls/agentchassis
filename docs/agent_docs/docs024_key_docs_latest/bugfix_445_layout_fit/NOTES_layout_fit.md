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
