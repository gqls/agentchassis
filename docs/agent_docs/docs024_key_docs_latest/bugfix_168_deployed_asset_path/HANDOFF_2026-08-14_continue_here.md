# HANDOFF — 2026-08-14 — ⛔ SUPERSEDED BY `HANDOFF_2026-08-14b_continue_here.md` — THE TITLE BELOW IS NOW FALSE

> **⛔ STOP — GO TO `HANDOFF_2026-08-14b_continue_here.md`.** This file's own title said the gates were
> *"observed unreached"*. That was true at 08:45:05Z and **false from 16:48:45Z**, when a gate was
> reached and passed. The title is kept rather than rewritten because the series is the record — but a
> reader who stops at the title of a `_continue_here` file gets the opposite of the truth, and this
> lane has already been bitten by a correction sitting below the fold. Everything below is HISTORY.
> Its §4 traps still hold except where 14b supersedes them.

# ~~HANDOFF — 2026-08-14 — the instrument is LIVE, the first sweep CONFIRMED the prediction, and the gates are now *observed* unreached. Read this file only~~

> ## 🟩 SUPERSEDED at 16:48:45Z — A GATE HAS NOW BEEN REACHED. Read this banner before anything below it
>
> **Owner instruction this session: "yes, clean a page"** — the live-content intervention §3.2 below
> records as *"the owner's"* and did not take. Taken now. `leopardessconsulting.co.uk/case-studies`
> asserted **"75,061 orchestration state records"** against a register fact of **2,578** (`gte`), i.e.
> a ~29x overclaim, flagged independently by BOTH `claims_unverified` `e713613f` and `stale_evidence`
> `3a5419a1`. One sentence deleted (36 chars, minimal deletion per the owner's 2026-08-06 ruling), on
> **both** stored surfaces, rerendered and deployed. Served page 24,558 -> 24,522 bytes, control 404s.
>
> | predicted, written into NOTES BEFORE the run | observed |
> |---|---|
> | arm is `gate_`-prefixed OR `resolved_all_gates_passed` | **`resolved_all_gates_passed`** |
> | not `scan_still_trips` / `gate_claims_still_present` / `page_absent` | none of them |
>
> **The measurement — 2026-08-14 (after): `0 | 0 | 0 | 10 | 17 | 3` of 30**, invariant `t` for **10/10**,
> zero `f`. Reach query: `refused_at_a_gate 0 | passed_all_gates 1 | uninstrumented 0 | total 30`.
>
> ### ⚠ WHAT THIS DOES NOT LICENCE — the standing instruction is UNCHANGED
> All three gates were **consulted and PASSED**. Every refusal arm is still at zero and still
> unobserved. **Do not describe any gate as having prevented anything.** The change is narrow and
> should be stated narrowly: the gates have gone from *never reached* to *reached once, and passed*.
>
> ### The refusal that is now cheap, and needs no intervention
> Had the sweep landed between the component edit (16:43:49Z) and the deploy (16:46:38Z), the
> published gate would have returned `gate_published_correction_unpublished` — the `bugs_open/262`
> case exactly. The window was ~3 minutes and the rerender closed it. **Any item whose page is edited
> but not yet redeployed produces that arm on the next sweep, at zero content cost.** Wait for it;
> do not manufacture it.
>
> ### ⚠ TRAP THAT NEARLY SHIPPED A NO-OP — now in `LANDMINES.md`
> `RerenderSinglePageAction` is *"Simple concatenation - no template re-rendering"* and contains **no
> `UPDATE page_components`**: it republishes stored `rendered_html` and **never regenerates it from
> `content_data`**. Clearing only `content_data` and rerendering republishes the claim unchanged, with
> a `COMPLETED` orchestration and a moved `deployed_at`. The audit reads both surfaces anyway
> (`rendered_html` for numbers, `content_data` for stats, `html||contentJSON` for the claim-granular
> gate), so **edit both**. Cheap check: `grep -c "UPDATE page_components" <the action you will fire>`.
>
> Sweep was **manual** (`84db99fc-5ebd-4bd7-9d1e-0272c7fd7557`), mirroring `fireTrigger()`;
> `scheduled_tasks.last_triggered_at` deliberately untouched. Being fleet-wide it also closed **4
> other lanes' genuinely-fixed items** early — named in NOTES. Full record: `NOTES_...md` (2026-08-14
> afternoon), owner's log: `README_where_we_are.md`.

> ## 🟢 RESULT, 08:45:05Z — the prediction written before the run held exactly. This supersedes the 07:55Z banner below
>
> The first instrumented sweep ran **2026-08-14T08:45:05Z**. Fleet has since moved to **`v1.0.1298`**,
> regression-checked at the binary (`gate_claims_still_present`=1 rc=0, fabricated control=0 rc=1).
>
> | predicted (written BEFORE the run) | observed |
> |---|---|
> | ~16–18 `scan_still_trips` | **18** ✓ |
> | ZERO `gate_%` | **0** ✓ |
> | ZERO `unreported:claims_unverified` | **0** ✓ |
>
> Plus `page_absent`=**2** and `evidence_base_absent`=**1** — the structured form of exactly what 08-13
> had to read out of prose. **Two methods, same answer.** So "no finding reached a gate" is now a query,
> not an argument. ⚠ **Still do not describe any gate as having prevented anything.**
>
> ### ⚠ AND IT FALSIFIED ONE LINE OF MY OWN LANDMINE WITHIN HOURS
> I filed *"an empty arm is never written, so `arm IS NULL` finds nothing"*. **Wrong.** The **9 closed
> items carry NO `arm` key** — frozen at closure (8× `2026-08-10`, 1× `2026-08-12`), because a terminal
> item is never re-swept. **`arm IS NULL` = "decided before 2026-08-14", a VINTAGE marker, not a gap.**
> The obvious "which revalidators lack arms?" query returns those 9 and reads as 9 uninstrumented items.
> **The gap check is `arm LIKE \'unreported:%\'`.** Corrected as a dated note on the entry, not a rewrite.
> **Bonus:** the key's absence now *proves* all 9 closures predate the instrument — the 08-12 residual
> that was called "not retrospectively measurable" is now simply visible.

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
| The three gates (copy-changed · claim-granular · published) | ~~**OBSERVED unreached** 2026-08-14 — 0 rows match `gate_%`~~ **SUPERSEDED 16:48:45Z: REACHED AND PASSED, once** (`resolved_all_gates_passed`, item `e713613f`). Still **LIVE on ≥`v1.0.1293`**; every REFUSAL arm remains at zero and unobserved |
| **`result.revalidation.arm` — the arm instrument** | **LIVE on `v1.0.1297`** (needle + both controls, one digest) + council APPROVED r1 |
| Council | **APPROVED first round**, `fe7dccb3-3038-4177-b77a-0cf620860556`, 10 seats, 7 abstained, 1 LOW advisory (actioned) |
| Fleet | **`v1.0.1298`** (`IMAGE_TAG` line 17); arm needle re-probed with control after the roll |
| `bugs_closed/262` (published gate) | CLOSED, live since `v1.0.1293` |

**The measurement this lane owes on every visit — 2026-08-14 (after the clean): `0 | 0 | 0 | 10 | 17 | 3`
of 30** (refused_gate1 / gate2 / gate3 / resolved / still_holds / unknown), invariant `t` for **10/10**,
zero `f` rows. ~~`0 | 0 | 0 | 9 | 18 | 3`, unchanged from 08-13~~ — superseded 16:48:45Z by the page
clean in the top banner. Queries: `RUNBOOK` § "the measurement this lane owes" and § "Where does the
ladder STOP?".

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

1. ~~NEXT ROLL: verify the arm field~~ ~~read the first instrumented sweep~~ **BOTH DONE — see the
   RESULT banner at the head of this file. This item is CLOSED.** The instrument is live, exercised, and
   its output agrees with an independent hand-reading taken before it existed. Nothing further is owed
   on it. ⚠ A zero gate count is the instrument working, **not** the gates approving.
2. ~~**The gates are still UNREACHED and no instrument changes that.**~~ **DONE 2026-08-14 16:48:45Z
   — see the top banner.** The owner authorised the intervention ("yes, clean a page"); leopardess
   `case-studies` was cleaned of a ~29x overclaim and the item closed at
   `resolved_all_gates_passed`. **This item is CLOSED.** What is NOT done, and is the successor: no
   gate has ever REFUSED. All three refusal arms remain at zero. **Do not describe any gate as having
   prevented anything** — a pass is not a proof of the refusal. The cheapest next observation is
   `gate_published_correction_unpublished`, which arrives free the first time a sweep lands between
   someone's component edit and their deploy; **wait for it, do not manufacture it.**
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
