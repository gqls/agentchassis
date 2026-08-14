# HANDOFF — 2026-08-14 — the gates are OBSERVABLE and the instrument is **LIVE**, and the archived-page "defect" was refuted by measurement. Read this file only

> ## 🔵 CORRECTION to this file, 07:55Z, before anyone read it — the arm instrument is **LIVE**, not pending
>
> I wrote the tables below saying "NOT LIVE until the next roll". **A roll happened while I was
> writing them:** the fleet is on **`v1.0.1297`** and the binary carries the arm code. Probed with
> both controls, so the check could have come out negative:
>
> ```
> NEEDLE_arm  "gate_claims_still_present" = 1 rc=0     ← unique to this change
> NEEDLE_sent "unreported:"               = 1 rc=0
> CONTROL_pos "the database is not the website" = 1 rc=0
> CONTROL_absent "zzz-not-in-any-binary-zzz"    = 0 rc=1   ← the probe discriminates
> ```
>
> 2 Running pods matched `-l app=agent-chassis`, **ONE digest** `sha256:2e89958a`, tag `v1.0.1297`.
> ⚠ That label is known to under-match this fleet (it has returned 2 while 56 ran the image), so the
> load-bearing claim is **digest uniformity across what the label returns**, not the pod count.
>
> **So the next daily sweep is the FIRST INSTRUMENTED ONE, and it is imminent — ~08:44Z, ~49 minutes
> after this was written.** The prediction, stated in advance so it is falsifiable: **~16–18 rows of
> `scan_still_trips`, ZERO `gate_%` rows, and zero `unreported:claims_unverified`.** If instead you see
> `unreported:claims_unverified`, the arm did not reach the record and edit 2 is wrong. Read it with:
>
> ```sql
> SELECT result #>> '{revalidation,arm}' AS arm, count(*) FROM site_work_items
> WHERE item_type='claims_unverified' AND result ? 'revalidation' GROUP BY 1 ORDER BY 2 DESC;
> ```

**Supersedes `HANDOFF_2026-08-12_continue_here.md` for state.** Its §4 traps still hold except where
corrected below; its banners are history. Working record: `NOTES_deployed_asset_path.md` (newest at
the bottom). Owner's log: `README_where_we_are.md`.

**Nothing is in flight. Nothing is half-applied. Everything below is committed.**

---

## 1. State — verified 2026-08-14

| thing | state |
|---|---|
| `claims_unverified` revalidator | **LIVE + PROVEN** |
| The three gates (copy-changed · claim-granular · published) | **LIVE on ≥`v1.0.1293`**, and **still never REACHED** — not once, by any item |
| **`result.revalidation.arm` — the arm instrument** | **LIVE on `v1.0.1297`** (needle + both controls, one digest) + council APPROVED r1 |
| Council | **APPROVED first round**, `fe7dccb3-3038-4177-b77a-0cf620860556`, 10 seats, 7 abstained, 1 LOW advisory (actioned) |
| Fleet | `v1.0.1297` (`IMAGE_TAG` line 17), digest `sha256:2e89958a` |
| `bugs_closed/262` (published gate) | CLOSED, live since `v1.0.1293` |

**The measurement this lane owes on every visit — 2026-08-14: `0 | 0 | 0 | 9 | 18 | 3` of 30**
(refused_gate1 / gate2 / gate3 / resolved / still_holds / unknown), invariant `t` for **9/9**, zero
`f` rows. Unchanged from 08-13. Queries: `RUNBOOK` § "the measurement this lane owes" and the new
§ "Where does the ladder STOP?".

---

## 2. What this session did, and the two things it got wrong

### 2.1 The arm instrument — answering the 08-13 handoff's open question

The 08-13 banner asked: **how would we ever observe these three gates?** The gates sit at the FOOT of
an 18-arm ladder, so every arm above them returns first and a gate's refusal counter reads `0`
identically whether it approved, refused nothing, or was never asked. Telling those apart needed a
LIKE over the **prose** of `reason`.

Now `result.revalidation.arm` names the rung that decided: 18 stable keys, `gate_`-prefixed for every
rung downstream of a clean scan. **Observation only — nothing branches on it**, no arm added, removed
or reordered, no predicate touched, and the load-bearing invariant was re-measured after the edit and
is unchanged. The four uninstrumented revalidators record `unreported:<item_type>`, never an empty
string, so a gap cannot read as an absence.

Commits: **`92b59138b`** (the field + 18 arms + tests), `bb05ce78a` (gofmt), **`ac6a86f58`** (the
correction in §2.2). All three are **in `v1.0.1297`** — see the correction banner at the head of this
file; they postdated `v1.0.1295`, which is what the body originally said.

### 2.2 ⚠ CORRECTION — the arm is a SNAPSHOT, not a history, and I walked into this lane's own landmine

I wrote, in the submission, the commit message and the code, that the field answers *"has a gate
**ever** been reached"*. **It does not.** `applyRevalidation` replaces `result.revalidation` wholesale
every sweep, so `arm` is the rung that decided an item on the **latest run** and nothing more.

Caught by the council's `debug_historian` seat — on an APPROVED round, as its one LOW advisory — which
**quoted this lane's own landmine back at me**: *"a per-row revalidation stamp is LAST-WRITE-WINS, so
GROUP BY it reads exactly like a run log and is not one"* (`LANDMINES.md`, filed by this lane
**2026-08-12**). Same column, same lane, three days apart. Corrected in `ac6a86f58`; logged in
`WRONG_CALLS.md`; new LANDMINES entry filed under the `arm` footprint and synced.

**Why the hook did not save me:** the `SessionStart` landmine hook matches entries against files
already **dirty** in the tree, and these were clean at session start. **Grep LANDMINES for the TABLE
and COLUMN you are about to write to, not only the file you are editing.**

**The per-RUN surface is the real history and already carries the key** —
`collected_data->'sweep'->'items'` in `orchestration_states`, one row per sweep, never overwritten.
But `[MEASURED 2026-08-14]` **exactly ONE sweep row is retained** (08-13 08:44:39Z) against 2,532
orchestration rows back to 07-13. Structurally right, effectively empty. **A one-row answer there is
not "the sweep has run once".**

### 2.3 ⚠ CORRECTION — the archived-page "defect" is NOT one, and the obvious fix would cause harm

The 08-13 banner flagged, as item 4, that `ScanDeployedClaims` has no page-status filter so an
archived page is judged. I set out to quantify it as a defect: **3 of 30 revalidated items sit on
non-`active` pages** (2 `still_holds`, 1 `resolved`). Then I fetched them, with a fabricated-URL
control per domain so the check could come out negative:

```
200 30997b  robot-hands.com/gripper-catalog.html                    ← archived AND SERVING
404  2886b  robot-hands.com/…-control.html                          (control)
404  2711b  leopardessconsulting.co.uk/for-engineering-teams.html   ← absent (same size as control)
302   143b  webdesign.uk/index-rejected-v1-20260806.html            ← never deployed
```

**An archived page can be serving 31KB to the public.** So the scan is RIGHT to look, and the code
comment's "NO PAGE-STATUS FILTER, and that is deliberate" is correct for a better reason than the
parity it cites. **Do not add a status filter to either end** — it would stop auditing a page that
really is asserting unsupported claims. `bugs_open/266` (owned by the 215 lane, fix APPROVED + LIVE on
`v1.0.1295`) reached the same conclusion from the producer side, and two of these three domains are
the consumers it notified. Consumer-side note appended to `266`; **no competing bug filed.**

**What the measurement does show:** the discriminator nobody reads is **"is it served"**, not "is it
archived" — `status`, `build_status` and `deployed_at` all failed to separate the serving page from
the two dead ones. Two items therefore park in `needs_human_review` for ever on unserved pages, so
**the gate-reachable population is 16, not 18.**

---

## 3. What is next

1. ~~NEXT ROLL: verify the arm field~~ **DONE — LIVE on `v1.0.1297`, needle and both controls in the
   banner above. What remains is to READ THE FIRST INSTRUMENTED SWEEP (~08:44Z).** Expect **~16–18
   `scan_still_trips`, ZERO `gate_%`, ZERO `unreported:claims_unverified`**; the last of those three is
   the one that would falsify edit 2. ⚠ **A zero gate count is the instrument working, not the gates
   approving.** For the first time that sentence is checkable rather than arguable — which was the
   entire point.
2. **The gates are still UNREACHED and no instrument changes that.** What they need is *an item whose
   page has genuinely been cleaned*, and nothing in the population produces one: 16 of 18 `still_holds`
   are correct refusals (the claims really are on the page) and the other 2 sit on unserved pages.
   **The remaining option from the 08-13 banner is the owner's:** construct a cleaned page, which is a
   live-content intervention. Not taken unilaterally. **Do not describe any gate as having prevented
   anything.**
3. `features_open/032` — the shared helper, and it now has a second consumer: **`arm` is set only by
   `revalidateUnverifiedClaims`**; the other four record `unreported:<item_type>`. Lifting arms into
   them belongs with lifting the copy-changed comparison. **Measure before building.**
4. §3.4's remaining `editquality` LOW from round 7: a before/after test for the SQL→Go locked-skip
   move ("emitted output is unchanged" is asserted, not demonstrated). **Still not done.**
5. §3.5's leftovers from the 08-11 handoff: §2.3 pin `ScanDeployedClaims` to its intended callers ·
   §2.4 the invisible backlog · §2.5 Decision 2's dedup half · §2.6 more sweep coverage · §2.7 the
   armed-but-inert cap at `check_image_source_unsatisfiable.go:167`.

---

## 4. Traps — additions to the 08-12 §4 list, which otherwise stands

- ⚠ **`result.revalidation.arm` is LAST-WRITE-WINS** (§2.2). Drop the words *ever*, *how often* and
  *rate*. The structural tell: **a `jsonb_build_object` assignment means no second row per item can
  exist, so no per-item key can carry a rate.**
- ⚠ **`resolved_all_gates_passed` carries NO `gate_` prefix.** A prefix-only reach query counts only
  the reaches where a gate REFUSED and **misses every closure**, inverting the reading. Use
  `arm LIKE 'gate\_%' OR arm = 'resolved_all_gates_passed'`. And `_` is a LIKE wildcard — escape it.
  `arm IS NULL` finds nothing; the gap check is `arm LIKE 'unreported:%'`.
- ⚠ **An `archived` page can be SERVING** (§2.3). Never conclude a page is dead from `pages.status`;
  curl it, **with a fabricated-URL control on the same domain**, or a catch-all 200 reads as live.
- ⚠ **The shared package would not compile for most of this session** — another session's uncommitted
  WIP in `palette_specialised_slots.go` (`undefined: colour`). Test against a clean tree plus your own
  files: `git archive HEAD | tar -x -C <dir>`, copy your files over, run there. 9/9 packages green.
- ⚠ **The council trigger refuses `plan.risks` as an ARRAY** — it must be one prose string. Refused
  client-side before credits, so it costs nothing but a retry.
- ⚠ **`- **status-evidence:**` appears in EVERY register entry (20×).** A script inserting "before the
  status-evidence line" inserted 20 times. Target the LAST occurrence when appending to the last
  entry, and check `wc -l` against the expected delta before committing. Caught by `grep -n`, restored
  with `git checkout` after confirming via `git diff --numstat` that all 61 added lines were mine.
- ⚠ **A bare `git stash` by another session deleted this lane's uncommitted tests on 08-12.** Still
  true. Commit before mutating; recover with `git checkout stash@{0} -- <path>`, never a bare `pop`.

---

## 5. Commits and correlations

`92b59138b` the arm instrument · `bb05ce78a` gofmt · **`ac6a86f58` the snapshot-not-history
correction** (`Council-Reviewed:`). Docs, landmine, register, WRONG_CALLS and the `266` note follow in
the docs commit after these.

| what | id |
|---|---|
| **arm instrument council — APPROVED at r1** | `fe7dccb3-3038-4177-b77a-0cf620860556` |
| claims_unverified council — APPROVED at r7 | `b67eb26a-14ef-45d7-b755-3e489fd57ef0` |
| voice_tells council (APPROVED r1) | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |

Verdicts saved verbatim, newest first: `VERDICT_2026-08-13_APPROVED_revalidation_arm.json`,
`VERDICT_2026-08-12_round7_APPROVED_claims_unverified_council.json`,
`VERDICT_2026-08-12_round6_*.json`, `VERDICT_2026-08-11_round5_*.json`.
