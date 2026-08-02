# NOTES — content_data link resolution (bugs_open/097)

Append-only, newest at the bottom. Evidence, commands, what the system actually
said — **and every misstep**, which is the part the next thread cannot rederive.

---

## 2026-08-02 — picking the bug, and the ownership check that mattered

`scripts/who-owns.py` returned "OWNED or recently active" for nearly every
candidate I tried, because it counts commits that merely TOUCH a bug file — and on
this tree every bug file is touched constantly. It is lagging by construction and
it says so.

What actually discriminated was grepping the **live `.jsonl` transcripts** for the
code symbols, not the bug numbers:

```
ctaFieldNames        max 6 mentions in any live session (a cross-reference)
resolve_internal_links   max 10
landmine-verifier    112 in 693556a1  -> bugs_open/163 IS being worked; dropped it
bugfix_165           180 in 806dfccd  -> dropped
bugfix_154           209 in 9de5c96a  -> dropped
```

Counting `bugs_open/NNN` alone is nearly useless: every session that runs
`ls bugs_open/` picks up all 60 numbers at once. **The signal is the SYMBOL, and
the discriminator is one session having a large count while everyone else has a
handful.**

## 2026-08-02 — the bug as filed is half-done, and reading to the bottom is the whole job

097's fix candidates are at the TOP of the file and are 6 days stale. Its live
state is in the **last** dated section, and it says the repair half shipped
(4 seams) and the detector half did not. Acting on the candidate list without
reading to the bottom would have rebuilt `RepairPageLinks` for the third time.

Then `component_link_repair.go` — committed by the **136 lane this morning**, hours
before I started — turned out to name my scope precisely and route it here:

> *"content_data. Same limit as 079's fix … The deployed artefact stays covered by
> the outbound rerender seam (repairOutboundPageLinks, bugs_open/097)."*

That is a clean handoff between lanes and it is why the scope is not a guess.
**Read the files your bug's siblings committed TODAY, not only the bug file.**

## 2026-08-02 — the measurement, and why the SQL version was not good enough

The SQL census (RUNBOOK R2) said 51 unresolved occurrences. Running the **shipping
code** over the same corpus (R1) said 52. The difference is small and the reason
matters: the SQL had to hand-reimplement `NormalizePagePath` and
`ClassifyLinkScope`, and a hand copy of a shared definition is exactly the drift
class this platform reviews for. The Go number is the one quoted anywhere.

```
components audited      : 885
components with findings: 13
REWRITE (target exists) : 19
PHANTOM (report only)   : 33
```

**The 872 components with NO findings are the load-bearing half of that result.**
They are the evidence that a name-based nomination with a value-based judgement
does not fire on prose, on assets or on off-site links — with no exclusion list
anywhere in the code.

## 2026-08-02 — MISSTEP: I nearly shipped two guards that made each other untestable

`TestFindingOrderIsStable` passed. I then mutated the code to prove the ordering
guard was load-bearing, and **it still passed**. Cause: I had written *two*
mechanisms guaranteeing one property — `sort.Strings(keys)` inside the walk and
`sort.SliceStable(findings, …)` at the exit. Either one alone makes the output
deterministic, so **deleting either left every test green** and neither was ever
demonstrated to do anything.

This is the memory-file shape *"two checks blind the SAME way AGREE with each
other"*, in the writing direction rather than the measuring one. A second
mechanism that cannot be distinguished from the first is not belt-and-braces — it
is a guarantee nobody can test.

Fixed by **deleting the redundant one** (the key sort) rather than keeping it as
decoration, after which the mutation reproduced immediately:

```
run 0: [a_url b_url c_url d_url e_url f_url]
run 1: [d_url e_url f_url a_url b_url c_url]     <- Go map randomisation, visible at last
```

The check: **after adding a guard, delete it and watch a named test fail.** If it
does not, either the test is vacuous or something else is already doing the job —
and both are worth knowing before the commit, not after.

## 2026-08-02 — MISSTEP: two mutations that "passed" had simply failed to compile

Two of the five mutation runs printed `FAIL` and I nearly recorded them as proof:

```
platform/orchestration/datahelpers/content_data_links.go:163:16: invalid operation: v[i] ... is not an interface
FAIL  github.com/gqls/agentchassis/platform/orchestration/datahelpers [build failed]
```

`[build failed]` is not a red test — it is no test at all. A malformed mutation
proves nothing about the guard, and the word FAIL in the output makes it look like
it did. Both were redone until they compiled and failed on an **assertion**.

## 2026-08-02 — the archived page that would have looked like a false positive

`robot-hands.com` cards link `/learning-center`, and the fix rewrites that to
`/learning-center.html`. Checking whether that was correct turned up:

```
 learning-center         | /learning-center.html        | active   | deployed
 learning-center-index   | /learning-center/index.html  | archived | deployed
```

`NormalizePagePath('/learning-center/index.html')` is `/learning-center`, so if the
index included archived pages the link would have RESOLVED and produced no
finding at all. It is excluded because `loadValidPagePaths` uses the shared
`linkablePageStatusPredicate` (`status NOT IN ('deleted','archived')`) — which the
offline census had to match exactly or the whole measurement would have been of a
different population. **Copy the predicate; do not retype it from memory.**

(That archived-but-still-served page is `bugs_open/098`'s shape — archiving does
not undeploy. Routed there, not worked around here.)

## 2026-08-02 — what I deliberately did NOT do, and why each was a judgement

- **Did not blank phantom values.** It is the `content_data` analogue of the
  unlink arm, and `link_repair.go`'s own header records that arm as unsettled by
  two council seats. Pinned by a test so reversing it is deliberate.
- **Did not touch the staged CTA precedence flip** (`ctafields.go`, trail
  `2525f980`). It belongs to the `cta_link_integrity` lane, carries 5 binding
  constraints from 5 seats, and inverting precedence is that round's job. Named in
  the submission so its reviewers are told rather than left to measure it.
- **Did not file `site_work_items`.** `bugs_open/083` (nothing drains `detected`)
  and `bugs_open/077` (no items whose handler has no remit) — the same reasoning
  `writeLinkRepairLog` already wrote down.
- **Did not write a migration.** The live damage clears on each page's next save.
- **Did not run `090` (needs_diagnosis).** Stated plainly per the owner ruling of
  2026-07-31: 097's mechanism was already CONFIRMED by diagnosis `9543aaf1`, and
  what I added is not a new root-cause claim but a census — I read the exact code
  (`ctaFieldNames`, `DeriveCTAURLFields`), read the live `input_schema` that hides
  the field, and ran the shipping function over all 885 production rows. That is
  first-hand verification of the same kind the loop would have performed, on a
  cause the loop has already confirmed once.
