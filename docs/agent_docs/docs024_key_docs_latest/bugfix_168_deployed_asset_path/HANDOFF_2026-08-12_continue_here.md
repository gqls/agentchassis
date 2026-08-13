# HANDOFF — 2026-08-12 — council APPROVED, gates LIVE on v1.0.1291, and `262` is FIXED. Read this file only

> ## 🔵 2026-08-13 UPDATE — the predicted run HAPPENED and REACHED NO GATE. Read this first; it supersedes item 2 of the night banner below
>
> **Nothing is in flight. Nothing is half-applied. Everything below this banner is committed.**
> Fleet: **v1.0.1295** (`IMAGE_TAG` line 17), needle re-probed `NEEDLE_262=1 rc=0` — the fix survives
> the fresh build. (Positive needle only; regression check, not yesterday's full control set.)
>
> **THE RESULT.** The daily sweep fired **2026-08-13T08:44:39Z** and decided **21** items. Every
> refusal counter is **still 0**; `resolved` still **9**; invariant `t` for 9/9. **Not one of the 21
> reached any gate.** From the verdicts' own reasons: **18 `still_holds`** = *"page X still carries N
> claim(s) the register does not support"* (the scan itself still trips), **2 `unknown`** = page absent
> or no readable content, **1 `unknown`** = site has no `evidence_base` spec.
>
> ⚠ **ORDERING VERIFIED, not assumed:** the run executed against **v1.0.1293**, which carries all three
> gates — 1293 rolled 08-12 19:13:54Z, 1295 only at 08-13 13:53:19Z, so 08:44:39Z sits inside 1293's
> window. The result is real, not an artefact of an old binary.
>
> > **CORRECTION to my own 08-12 wording.** I called this run *"the first that can exercise either
> > gate"*. **The run is necessary and nowhere near sufficient.** All three gates sit **downstream of a
> > clean scan**; a clean scan happened **0 times in 21**. The zero has now had **four** meanings in two
> > days: never shipped · shipped but never asked · asked-and-approved (never true yet) · **the code ran
> > and the ladder stopped above it.** Only reading the REASONS distinguishes them — no count can.
> > **Never count sweep runs as exercises of a late-ladder arm.**
>
> **STANDING STATE: the three gates are ARMED BUT INERT and remain UNOBSERVED.** What will ever reach
> them is an item whose page has genuinely been cleaned; the population is dominated by findings whose
> claims are still on the page (the sweep working correctly). Every closure to date predates gate 2.
> **Do not describe any gate as having prevented anything.**
>
> ### What is next (unchanged from the night banner except item 1)
> 1. ~~Watch the 08-13 run~~ **DONE, above.** The open question it leaves: **how would we ever observe
>    these gates?** Either instrument the arms (count reaching, not just refusing) or find or construct a
>    cleaned page. Worth a decision before more gates get built on the same pattern.
> 2. `features_open/032` — the shared helper. **Measure before building.** Still open, untouched.
> 3. §3.5's leftovers; §3.4's remaining `editquality` LOW (before/after test for the SQL→Go move).
> 4. Incidental live confirmation of a known landmine: `page index-rejected-v1-20260806 still carries
>    14 claim(s)` — `ScanDeployedClaims` has **no page-status filter**, so archived pages are judged.

> ## 🟢 NIGHT UPDATE (2026-08-12, later) — `262` is CLOSED, and the gates are LIVE but UNASKED. Read this before §1
>
> **1. `262` verified live and CLOSED** → `bugs_closed/262…` (`d142fcd27`). Fleet on **v1.0.1293**,
> 98 Running chassis pods, ONE digest `sha256:4717bcb3`, rolled 19:13:54Z.
> `NEEDLE_262=1 rc=0 · NEEDLE_262b=1 rc=0 · NEGCONTROL_oldselect=0 rc=1 · CONTROL_pos=1 rc=0 ·
> CONTROL_absent=0 rc=1`.
> ⚠ **Stated limit:** `ce8733262`'s only deletions in the shipped file are two *function signatures*,
> so **this commit has no removed-string control of its own**. `NEGCONTROL_oldselect` proves the
> binary is newer than `ea18664f3`, **not** newer than `ce8733262`. A negative control does not
> transfer forward to later commits — what proves this one shipped is the two present needles.
>
> **2. ⚠ ALL THREE GATES READ `0` REFUSALS, AND THE THREE ZEROES MEAN THREE DIFFERENT THINGS.**
> The last sweep ran **08-12 at 08:44:34Z**. The claim-granular gate was committed **12:56Z** and the
> published gate **17:42Z** — *both after it*. So **neither has ever been asked a single question**;
> only the copy-changed gate has (30 item-decisions, never refused). Driver is
> `scheduled_tasks.review-queue-revalidate-daily` (86400s, `enabled=t`).
> **The next run, ~08-13 08:44Z, is the first that can exercise either**, against 21 open items.
> **Do not quote either zero as "the gate approved" until then.**
> ⚠ **`result->'revalidation'->>'at'` is LAST-WRITE-WINS, not a run log** — every open item carries the
> identical stamp. I read a run history off it, said 08-11 was skipped, and **withdrew it**
> (`WRONG_CALLS.md` + LANDMINE). `max(stamp)` vs the ROLL time is the question that survives.
>
> **3. §3.4's `bug_historian` advisory is ACTIONED — the producer claim HOLDS, the thing beside it
> does not.** The rerender paths neither file these items nor write `page_components` at all (opened,
> not grepped). **But gate 1's premise is violable:** `page_components.updated_at` is bumped with no
> copy change by `fix_component_template_action.go:853` and, page-wide, by `v3_site_actions.go:4205`.
> **Not a live defect** — the claim-granular gate answers FIRST and is immune, and neither writer has
> ever fired (`repair_page_component_status` **0** of 5,547 orchestration rows against a control of
> **28**; `reviewed_at` NULL on all 1,458 components). **This is the mechanism-level reason the
> owner's KEEP-BOTH ruling is right, where the original argument was reversibility alone.** Full
> entry in CQ-021 + `LANDMINES.md`.
> ⚠ **Residual:** all 9 existing closures predate gate 2, so they rest on gate 1 alone, and that is
> **not retrospectively measurable** — `updated_at` keeps no history.
>
> **4. Method warning for this area — RUN THE CONTROL BEFORE BELIEVING A ZERO.** Three discriminators
> came out **inert** in one session: `reviewed_at ≈ updated_at` (NULL on all 1,458),
> `workflow_plan->>'agent_type'` (NULL on all 5,547), and the revalidation stamp as a run log. Each
> gave a clean, actionable-looking number that could not have come out otherwise.
>
> ### What is next now
> 1. **~08-13 08:44Z: watch the first run that exercises gates 2 and 3.** Queries in §1.
> 2. `features_open/032` — the shared helper. **Measure before building.** Still open.
> 3. §3.5's leftovers; §3.4's remaining `editquality` LOW (before/after test for the SQL→Go move).

> ## 🟢 EVENING UPDATE — the owner's decision is TAKEN, and `bugs_open/262` is FIXED. Start here
>
> **OWNER DECISION (2026-08-12): KEEP BOTH GATES.** The ANDed-gates question in §3.3 below is
> **CLOSED** — do not re-open it as an open item. The reasoning that settled it: `refused_by_gate` is
> **0**, so the copy-changed gate has never actually blocked a closure; keeping it costs nothing
> observable and is the reversible choice, while removing it could not un-close anything later.
>
> **THERE ARE NOW THREE GATES**, and `resolved` requires all three:
> 1. **copy-changed** (owner, 08-09) — an examined component edited after the finding was filed;
> 2. **claim-granular** (council r6) — every cited text gone from the slot it was cited from;
> 3. **published** (`262`, 08-12) — the page was published *after* its copy last changed.
>
> ### `bugs_open/262` — taken on and fixed (`ce8733262`), OPEN until it rolls
>
> Owner said take it if unowned. **Ownership checked three ways first**: `who-owns.py` clean, **0**
> open work items matching, and the three live transcripts naming the file were all incidental (two
> `ls` listings, one `git status`). **A filename in a transcript is not ownership.**
>
> Four refusal arms, all non-terminal and deliberately spelled differently so *"I could not look"*
> never shares a spelling with *"it has not shipped"*: page row unreadable · never deployed ·
> `deployed_at` precedes the newest examined component update · `build_status` not `deployed`.
> ⚠ **`build_status` and the timestamp are checked SEPARATELY — neither implies the other.**
> ⚠ **Equal timestamps RESOLVE** — a same-instant publish is a real ordering, and refusing it would
> strand items whose clocks agree. Placed **last** in the ladder so the more informative refusals
> answer first. **Five mutations, five correct failures.**
>
> **Next roll must verify it.** New needle for the standing pod-probe:
> `"the database is not the website"` — expect **1** once rolled, and it is currently **0**.
>
> ### ⚠ A BARE `git stash` BY ANOTHER SESSION DELETED MY UNCOMMITTED TESTS MID-RUN
>
> `go vet` failed against call sites updated ten minutes earlier and watched pass, while
> `git status` showed the file **clean**. `git stash list` named it — `stash@{0}: WIP on
> 087_towards_multiple_domains`. **A pathspec-less `git stash` takes the whole tree**, and unlike a
> `git add -A` sweep — which at least *commits* your work where `git log` can find it — a stash
> leaves status clean and the file back at HEAD. Recover with
> `git checkout stash@{0} -- <your path>`, **never a bare `pop`** (it applies their whole stash).
> Merged into the existing `git stash` LANDMINE entry, which covered only the *popper's* side.
>
> > **CORRECTION carried from earlier in that session:** I reported five mutation results where two
> > ("build failed", blamed on my own `sed`) **had never run** — the build failure was the stash
> > arriving mid-run. Re-run after committing: all five fail exactly the test written for them.
> > **A mutation loop is the worst place to be robbed: it expects the file to change under it, and
> > the restore step hides the theft.** Commit before mutating.
>
> ### What is actually next
>
> 1. **Verify `262` on the next roll** (needle above), then close it — the bar is fixed AND live.
> 2. `features_open/032` — lift the copy-changed comparison into a shared helper before `voice_tells`
>    gets a second bespoke gate. **Measure before building**; the claims answer may not transfer.
> 3. Unactioned advisory from the approving round: check the single-producer claim against the
>    **rerender** paths (`rerender_page_sections`, `rerender_single_page`) — 016b §9 case `093`.
> 4. §3.5's leftovers, noting §2.1 is DONE and §2.4's stated blocker was FALSE.


**Supersedes `HANDOFF_2026-08-11_continue_here.md` for state** (its banners are history; its §4 traps
and §5 open-items list still hold except where corrected below). Working record:
`NOTES_deployed_asset_path.md` (newest at the bottom). Owner's log: `README_where_we_are.md`.

**Nothing is in flight. Nothing is half-applied.** Three things are open, all named in §3.

---

## 1. State — verified 2026-08-12

| thing | state |
|---|---|
| `claims_unverified` revalidator | **LIVE + PROVEN** |
| **Owner's copy-changed gate** (2026-08-09) | **LIVE**, held on 8/8, **still `[UNEXERCISED]`** — `refused_by_gate` = 0 |
| **Claim-granular gate** (council r6) | **LIVE on v1.0.1291**, not yet observed firing |
| **Shared-loader SELECT/Scan guard** | **LIVE**, held by a test rather than a comment |
| **Council** | **APPROVED round 7**, 2026-08-12 14:34:14Z, after **six** REVISE |
| `bugs_closed/168` (asset path) | CLOSED, live since v1.0.1229 |
| Covered item types | 6 |

**Deploy proof, and it is the strongest this lane has had** — 56/56 Running pods, ONE digest
`sha256:382a523a`, tag **v1.0.1291**:

```
ownergate=1 claimgate_NEW=1 parityfix=2 NEGCONTROL_oldselect=0 rc=1 CONTROL_pos=2 CONTROL_absent=0
```

⚠ **`NEGCONTROL_oldselect` is a TRUE negative control — this lane's first.** The loader guard
replaced a literal `SELECT id::text, site_id, item_type, …` with `strings.Join(parkedReviewItemColumns, ", ")`,
so that literal is **a string the change REMOVED** (present at `ea18664f3~1`, absent at `ea18664f3`).
That is what `bugs_open/153` asks for, and the RUNBOOK previously recorded it as unavailable here.
**Its absence proves the binary is NEWER than the commit**, which `CONTROL_absent` (a fabricated
string) never could. Keep using it. The `rc=1` on both zeroes is what separates "grep ran and found
nothing" from "I could not look".

### The measurement this lane owes on every visit — re-run it

```sql
SELECT count(*) FILTER (WHERE result #>> '{revalidation,reason}' LIKE '%register moved, not the page%') AS refused_by_gate,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='resolved')    AS resolved,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='still_holds') AS still_holds,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='unknown')     AS unknown
FROM site_work_items WHERE item_type='claims_unverified' AND result ? 'revalidation';

-- the load-bearing invariant: EVERY closure must show the copy moved. Zero `f` rows, always.
SELECT (result #>> '{revalidation,evidence,newest_component_update}')::timestamptz
         > (result #>> '{revalidation,evidence,item_filed_at}')::timestamptz AS copy_actually_changed,
       count(*) FROM site_work_items
WHERE item_type='claims_unverified' AND resolution_path='auto:revalidated' GROUP BY 1;
```

**2026-08-12: `0 | 8 | 19 | 3`, invariant `t` for 8/8.** (Population has since moved to 9 `complete`.)
⚠ **Do not describe either gate as having prevented anything until it is observed refusing.** Both
are pinned by unit tests, not by observation.

**NEW, now that the claim-granular gate is live — watch for its first refusal:**
```sql
SELECT count(*) FROM site_work_items WHERE item_type='claims_unverified'
  AND result #>> '{revalidation,reason}' LIKE '%STILL in the component they were cited from%';
```

---

## 2. What the seven rounds actually taught, and it is not what it looks like

**The round that passed was the SMALLEST submission of the series** — 57,989 → **45,943 bytes**, with
the council narrative stripped out of the rationale, the risks, all eight edit rationales and the
evidence list.

Rounds 4, 5 and 6 each answered a seat by **adding** material, and each drew fresh objections from a
*different* seat on the material it had just added. Round 6 is the clean demonstration: the paragraph
written to fix `guardian`'s round-5 contradiction drew a gating HIGH from `prior_art_librarian` **and**
a MEDIUM from `editquality`. **The submission's prose had become the objection surface rather than
the code.** `constitution` named it directly ("dramatic ALL-CAPS meta-narrative … self-congratulatory
measurement claims") and was right.

**So: when a council round revises on evidence rather than design, cut before you add.**

Two other things that generalise:

- **"I cannot verify this" is answered by a query, not by an argument.** Every gating objection in
  rounds 4–6 named the exact corpus it would accept. Each time the fact existed and had simply never
  been handed over. Round 6's gate was a CLAUDE.md owner ruling with **zero** `doc_notes` trace; the
  fix was to make it visible (file the landmine, run `landmines-sync.py --apply`), not to assert it
  harder. See the fleet landmine *"CLAUDE.md's owner rulings are invisible to council seats"*.
- **A seat can be right about the defect and wrong about the remedy, and the remedy is the half you
  can measure.** `compliance` asked for the cited **snippet** to be compared; measured against the
  population where the answer is known, snippet saw a present claim **18/41** against the token's
  **40/41**, and in a gate a missed match reads as "the copy changed", which **grants** closure. Its
  literal suggestion would have failed open on ~56% of claims.

---

## 3. What is open

### 3.1 `bugs_open/262` — the real defect the APPROVING round found

Both gates judge **`page_components`, the database**, as ground truth for closing a finding about
**what a live site asserts**. Neither the scan nor either gate reads `pages.build_status` or
`pages.deployed_at` (`grep -c` → **0** in both files; both columns exist and are populated). So a
component edited in the DB but not yet rerendered/deployed satisfies both gates and the item closes
while the served page may still carry the claim.

**[MEASURED 2026-08-12] 2 of the 9 `complete` items sit on pages whose newest unlocked component
update is later than `deployed_at`.** That does **not** prove those pages still carry the claims — it
proves the closure's evidence **cannot show they do not**.

Fix candidates are ranked in the bug file; candidate 1 (refuse to close unless the page deployed since
the copy changed) is ~one column on a row the sweep already loads. ⚠ **Do not "fix" the emit side to
match** — parity is deliberate, and the two ends differ in *consequence*, not predicate.

### 3.2 `features_open/032` — `voice_tells` has the same hole and neither gate

Deliberate (its surface is style, not truth) but now tracked rather than an accepted risk with nothing
behind it. `reuse_agent` wants the copy-changed comparison lifted into a **shared helper** before a
second bespoke gate gets written. **Measure before building** — the claims answer may not transfer.

### 3.3 The owner's open question — the two gates are ANDed

The claim check is the **stronger** evidence: a token that has left its slot *proves* the copy moved,
where a timestamp only asserts it. It could **stand in for** the timestamp rather than join it.
Requiring both refuses a genuinely-fixed page whose `updated_at` was never bumped. Left ANDed because
**adding** a condition in front of an owner-mandated control needs no ruling; **removing** his
comparison would. One-line change either way. Recorded in `README_where_we_are.md` and in CQ-021.

### 3.4 Unactioned advisory objections from the approving round

- `bug_historian` MEDIUM: the single-producer claim should be checked against the **rerender** paths
  (`rerender_page_sections`, `rerender_single_page`), not only current call sites — 016b §9 case `093`
  is that exact shape in this exact area. **Not done.**
- `editquality` LOW: "emitted output is unchanged" for the SQL→Go locked-skip move is asserted, not
  demonstrated by a before/after test.

### 3.5 Still open from the 08-11 handoff §5

§2.1 claim-granularity — **DONE, this is the gate that shipped.** §2.2 the two-standards asymmetry →
now `features_open/032`. §2.3 pin `ScanDeployedClaims` to its intended callers · §2.4 the invisible
backlog (⚠ **its stated blocker was FALSE** — `code_symbols` indexes 700 package-level vars; whether
to file remains a judgement) · §2.5 Decision 2's dedup half · §2.6 more sweep coverage · §2.7 the
armed-but-inert cap at `check_image_source_unsatisfiable.go:167`.

---

## 4. Traps specific to this lane (corrections to the 08-11 list marked)

- ⚠ **Use the NEGATIVE control now that one exists** (§1). `CONTROL_absent` is fabricated and only ever
  proved grep can return 0.
- ⚠ **`grep -a` on `/proc/1/exe`, never `strings`**, and **capture the exit code** beside every count —
  `n=${n:-0}` converts "I could not look" into "it is not there". Filter to
  `--field-selector=status.phase=Running`: a completed job pod cannot be exec'd at all.
- ⚠ **Prove the fleet is ONE binary by DIGEST, never by a replica count.** `-l app=agent-chassis`
  returned 2 pods while 56 run that image. The pod count is churn; digest uniformity is the invariant.
- ⚠ **Comparing a work item's flagged text against live copy PAGE-WIDE always finds it** — bare
  `unregistered_number` tokens are 1–4 chars and match any markup. Scope to the slot; print the
  strings, never the count. Full entry in `LANDMINES.md`.
- ⚠ **Query a ruling's WORDS, never its date.** `body ILIKE '%2026-07-29%'` → 130 rows and means
  nothing; the ruling's actual phrase returned 0 until it was synced.
- ⚠ **The revalidation stamp key is `at`, NOT `checked_at`.**
- ⚠ **`uncovered_backlog` CANNOT confirm an adoption** — confirm at the per-type map.
- ⚠ **A dispatch of this sweep CANNOT be scoped** — both filters read from step config; the live
  `sweep` step has no `input_mapping`.
- ⚠ **The council refuses >8 edits CLIENT-SIDE** (`097_TRIGGER:101`), before credits are spent.
- ⚠ **`landmines-sync.py --apply` before `landmines-verify-dispatch.sh`** — CLAUDE.md's ordering.
- **Registering a revalidator is the whole wiring** — `coveredItemTypes()` derives from
  `reviewRevalidators`; `TestRevalidatorCoverageIsDeliberate` pins the set.
- > **CORRECTED 2026-08-12:** the 08-11 handoff said `platform/orchestration/actions` has failing
  > tests that are not this lane's (`TestEveryCheckProducedItemTypeIsClassified`). **It passes at
  > HEAD** — another session fixed it. Full package is green. Caveat retired.
- **`/tmp` is a near-full 16G tmpfs** — `TMPDIR=/home/ant/.cache/buildtmp go build ./...`.
- **Use `git commit -F <file>`**; backticks in `-m` execute.

---

## 5. Commits and correlations

`ef80216be` voice_tells · `4030cadb9` claims_unverified + CQ-021 · `9a9fef332` the owner's gate ·
**`58bede8d5` the claim-granular gate** · **`ea18664f3` the loader guard + parity fix** ·
`a3ccb3433` round 7 submission + `features_open/032` · **`555b09283` APPROVED, live proof, `bugs_open/262`**
· `d8066fcfa` the fleet landmine on invisible rulings.

⚠ **`git log` the FILE, not my commits** for two entries: the LANDMINES + WRONG_CALLS entries of
2026-08-12 landed in the 215 lane's `f8ca05594`, swept in the ~90 s before my own commit.

| what | id |
|---|---|
| **claims_unverified council — APPROVED at r7** | `b67eb26a-14ef-45d7-b755-3e489fd57ef0` |
| voice_tells council (APPROVED r1) | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |

Verdicts saved verbatim: `VERDICT_2026-08-11_round5_*.json`,
`VERDICT_2026-08-12_round6_*.json`, `VERDICT_2026-08-12_round7_APPROVED_*.json`.
