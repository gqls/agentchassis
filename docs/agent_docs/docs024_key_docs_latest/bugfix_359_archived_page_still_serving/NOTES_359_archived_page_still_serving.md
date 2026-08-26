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

---

## 2026-08-26 (later) — the census is a command now, and three measurements that shape the fix

### `scripts/audit-archived-still-serving.sh` — candidate 1, discharged

The census is in the repo rather than in this file, because the two controls are the whole
value and a paragraph describing them is not runnable. `--self-test` proves all eight verdict
rows with no cluster. Exit codes follow the check-fleet convention and **a control failure is
exit 2 — a refusal, never a pass.**

It also carries the control that caught my own misstep this morning: it computes the population
size with one query and the audited-row count from the loop, and **refuses if they disagree**.
A loop that shells out can be truncated silently at exit 0, and a census that reports only what
it found cannot be told from one that stopped early.

First run, `[MEASURED 2026-08-26]`: **population 39 · 7 archived AND serving · 32 correctly
absent · 0 unjudgeable.** Exit 1. Reproduces the hand census exactly.

### Measurement A — none of the seven is reachable from its own site

```
site_nav_items rows: 0 for all 7        link_registry inbound: 0 for all 7
```

So every one of them is orphaned from the site's own structure and reachable only by direct URL
or by a search engine that indexed it before we retired it. That is the shape that makes this
class invisible: the site itself looks correct, because from inside the site the page is gone.

### Measurement B — none of the seven is listed in its site's `sitemap.xml`

Checked live against each site's own sitemap (35/25/51/37/41 `<loc>` entries respectively): **0
of 7 appear.** Worth having, and it lowers the severity honestly — we are not actively inviting
indexing, we are failing to withdraw something already indexed. `medium`, not `high`.

### Measurement C — the re-deploy seam is already closed, so a retraction will stick

`bugs_open/266`'s `ARCHIVED_PAGE_GUARD` is live at both deploy seams
(`git_deployer_actions.go:81,103` and `v3_site_actions.go:899,911`, with
`archived_page_guard.go` shared between them). That matters because LANDMINES records that
retraction used to be **self-undoing** — delete the file and the next refresh republishes it,
while a post-delete `curl` still shows 404 at the moment you look. It is not self-undoing now.

### Adjacent finding — the guard that will protect my own change is 23% blind

Enabling a check means adding its name to `agent_definitions`, and an **unregistered name
hard-fails the whole discovery step** (`discovery_checks.go:198-216`, `bugs_open/149` B4) and
takes the run's already-collected findings with it, because the return precedes `tx.Commit()`.
The safety proof for making that fatal is `liveConfiguredChecks` in
`discovery_checks_registration_test.go` — a hand-maintained fixture of every name the live
agents are configured with.

`[MEASURED 2026-08-26]` **the live agents configure 82 distinct check names; the fixture asserts
63.** Nineteen are missing, including every name of a **fifth** agent
(`acceptance-discovery-agent`: `build_prerequisites`, `heading_promise`, `structure_floor`) and
`page_content_divergence`, the second half of the very agent I intend to host this check on.

I verified all 82 resolve today (dumped `discovery_checks.Names()` from a throwaway test and
diffed), so **there is no production risk right now** — this is a pure under-assertion. But the
fixture's own header says it exists so that a rename fails *in the test* rather than in the
fleet, and it currently cannot see 23% of the roster. The file has been here before: it records
finding `literal_markdown` live and unlisted and says leaving a known gap "would be the same
defect one level up". Refreshing it by UNION is part of this lane's commit.
