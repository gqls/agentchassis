# HANDOFF — bug 122 lane. START HERE. Written 2026-08-12.

Supersedes `HANDOFF_2026-08-11_continue_here.md` for **state**. That file stays the reference
for the sweep episode (cost table, the two corrections, the guard arithmetic) and its banner is
now marked RESOLVED. The 08-10 file remains the delivery evidence for what originally shipped.

**Nothing in this lane is on fire.** The queue is empty, no site is locked out, and the one
running mechanism costs nothing until 2026-08-16. Read §1, then go to §3.

---

## 1. State, measured 2026-08-12 12:39Z on chassis `v1.0.1290`

| thing | state |
|---|---|
| `page_rerender` queue | **drained**: `triaged` 446 → **0**, `complete` **2,803** (+542 overnight), `failed` 66 → 15, `detected` 12 |
| sites locked out of the sweep's guard | **0 of 22** (was 8 at the moment of stopping) |
| `improvement-sweep` | **disabled** 2026-08-11 18:00:39Z after 5h29m. Do not re-enable without reading §2 |
| `site-discovery-rotation-quality` | **enabled** at 10800s (migration `395`) — and **inert until 2026-08-16**, see §2 |
| `design` / `completeness` rotations | still disabled, deliberately |
| `site-render-audit-rotation` | enabled, untouched, measured zero LLM spend — this is what finds our contrast failures |
| 226 `contrast_failure` items | still **parked** at `deferred`, all stamped `parked_by='migration_389'`, none outside the park |

Provenance note: the `build provenance` startup line had already rotated out of `--tail=3000`,
which is normal on this service. Both replicas are on `v1.0.1290`, started 2026-08-11 21:53Z.

---

## 2. The owner's two decisions of 2026-08-12, and what they mean

**Decision A — detection: "enable the discovery rotations, slowly."** Not raise the sweep's
guard, not make its discovery step conditional. Done as migration `395`: `quality` only,
`enabled=true`, cadence `3600s → 10800s`.

**⚠ THE THING THAT WILL MISLEAD YOU: it is enabled and it will dispatch NOTHING until
2026-08-16 09:49Z.** All 22 active sites were already stamped for `quality-discovery-agent`
between 2026-08-09 09:49Z and 2026-08-10 16:39Z (the window it was briefly on before the owner
switched it off), and the due window is 7 days. A fire with no due site dispatches nothing,
costs nothing, and advances `last_triggered_at` exactly like a working one. **Do not "fix" this.
It is correct and it is waiting.**

Why the cadence is not the cost dial, which is the other thing to understand before touching it:
the `pre_query` is `LIMIT 1` per fire against a 7-day window **and stamps
`last_selected_at = now()` in the same statement that selects the site**. So the work rate is
bounded by the period — 22 sites / 7 days ≈ **3.1 passes/day** however often it polls — and the
interval only sizes the initial ramp (3600s → whole fleet in ~22h; 10800s → ~8/day). That
stamp-on-selection is also why this mechanism does not suffer the sweep's `ORDER BY
sites.updated_at ASC` starvation (IMP-010): it cannot re-pick a site it just examined.

**This restores DETECTION, not repair, and that is deliberate.** Findings are born `detected`
and the only triage carrier (`improvement-sweep`) is off, so they accumulate unconsumed. Two
consequences to carry: `detected` rows **count toward the sweep's ≥50 guard**, so anyone
re-enabling the sweep later must re-read the guard census knowing this rotation has been raising
the numbers it reads; and our 226 parked rows are unaffected, because `deferred` is not
`detected`.

### 📅 2026-08-16 — the dated action this lane owns

When the ramp starts, price it and check its fairness. Both queries, the one-UPDATE stop, and
the statement that adds the other two rotations are at the foot of
`docs/agent_docs/sql_for_agents/395_enable_quality_discovery_rotation_slow_ramp.sql`. Two rules
from the sweep episode apply directly:

- **Count calls AND tokens.** Calls/hour did not discriminate the sweep's 3.2x, because each
  fire is a whole site pass. Baseline to compare against: **~248k input tokens/h** with nothing
  running; the driven sweep was **~806k/h**.
- **A cost estimate is not a measurement.** The `[ESTIMATED — UPPER BOUND]` figure in `395`
  (≤ ~180k input tokens per pass, derived from sweep fires which also did triage and dispatch)
  exists to be replaced by the real number, not quoted.

**Why there is no measurement yet, on purpose:** I considered backdating one stamp to price a
single pass immediately and decided against it. The 08-16 ramp yields the *same* single-pass
figure for free; a probe buys lead time only, and buying lead time with an unrequested LLM
spend four days after a 3.2x cost surprise is not this lane's call to make. What waiting costs
is **visibility**, not data — which is why the date is written here rather than solved by
spending.

**Decision B — next code task: register a `contrast_failure` verifier.** That is §3.

---

## 3. NEXT: the `contrast_failure` verifier (unblocks the 226)

**Why this and not `bugs_open/242`:** it is the only thing standing between the 226 parked items
and release, and the reason the park exists changed on 2026-08-12. It is *not* waiting on
`bugs_open/213` any more — see §4.

### What is established (read, not inferred)

- `contrast_failure` has **no registered verifier**. It is filed at
  `write_render_audit_findings_action.go:258` and appears at **no** `RegisterVerifier*` call
  site. The 12 registered types are `orphan_element_refs`, `truncated_component`,
  `unbuilt_internal_link`, `revenue_shape_cta`, `missing_conversion_path`,
  `content_duplication`, `empty_section`, `page_canonical_collision`, `literal_markdown`,
  `hardcoded_section_colors`, `decision_regression`, `dead_fragment_link`.
- Consequence, which is the whole point: `verifyBeforeComplete` resolves via
  `checks.GetVerifier(itemType)` (`complete_work_item_verification.go:70`) and an
  **unregistered type completes untouched** — documented at `:16`, no `_verification` key
  written. So promoting the 226 today mints 226 completions that are **ungraded by
  construction**. Confirmed empirically on a sibling type: 14 `dark_section_audit` items
  completed post-roll on 08-11, **0 of 14** carry a `_verification` key, against **9 of 9** for
  `hardcoded_section_colors`.

### The template to copy

`check_hardcoded_section_colors.go:60-89` — the first opt-in to the `Grades` remit test, and the
closest analogue. Two things to lift from it:

- `RegisterVerifierWithPolicy("<type>", VerifyFn, VerifierPolicy{Grades: remitTest})`
- **The remit test is a POSITIVE SHAPE MATCH on the spec, never a producer blocklist.** Its own
  comment explains why and it is the right reasoning: live agent config can mint a producer with
  no code change, so a producer list cannot cover a producer that does not exist yet.

### The spec is fully specified — all 226 rows carry every key

```
selector "A.A" · page_name "/blog/….html" · affected_url "https://…" ·
fg "rgb(26, 31, 46)" · bg "rgb(15,18,24)" · ratio 1.14 · need 4.5 · font_px 17 ·
text_sample "…" · category "contrast" · fix_type "contrast_fix" · run_id … ·
acceptance_test "computed contrast for elements matching A.A on /… is at least 4.5:1 —
                 a single-selector, single-page measurement, not a site re-audit"
```

`acceptance_test` already states the predicate in words, which makes the remit test easy: match
on `category='contrast'` **and** the presence of `selector` + `page_name` + `need`. Verify that
partition could come out otherwise before trusting it — census the keys per producer, the way
`gradesHardcodedColourAggregate`'s `[MEASURED]` note does, rather than asserting it.

> **⚠ TRAP, found while reading the specs: `fg` and `bg` are formatted INCONSISTENTLY** —
> `"rgb(26, 31, 46)"` has spaces after the commas, `"rgb(15,18,24)"` does not, **in the same
> row**. Any verifier that string-compares a recorded colour against a computed one will report
> a false `defect_persists`. Parse to numeric triples; never compare these as strings.

### The one real design fork — decide this before writing code

The measurement must be a **computed** one. A source-level read cannot settle contrast: a CSS
`var(--x,#fallback)` literal is present in the source and never applied (own landmine), so
grepping the stylesheet answers a different question from the one the item asks. The verifier
runs in Go, in-process, on the completion path; the contrast tool this lane used is **Python**
(`cmd/contrastscan` does not exist — register VIZ-010). So:

1. **Route verification through the same browser path that produced the finding**
   (`render-audit-agent` / browser-runner). Highest fidelity, and it grades exactly what the
   audit graded. Cost: an outbound browser run on the completion path — and
   `verifier_coverage_test.go:171` records a **standing objection** to putting an outbound probe
   on the completion path. Read that objection before choosing this; it may already decide it.
2. **Re-implement the computation in Go against the served page.** No new dependency on the
   completion path, but it is a second implementation of a measurement we already have, and a
   disagreement between the two would be invisible.
3. **Narrow, honest predicate:** assert the specific recorded pairing (`fg` on `bg` for
   `selector`) is no longer what the page computes for that selector. Weaker than the full
   acceptance test, but it refuses the false-complete case, which is the whole reason the park
   exists.

**Recommendation: cost (1) against that standing objection first**, because if the objection
holds, the fork collapses to (3) and the work is small. Do not start by writing the verifier.

### Process obligations for this change

Platform code touching a **shared registry**, so:

- **Register it** in the concept register (a new callable verifier is exactly the bar) and drop
  any matching line from `102_coverage_ratchet.txt`.
- **Council gate**, before or alongside the commit. Use `Council-Submitted: <corr>` if you
  commit before the verdict lands; **never** write `Council-Reviewed:` on a verdict you have not
  read. The `commit-msg` trailer gate rejects a non-UUID value — a prose trailer like
  "not applicable" is blocked, so on a docs-only commit just omit it.
- **Image before behaviour**: Go is inert until a chassis rebuild and roll. `make build-<svc>`
  builds from committed HEAD.

---

## 4. What changed about the park, and what 213's lane has been told

The park's trigger was **wrong** and is corrected. I had told `bugs_open/213` "they unpark when
you close this"; withdrawn in their file on 2026-08-12. `contrast_failure` has no registered
verifier *either*, so closing 213 does not make promotion safe. **The real trigger is "a
`contrast_failure` verifier exists, or someone rules the type needs none".** 213's lane has been
told explicitly that they are no longer blocking us, so they do not hold their closure for our
226.

The restore itself is unchanged: one UPDATE at the foot of migration `389`, predicated on
`spec->>'parked_by' = 'migration_389'` so it cannot disturb another session's deferred rows.
Row-level backup at `scratchpad/backups/backup_park_contrast_failure_20260811.tsv`.

**Contribution left in 213's file, worth knowing if you touch verifiers:** the design-audit
producer straddles *both* holes — one item_type registered (theirs, fixed) and one
(`dark_section_audit`) unregistered, where completion is untouched by construction. And
`VerifierDeclaresRemit` returns false for an unregistered type (`verifiers.go:174`), so
`cmd/verifier-remit-check` will not flag it — the class detector is blind to the class it most
resembles. Whether `dark_section_audit` *should* be verified is marked NOT ESTABLISHED and is
their call or the owner's, not ours.

---

## 5. Also still open in this lane

| | item | status |
|---|---|---|
| 1 | `bugs_open/212` §8 — component-painted grounds (~24 failures) | **Owner's.** Architecture, not a bug patch. Unchanged |
| 2 | `bugs_open/242` — the silent 25-page cap | **Open, unstarted.** Cheapest fix is parity with the drain one step downstream: put `pages_total` into the summary the adapter returns, so a capped sweep cannot read as a complete one. All 19 overnight sweeps measured **at most 25 pages** and none said whether that was all — **226 is a floor, not a census** |
| 3 | Free cross-check | if a lane re-renders robot-hands `/selection-guide.html`, the audit filed `info-card-grid__card-link` + `__eyebrow` failures there and migration `368` should close both. **Grade at the next audit, never at the item status** |

---

## 6. Standing traps this lane has paid for

- **Grade per selector, never by fleet total.** It rose 109 → 112 while every targeted failure
  closed.
- **A filed count is not a found count.** "34 findings" was 171 firm — 111 dropped by a cap.
- **Read the selection before asserting it excludes your rows.** Two wrong causes written down
  for the 220 stalled items; each died to one further query.
- **A pathspec commit still takes a same-file passenger**, and the check is per-commit, not
  per-session. It happened again on 2026-08-11: a `WRONG_CALLS.md` entry of mine went out inside
  another session's commit `0af98afb2`.
- **`pages.sections` is an array of plain strings**; an object-shaped census returns 0 rows
  silently.
- **NEW — never run `run-migrations.sh --apply` on this tree.** The 2026-08-12 dry run listed
  **12 pending files from other threads**, including one whose guard refuses because on an older
  binary it *"deploys the WRONG asset bytes"*. `--apply` takes every pending file in order.
  Apply your single file by hand (`psql -v ON_ERROR_STOP=1 -f -`), then
  `--record-only <file> --note "<why>"`.
- **NEW — a call count does not price an LLM loop.** The sweep held 93–184 calls/h (a no-sweep
  busy hour was 134) while running **3.2x** the input tokens, because each fire is a full site
  pass. Threshold the expensive unit, not the cheap one.
- **NEW — put a row COUNT you could be wrong about in your post-check.** Migration `395`'s
  post-check printed `0 site(s) due`, which is the only reason anyone knows the rotation is
  inert for four days. Asserting "the UPDATE succeeded" would have reported success.
