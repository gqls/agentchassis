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
