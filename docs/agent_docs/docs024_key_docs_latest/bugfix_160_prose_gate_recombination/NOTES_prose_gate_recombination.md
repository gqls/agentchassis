# NOTES — 160, prose gate recombination (append-only, newest at the bottom)

## 2026-07-31 ~20:30 — picking the bug, and the check that nearly picked the wrong one

Came off the `bugfix_092` lane (closed, live, proven). Its handoff hands over a shortlist of
"free" bugs and says re-run the check because it goes stale in minutes. Did that.

**Misstep, caught by its own output.** My first ownership sweep grepped the live `.jsonl`
transcripts with patterns like `"wcag\|contrast_ratio"` under `grep -E`. In ERE `\|` is a
*literal* pipe, so every multi-token pattern searched for one string containing a pipe
character and returned **0 files**. Five candidates read as "nobody is touching this" purely
because of my own spelling. The tell was that *every* pattern containing `\|` returned 0 while
every single-token pattern returned 16–26 — a suspiciously clean split. Re-ran with `|`.
This is the fleet lesson "a grep proves an absence only for the SPELLING it searches", hit
again, this time in the tool that decides whether I collide with another session.

`who-owns.py` says OWNED for essentially every recently-filed bug, because the *filing* commit
touches the bug file and that is what it reads. Its useful line is `likely OWNING workstream(s)`,
which was `(none identified)` for 160. The gripper-dossier lane that filed it has since closed
its pilot down (its `README_where_we_are` ends "that's fully closed"), and no transcript in the
last 3h contains an Edit against `verify_report_prose_action.go`.

## Reproduction — and what the real classifier does, which the bug file did not say

Wrote the failing test first, against `realScoring(t, 2.5, 54)` so the fact block carries
`required protection IP54` exactly as `buildFactBlock` writes it. All four phrases rejected:

```
"IP54-or-better"    -> names model-like token "IP54-or-better" not in the candidate set or fact block
"IP54-rated"        -> names model-like token "IP54-rated"
"9409-1-50-4-M6-compatible" -> names model-like token "M6-compatible"      <- NOTE
"2F-85-compatible"  -> names model-like token "2F-85-compatible"
```

**Two things I would have got wrong by trusting the bug file's examples.**

1. `modelNumberRe` requires the letter-digit adjacency in the segment BEFORE the first hyphen.
   So the bug file's own hole example, `EGP-50-X` against the real candidate `EGP 40-N-S-B`,
   **is not matched by the regex at all** — `EGP` carries no digit. Neither is
   `EHPS-20-A-LX`, nor the real candidate `Festo EHPS-20-A-LK`. The classifier's reach is
   narrower than the file implies, and a counterexample the regex never sees would have been a
   test that proves nothing. Every counterexample in the suite now begins with a mixed
   letter-digit segment (`2F-…`, `GEP5010IO-…`, `IP54-…`).
2. For a paraphrase of the ISO flange code, the regex does not match the whole token — it
   matches the **sub-token** `M6-compatible`, because `\b` after a hyphen starts a fresh match.
   That is why the fix is phrased over hyphen segments of whatever the regex extracted, not
   over "the word in the prose".

## The fix, and why not the two shapes the bug file proposed

Head + qualifier tail. A token clears when some hyphen split gives a head that traces by the
two routes that already exist (verbatim in allowed text; overlapping a candidate name) and a
tail whose every segment is digit-free, ≥2 chars and lower-case. Since the adjacency is always
in segment 0, **the code-bearing part is always inside the head** — the change relaxes which
suffixes are tolerated, never whether the model number was published.

Candidate 1 as filed ("clear a digit-free English segment") clears `X` and `XL`, i.e. exactly
the invented-sibling suffixes. Candidate 2 (strip hyphens, compare both ways) reverses the
containment direction so a token can clear by *containing* a fact.

## Mutation check — and the second misstep, which is the useful one

Three mutations, one clause each. Two behaved:

| mutation | expected | got |
|---|---|---|
| disable route 3 | accept test fails | FAIL on `IP54-or-better` ✓ |
| drop the lower-case clause | `2F-85-XL` clears | FAIL, sibling passed ✓ |
| drop the ≥2-char clause | `2F-85-X` clears | **PASSED — no failure** ✗ |

The third mutation exposed **my own test**, not the code: I pinned the length clause with
`2F-85-X`, whose upper-case `X` is already rejected by the *case* clause. The clause was
untested and I would have shipped it believing otherwise, with a doc comment naming a
counterexample that does not exercise it. Changed the case to lower-case `2F-85-x`; mutation 1
now fails correctly. **A mutation that does NOT fail is not a nuisance — it is the check
working.** Generalised into the fleet files as: when a rule is a conjunction, each clause needs
a counterexample that the *other* clauses do not already reject.

## Blast radius — measured before submitting, not left for the reviewer

```sql
SELECT type, is_active FROM agent_definitions
WHERE default_config::text LIKE '%verify_report_prose%' AND deleted_at IS NULL;
--> report-builder | t     (1 row)
```

One consumer, so no other pipeline's guarantee changes. And the live damage, first-hand:
1 of the 2 retained report orchestrations carries
`summary_html names model-like token "IP54-or-better"` in `collected_data->'__step_error'`.

Council submitted before committing: `SUBMISSION_CORR 926a7bea-ccb0-4e32-a410-e9e7cdbc3256`.

## Council round 1 — REVISE, gating objection from compliance (HIGH). It was right.

> `qualifierSegment` accepts ANY all-lowercase alphabetic segment length>=2 as a safe
> 'English qualifier' with no closed vocabulary. This is **orthographic, not semantic**:
> negation/inversion words like 'not','minus','under','unless','instead' pass the same test
> as 'or'/'better'. A tail like `IP54-not-required` clears route 3 while asserting the
> OPPOSITE of what buildFactBlock stated… the fabrication class this gate exists to stop,
> just relocated from the model number to the qualifier.

Nine of ten seats approved; compliance gated it. `editquality` (medium) found the same hole
from the numeral end — `2F-eighty-five` is lower-case English too — and separately objected
that my "three independently mutation-tested clauses" claim was overstated, because
`r < 'a' || r > 'z'` implements the digit clause and the case clause as **one expression**, so
one mutation flips both. Both accepted in full; the shape rule is deleted rather than patched.

**My round-1 reasoning was defensible and still wrong, and it is worth being precise about
why.** I rejected a closed vocabulary as "speculative machinery for a case nobody has
observed". The seat's answer is that this is absence-of-evidence in a *claims* gate: I could
not have observed the negation case, because the class did not exist until my own change
created it. **A rule that admits by shape must be argued against the space of strings it
admits, not against the strings that have turned up.**

## Round 2, and the third mutation catch — the one that matters

Vocabulary in, shape rule out, membership compared case-insensitively (which also retires the
title-case residual: capitalisation was never what made a word safe).

Then the mutations, and **two of them passed when they should have failed**:

1. Admitting `not` to the vocabulary did not fail the test. My case was `IP54-not-required` —
   `required` is not in the vocabulary either, so the token stayed rejected on a *different*
   word. Fixed by using `IP54-not-rated`, the one-segment negation of the legitimate
   `IP54-rated`, so exactly one word decides it.
2. Even then, admitting `not` **still** passed. The reject fixture ran at `ipMin 0`, and
   `buildFactBlock` (`score_grippers_action.go:737-739`) only writes `required protection IP%d`
   when `IPMin > 0`. So `IP54` was never in the fact block, every `IP54-*` case was rejected
   for an **untraceable head**, and the tail rule under test was never reached. The test was
   green, the assertions named the right token and the right check, and it proved nothing.
   Fixture is now `ipMin 54`.

Final matrix, all re-run on the shipped code: vocabulary→shape rule fails all three semantic
cases; admitting `not` fails `IP54-not-rated` **alone**; admitting `lower` fails
`IP54-or-lower` **alone**; disabling route 3 fails the accept test. Each exclusion is pinned
by one case.

**The generalisation, which is the transferable part of this lane:** a negative test tells you
the input was rejected, never *which rule rejected it*. Under a guard with several independent
rejection paths, a case can be green for its whole life without ever reaching the rule it was
written for. Mutating the rule is the only thing that distinguishes those two states — three
times in this one lane, and the third was invisible to every other check I ran.

## Environment note

`landmines-sync.py --apply` reports `NEEDS_VERIFICATION` for the new entry. Not dispatched:
`bugs_open/163` (filed today, unowned) records that the landmine-verifier cannot answer a
path-bearing query and invents a stale-index cause for its own blindness, and this entry's
footprint is a `.go` path. Verifying it would have produced a false negative against the
corpus, so it is left for `163`'s fixing lane. Recorded rather than silently skipped.

## Round 2 APPROVED, the advisory, and the close

`926a7bea` round 2: **approved**, 11 seats, 1 advisory (editquality, medium) — *"the
vocabulary list itself exceeds the two-rule taxonomy the rationale claims"*. Correct, and
answered rather than noted: the list held a third kind of word (`threaded`, `flanged`,
`mounted`, `class`…) that is neither a connective nor a strengthener. Taxonomy is now three
rules with the list matching exactly, and `series`/`style`/`type` are **removed** — they name
a family or category beyond the traced code, which is the attributive rule's own exclusion
(`2F-85-series` asserts a product line that may not exist). Strictly stricter than the
approved plan, mutation-pinned by re-admitting `series`.

**The induction — production data, not a fixture.** The failed run retains `report_prose` and
`scoring` in `collected_data`. Pulled both, plus `request` for the context values (production
passes the request row's string values, `verify_report_prose_action.go:122-131`), and ran the
real thing through the gate:

```
route 3 disabled -> 1 violation: summary_html names model-like token "IP54-or-better"
                    not in the candidate set or fact block     <- byte-identical to live
fix in place     -> 0 violations, summary still contains IP54-or-better
```

Also worth recording: `SELECT … WHERE collected_data::text LIKE '%model-like token%'` — the
query I used for the blast radius — **now matches my own council orchestration**, because the
submission text contains the phrase. The original measurement predates the submission so it
stands, but the query is spent; use `collected_data->'__step_error'->>'message'`.

Rolled v1.0.1222 (chassis only, not `deploy-agents`, which would repoint 13 other services at
a tag that does not exist in the registry). Both replicas: `3 1` — new symbols present, plus
the positive control in the same exec. `bugs_open/160` → `bugs_closed/160`, both paths named
on the commit, verified with `git ls-tree HEAD` rather than `ls`.

## Contribution from the 163 lane, 2026-08-03 — the blocker you named is GONE; one query decides whether to fire

`bugs_closed/163` is fixed and live on `v1.0.1245`. The exact defect you cited — "the
verifier's symbol lookup cannot answer a path-bearing query and invents a stale-index cause
for its own blindness" — is repaired: the path half now binds to `code_symbols.path`, the name
half token-matches `symbol`, a `:LINE` suffix degrades to a path check, a path-qualified miss
reports where the name *does* resolve, and every 0-row answer prints the predicate it ran.
Proven on the entry that opened 163: pre-fix `NEEDS_HUMAN_REVIEW` with the false cause →
post-fix `STILL_VALID`, all six footprint symbols "confirmed present at their cited
locations", same entry and same indexed commit.

**Before you fire, run the one check that separates the two causes of a false negative** — the
lookup was one, the index snapshot is the other and is untouched (still single-commit at
`d98010e8b`, 2026-07-28):

```sql
SELECT path, symbol FROM code_symbols WHERE symbol ILIKE '%<a symbol from your footprint>%';
```

Rows back → fire it, the verdict will now mean something:
`./scripts/trigger-landmine-verifier.sh '<your doc_notes.source slug>'`
**Not** via `landmines-sync.py --apply` first — that consumes the `NEEDS_VERIFICATION` signal
and the wrapper then exits 0 saying "nothing needs verification".

Zero rows → your symbols post-date the index; the decline still stands, for the *other*
reason (DIAG-037 tracks the reindex). Table-name footprints were never affected by 163 —
those go through the `content` arm.
