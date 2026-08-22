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
