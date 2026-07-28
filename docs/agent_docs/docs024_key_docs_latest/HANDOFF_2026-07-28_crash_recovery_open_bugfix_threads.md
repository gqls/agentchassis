# HANDOFF 2026-07-28 (evening) — crash recovery, and the owed bugfix items it cleared

**Why this file exists.** The owner's machine crashed mid-session. He asked a fresh
session to pick up the open bugfix threads and continue them. This is where that got to,
what is genuinely finished, what is finished-but-unwitnessed, and the one thing waiting
on him. Written for a cold start: assume the reader knows nothing about this session.

**Everything below was true at ~18:30Z 2026-07-28.** This tree moves in minutes — re-run
`git status`, `git log`, and any figure you intend to repeat.

---

## 0. The one thing waiting on the owner

**Two `image-build-handler` error handlers are disabled and the choice of what to do
with them is his.** He has been reminded in chat; if he has not answered, ask again
before doing anything else with 086.

| step | action | if it fails, today |
|---|---|---|
| `mark_work_item_complete` | `update_work_item_status` | orchestration fails; the `site_work_items` row is left neither complete nor failed |
| `flag_rebuild` | `flag_page_image_rebuild` | same, and the generated imagery stays deployed and unreferenced |

The three options, as put to him:

1. **Re-point both at `mark_work_item_failed`** — *recommended*. The agent already
   declares that lane (`mark_work_item_failed → complete_error`) and all five of its
   `call_*` steps use it; these two alone routed to `complete`.
2. **Leave disabled** — status quo since 07-26. Loud generic failure, no work-item record.
3. **Re-enable as authored (`→ complete`)** — argued against: a silent `flag_rebuild`
   failure means imagery generated, deployed and never referenced, which is exactly
   `bugs_open/114`. This option manufactures that bug on demand.

It is a **DB config change — live immediately, no roll**. Revert is one rename; the
pre-change snapshots are verified present (see §2). These two are the only ones of the
ten with live traffic, which is why they are the item with a clock on them.

---

## 1. What was recovered from the crash

The crashed session left an uncommitted edit to `bugs_open/131` and a fleet-wide
`IMAGE_TAG` bump another session had left dirty. The 131 edit was coherent, not garbled —
committed as `1c2e7f614`.

Its subject was **131 item B, check-side**: the `no_horizontal_overflow` check could not
see content that was *cut off* rather than *scrollable*, so it passed pages whose content
was clipped. Fix `5042d5ecb`. Council **APPROVED**, corr `845893c9`, *"approved with 3
advisory objection(s) — none high-severity"*, decided `14:42:27Z`. The verdict post-dates
the commit by three minutes, so it joins by **note, not by a `Council-Reviewed:` trailer**.

---

## 2. bugs_closed/086 — the per-handler audit, DISCHARGED (`7e3a6d89d`)

The review the owner made a condition of seed 220. **Verdict: 7 stay disabled, 2 want
re-pointing (§0), 1 is another workstream's.** Nothing was changed; it is a reading.

**Two things the original three-way split had missed:**

1. **Only `image-build-handler` has a real error terminal**, and its two disabled
   handlers are precisely the ones that bypassed it. The other six affected agents
   (`blog-content-planner`, `content-gap-planner`, `site-adoption-agent`, `spec-updater`,
   `tool-improver`, `webdesign-agent`) have **no error terminal at all** — every terminal
   they own is a `complete_workflow`. For them `error_step: complete` was the author
   reaching for the only terminal that existed, so disabling is right and there is
   nothing better to point at without designing a new lane.
2. **The premise moved.** The ruling rested on *"none has fired in 30 days"*. The two
   `image-build-handler` steps ran **4 times each on 07-28** — all `COMPLETED`,
   `__step_error` NULL, so the *failing* branch still has not been exercised, but the
   exposure is now real rather than theoretical.

**Containment re-verified live: 0 / 10 / 44**, and it **survived a 181-agent bulk re-seed
at `14:25:02.999304Z`** — worth stating because "config re-seed clobber" is a recorded
landmine on this fleet and here it did not bite.

**Revert is safe:** seed 220's snapshots are real — 7 rows in `agent_definitions_backup`
with `snapshot_reason LIKE '220_%'`, taken `2026-07-26 18:32:26.229Z`.

**Contributed, not acted on:** `tool-improver.note_refusal` sits on the refusal branch
and its `next_step` and `error_step` were **both `complete`** — the handler drew no
distinction at all. `scripts/who-owns.py 126` puts that path with the **oufe workstream**
(ACTIVE). The finding is written into `bugs_open/126`; do not decide it for them.

**Loose end, stated rather than tidied:** the 45→44 drift is only partly resolved. The
seven snapshotted agents went **+1** (`tool-improver.update_component` gained a handler,
`6e29d6d19`), which makes the expected total 46, not 45. `[UNRESOLVED]` **two handlers
are unaccounted for among the other 12 agents**, and there is no baseline to diff them
against. **The cheap fix for next time:** `snapshot_agent(<type>, 'baseline')` on those
12 — one call each, and the next drift becomes a two-table diff instead of an open
question.

---

## 3. bugs_open/131 item B — LIVE, and deliberately not called proven

**State: the clause is in the running binary. It has never been seen firing.**

Rolled first as `v1.0.1190` at `16:56:18Z` via a new makefile target (§4). A later fleet
deploy moved it to **`v1.0.1192`** (pod `browser-runner-adapter-8f74cbd95-nj866`, started
`18:23:07Z`). **Re-grepped after that roll rather than assumed** — positive control 1,
fix marker 1, negative control 0. **Do not go looking for `v1.0.1190`; it is gone.**

> **Why a rebuild was needed and a redeploy would have been a no-op.** `v1.0.1188` and
> `v1.0.1189` were the **same image id** (`bb9cb4a8b649`), built `13:43:31Z` — 56 minutes
> *before* the fix commit at `14:39:22Z`. **A retag is not a rebuild.** Rolling 1189
> again would have restarted the pod, looked exactly like a successful deploy, and
> shipped the identical binary.

**The pod-grep recipe for this container** (CLAUDE.md's `strings` recipe silently returns
0 for everything here — `strings` does not exist in the image):

```
kubectl -n ai-persona-system exec <pod> -- sh -c \
  "grep -c 'while the content stays cut off' /app/browser-runner-adapter"   # fix marker
kubectl -n ai-persona-system exec <pod> -- sh -c \
  "grep -c 'no_horizontal_overflow' /app/browser-runner-adapter"            # positive control
kubectl -n ai-persona-system exec <pod> -- sh -c \
  "grep -c 'zzz_not_a_real_marker_zzz' /app/browser-runner-adapter"         # negative control
```

All three must move together. A `grep -c` that finds something proves the file is
greppable, not that the binary is the one you think.

### Do NOT re-run the two named cases and call it verified

The bug's own text asks for re-verification against `/` and `/about.html` at 390px.
**Following that literally now produces a green result that means nothing:**

- `/about.html` was **never** a real failure — premise corrected at ~15:45; the table
  already sits in `div.pc-table-wrapper` with computed `overflow-x: auto`.
- `/`'s residual cut (`div.brief-explanation__stat`) was **fixed page-side at ~15:50**
  via `flex-wrap: wrap`.

Both are clean with or without the new clause. A pass proves deployment, which the
pod-grep already proved. See `[[verify-the-failing-branch]]`.

### What will witness it, with no manual dispatch

The clause is on an **actively exercised path**: tool acceptance runs it as
`{"id": "no_overflow", "tier": 4, "type": "no_horizontal_overflow", "profiles": ["mobile"]}`.
14 orchestrations reference it, most recently `15:17:51Z` — i.e. **before** the roll.
**0 runs since.** The next acceptance run at `mobile` exercises the new code.

```sql
SELECT count(*), max(updated_at) FROM orchestration_states
WHERE collected_data::text LIKE '%no_horizontal_overflow%'
  AND updated_at > '2026-07-28 16:56:18+00';
```

**The tell:** a `CheckResult` with `pass:false` **and** a populated `culprit` /
`component` / selector, on a page whose `scrollWidth` is clean. The old clause could only
fail on `scrollWidth - clientWidth > 2`; nothing else produces that shape.

**Caution for whoever confirms it:** a false positive here becomes an `improve_tool`
fixer aimed at a correct page (`bugs_open/126`). If the first flag looks wrong, check for
a horizontally scrollable ancestor before treating it as a page defect — that escape is
what the three filters exist for.

**Item B check-side stays OPEN**: live and unwitnessed, not fixed and live.

---

## 4. New machinery — `make deploy-<service>` (`35c8277a8`)

**`deploy-agents` is all-or-nothing** and that is a live hazard, not a style point. It
seds *every* service's kustomization to `$(IMAGE_TAG)` and applies them. Measured while
rolling the browser-runner fix: **of the 14 backend services, exactly 2 had `v1.0.1190`
in the registry.** Running it would have pointed twelve healthy deployments at images
that were never pushed and ImagePullBackOff'd them together.

```
make deploy-<service>                                  # deploys ONE service at $(IMAGE_TAG)
make deploy-browser-runner-adapter IMAGE_TAG=v1.0.1190 # what was actually run
```

Mirrors the build side's `build-%-ref` / `build-%-tree` pattern rules. Explicit targets
keep priority in make, so `deploy-agents`, `deploy-infrastructure` and the numbered
`deploy-0NN` targets are untouched — verified with `make -n deploy-agents`.

**The registry pre-flight is the load-bearing part and it earned its place on its first
dry run.** `push-*`/`deploy-*` are git-blind, so nothing downstream of the build checks
that the tag exists. The dry run resolved `IMAGE_TAG` to **`v1.0.1191`** — another
session had bumped it mid-task (`02113a3a9`) — against an image built at `v1.0.1190`.
The guard refused instead of rolling the deployment onto a missing image.

> **Register this if it is not already there** — it is a new callable mechanism and the
> concept register's bar is "another workstream could call this and would not know it
> exists". Category: build/deploy. Not done in this session.

---

## 5. Memory corrections made

- **`mig 221 IS applied`** — 2026-07-26 21:05:45Z, `UPDATE 3`. The index had said
  "still UNAPPLIED". The topic file had **contradicted itself since the day it was
  written**: lines 18–20 and 77 recorded the apply correctly, a stale landmine at line 66
  said the opposite, and it was the stale version that reached the index. Both fixed.
  Check before repeating any "unapplied" claim:
  `SELECT filename, applied_at FROM schema_migrations WHERE filename LIKE '221%';`
- **086 index line** now records the audit as discharged with the owner call outstanding.
- **`execution_path` is DEAD** — 0 of 2225 rows populated. Use `processing_history`.
  Added to the 086 line because it is the column a step-history question naturally reaches
  for and it answers with silence.

---

## 6. Landmines this session paid for

- **A retag is not a rebuild.** Two tags sharing an image id look exactly like two
  builds until you compare image ids or timestamps. `docker images <repo> --format
  '{{.Tag}}\t{{.ID}}\t{{.CreatedAt}}'` settles it in one call.
- **`strings` does not exist in the browser-runner container** — CLAUDE.md's verify
  recipe returns 0 for *everything* there. Caught only by a positive control.
- **The tag can move between your build and your deploy** on this tree. Pin it:
  `IMAGE_TAG=<what you actually built>`.
- **A fleet deploy can roll your service out from under you.** browser-runner went
  1190 → 1192 without this session touching it. Re-grep after any roll you did not do.
- **COUNT the baseline before believing an `EXCEPT` / `NOT EXISTS` diff.** An empty
  baseline returns "nothing lost" and reads as reassurance. This nearly became a finding
  in the 086 audit — logged in `WRONG_CALLS.md`.
- **`snapshot_agent` is OVERLOADED.** One-arg writes to `agent_definitions`
  (`is_snapshot=true`); **two-arg — the one seeds call — writes to
  `agent_definitions_backup`.** Looking in the wrong table produced a confident and
  wrong "the safety net does not exist".

---

## 7. Commits from this session

```
1c2e7f614  docs(131 B): crash residue — council APPROVED, status recorded
e92198aa8  build(131 B): browser-runner-adapter v1.0.1190 — the image that carries the fix
7e3a6d89d  audit(086): the owed per-handler review — 7 disabled, 2 re-point, 1 oufe's
35c8277a8  feat(makefile): deploy-<service> — single-service deploy
71c80bc43  bug(131 B): check-side LIVE on v1.0.1190 — deployed, not proven
bcc396b6b  bug(131 B): tag moved to v1.0.1192 and the fix survived — re-grepped
```

---

## 8. Where to go next

1. **Get the owner's answer on §0.** It is the only blocked item and it is cheap to apply.
2. **Watch for the first `no_horizontal_overflow` firing** (§3 query). One run closes the
   check-side of 131 B or reopens it with real evidence.
3. **Register `make deploy-<service>`** in the concept register (§4).
4. If picking up something new instead, the fleet cold-start list is
   `OPEN_THREADS_RESTART_LIST.md` — but treat its figures as stale and re-ground them;
   it was last refreshed 2026-07-27 22:00Z and this tree has moved a long way since.
