# 267 — the bundle advises a whole-file re-read that cannot fit, and the loop spends an iteration obeying it

**Filed:** 2026-08-12, `silent_hero_logo_readers` lane. Found by the run that verified
`bugs_closed/261` — this is the defect that was **behind** it.
**Status:** OPEN, not started. Diagnosed first-hand from artefacts; no code written.

> **Numbering:** 262–266 were all taken by other sessions while `261` was being closed **this
> afternoon**. Checked 267 free at filing. Resolve by slug.

---

## 1. Symptom

A `090` run reaches `UNVERIFIABLE` reporting *"body omitted, too large for the bundle"*, having spent
an iteration asking for a whole file that could never have fitted.

Worked case, run `eddaf1af-b44d-4bc0-8485-5885056042cd`, iteration 3:

```
### platform/orchestration/coordinator.go
_(body omitted — 169139 chars, and 0 of the 60000-char body budget is already spent.
  It was found; it did not fit. Put THIS SYMBOL ALONE in next_scope to read it whole.)_

> **This section is INCOMPLETE.** 0 of 1 in-scope symbol(s) rendered with a body.
```

**The marker is honest and the accounting is correct** — this is `bugs_closed/164` working as
designed, separating "did not fit" (coverage) from "could not be read" (defect). The defect is not
the report. **It is that the loop was ADVISED to make a request that was arithmetically impossible.**

## 2. Root cause — advice offered unconditionally for an outcome that is conditional

Two places in `diagnose_assemble_bundle_action.go` tell the model to re-request a whole file:

- the sibling-signatures section: *"put the bare file path in next_scope to see it whole"*
- the over-cap marker itself: *"Put THIS SYMBOL ALONE in next_scope to read it whole."*

Neither is conditioned on the file fitting. `coordinator.go` is **169,139 chars against a
60,000-char budget** — 2.8×. **No arrangement of scope makes that request succeed**, and the second
message is the sharper failure: it is printed *by the cap*, which at that moment knows both numbers,
and it still advises the re-request. Alone or not alone, it will not fit.

The cost is not just a wasted step. The loop had **3 iterations**; it spent one of them on this, and
its final verdict then asked for exactly the right four symbols — **as `next_scope`, with no
iteration left to spend it**:

```
"next_scope": ["…coordinator.go:persistAwaitingStateWithRetry", "…:storeActionResult",
               "…:processAwaitResponse", "…:applyResponseToState"]
```

Those four bodies total ~13,400 chars — they fit inside the 60,000 budget four times over. **The
loop worked the answer out and ran out of road.**

## 3. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Do not advise what the numbers already refute.** The over-cap marker has `chars` and `budget`
   in hand. When `chars > budget`, drop the "read it whole" sentence and instead name the file's
   largest symbols — the bundle already computes per-file symbol lists for the sibling section, so
   the data is in scope at that point. This closes the door rather than warning about it.
2. **Make the sibling section's advice conditional too** — *"put the bare file path in next_scope"*
   should not be printed for a file whose whole-file size exceeds the body budget.
3. Weaker: raise the cap. Rejected — 169k against 60k is not a tuning problem, and a bigger cap
   spends the verdict model's context on a file it was never going to read.

## 4. Verification

The disconfirming observation is cheap and specific: a bundle whose scope is a bare path to an
oversized file should **not** contain the string "read it whole". Confirm the advice still appears
for a file that *does* fit — otherwise the fix is a blanket removal rather than a conditional one,
and that is the failure mode to guard against.

## 4b. MEASURED 2026-08-12 (~20:05Z) — the `[UNMEASURED]` above is now discharged

Over all **486** bundles ever assembled:

| shape | iterations | runs |
|---|---|---|
| a body omitted for size (any) | 49 | 26 |
| …and the bundle also says "read it whole" | 44 | — |
| **zero bodies rendered, cause = CAP only** | **6** | **5** |
| zero bodies rendered, cause = resolver only (`bugs_closed/261`'s class, now fixed) | 6 | — |
| zero bodies rendered, both causes | 1 | — |

**The load-bearing number is 6 iterations across 5 runs** — an iteration that returned *no code at
all* because everything in scope was too big. At 3 iterations per run that is a third of the run's
budget, spent on a request the bundle itself advised and its own arithmetic refuted.

⚠ **Do NOT quote the 44.** That is co-occurrence — "read it whole" appears somewhere in the bundle
and *a* body was omitted — and the phrase is printed by the over-cap marker itself, so the two are
nearly the same event by construction. It is not 44 wasted iterations.

⚠ **The separation matters and the naive query gets it wrong.** `**This section is INCOMPLETE.** 0 of
N in-scope` is emitted for BOTH causes. Filtering only on that gives 13 iterations and silently
folds `261`'s resolver failures into `267`'s cap failures — two different bugs, one of them already
fixed. The discriminators are the marker words: `did not fit` (cap) versus `could not be read`
(resolver). My first count made exactly this mistake and read 13.

```sql
SELECT count(*) FILTER (WHERE body ~ '\*\*This section is INCOMPLETE\.\*\* 0 of [0-9]+ in-scope'
                          AND body LIKE '%did not fit%' AND body NOT LIKE '%could not be read%') AS cap_only,
       count(*) FILTER (WHERE body ~ '\*\*This section is INCOMPLETE\.\*\* 0 of [0-9]+ in-scope'
                          AND body LIKE '%could not be read%' AND body NOT LIKE '%did not fit%') AS resolver_only
FROM diagnosis_artifacts WHERE kind='bundle';
```

**Live confirmation while this file was being written:** the seeded re-run
`36bd1b42-29b5-4094-9264-94ea80c6194a` hit the cap on its **iteration 2**, having been given a
perfectly-sized seed for iteration 1. So this is current behaviour, not history.

## 5. Why this was invisible until today

It sat **behind** `bugs_closed/261`. Until this afternoon the code tier could not resolve
receiver-qualified method names at all, so a scope naming those methods failed on *resolution* and
never reached the *cap*. Fixing the resolver is what let a run get far enough to hit this.

**That is the expected shape of removing a blocker, not a regression** — but it is exactly the shape
that gets misreported as one. The 261 fix is independently proven: same run, iteration 1, **12 of 12
bodies rendered on the identical scope where 3 of 12 rendered pre-fix**.

## 6. Relations

- `bugs_closed/261` — the resolver defect in front of this one; its §8 records the neighbouring
  sibling-signature cap issue.
- `bugs_closed/164` — the omitted-vs-unreadable distinction. **Working correctly here**; this bug is
  about the advice, not the accounting.
- `bugs_open/236` — the case still waiting on all of this. Re-run seeded with the four symbols is in
  flight as `36bd1b42-29b5-4094-9264-94ea80c6194a`.
- `bugs_open/174` — `seed_scope` confiscated in transit; the reason that seeded re-run checks for the
  "This scope was NOT chosen" marker rather than assuming its seed arrived.
