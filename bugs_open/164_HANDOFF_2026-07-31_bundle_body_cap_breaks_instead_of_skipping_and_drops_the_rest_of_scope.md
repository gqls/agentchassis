# 164 — the diagnosis bundle's body cap `break`s instead of skipping, so one oversized symbol silently drops the whole rest of scope

**Filed 2026-07-31 by the `bugfix_145` lane, AT THE COUNCIL'S EXPLICIT REQUEST.**

> **STATUS 2026-07-31 (evening) — FIX COMMITTED, NOT YET LIVE. Still OPEN, and it stays
> open until a chassis image carrying it rolls** (a fix committed but inert leaves the
> defect reproducible in production — that is the `/bugs_closed/` bar).
> Owned by `docs024_key_docs_latest/bugfix_164_bundle_body_cap/`.
> Council `SUBMISSION_CORR 75f3cd52-316c-4cb3-a55d-1b1c3f316214`.
>
> **The `[UNMEASURED]` marker below is RESOLVED, and the answer is that it has fired
> repeatedly** — see "MEASURED, 2026-07-31" after the Measured section. The filing was
> right to refuse to guess, and right to make measuring the first task.
>
> **Fix candidates 1 and 2 are built** (`continue` + inline per-symbol marker + a
> conditional coverage line), **plus the sibling read-failure path in the same loop**,
> which had the identical silence six lines up. Candidates 3 and 4 DECLINED with
> reasons — candidate 3 would *reduce* coverage in the common case; see the lane's
> `PLAN` §5. Verify with `VERIFY` at the bottom of this file after the next roll.

I found this while fixing `bugs_open/145` in the same function, disclosed it in my
submission's `risks` as "an adjacent pre-existing wart I am not fixing", and left it for a
reviewer to rule on. The council gate's **`bug_historian` seat objected to exactly that
choice** (corr `bce4caab`, round 1, severity **medium**), and it is right:

> "The plan's own risk section discloses, but declines to fix, an adjacent bug in the exact
> file/loop it is editing… This is not a hypothetical near-miss: it is a byte-for-byte match
> to the indexed transferable pattern **'A hard cap that silently discards its input's tail
> rewrites meaning — and the tail is whatever was composed LAST (2026-07-20)'** (016b §9),
> which is precisely the silent-drop-during-render shape this council exists to catch. The
> author found it while auditing the very function this edit modifies, named it accurately,
> and then left it for 'a reviewer to say so' rather than filing it."
>
> `missing`: "No work item or bug filing accompanies the disclosed maxBodyChars/break
> silent-truncation defect… it should be filed now (referencing the 016b §9 hard-cap
> pattern) rather than left as a reviewer's-discretion footnote."

**So this file exists because a review seat caught a thread (me) doing the thing the seat
was built to catch. Recording that plainly, because the seat's value is the reason to keep
paying for it.**

## The defect

`platform/orchestration/actions/diagnose_assemble_bundle_action.go:197-213` — the loop that
renders in-scope code into the diagnosis bundle:

```go
b.WriteString("## In-scope code\n\n")
total, truncated := 0, false
included := 0
for _, sym := range scope {
    body, err := analysis.ReadSymbolBody(repoRoot, anaOut, sym)
    if err != nil {
        logger.Warn("diagnose_assemble_bundle: could not read body", zap.String("symbol", sym), zap.Error(err))
        continue                                    // ← an ERROR skips and carries on
    }
    if total+len(body) > maxBodyChars {
        truncated = true
        break                                       // ← a CAP HIT abandons the whole rest of scope
    }
    fmt.Fprintf(&b, "### %s\n```go\n%s\n```\n\n", sym, body)
    total += len(body)
    included++
}
```

The two failure paths are inconsistent, and the more common one is the harsher: a symbol
that cannot be READ is skipped (`continue`) and its siblings still render, but a symbol
that merely does not FIT ends the loop (`break`) and **every remaining scope entry is
dropped**, however small.

## Why that is worse than it looks

- **Scope is SORTED, so what gets dropped is not random.** `nextScope`/`namedScope` end with
  `sort.Strings(next.Symbols)` (`pkg/diagnose/loop.go:390`, `:416`). So the casualties are
  whatever sorts last — an alphabetical tail, not a "least relevant" tail. One 60,000-char
  file under `internal/` can silently evict everything under `pkg/` and `platform/`.
- **The verdicter is never told.** `truncated` is set, but the bundle text gets no marker at
  this site — contrast `siblingSignatures`, which does write `_(further files omitted — cap
  reached)_` when *its* cap trips (`:604`). So the model reads a bundle that looks like the
  complete answer to its own `next_scope` request.
- **It corrupts the loop's control signal, not just its content.** The diagnosis loop's
  convergence guard measures whether scope NARROWS. A verdict formed on a silently truncated
  bundle names its `next_scope` from what it could see, so the loop can converge on the
  wrong region and record it as progress. This is the mechanism that makes it a *loop* bug
  and not only a rendering bug.
- **One oversized body is enough**, and `maxBodyChars` defaults to 60,000 while
  `ReadSymbolBody`'s whole-file branch is unbounded at the read (see `bugs_closed/145`) —
  and the bundle **advertises** the whole-file form to the model at `:597` ("put the bare
  file path in next_scope to see it whole"). So the platform invites the model to request
  the exact input that triggers this.

## Measured

`[UNMEASURED — deliberately, and here is why]` I have **not** established how often this
fires in production. The honest reason: attributing a past truncation needs the bundle
artefacts joined against the scope that produced them, and `truncated` is not surfaced per
run in a way I could query in the time I had. **Do not quote this file as evidence that it
has fired.** What IS established is the code path, the sort order, and the absent marker —
all above, all first-hand. The first thing the fixing thread should do is measure it:

```sql
-- bundles at/near the cap, as a rate, with their scope size
SELECT date_trunc('day', created_at) AS d, count(*) AS bundles,
       count(*) FILTER (WHERE length(body) >= 59000) AS near_cap
  FROM diagnosis_artifacts WHERE kind='bundle' GROUP BY 1 ORDER BY 1 DESC LIMIT 14;
```

(`length(body) >= 59000` is a proxy, not the flag — check `metadata` for a `truncated`/
`symbols_in_scope` vs `included` pair first; `:304` and `:328` log both, so the fields may
already exist on the artefact.)

## MEASURED, 2026-07-31 — the `[UNMEASURED]` marker above is RESOLVED

> **CORRECTION to this file's own guidance:** the suggested `length(body) >= 59000` proxy is
> wrong in **both** directions and should not be used. `body` is the WHOLE bundle (runtime
> evidence + schema + signatures), so a fat untruncated bundle scores as truncated; and the
> three *worst* real cases have `body_chars = 0`. `metadata->>'truncated'` is the real flag and
> was on the artefact all along — the fields the filing guessed "may already exist" do exist.

Live `clients_db`, all-history, reported as a **rate with its window** because
`diagnosis_artifacts` is retention-clocked (`bundle_retention_days` default 30) and a bare
`count(*)` is "still retained", never a census:

```
 bundles | truncated | pct |   first    |    last
     254 |        18 | 7.1 | 2026-07-09 | 2026-07-31
```

Per-bundle damage (`symbols_in_scope` − `symbol_count`), worst first:

| corr | iter | in_scope | included | dropped | body_chars |
|---|---|---|---|---|---|
| `c16ee494` | 5 | 18 | 4 | **14** | 57,395 |
| `954d8da9` | 2 | 18 | 4 | **14** | 50,831 |
| `2a656f25` | 2 | 18 | 5 | 13 | 56,330 |
| `65103331` | 4 | 7 | **0** | 7 | **0** |
| `f9bcee6f` | 4 | 7 | **0** | 7 | **0** |
| `f9bcee6f` | 5 | 7 | **0** | 7 | **0** |

**The three `included=0, body_chars=0` rows are the headline.** That combination can only mean
the FIRST body alone exceeded the 60,000-char cap, so the loop broke before rendering anything.
Note `f9bcee6f` hit it on iterations 4 **and** 5 — twice in one run, and the loop converged
regardless. Read at the artefact rather than inferred from the counter:

```sql
SELECT substring(body from position('## In-scope code' in body) for 220) FROM diagnosis_artifacts
 WHERE kind='bundle' AND (metadata->>'truncated')::bool AND (metadata->>'symbol_count')::int=0;
```

returns, for all three:

```
## In-scope code

## Same-file signatures (siblings of the in-scope symbols — …)
```

A heading promising the in-scope code, then straight to the next section. **The verdicter was
handed an empty evidence section with nothing saying why**, and could not distinguish "no code
in scope" from "seven symbols dropped".

### Blast radius, measured rather than left for a reviewer

- **Three char-budget cap sites exist in the whole repo** and all three are in this one file:
  `:208` (this bug), `:521`, `:605`. **The other two already write a marker before breaking.**
  This loop was the sole deviation from a convention its own file established twice — so the
  fix is a reuse of that convention, not a new mechanism, and needs no architecture seam.

  > **CORRECTED 2026-07-31, by this bug's own council round — "three sites repo-wide" is true
  > only for the SHAPE I greped.** The pattern keyed on `+ len(x) > cap`, which structurally
  > cannot see a **count** cap or a **slice reslice** (the reslice form has no comparison
  > operator at all). The `bug_historian` seat asked for exactly that complement; re-run for it,
  > `diagnose_load_runtime_action.go:945` is a genuine fourth instance — **now
  > `bugs_open/172`**, and it sits inside the file `bd003f67a` had cleared by name. The
  > remaining count-shaped sites were triaged and are sound: `diagnose_route_action.go`
  > (`:359`,`:388`,`:451`,`:466`) reports its drops, `diagnose_read_repo_files_action.go:136`
  > fails LOUD with an error, and `diagnose_run_checks`/`diagnose_code_lookup`/
  > `diagnose_load_runtime:454` still report as `bd003f67a` recorded.
  > **State the pattern beside the conclusion** — see 016b §9 *"an audit clears the file it
  > searched, not the file it read"*, which this correction is the source of.
- **One live consumer**: `diagnose-agent` is the only active agent invoking the action; it reads
  neither `bundle.truncated` nor `bundle.symbol_count`, and does not override `max_body_chars`
  (so production runs the 60,000 default). Nothing in Go reads the flag either. The only
  consumer of the change is the verdict LLM reading the bundle TEXT.
- `bd003f67a` (2026-07-20) audited **this file for this shape** on a `bug_historian` objection,
  found `workflowRefsFromRuntime`, and recorded "audited platform-wide by SHAPE rather than by
  instance". **It did not examine this loop, 300 lines above, in the file it was editing.**

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **`continue`, not `break`, plus a per-symbol refusal line in the bundle.** Skip the body
   that does not fit, keep going, and write `### <sym>\n_(body omitted — %d chars, cap %d)_`
   where it would have gone. The model then sees exactly which symbols it asked for and did
   not get, and can re-ask for them singly. **Makes the bad state unrepresentable:** there is
   no longer a silent difference between "not in scope" and "did not fit".
2. **Report `included` vs `len(scope)` in the bundle text**, not only in the log line at
   `:304`. A count the verdicter can read is the difference between a truncation it can
   route around and one it cannot see.
3. **Budget per symbol rather than first-come-first-served** — `maxBodyChars/len(scope)`
   with slack, the allocator `siblingSignatures` already uses for its per-file share
   (`:590-600`). Reuses an in-file convention instead of inventing one, and stops the
   alphabetically-first symbol spending the whole budget.
4. Bound the read itself so a single body cannot approach the cap. Weakest, and it is the
   candidate `145` already declined for being a knob rather than a boundary — but worth
   noting the two bugs meet here.

Candidates 1 and 2 together are ~10 lines and testable without the cluster.

## How to verify a fix

- A scope of `[big.go, small.go]` where `big.go` alone exceeds the cap must still render
  `small.go`, and must name `big.go` as omitted **in the bundle text**, not only in a log.
- Negative control: an unchanged scope that fits entirely must produce a **byte-identical**
  bundle to before the fix — otherwise the change moves every existing diagnosis's baseline.
- Assert the mechanism fired: a test that passes when the cap is never hit is asserting
  nothing. Force the cap, then assert on the omission marker's presence AND on the
  subsequent symbol's body being present.

## Related

- **`bugs_closed/145`** — same function, same loop; the boundary fix that exposed this.
  `145`'s `risks` section is where I first named it, and its `NOTES` records the decision to
  split rather than fold it in.
- **016b §9**, "A hard cap that silently discards its input's tail rewrites meaning — and
  the tail is whatever was composed LAST (2026-07-20)" — the indexed pattern this matches.
  Add this instance to that entry when fixing.
- Family: `bugs_open/012` (an `output_tokens == max_tokens` truncation persisted as success)
  and MEMORY's *"a `complete` work item is not a repaired artefact"*. Every member of this
  family is a cap that reported success.

## VERIFY after the next chassis roll (this is what closes the ticket)

1. **Prove the binary carries it, not the tag** (`bugs_open/153`: a roll is not evidence, and
   the image has no provenance). Grep a string the change ADDED plus a positive control in the
   SAME exec, so a wrong-pod or wrong-path mistake cannot read as a pass:

   ```bash
   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
     'strings /app/agent-chassis | grep -c "body omitted"; strings /app/agent-chassis | grep -c "In-scope code"'
   ```
   Expect `≥1` and `≥1`. A `0` on the first with `≥1` on the second means the image predates
   the commit. Do it on **both** replicas.

2. **Induce the cap on a real run** — the mechanism must be seen to fire, since a green run
   proves nothing here (every assertion below is vacuous when no body is oversized). Fire a
   diagnosis whose `seed_scope` names a bare file path for a large analysed file (the bundle
   advertises that form at `:597`, and the whole-file branch is unbounded), then:

   ```sql
   SELECT (metadata->>'symbol_count')::int      AS included,
          (metadata->>'symbols_omitted_size')::int AS too_big,
          (metadata->>'symbols_unreadable')::int   AS unreadable,
          body LIKE '%body omitted%'          AS marker_present,
          body LIKE '%This section is INCOMPLETE%' AS coverage_line
     FROM diagnosis_artifacts WHERE kind='bundle' AND correlation_id='<corr>'
    ORDER BY iteration;
   ```
   Required: `too_big ≥ 1`, `marker_present`/`coverage_line` **true**, and — the regression
   assertion — a symbol sorting AFTER the oversized one still present with a body.

3. **Negative control on the same roll**: a bundle that hit no cap must contain none of
   `body omitted` / `body unavailable` / `This section is INCOMPLETE`. If a clean bundle carries
   them, the conditional went unconditional and every diagnosis's baseline has moved.

⚠ **`symbols_omitted_size` and `symbols_unreadable` are NULL on every bundle written before
this ships** — older rows return NULL, not 0. `COALESCE` or filter on `created_at`, or a
"drop to zero" will really be an absence of data.

⚠ **Do not "fix" `TestBundleBodyCap_FittingScopeIsByteIdenticalToThePreFixFormat`**: it passes
against the pre-fix code too. That is what makes it a control — it asserts the OLD bytes survive
for a scope that fits. The other three tests are the ones that fail pre-fix (verified by
reverting the action to HEAD and re-running).
