# NOTES — bugs_open/150 (append-only, newest at the bottom)

## 2026-07-31 — picking the bug

Started from "take the next bug nobody is working on". With ~30 concurrent sessions,
`who-owns.py` answers **OWNED or recently active** for essentially every candidate — it
fired on 7 of the 7 I tried — so on this tree its verdict line is close to non-discriminating
and the useful part is the *body*: which workstream, how many mentions, which commits. What
actually separated the candidates was grepping the live `.jsonl` transcripts for each bug's
**code symbols**, not its number.

Ruled out, with the evidence:

- **071** (validate gate discards link findings) — session `90e5c832` had **71** hits on
  `validate_page_content`, 31 on `phantom_link`, 2 on `writeValidationFailureLog` in its
  live transcript tail. That is 071's exact surface, being worked right now.
- **093** (stat audit, one guarded call site) — already built and live; its own tail says it
  is blocked on `bugs_open/083`'s missing cadence, not on code.
- **128** (`image_url_404`) — its own header records the obvious fix as **measured and
  refuted**: only 9 of 79 masked paths have an assets row carrying the basename while 73 of
  79 serve 200, so a path predicate would flag ~70 working images. Nothing in the DB records
  which static files deployed, so the check cannot answer its own question without new
  plumbing. Not a session's-worth of clean work.
- **132** (raw B2 JSON instead of a 404 page) — the fix is Cloudflare-side. Worth recording:
  the file says *"no wrangler.toml, no worker JS anywhere under ~/projects/sites or
  ~/projects/agentchassis"* and **`scripts/cloudflare/worker.js` does exist** — but it
  returns `new Response('Not found', {status: 404})` on a B2 miss, which is **not** what the
  live edge serves, so the repo copy is stale rather than authoritative and the bug's
  conclusion stands. Contributed back to 132 as a dated note.

## 2026-07-31 — verifying 150 was still real

Read the live rows, not the seed. All three load-bearing claims held: three agents carry the
promoter, the promotion is site-wide and unfiltered, and `check_has_findings` still reads
`triage_result.has_items == true`.

Two things the file did not have, found by reading around it:

1. **Four live consumers of `has_items`, across three actions** — `build-dispatch-loop`,
   `improvement-loop`, and `site-work-orchestrator` twice. Three of them read their own
   loader's output and are correct. That measurement is what decided the fix shape: adding a
   key rather than redefining one.
2. **A second route to the same false claim.** `check_audit_pass_limit` sends a site with
   `get_audit_pass_count(site) >= 3` **straight** to `complete_clean` — no discovery, no
   triage, and the site is told it is clean. `[MEASURED]` 0 of 25 sites are at the limit
   today, so it is latent, not live. Recorded, not fixed.

## 2026-07-31 21:12 — the control run, and what it corrected

Fired one sweep at vetcomparison.uk (0 `detected`, 12 actionable at fire time) on the
**pre-fix** binary, v1.0.1218. Orchestration `911ecdd8-140f-402f-99fd-aa89700afed2`.

```
call_design_audit.response.triage_result = {"promoted": 24, "has_items": true}
call_site_review.response.triage_result  = {"promoted":  3, "has_items": true}
triage_result (the parent's own)         = {"promoted":  0, "has_items": false}
current_step = complete_clean            status = COMPLETED
27 rows triaged in the run's window;  0 improvement_rerender% items created
```

**The bug reproduced exactly, on a second site, on a second day.** The file's
`[INFERRED from a single run]` marker is discharged: this is the second observation, and
`orchestration_states` retention had destroyed every earlier one.

**It also corrected the file.** §Confidence names
`site-review-agent.write_strategic_findings` as the one escape hatch — a step that could
create `detected` items *after* the last child triage — and says it created none in the
observed run. Here it **did** create some: site-review promoted 3 of its own. The parent
still saw 0, because the child triages after it writes. So the hatch opens and does not
help, and the defect is more robust than the file allowed for.

Worth keeping: the site had **three other orchestrations running concurrently** (page
builds), so querying `orchestration_states` by `site_id` surfaced the wrong runs. Filter by
`owner_agent_type='improvement-loop'`.

## 2026-07-31 — missteps

- **I planned migration number 279 from an `ls` taken ~40 minutes earlier.** By the time I
  wrote the file, other sessions had taken **279 and 280**. Caught by re-running the listing
  in the same command that created the file; the migration is **281**. Same class as the
  session-start `git status` going stale, one directory over. Logged in `WRONG_CALLS.md`.
- **Same class, second instance the same hour:** the plan said vetcomparison.uk had 2
  actionable items; at fire time it had 12. Nothing depended on the number, but it is the
  same lesson — a count read during planning is not a count at execution.
- `/tmp` is a 16G tmpfs at 93%, ~11G of it other sessions' scratch. One command died with
  ENOSPC mid-investigation. I deleted nothing that was not mine; keep tool output small when
  it bites.

## 2026-07-31 — what shipped, and what did not

Committed `337fdd9af` (Go half + six tests + register WDS-015), trailer
`Council-Submitted: 757cc7be-8551-4e43-9d1e-705b0977be1d`. Migration **281 written and
deliberately NOT applied** — on a chassis predating the Go half the new field resolves to
nil and *every* run takes the clean branch, which is worse than the bug. The owner's call
for this session was to commit and wait for another session's roll, so **150 stays open**:
the bar is fixed AND live.

## 2026-07-31 — the hazard I created, and caught by reading the runner

I left migration 281 unapplied on purpose (the Go half has not shipped) and wrote a banner
saying so. Then I read `scripts/migration/run-migrations.sh`: **`--apply` takes EVERY pending
file in number order**, and it is another session that runs it. `schema_migrations` has
recorded nothing since **273**; the runner lists **67 pending**. So "written but not applied"
is not a state the directory can hold, and a banner addressed to a human who is not reading
my file protects nothing.

Renamed to `281_..._HOLD.sql`. The runner's `SIDECAR_RE` (`_[A-Z][A-Z0-9_]*\.sql$`) excludes
an UPPERCASE-suffixed file from `--apply` while still **listing** it under *"Sidecars
(hand-run only, NOT applied by this runner)"* — held back visibly rather than hidden.
Verified: `--no-probe` shows 281 under Sidecars and not in the Pending 67.

The general form is now a landmine: **a migration's guard checks for DRIFT, never for
ORDER.** A `WHERE` clause that refuses a changed row still applies happily at the wrong
moment, and the applying session sees an ordinary successful run.

`sql_for_agents/278` — `bugs_open/154`'s config half, same two-part shape, same banner — is
exposed identically and sits in the same pending 67. Told that lane in their bug file rather
than renaming their file for them.

## 2026-07-31 — my LANDMINES append was swept into another lane's commit

Committed my rename with a pathspec naming LANDMINES.md and git reported only 2 files. The
entry was already at HEAD: the bugs_sweep/111 lane's `f076f4bd1` ("close(111): footer Contact
heading…", 22:28:32) had taken it — 33 LANDMINES lines and 55 WRONG_CALLS lines in a commit
about a footer.

Nothing was lost and forward-only holds, so this is recorded rather than repaired. It is the
exact scenario CLAUDE.md describes: **a pathspec commit stops me sweeping up others' work; it
cannot stop a session running `git add -A` from sweeping up mine.** The practical lesson for
this lane's remaining work is the one already in that file — commit each coherent piece
immediately and narrowly, and expect append-only fleet docs (`LANDMINES.md`, `WRONG_CALLS.md`)
to be the most contended files in the tree, because every lane appends to them.
