# HANDOFF — 2026-08-21, fresh chat starts here: **step 5's precondition is nearly met.** The `?` marker, its adoption gate and three of the four blocking wires are LIVE. What remains is the flip itself, two verifications waiting on traffic, and one named residual.

**Supersedes `HANDOFF_2026-08-20_continue_here.md`** (whose §3 census is now mostly dispositioned)
and, transitively, `HANDOFF_2026-08-18b_continue_here.md`. Both stay readable for the step-1→4
evidence trail. **Do not take the older step-5 plans at face value** — the population they describe
has changed under them twice today.

**Companion file:** `HANDOFF_2026-08-21b_continue_here.md` is the `bugs_open/306` session's SHORT
SUPPLEMENT, not a competing "start here" — it corroborates 537's verification from their own
measurements and carries the handler-standardisation trail (migrations 519–540). **This file is the
primary.** We nearly wrote two documents to this same path within ten minutes of each other; if you
are about to add a third, check what exists first.

**Read in this order:** this file → NOTES `## 2026-08-21` entries (there are eight; they are the
day's record) → `WRONG_CALLS.md` 2026-08-21 entries (three, all mine) → `bugs_open/330` and
`bugs_open/334`.

---

## 1. THE ANSWER TO "CAN WE CLOSE THIS LANE?" — NO, AND HERE IS EXACTLY WHAT IS LEFT

**The lane exists to do one thing: make the resolver never guess (RFC_029 §9 D2 Phase 2).** That
flip — `findFieldRecursive`, conflicting candidates resolve to **nothing** — **IS STILL UNBUILT.**
Everything shipped today was its *precondition*, not the thing itself.

| # | what | state |
|---|---|---|
| **1** | **THE FLIP** — conflicts → refusal at the marked sites (`unified_extractor_search_test.go` header names them) | **CORRECTED ~19:1xZ (updated ~19:3xZ): BUILT + COMMITTED `5fe010ada` (Go — inert until a roll), council corr `26186633` at ROUND 3 (R1 + R2 both REVISE, bug_historian gating; commit unchanged throughout, only the rationale grew) — owned by the active [324079] session; see NOTES ~17:1x/~19:0x. Do NOT start it again.** ~~❌ NOT BUILT~~ |
| **2** | Retire the read-side tolerance in `setRenderContextScalarsFromData` (2nd `if`) + the "old tree"/"both present" cases of `TestRestoreAcceptsBothSpellingsAcrossTheRoll` | ❌ NOT DONE — the comment above that function names step 5's commit as its owner. **Use the two sound grounds in §5, NOT the retention argument** |
| **3** | `bugs_open/330` candidate 2 — an unmarked wired-but-empty field still falls through to the search | ❌ OPEN, gated on the **269-pair / 75-agent unsampled remainder** (330 §9) |
| **4** | A **standing** form of 537's guard | ❌ NOT BUILT — named residual, see §4 |
| — | `?` OPTIONAL-EXPLICIT marker | ✅ LIVE `v1.0.1321`, council APPROVED r3 |
| — | Adoption gate + acks + **daily CronJob** | ✅ LIVE and proven in-cluster |
| — | Migrations 512 / 515 / 516 / 537 (+ the 306 lane's 519–540) | ✅ ALL APPLIED |

**So: not closeable.** But the four blocking conflict classes named in the 08-20 handoff are now
either fixed or wired, and the flip is the next build rather than a distant one.

---

## 2. WHAT SHIPPED TODAY (all verified at the artefact, not at the tag)

- **The `?` OPTIONAL-EXPLICIT marker** — `"field?": "<ref>"` on `ExtractActionInputs`' step-config
  surface: that path or **ABSENCE**; never the whole-tree search, the nested-object fallback or the
  deprecated bridge; unlike `!`, a miss is not an error (Defaults stand, Optional fields are absent,
  Required fields fail ordinary validation). `!` beats `?` on a collision.
  **LIVE `v1.0.1321`**; council **APPROVED** round 3 (`5f82423b`).
- **One shared marker parser** — `datahelpers.MarkedConfigKey`, now used by BOTH surfaces
  (`input_mapping.go`'s two sites call it). It implements *input_mapping's* algorithm because that
  surface has **77 live `?` keys** and this one had zero. Pinned by
  `marked_config_key_parity_test.go` against all **56** live marker spellings.
  **Rides `v1.0.1322`** (see §3 — its live check is OWED).
- **The adoption gate** — `config-key-audit --optional-explicit-wires [--report]`, acks file at
  `architecture_review/optional_explicit_wire_acks.json`, **CronJob `optional-explicit-wires-check`
  daily 06:30 UTC**, image `v1.0.1321`, proven in-cluster by a manual `--from=cronjob` run
  (Succeeded, 4 wires / 4 acknowledged, `doc_notes` row written — checked at the POD, because
  `ImagePullBackOff` reports as a Job still RUNNING here). **5 wires, 0 unacknowledged** as of 18:0xZ.
- **Migrations applied:** `516` (both tool build steps → `related_pages?`, = `bugs_open/330`'s fix)
  and `537` (bdl `mark_complete` → `commit_sha?`, = `bugs_open/334`'s resolver half).

---

## 3. ✅ DONE 2026-08-21 ~19:1xZ (NOTES ~19:1x entry) — bdl 25 COMPLETED / 0 FAILED under real demand; the `RESOLVER_MAPPING_BYPASSED` zero discriminates NOTHING (class already 4 days silent). Was: OWED NOW — the `v1.0.1322` roll carries a live-surface refactor and it has NOT been checked

**`v1.0.1322`, both pods, one digest `sha256:68075cf5…`, up 2026-08-21 16:54Z, build revision
`bac189921` (which is HEAD).** It carries `383d1afbc` — the shared-parser extraction, which **edits
`input_mapping.go`, the resolver surface with 77 live `?` keys used by `build-dispatch-loop`,
`diagnose-dispatch-loop` and `report-dispatch-loop`.**

The change is behaviour-identical *by construction* (that surface's own algorithm, moved) and pinned
by the parity test — but **that is an argument, not a measurement, and this lane's whole subject is
the difference.** OWED, and it is the first job for the next session:

```sql
-- 1. are the three dispatchers still completing work after the 16:54Z roll?
SELECT owner_agent_type, status, count(*)
  FROM orchestration_states
 WHERE owner_agent_type IN ('build-dispatch-loop','diagnose-dispatch-loop','report-dispatch-loop')
   AND created_at > '2026-08-21 16:54:34Z'
 GROUP BY 1,2 ORDER BY 1,2;
-- Expect COMPLETED rows and no new FAILED class. Zero rows = NO DEMAND, not a pass.

-- 2. the mapped-field instrument must not have lit up
SELECT context->>'field', count(*), max(occurred_at)
  FROM agent_error_log
 WHERE error_code='RESOLVER_MAPPING_BYPASSED' AND occurred_at > '2026-08-21 16:54:34Z'
 GROUP BY 1 ORDER BY 2 DESC;
```
⚠ **A control I set for this and got wrong** — do not repeat it: I tried to prove the ancestry test
discriminates by picking a "later" commit, but **the build commit IS HEAD**, so no later commit
exists and the control failed by construction rather than by the test being broken. If you need a
negative control for `git merge-base --is-ancestor`, use a commit that is genuinely **not in this
history** (`git rev-list --all --not HEAD`), or note that the same test was proven discriminating
against `v1.0.1321` earlier the same day.

---

## 4. THE TWO VERIFICATIONS WAITING ON TRAFFIC — baselines are BANKED, do not re-derive them

Both fixes are applied and **neither is proven.** "Applied is not proven" is this lane's own
standing lesson and it has bitten twice today.

### 4.1 `bugs_open/330` — **FIRST READ ~19:1xZ: demand arrived (4 runs 18:47–18:55Z), RIGHT on 3 of 4 legs; the with-pages negative control is still demand-starved. n=4 < the ≥5 bar — not a close. Full read: 330 §10 + NOTES ~19:1x.** Migration 516, applied 2026-08-21 ~16:55Z

| figure (pre-apply) | value |
|---|---|
| tg/`related_pages` conflict rows | **8**, last `2026-08-21 14:15:34Z` |
| tool-generator runs in that window | **8** (≈1 conflict row per run) |
| cross-link items on webdesign.co.uk | 32 rows, 9 tools, 2 pages, **0 complete** |

**As of 18:2xZ: still ZERO tool-generator runs since the apply.** The 306 session independently puts
its last run at **16:40:11Z — before the 16:55Z apply**. So there is nothing to read yet, and a
"quiet class" reading here would be pure demand starvation. It runs ~8–16×/24 h, so allow hours.

⚠ **THE TRAP 516 CREATED FOR ITSELF:** `related_pages` **was** tool-generator's own instrument-alive
control, and 516 removed it. A post-apply zero on this class can no longer be checked against a
sibling class in the same agent — **the control must come from ANOTHER agent's live class**. Without
that, a zero is indistinguishable from a dead recorder. Full four-query verification (demand →
class → instrument control → **the artefact**) is in NOTES `## 2026-08-21 (~16:5xZ)`.

### 4.2 `bugs_open/334` — migration 537, applied 2026-08-21 ~15:5xZ

| figure (pre-apply) | value |
|---|---|
| bdl/`commit_sha` conflict rows | **881**, 2026-08-19 20:40:07Z → **2026-08-21 15:32:44Z** |
| items completed carrying a sha, last 24 h | **594** ← the demand control |

**The falsifier, stated so nobody reads a double zero as success:** after fresh multi-iteration loop
demand, conflict rows should read **0 WHILE items keep recording `result.commit_sha`** at roughly
594/24 h. **If BOTH go to zero, the wire is DROPPING the field rather than declaring it** — that is
the failure mode, and it is why the item count is the control rather than a second conflict class.

**FIRST READ, 2026-08-21 ~18:1xZ (≈2¼ h of demand) — the signal is RIGHT, on both halves:**

| metric | value | reading |
|---|---|---|
| bdl/`commit_sha` conflict rows since apply | **0** | pre-rate ≈ 20/h over 43 h, so ~45 were expected |
| items completing WITH a sha since apply | **22** (page-rerender 22 of 25) | the field is still being WRITTEN — not the double-zero failure |
| `tool-generator` items with a sha | **0 of 2** | **the prediction holds** — it never produced one |

**INDEPENDENTLY CONFIRMED ~18:2xZ by the `bugs_open/306` session, on a stronger sample than mine:**
**263 conflict rows in the 9 h BEFORE the apply, 0 since, against 19 real bdl runs post-apply** —
genuine demand, not a quiet window. Their per-handler spot-check: page-rerender 28/31 (3 legitimate
skips), css-patch-agent 2/2, tool-generator correctly 0/4, page-build-handler 0/4 *with all four
traced to the no-sections path*. **537 is VERIFIED**, by two sessions, two routes, neither assuming.

**Two things that look alarming and are NOT, both checked rather than assumed:**
- **`page-build-handler` recorded 0 of 4**, against a ~62% base rate (32/52 pre-fix). Checked at the
  reply, which is upstream of this wire: `handler_result.response ? 'commit_sha'` is **false** on
  those runs — **the handler sent no sha**, so no wire could have supplied one. Previously the search
  would have borrowed a sibling's. This is the fix behaving as designed, not a drop.
- **The instrument-alive control reads ZERO fleet-wide.** `agent_error_log` is alive (11 UNKNOWN,
  3 TIMEOUT, 1 PROCESSING_FAILED in the window) but there are **no `RESOLVER_*` rows of any class**.

⚠ **THAT SECOND ONE IS STRUCTURAL AND THE NEXT SESSION SHOULD PLAN FOR IT: as this lane succeeds, it
destroys its own instrument-alive control.** Steps 1–4 plus 512/516/537 have now silenced every
conflict class that was firing, so "some other class is still writing rows" is no longer available as
proof the recorder lives — and it will not come back. From here, a zero needs either **a deliberate
positive control** (provoke one known conflict on purpose and confirm the row appears) or
**mechanism-level evidence** (as step 4 used: the colliding key is no longer in the tree, so the
conflict is *unrepresentable* rather than merely unobserved). Do not accept a bare zero again.

⚠ **`tool-generator` should now record NO `commit_sha`. That is the FIX, not a regression** — it
never produced one; its single recorded sha belonged to a different work item in a different loop
iteration (item `cc1db035`, parent `62df9f7a`). Written into 537's header so a dashboard watcher
does not revert it.

---

## 5. WHEN YOU BUILD THE FLIP (step 5) — what is already settled, so you do not re-derive it

- **Flip sites** are named in the header of
  `platform/orchestration/datahelpers/unified_extractor_search_test.go`.
- **In the SAME commit, retire the read-side tolerance** in `setRenderContextScalarsFromData` (the
  second `if`) plus the "old tree"/"both present" cases of
  `TestRestoreAcceptsBothSpellingsAcrossTheRoll`. **The retention argument for this is UNSOUND — do
  not repeat it.** Two sound grounds instead:
  1. **Zero non-terminal pre-roll orchestrations** — re-run at the time; it only gets safer.
  2. **`buildRerenderBaseData` writes the NEW key fresh** from its `pageName` argument, and the
     tolerance's first branch `continue`s whenever `current_page_name` is present, so stored
     `page_components` rows never reach the second branch.
     > **CORRECTED 2026-08-21 ~19:5xZ (by the retiring session, from the code + a measurement):
     > the "never reach the second branch" half is FALSE as stated.** The short-circuit is
     > per-MAP, and the rerender path merges base / stored contentData / resolved_data as THREE
     > separate calls (`rerender_page_sections_action.go:628-631`) — all 18 live stored rows
     > carry the old key WITHOUT `current_page_name`, so in their own merge call the second
     > branch WAS reached and DID adopt. What saves the retirement is the measurement, not the
     > stated mechanism: **all 18 stored values agree exactly with their page's own name**
     > (0 differ), so the adoption was value-neutral — and the same fact exposes what the
     > tolerance really held open: a FUTURE stale stored string would have CLOBBERED the fresh
     > base identity (085's shape). Retired on the measured grounds in
     > `COUNCIL_SUBMISSION_2026-08-21_retire_read_side_tolerance.json` (corr `e05ea6f9`,
     > commits `e5c1b3c15` + `9970eb71c`).
- **The precondition is "zero conflict WARNs OR every observed field/caller pair explicitly
  mapped."** Re-run the census before trusting any of it (RUNBOOK, "the demand-control join").
- ⚠ **"Zero conflict WARNs" can NEVER establish that the search is safe.** A conflict row requires
  candidates to DIFFER; a tree with one match — or several that agree — substitutes **silently**.
  `bugs_open/330` was the worked case. Say this out loud in the design rather than inheriting the
  precondition's "or" branch as if it were sufficient.

**The residual worth building alongside it (§1 item 4):** 537's guard runs **once, at apply time**.
The handler population is per-item and **dynamic**, so a new commit-producing handler appearing
afterwards that does not expose at `handler_result.response.commit_sha` will simply never record
`result.commit_sha` — no error, no row, nothing to notice. The query is already written, in 537's
guard; it wants the daily-CronJob treatment the `?` gate just got. Same shape as RFC_022's
optional-key budget.

---

## 6. TRAPS EARNED TODAY — all three cost something, all three are in `WRONG_CALLS.md`

1. **A correction is only a correction where it is READ.** I fixed a landmine heading whose stale
   "NOT YET" had caused *two* false council objections — by appending the fix to the END of the
   heading, where truncation hides it. The next round objected that the entry "still shows the
   generic text". Put the state in FRONT of the explanation: the first 80 characters are what a
   snapshot, a listing or a `left(body,90)` shows.
2. **A demand census counts the defect when the defect is what supplies the value.** Sizing which
   handlers would lose `commit_sha`, I censused "who records one today" and reported three gaps. One
   (`tool-generator`) was `bugs_open/334` caught in the act — the search had glued a sibling
   iteration's commit onto an unrelated item. **Ask whether the handler's OWN tree can produce the
   value** (`orchestration_states` for that `owner_agent_type`), never whether an item recorded one.
3. **A test that pins a guard's INPUTS does not pin the guard.** My vacuity-guard test asserted the
   detector's inputs while its comment claimed a mutation proof of the branch; the branch sat behind
   `os.Exit`, and the mutation PASSED when I finally ran it. Extract the decision into a function so
   it can be called. **An unverified mutation-proof claim is worse than none** — the next reader
   treats the branch as covered.

Plus two more, both from council seats and both correct:
- **"All-time" counts are not sourceable here.** `orchestration_states` retention is per-status; its
  oldest row of any kind is 2026-07-19 and a given agent's window may be **one day**. State the
  window and re-measure inside it.
- **The gate caught its own author.** I applied 537 without writing its acks entry, so the gate went
  red on my own wire — the "wire and ack in ONE commit" rule I had pressed on two other lanes. A
  guard that only fires on other people is not evidence of anything.

---

## 7. COUNCIL / OWNERSHIP STATE

- **Nothing owed.** `5f82423b` (marker) APPROVED r3 · `101ed0c6` (516) APPROVED r3 · `2716123d`
  (537) **REVISE — answered in the file's header by four checks**, applied afterwards deliberately.
- **Three other sessions are active in this territory** and all three behaved well; route by BUG
  FILE, not by memory: `bugs_open/286` owns 331/TL-047 + migrations 496/532 · `bugs_open/306` owns
  the `334` handler standardisation (519–540) · `webdesign_tool_rebuilds` is the *finder* on 331,
  not its author. **Two sessions correctly refused to confirm an ack for work they did not own** —
  `git log` on the FILE is what settles authorship, and I routed one ack wrongly twice before that.
- **Provisional ack now CONFIRMED**: `create_tool_component.replace_existing` was confirmed
  first-hand by `bugs_open/286` (commit `24ba20ed9`).
