# HANDOFF — 2026-08-12 — council APPROVED, both gates LIVE and PROVEN on v1.0.1291. Read this file only

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
