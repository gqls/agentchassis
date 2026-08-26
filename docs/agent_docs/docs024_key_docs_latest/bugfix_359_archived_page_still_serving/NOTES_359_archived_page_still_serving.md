# NOTES — bugs_open/359, a retired page that is still serving

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-26 — lane opened; ownership checked, bug re-validated at the artefact

### Why this bug, and the ownership check that chose it

Swept `bugs_open/` (131 files) for a bug that is genuinely OPEN, genuinely UNOWNED and
framework-shaped. Most of the low-numbered "OPEN" files are **fixed and live** and stay in
`bugs_open/` only by the owner's 2026-08-06 ruling — 181, 207, 210, 217, 220, 221, 222, 230,
233, 244, 247, 249 all read that way on inspection, so "OPEN" in the index is not a
work-available signal and the status header has to be opened.

`348` looked like the next candidate by number and is **superseded**: its own banner says
*"Do not fix from this file — `bugs_open/344` OWNS THIS"*, and 344 is in `bugs_closed/`.
Recording that here so the next sweep does not re-pick it.

Three checks agreed on 359:

| check | result |
|---|---|
| `scripts/who-owns.py 359` | `likely OWNING workstream(s): (none identified)` — the only commit touching the file is the filing commit `3f2891c75` (2026-08-22) |
| live transcripts, `find -newermt 2026-08-25` | one hit, in the FILING lane's own close-out ("filed `bugs_open/359`, ready to close") — nobody fixing |
| `site_work_items` open rows matching `archiv`/`retract` | none of this shape |

### The bug is STILL VALID — re-measured today at the artefact, with controls

`scratchpad/census_359.sh` (kept in the RUNBOOK): every `status='archived' AND deployed_at IS
NOT NULL` page, probed live, with **two controls per domain** — an invented URL that must be
non-200 (a parked/catch-all domain 200s every path) and a known-good `active`+`deployed`
sibling that must be 200 (a dead origin makes every target read "correctly absent", which for
THIS question is a false all-clear).

`[MEASURED 2026-08-26]` **39 archived+deployed pages; 7 serving 200; 32 correctly 404.
Every domain's invented control = 404 and every domain's sibling control = 200**, so both
readings are real in both directions.

```
ai-agent-orchestration.com  /llm-cost-calculator.html
finetuning.uk               /tools/password-entropy.html          (archived 2026-08-25)
fundamentallyai.com         /blog/ai-readiness-checker-guide.html
fundamentallyai.com         /tools/llm-cost-calculator/index.html
leopardessconsulting.co.uk  /our-approach.html
robot-hands.com             /gripper-catalog.html                 (serving since ≥2026-08-14)
robot-hands.com             /news.html
```

**The population MOVES, which is itself a finding.** The bug filed `loancalculator.co.uk`
`/blog/loan-faqs.html` and `/blog/jargon-buster.html` as serving on 2026-08-22; both 404 today.
Five of today's seven are new since that sample. So this is not a fixed backlog of three pages
to sweep — it is a flow with no meter, and a one-off cleanup would read as a fix and leave the
mechanism untouched.

`robot-hands.com/gripper-catalog.html` is the load-bearing datum and it has grown: 30,997 bytes
recorded in `bugs_open/266` on 2026-08-14, still serving 2026-08-26. **Twelve days**, and
nothing raised anything.

### Misstep 1 — `kubectl exec -i` inside a `while read` loop ate the loop's own input

First census run printed exactly ONE row and exited 0, which reads as "only one archived page
exists". It is not: the loop was `done <<<"$ROWS"`, and the per-domain sibling lookup inside the
body is `kubectl -n … exec -i … psql`, whose `-i` consumes stdin — i.e. the rest of the
here-string. The loop then saw EOF after one iteration.

**Why it is worth writing down:** the failure is silent and exits 0, and the truncated output is
a plausible answer to the question being asked. Any `while read` loop whose body shells out to
`kubectl exec -i`, `ssh`, or `psql` has this bug.

The fix is both halves: `</dev/null` on every in-loop `kubectl exec -i`, AND feed the loop from
a process substitution (`done < <(printf '%s\n' "$ROWS")`) rather than a here-string. The tell
that caught it: the row count did not match the `SELECT count(*)` I had run one command earlier.
Logged in `WRONG_CALLS.md`.

### Architecture read (first-hand, today) — what a fix has to fit

- Checks live in `platform/orchestration/actions/discovery_checks/`, 77 non-test files, each
  registering itself from `init()` against `registry.go`'s `DiscoveryCheck` interface. Per-SITE.
- Enablement is **DB config**: the check's name goes in
  `agent_definitions.default_config.workflow.steps.<step>.config.checks`. The runner **hard-fails
  on a name the binary does not register**, so the migration must be held until the image rolls.
- `availability-discovery-agent` (`run_checks`: `site_unreachable`, `page_content_divergence`) is
  driven by the `site-discovery-rotation-availability` scheduled task: **enabled, 300s, one site
  per pass, each site at most every 4h** — a live recurring driver, which is the thing
  `bugs_open/230` had to build for the other rotations.
- Outbound probing from the discovery path is already precedented **three times over**
  (`check_backend_unreachable`, `check_backend_entry_orphaned`, `check_tool_acceptance`) and
  `verifier_coverage_test.go` records the standing objection that only the COMPLETION path is
  closed to probes, with the sanctioned alternative: self-clear via `CheckResult.Resolved` on
  the discovery path.
- `redirects` holds **0 rows fleet-wide** `[MEASURED 2026-08-26]`, so "retired behind a 301" is
  not a live pattern to reason from — only one to leave room for.
- None of the 7 serving pages has a same-`url` sibling row, so the retraction action's
  active-page collision guard would not refuse any of them `[MEASURED 2026-08-26]`.
