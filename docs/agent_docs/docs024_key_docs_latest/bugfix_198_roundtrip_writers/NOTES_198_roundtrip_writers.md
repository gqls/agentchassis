# NOTES — bugs_open/198 (two-writer stylesheet clobber)

Append-only, newest at the bottom. Evidence, commands, and every misstep.

---

## 2026-08-21 — the prevention half built and shipped (bugfix-198 lane)

### Starting state, and how much of it was already someone else's

The owner dispatched this session at `bugs_open/198`. First action was `who-owns.py` and a
read of the bug file — which by then ran to **982 lines across five lanes**, and had gained
five commits *that same day* from three other sessions. Re-reading it before writing
anything was the single most useful thing I did: **most of what I would have "discovered"
was already recorded**, and one thing I planned to do had already been done by someone else
while I was planning it.

What other lanes had done today, before I touched anything:

- fleet backfill of every empty theme row (9 rows) — ROUND-2 candidate 1, **done as data**;
- remortgagecalculator.uk and loanzy.uk clobbered at 10:27Z and restored the same morning;
- cookly.uk restored (18,047 B served) by the `news_editorial_features` lane;
- a `stylesheet_gutted` discovery check built (`e34b33a36`, IMP-055) with its enabling
  migration held at `541_..._HOLD.sql`, council `d3187418`;
- the two-clause `-ink` staleness check, and its correction.

Their §7 states plainly what was left: **candidate 2 (deploy-side shrink guard), the birth
guard, and candidate 6.** That is the scope I took, and it is why this session built no
restores.

### Measurements taken first-hand (not inherited)

Census via the exact `load_current_css` JOIN, 2026-08-21 ~15:00Z:

| | |
|---|---|
| linked theme rows | 22 |
| healthy | 19, at **13,650–26,917 bytes** |
| armed | 3, at **0–1,649 bytes** |
| nothing in between | 2,381 → 13,650 is empty |

That gap is what makes a 4096-byte floor defensible rather than arbitrary, and it is the
first thing to re-derive if the census ever changes shape.

Watched loanzy.uk mid-loop at 15:14–15:25Z: theme row **v13 → v15 → v26** while I queried
it, 14 patches appended in a day onto a base born empty on 08-18. Its first 600 characters
were *pure patch accumulation* — two blank lines, then provenance comment / rule, repeating.
No `:root`, no layout. That is the clobber's fingerprint, and it is unambiguous.

> **CORRECTED, same session, hours later.** I reported loanzy as "mid-clobber-loop, live" in
> my own summary. By the time I came to act on it, another lane had restored the row and the
> queued items were appending to a TRUE base — the same mechanism, running benignly. **The
> site was never going to need my intervention.** I had not mis-measured; I had let a
> measurement age across a decision boundary on a tree where nine lanes are working. The
> honest version: *loanzy was mid-loop at 15:14Z and restored by ~15:25Z.* A live-state claim
> needs its timestamp attached to the claim, not to the session.

### The three holes the Plan agent found in my design, all real

1. **A shrink-guard refusal would have minted `complete`.** `deploy_css.error_step` pointed
   at `complete_error`, which is a *success-labelled* `complete_workflow`, so the dispatch
   loop's `complete_work_item` stamps `complete`. The guard would have fired correctly and
   the ledger would have said it did not. This is now a LANDMINE entry, because it
   generalises: **any** guard routed to an `error_step` on this platform inherits it.
2. **`site_count = 1` alone does not protect a library theme** — a seed theme linked by
   exactly one site would be overwritten with that site's render. Added `origin <> 'seed'`.
3. **The refusal-forever worry was real but inverted.** A `needs_human_review` item is never
   re-promoted (the promoter selects `status='detected'` only), so no queue balloon — but
   `idx_swi_dedup` does **not** exclude that status, so a parked row HOLDS its dedup key and
   the finding cannot re-file even after the base is healed. Hence `result_fields.parked_by`
   and an explicit unpark sweep in the RUNBOOK. Without the marker the sweep would be
   approximate, which for a status humans also set by hand is not good enough.

### Proofs, and what each could have come out as

- **543's persist UPDATE, in a rolled-back transaction against live rows.** A real
  25,202-byte value onto dartsonline: `UPDATE 1`, v5→v6, and `md5(css_content)` equal to
  `md5(value)` exactly — byte-identical, which is the property that makes "deploy the whole
  row" safe. Then `UPDATE 0` four times: shared row, 100-byte fragment, unchanged content,
  seed row. **Four negatives and one positive from the same statement** is what makes it
  evidence rather than a demonstration.
- **Two Go mutations RUN, not asserted.** Deleting the `enforceFileShrinkFloor` call failed
  three tests. Measuring the **unprefixed** path failed its dedicated test *and let the
  clobber commit straight through* — which is the point of having that test: every lookup
  404s, every 404 reads as "new file", and the guard logs that it ran. Source restored
  byte-for-byte after each.
- **Built and tested from a clean `git archive HEAD` tree plus only my files.** The working
  tree fails an unrelated test (`render_context_step_boundary_resolver_test.go`) from
  another session's in-flight work; on HEAD it passes. Without the archive check I could
  have spent the evening on someone else's failure, or worse, assumed mine caused it.

### Missteps

**1. I checked a migration number was free with a query that could never have found one.**
`WHERE filename LIKE '54[23]%'` — SQL `LIKE` has no character classes, so that matches the
literal string `54[23]` and returns zero rows regardless of truth. Caught only because the
*same* query returned zero again ten minutes later, when the runner had just printed
`recorded` for both files: two contradictory answers from one query. The conclusion was right
for an unrelated reason (an `ORDER BY DESC LIMIT 5` listing I had also run), which is exactly
why it survived — **a worthless check that agrees with a good one is invisible.** Full entry
in `WRONG_CALLS.md`.

**2. I nearly ran `run-migrations.sh --apply`.** The dry run listed a large pending backlog
belonging to other lanes, several flagged by the script's own lint as replay hazards
("already applied and merely unrecorded"). Applying mine would have applied all of theirs.
Used apply-by-hand + `--record-only` instead. The non-obvious part is that **recording is not
bookkeeping**: my migrations carry probe guards that `RAISE` on re-application, so leaving
them unrecorded would have made the next `--apply` abort and block every later migration in
the queue, including other lanes'.

### What is deliberately NOT proven

The **witnessed live refusal**. It is proven in-transaction, by config probe, and by the
evaluator's own test suite — but not observed on a real dispatch. I chose not to induce one:
the only sites that would exercise the refusal arm are finetuning.uk and gaswholesalers.com,
both live, and a gate that failed would clobber them. That is the bug file's own closure bar
and it stays owed rather than being quietly claimed.

Likewise the shrink floor is **committed and inert** — it needs both a chassis and a
git-adapter roll. Post-roll probes are in the RUNBOOK §7 with their negative controls.

---

## 2026-08-21 (later) — the arm I missed, and the council round

### A graph query found a defect that reading my own edits could not

After 542 applied I ran an edge-resolution query over the whole workflow — every
`next_step`, `error_step`, `config.then_step`, `config.else_step`, each target resolved
against the step map. **I wrote it to catch a DANGLING edge after the rewire.** Every edge
resolved; the query "passed". Reading the 18-row table it printed is what showed:

```
check_saved | else | complete_error | ok
```

`check_saved` is not an `error_step` — it is a `conditional_branch`, and its refusal travels
on `config.else_step`. My 542 rewired the three `error_step`s and never touched it. So the
door 318 built on purpose — the guarded append matching zero rows when the model returns an
empty or oversized `css_added`, i.e. **the founding 2026-08-04 failure mode** — still landed
on `complete_error` and still minted `complete`.

**Reading the steps you edited cannot find the step you did not edit.** My 542 verify block
asserted every edge I had changed and was green; it had nothing to say about the one I had
not thought about. Migration 546 fixes the arm AND promotes the edge-resolution query into
the verify block as a post-condition, so a future migration that orphans a step fails at
apply rather than at runtime.

Worth noting what kind of check this is: it is a **structural** check over the artefact, not
a check of my intent. That is why it could contradict me. A verify block written from the
diff can only ever confirm the diff.

### Council: APPROVED round 1, and one objection was right about my reasoning

Six advisory objections, none high-severity. Four were checkable and I checked them rather
than accepting them:

- **the installed SQL string had never been executed.** `debug_historian` pointed at the
  landmine: step SQL is DATA to a migration's verify and parses only when the step RUNS. My
  542 proof was of the gate's *arithmetic*, via an equivalent hand-written SELECT — not of
  the string that actually shipped. Extracted the verbatim live query and ran it:
  dartsonline `26917 / 1`, finetuning.uk `1649 / 2`. `PREPARE` alone proves it parses; the
  execution proves both gate inputs are real. **This was the single most valuable objection
  of the round** and it cost one query to close.
- three `DisallowUnknownFields` sites fleet-wide, **none on the git path** (guardian).
- `bugs_closed/072` holds no prior persist-at-render or shrink-floor proposal (prior_art).
- adapter HTTP client timeout is 20s; one GET per opted-in file per commit (guardian).

**And the one that corrected me.** The `architecture` seat: I argued RFC_022's *shape*
exception covered the new key. It does — and it is a different check from the *accumulation*
gate, which the owner's ruling defines separately and which fires on the count alone
(17 carriers, 10 → 11 keys). **I used one gate to answer the other.** The remedy is the
estate's own: an ack in `optional_key_budget_acks.json` at 11 pointing at the review,
`check.py` mirrored, parity test green, overlay re-applied and verified at the ConfigMap
with a control. With the caveat that keeps it honest: `git_commit` is uncounted, so that ack
has **no automatic enforcement behind it** — it is a recorded judgement, not a live baseline.

### Second misstep of the session, same shape as the first

Both of today's errors were **checks that could not fail**: the `LIKE '54[23]%'` ledger query
(no character classes in `LIKE`, so it returns nothing whatever is true), and the 542 verify
block (a literal-match assertion over a string I had just written, which cannot discover a
step I never considered). Different mechanisms, same failure mode — *an instrument that
agrees with me by construction*. The transferable habit is the one the working-docs rules
already state and I applied twice too late: **before recording a check as passed, name the
result that would have failed it.**
