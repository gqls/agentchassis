# HANDOFF 2026-08-22 — `bugs_open/260` lane: CLOSED. Read this only for what it spun off.

**Supersedes `HANDOFF_2026-08-19_continue_here.md`** (that one asked for work that is now done and
live). **This lane needs no continuation.** The file exists because the session that closed it went
on to find two unrelated defects and file them, and a reader arriving here should be routed, not
re-briefed.

---

## 1. State of the 260 work: DONE, LIVE, RE-VERIFIED

`bugs_closed/260` — the silent regex/handlebars fallback in `RenderTemplate` is deleted; template
execution failure now returns an error. Council-approved (three correlations, all APPROVED, every
commit carrying a trailer). Register entry **STY-057**. Architecture note `RFC_041`.

**Re-verified on `v1.0.1326`, both replicas, 2026-08-22** — see NOTES for the table. The method that
works, because the obvious ones do not:

```bash
# ONE pass. Seven separate exec-greps over a ~100MB binary exceed the 2-minute tool timeout.
kubectl -n ai-persona-system exec <pod> -- sh -c 'grep -aoF \
  -e "Go template execution failed, using regex fallback" \
  -e "Failed to parse template, using fallback" \
  -e "refusing to emit output that was not executed" \
  -e "component template execution failed" /proc/1/exe | sort | uniq -c'
```
The first two are **removed-string controls** — real removals from the pre-260 binary, so they can
come out either way. Expect them absent and the rest present.
**Do not look for `build provenance` in the logs on a pod more than an hour or two old** — it is a
startup line and it scrolls; absent there means "not in range", never "unstamped".

**Residual, and it is not this lane's:** `bugs_open/342` (absent-required-field reporting; 6 of 15
call sites unwired, no refusal anywhere) — **another lane took it 2026-08-22**. Do not start on it.

---

## 2. What this lane spun off, and who has it

### `bugs_open/354` — a workflow that ends at its ERROR TERMINAL is recorded `COMPLETED` with `error` NULL
**OWNED by `bugs_open/307 [e24299]` since 2026-08-22. Do not compete — contribute into the file.**
They asked before starting rather than trusting `who-owns.py`, which reads commits and is blind to a
session mid-fix.

Everything is in the bug file. What a passer-by should know: it is a **legibility** defect, not a
routing one; `agent_error_log` holds every one of the failures and retains a month, so nothing is
being lost, only filed where no consumer joins; and the 090 verdict (CONFIRMED, iteration 1) named
the scope a fixer needs — `routeToErrorStepOrFail` and `failWorkflow`, the sibling arms that DO fail
the workflow.

**Their findings, already folded in and better than mine** — migration `466`'s status vocabulary; the
`database-cleanup` arm-3/arm-4 trap that would make a new terminal status unreapable; and the
discriminator that actually works (`terminal declares outcome:"error"` AND `__step_error` present —
36/36 dishonest, 0/13 honest). **Three structural rules they tried and disproved are recorded in §5
so nobody re-walks them.**

### `bugs_open/361` — `component-render-check` is permanently red because a growing library manufactures "NEW" findings
**OPEN, UNOWNED.** `who-owns.py 140` puts the originating lane at quiet-14d.

Nothing is broken. Twelve consecutive red days (last green 2026-08-09) because the ratchet
(`rendercheck.go:759`) is a flat key-level diff against a baseline cut once on 2026-08-04, and 109
components have been created since. 37 of the 38 components in today's 227 NEW findings postdate the
baseline; the one exception had its template rewritten. **Zero regressions — and all 227 still real
debt.**

**⚠ THERE IS AN OWNER DECISION HERE AND IT HAS NOT BEEN TAKEN.** Regenerating the baseline banks 227
real findings as "known". That is a debt judgement, not a code judgement, and it was deliberately
left to the owner. **Do not regenerate on your own initiative.** The code fix (candidate 1: a
component-scoped ratchet, ~6 lines) is separate and does not need that decision.

**If you deploy anything here:** `make deploy-component-render-check` ships **nothing** on its own —
the tag is pinned in the overlay, and both make and `kubectl apply` report success anyway. Bump
`newTag` in the same commit as the rebuild and read the artefact, not the make target.

---

## 3. Two landmines added today — read these before touching a CronJob or messaging a peer

Both in `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`:

- **A CronJob's Job listing shows only the last N failures**, so any outage reads as about N days
  old. `failedJobsHistoryLimit: 3` had collected nine of twelve failures, their pods and their
  events. `lastScheduleTime` was 132 minutes old the whole time and is **not** a health signal. Use
  `status.lastSuccessfulTime`, or the check's `doc_notes` series where the **count** is the tell —
  `backoffLimit: 1` means a failing run writes two rows and a passing run one.
- **Two live sessions can share a lane name, and "active N ago" points at the WRONG one.** A lane
  mid-investigation runs long tool calls and looks stale. **The remedy is verified: the incoming
  message's `from` socket path works as `to`.**

---

## 4. Standing hazards this session hit three times, stated so the next one expects them

**Same-file passengers, in both directions, on `LANDMINES.md` / `WRONG_CALLS.md` / the concept
index.** My index correction rode into another session's commit; two other sessions' entries rode
into mine. **`git diff --numstat` before committing is not sufficient** — the window between the
check and the commit is the exposure. Read the **insertion count and the scope block the hook prints
after the commit**; that caught all three. And holding an edit back does not avoid the trap, it only
decides which side of it you are on. Forward-only holds; each was recorded in a follow-up commit
naming the sha so the other session can find what they shipped.

---

## 5. If you are picking this up cold

There is nothing to do on this lane. In order of what is actually outstanding on the estate:

1. **The `361` baseline decision** — owner's, not yours.
2. **`361`'s code fix** — unowned, ~6 lines plus a mutation proof of both arms (a fix verified only
   on the "unbaselined component does not fail" arm is a fix that turned the check off).
3. **`354`** — owned; contribute, do not compete.

Commits from this session, newest last: `a15fa0d8c` `2a1178103` `1981c23d8` `c0484618c` `8cc994b12`
`f0db82afc` `b17f69eab` `b36136e18` `56baf6202` `d64c1b8a6` `a4a5f1d3c`.
