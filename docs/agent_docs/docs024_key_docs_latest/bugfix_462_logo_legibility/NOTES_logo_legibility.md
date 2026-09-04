# NOTES — 462, logo legibility

Append-only, newest at the bottom. Missteps are the point, not an appendix.
Technical log; the plain-English history is `README_where_we_are.md`.

---

## 2026-09-04 — lane opened by handover from `bugfix_417_logo_text_policy`

**How the handover went, because the mechanics are reusable.** The owner opened a dedicated session
for 462. `ListAgents` showed `bugfix 417` **live and busy**, and `git log` showed it had committed
the sweep, `bugs_open/462` §8, the RUNBOOK section and its NOTES **within the previous three
minutes** — the newest of them 32 seconds before I started reading. So the first action was not to
edit anything: it was to ask, naming the exact files I intended to claim.

That was right, and by more than politeness. The reply carried **ten commits made after the four I
had read**, and a measurement taken after all of them that changes the routing design (§9c). Had I
started from the four commits I found, I would have designed a filer that routes every legibility
finding at an image generator — and one of the two live findings is a **human's uploaded logo with
no prompt to regenerate from**.

> **The transferable bit: on a shared tree, "I read the bug file and the git log" is a snapshot of a
> lane that is still moving.** The peer's knowledge was ~30 minutes ahead of its own commits. Ask.

**What I re-verified rather than inherited on trust** (a takeover is exactly when to re-run the
motivating case, per `bugs_open/462` §7):

- `--self-test`: 6/6, including both preserved websitepromotion PNGs. Arm A fires on the
  pre-regeneration mark, arm B on the post-regeneration one.
- Live, 13:25Z: fires on `websitepromotion.co.uk` (arm B, 6.7% of ink clears 3:1, median 1.01:1) and
  `mortgagecalculator.co.uk` (arm A, max 2.39:1); `seotools.co.uk` passes at 29.3%. Byte counts and
  md5s match §8b exactly, so these are the same artefacts §8 measured.

### MISSTEP — I nearly filed a false defect against the script I had just inherited

I ran `… | tail -40; echo "EXIT=$?"` and read `EXIT=0` against a run that had printed two FINDINGs,
while the script's own docstring says *"exit 0 = every logo measured and legible"*. That looked like
a real bug in a detector whose whole job is not to report false passes.

**It was my measurement, not the script.** `$?` after a pipeline is the status of the **last**
command — `tail` — which exits 0 regardless. Run bare, the script exits **1**, exactly as documented.

The cheap check was one command: `./scripts/audit-logo-legibility.py --site X >/dev/null 2>&1; echo $?`.
Cost: nothing, because I checked before writing it down. **Recorded because the near-miss is the
useful part** — the instrument I was using to judge the detector was the thing that was wrong, and a
"defect" found in the first ten minutes of a takeover is exactly when that is most likely.
Now in the RUNBOOK where the command is.

### Routing — what I went looking for, and what I found instead

I expected the routing question to be "which handler", answered by finding the agent that
regenerates logos. It is not that question. What the search actually turned up:

1. The live repair unit is **`needs_imagery` / `needs_imagery:site:-:logo`**, not `needs_logo`.
   `needs_logo` reads as 2 live rows, both cancelled — and **13 complete in the archive**. The live
   table is a rolling window; I would have called a working mechanism dead if I had stopped at it.
2. `needs_human_review` — the obvious "tell a human" destination — is the estate's **documented dead
   queue**, and I measured it at **1,439** rows against the 370 recorded in
   `revalidate_review_queue_action.go` on 2026-07-25. Routing there would reproduce this bug.
3. There **is** a reversible remedy: `ingest_staged_asset` (built for `bugs_open/131`, uploads to a
   new key, never overwrites, records the previous path in `assets.alterations`). It cannot be
   automated — it needs a person with a file.
4. There is **no legibility criterion anywhere in generation**: the adapter's only fail-closed
   statistic is `keyGroundMinBorderKeyed = 0.95`. So routing at the generator asks it to draw again
   with no new constraint. §6 already measured how that goes.

Put together with the 417 lane's provenance measurement, the filer has **zero** legitimate targets
today. Written up as `bugs_open/462` §9, with the fork put to the owner rather than decided.

**What I have NOT done and am not claiming:** nothing files, nothing is scheduled yet, and option (a)
— the render-audit home that does not cache the theme token — is untouched. The sweep is still a
hand-run tool, which is the state §9e calls decaying.

### A doc-writing misstep, same session

Two `cd`-prefixed heredoc calls in a row: the shell's working directory **persists between Bash
calls**, so the second `cd <relative path>` failed from inside the directory it had already entered,
`&&` short-circuited, and `NOTES` was silently not written while the file after it (no `&&`) was.
`ls` caught it. **Use absolute paths in write commands**; a partial failure in a chain of writes
looks exactly like a success if you only read the last line.
