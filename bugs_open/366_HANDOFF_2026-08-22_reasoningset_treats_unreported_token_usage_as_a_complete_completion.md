# 366 — `cmd/reasoningset` treats UNREPORTED token usage as a completed answer, so a cut row can enter the eval corpus unflagged

**Filed 2026-08-22** by the `bugfix_305_negation_gate` lane. **Not our tool** — filed because a
council seat was right that a finding left in a submission's prose is a finding that gets lost.

> **Provenance, stated plainly:** this came out of `bugs_open/305`'s reuse follow-up
> (council `a696e2a3`, APPROVED round 1). The `reuse_agent` seat raised it at **medium**:
> *"'that finding is reported to them' is not a tracked artifact (no work item, no doc_note) — it can
> be lost with nothing here to show it was ever raised."* This file is that artifact. The seat's
> founding failure mode is two paths solving one problem with nobody unifying them, which is exactly
> what an untracked report produces.

## 1. What is wrong

`cmd/reasoningset` decides which historical LLM calls may enter a reasoning eval/training corpus. Two
of its exclusion checks implement CLAUDE.md's rule that `output_tokens == max_tokens` means the
completion was CUT:

- `commonExclusions` — `cmd/reasoningset/main.go:684`
- `judgeInput` — `cmd/reasoningset/main.go:706`

Both are written the same way:

```go
if p.OutputTokens != nil && p.MaxTokens != nil && *p.MaxTokens > 0 && *p.OutputTokens >= *p.MaxTokens {
    return false, "truncated"
}
```

**The nil-checks are the defect, not the arithmetic.** When usage was never reported the guard is
skipped, so the row is NOT excluded — and "the provider told us nothing about this call" is silently
handled as "this call finished normally". Those are different claims. `commonExclusions` does catch
the case where the truncation produced an *error* (`error_message LIKE 'response truncated%'` on the
`Success == false` path), so the hole is specifically:

> **`success = true`, usage unreported, and the completion was actually cut.**

## 2. Why this is not theoretical

Truncation on this estate is real and current. `llm_call_log` for one step, `[MEASURED 2026-08-22]`:
**3 of 70** `rewrite_negations` calls were cut at exactly `output_tokens = max_tokens = 2000`. Those
three were caught structurally (`stop_reason=max_tokens` → a typed error → `success=false`), which is
the healthy path. The hole is the other shape, and this estate has already recorded that it exists:
**MDL-038's own verify-later note says "truncated calls can log `output_tokens` NULL, so that census
under-counts"** — i.e. a NULL-usage truncated row is a known species, and it is precisely the one
these two guards wave through.

## 3. The fix, and why it is one line each

`platform/aiservice.ClassifyTruncation` now exists for exactly this (register **MDL-043**, council
`a696e2a3`, committed 2026-08-22, inert until the next roll). It returns three states —
`TruncationAtCeiling`, `TruncationBelow`, `TruncationUnknown` — with **Unknown as the zero value**, so
the state that cannot be vouched for is the one you get by default rather than the one that says
"fine". `Truncated()` and `Complete()` are deliberately not complements, so `if !s.Truncated()` reads
as the mistake it is.

The **decision** this file cannot make for you: what SHOULD an unknown-usage row do here? Excluding it
shrinks the corpus by rows that are probably fine; including it admits rows that are probably cut.
That is an eval-quality judgement belonging to whoever owns the corpus, which is why this lane did not
patch it — changing which rows reach an eval set is not a refactor.

⚠ `cmd/reasoningset/main.go` also carries its own house rule right above the site:
*"a new lane must call this, not reimplement it"* — the same doctrine, one layer down. Whatever is
decided, decide it in `commonExclusions`/`judgeInput` rather than at a caller.

## 4. What has NOT been measured

`[UNMEASURED]` how many corpus rows are actually in the hole — `success = true` with NULL/0 usage.
Deliberately left to the owning lane, because the query's shape depends on which corpus build is
authoritative, and a number produced against the wrong extract would be worse than no number. The
query is otherwise trivial:

```sql
SELECT count(*) FROM llm_call_log
 WHERE success AND (output_tokens IS NULL OR output_tokens = 0);
```

Run it before sizing the fix; a zero result downgrades this to a latent hazard rather than live damage,
and that is a fine outcome to record.

## 5. Verification when it is fixed

Not "the code now calls the helper" — that is a source scan and this estate has been bitten by those.
Prove it at the behaviour: construct a provenance row with `Success=true` and nil `OutputTokens`, and
assert the exclusion decision is whatever was chosen, with a **mutation** that flips it. The helper's
own tests are mutation-proven (`platform/aiservice/truncation_state_test.go`) and are the pattern.

**Sources:** `cmd/reasoningset/main.go:684,706`; `platform/aiservice/truncation.go`; register MDL-043,
MDL-038; council report `a696e2a3-311b-4490-b862-f5cdfc1bc169` (the `reuse_agent` medium objection);
lane `docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/`.

---

## 2026-09-03 — FIXED AND COMMITTED. §4's `[UNMEASURED]` question is ANSWERED: the hole is EMPTY

Picked up unowned by the `bugfix_361_render_check_ratchet` lane (the 305 lane that filed this
closed on 2026-08-24; no live session on it). Guards re-read at HEAD first — unchanged, still
`main.go:684` and `:706` as filed.

### §4's question, answered — and the query in §4 gives the WRONG answer

⚠ **Run as written, §4's query returns 882. The correct answer is ZERO.** §4's own warning is
what catches it (*"a number produced against the wrong extract would be worse than no number"*),
and the query it is attached to **is** that number: it counts all of `llm_call_log`, while this
tool ingests only `step_name ~ '^(verdict|review_|propose|repropose|reframe)$'` plus
`score_relevance` (`extract.sql:48,215`). The 882 are `med-price-collector` (647) and
`business-intel` (177) rows that never reach the corpus. Scoped correctly `[MEASURED 2026-09-03]`:

| over the REAL corpus predicate | rows |
|---|---|
| corpus rows / successful | 2,191 / 2,013 |
| **`success` + usage unreported — the filed hole** | **0** |
| ...of those, with a ceiling recorded | 0 |
| `success` with **no ceiling recorded** | **161** |
| caught at the ceiling today | 2 |

**So this is a latent hazard, not live damage — exactly the outcome §4 said "is a fine outcome
to record".** Fixed anyway: it costs a few lines and the guard is what stops the population
being non-zero later.

### The decision §3 reserved, taken on evidence rather than by picking a side

§3 framed it as a trade: excluding unknown-usage rows shrinks the corpus by rows that are
probably fine, including them admits rows that are probably cut. **The measurement says it is
not a trade, because `TruncationUnknown` has two causes and only one is suspicious:**

- **a ceiling WAS recorded and the provider still reported no usage** — anomalous, since the
  ceiling proves the call went through the configured path. **Excluded**, with its own reason
  `usage_unreported` rather than borrowing `truncated`: we do not know it was cut, and §3's
  whole point is that those are different claims. Population **0**.
- **no ceiling recorded at all** — never answerable, no signal either way. **Not excluded.**
  Population **161**: 8% of the corpus's successful rows, every one a `score_relevance` call
  from the Mar–May logging regime that did not record `max_tokens`, averaging **3,092** output
  tokens, none with an empty response.

**A blanket "exclude everything Unknown" — the obvious reading of §3 — would have deleted those
161 substantial rows to guard a population of zero.** That is why the ceiling is tested
separately rather than leaning on `ClassifyTruncation`'s verdict alone; the shared classifier is
right that it cannot distinguish these, because the distinction is about LOGGING provenance, not
about the completion.

### Also fixed: the duplication that made this a two-place bug

§3's ⚠ note is right and it went further than the file says. `judgeInput` carried its own copy
of **three** rules `commonExclusions` already owned — the truncation guard, the blinded-marker
scan and the failure check — in a file whose comment on `commonExclusions` says *"a new lane must
call this, not reimplement it"* and records that the lanes had already drifted once. That copy is
why a one-line defect had to be fixed in two places. `judgeInput` now delegates and keeps only its
own `<no value>` rule (`bugs_open/016`), which still outranks the shared ones.

One visible difference, pinned by test: a row that is both blinded and truncated now reports
`blinded_docs` where it reported `truncated`. Rows are **flagged, not dropped**, so the admission
decision is identical and only the label moves.

### Tests and verification

First tests this tool has ever had (6 cases, `main_test.go`). Four mutations, each failing only
what it should — **restoring the old nil-guarded check fails the usage-unreported arm**, i.e. the
filed defect is caught by the test written for it; excluding all Unknown fails the 161-row
protection; un-delegating fails the convergence pin; admitting at-ceiling fails the arm that
already worked. Verified against HEAD `b55f837ef` with `scripts/verify-head-builds.sh --test`.

Out of council scope (`in_council_scope` refuses `cmd/reasoningset/`, with `platform/aiservice`
as a passing control), so no submission was owed.

**Stays OPEN** per the fixed-AND-live bar: this is an offline corpus tool with no image and no
CronJob, so "live" means the next corpus build runs it. Whoever builds one next should confirm
`usage_unreported` appears zero times and the row count does not drop by ~161.
