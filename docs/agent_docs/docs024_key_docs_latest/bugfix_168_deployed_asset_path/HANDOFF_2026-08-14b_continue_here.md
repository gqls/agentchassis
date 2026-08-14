# HANDOFF — 2026-08-14b — A GATE HAS BEEN REACHED AND PASSED. Read this file only

**Supersedes `HANDOFF_2026-08-14_continue_here.md` for state** — that file's TITLE says the gates are
"observed unreached", which was true at 08:45Z and is false from 16:48:45Z. Its banners are history;
its §4 traps still hold except where superseded below. Working record:
`NOTES_deployed_asset_path.md` (newest at the bottom). Owner's log: `README_where_we_are.md`.
Read-out: `SUMMARY_2026-08-14b_deployed_asset_path.md`.

**Nothing is in flight. Nothing is half-applied. Everything below is committed.**
**Nothing is blocked and nothing needs a decision.**

---

## 1. State — verified 2026-08-14 17:04Z

| thing | state |
|---|---|
| `claims_unverified` revalidator | **LIVE + PROVEN** |
| **The three gates** (claim-granular · copy-changed · published) | **REACHED AND PASSED, once** — `resolved_all_gates_passed` on item `e713613f`, 16:48:45Z. **No gate has ever REFUSED**; all refusal arms are 0 and unobserved |
| `result.revalidation.arm` — the arm instrument | **LIVE**, council APPROVED r1, exercised twice |
| Fleet | **`v1.0.1299`**, 17 pods, **ONE digest** `sha256:2d247078…` (makefile `IMAGE_TAG` line 17 moves ahead of the fleet — read the pod, not the file) |
| Daily sweep anchor | **`last_triggered_at` = 2026-08-14 08:45:04Z — UNMOVED.** The 16:48 sweep was a manual dispatch, deliberately not a schedule wind-back |
| `bugs_closed/262` (published gate) | CLOSED, live since `v1.0.1293` |
| ⚠ Fleet LLM capability | **DOWN until 2026-09-01** — monthly spend cap, landmine filed by the bugfix_213 lane. Nothing on this lane needs it |

**The measurement this lane owes on every visit — 2026-08-14 (after): `0 | 0 | 0 | 10 | 17 | 3` of 30**
(refused gate-claims / gate-copy / gate-published / resolved / still_holds / unknown), invariant `t`
for **10/10**, zero `f` rows. Reach query: `refused_at_a_gate 0 | passed_all_gates 1 |
uninstrumented 0 | total 30`. Arms: `scan_still_trips 17 · resolved_all_gates_passed 1 ·
page_absent 2 · evidence_base_absent 1 · <no arm> 9` (the 9 are frozen pre-instrument closures — a
**vintage** marker, not a gap; the gap check is `arm LIKE 'unreported:%'`).

Queries: `RUNBOOK` § "the measurement this lane owes", § "Where does the ladder STOP?", and the new
§ "Cleaning a page so the gates can be reached".

---

## 2. What this session did

**Owner instruction: "yes, clean a page"** — the live-content intervention the previous handoff's §3.2
recorded as *the owner's* and did not take. Taken, and it closed that item.

### 2.1 The target, chosen against the ladder rather than by convenience

`leopardessconsulting.co.uk/case-studies`, item `e713613f-178e-4023-90e0-75edacee7ba6`, page
`ff5e75e3-…`, component `9c9aaed8` slot `case-studies-list`. It asserted **"75,061 orchestration state
records"** against register fact `C4-orchestration-state-records` = **2,578**, tolerance `gte` — a ~29×
overclaim, and **flagged independently by two checks**: `claims_unverified` `e713613f` and
`stale_evidence` `3a5419a1`. Not a false positive.

Selection criteria and the four rejected candidates are in the RUNBOOK. The one to carry:
**four of the six one-finding items were the WRONG act.** `finetuning.uk/privacy-policy` matched
**"16"** inside *"we do not knowingly collect personal data from anyone under the age of 16"* —
deleting that damages a legal notice. **Read the `snippet`, never just the `matched`.**

### 2.2 The act, and why it is not "writing content"

**One sentence deleted, 36 chars** — `75,061 orchestration state records. ` — no rewrite, no
substituted figure, no connective repair needed. That is the owner's 2026-08-06 *"minimal deletion is
not writing"*, and it is what the check's own `fix` text asks a human to do ("either an evidence_base
fact row … or removal from the copy"). Then rerendered and deployed through the framework.

Verified at the artefact, cache-busted, with a fabricated-URL control: **24,558 → 24,522 bytes**
(exactly −36), `75,061` count 0, control 404. Not from the orchestration status.

### 2.3 ⚠ CORRECTION to my own plan — the one-surface fix would have shipped a no-op

I planned **one** edit (`content_data`) believing the rerender regenerates `rendered_html` from it.
**It does not.** `RerenderSinglePageAction` is *"Simple concatenation - no template re-rendering"* and
contains **no `UPDATE page_components`** — it republishes stored per-component HTML. The claim would
have gone out unchanged behind a `COMPLETED` orchestration and a moved `deployed_at`, and the next
audit would have re-reported it, which reads as the audit being broken rather than as nothing having
happened. **Caught by reading the action before firing it, not by a symptom.** Filed in `LANDMINES.md`
(synced) with the one-line check.

Both edits ran under `DO/RAISE` guards that were **induced before being trusted**: run with the wrong
delta each aborted *after* performing the UPDATE (it can only report the true delta by having done the
work) and the rows came back byte-identical. A verify block of bare `SELECT`s could not have stopped
either COMMIT.

### 2.4 The result, against a prediction written into NOTES before the run

| predicted | observed |
|---|---|
| arm is `gate_`-prefixed OR `resolved_all_gates_passed` | **`resolved_all_gates_passed`** |
| not `scan_still_trips` / `gate_claims_still_present` / `page_absent` | none of them |

The closure records all three gates' inputs, so it states what it rests on:
`flagged_texts 1 / "all absent from their own slot"` · `item_filed_at 2026-08-08T17:13:14 <
newest_component_update 2026-08-14T16:43:49` · `deployed_at 2026-08-14T16:46:38 > newest_component_update,
build_status deployed`.

### 2.5 Side effects, which I caused and therefore own

The sweep is **fleet-wide (500 items, oldest-first)**, not per-item. Besides the target it closed
**4 further items** other lanes had genuinely fixed the same day, hours before the next daily would
have: `needs_page` on ai-agent-orchestration.com (model-directory) and two on webdesign.co.uk
(tool-seo-injector, tool-shadow-stacker), plus `unresolved_cta` on webdesign.uk/index. All
`auto:revalidated` and individually reversible. Nothing closed that the daily would not have.

---

## 3. What is next

1. **The first REFUSAL is the only observation still owed, and it is FREE.** Had the sweep landed
   between the component edit (16:43:49Z) and the deploy (16:46:38Z), the published gate would have
   returned `gate_published_correction_unpublished` — the `bugs_open/262` case exactly. The window was
   ~3 minutes; the rerender closed it. **Any item whose page is edited but not yet redeployed produces
   that arm on the next sweep, at zero content cost. WAIT FOR IT — do not manufacture one.** A second
   live-content change to chase an observation is a worse trade than patience, and the owner's
   authorisation was for *a* page, not a standing licence.
2. `features_open/032` — the shared helper. `arm` is set only by `revalidateUnverifiedClaims`; the
   other four record `unreported:<item_type>`. Lifting arms into them belongs with lifting the
   copy-changed comparison. **Measure before building.**
3. Round 7's remaining `editquality` LOW: a before/after test for the SQL→Go locked-skip move
   ("emitted output is unchanged" is asserted, not demonstrated). **Still not done.**
4. Leftovers from the 08-11 handoff §3.5: §2.3 pin `ScanDeployedClaims` to its intended callers ·
   §2.4 the invisible backlog · §2.5 Decision 2's dedup half · §2.6 more sweep coverage · §2.7 the
   armed-but-inert cap at `check_image_source_unsatisfiable.go:167`.

---

## 4. Traps

**New this session** (the first is in `LANDMINES.md`; the rest are lane-local):

- ⚠ **The assemble rerender never regenerates `rendered_html` from `content_data`** (§2.3). Cheap
  check before dispatching anything: `grep -c "UPDATE page_components" <the action>.go` — **0 means it
  does not write that table**, whatever its name suggests. And `ScanDeployedClaims` reads BOTH surfaces
  (`rendered_html` for banned/number scans, `content_data` for stat scans) while the claim-granular
  gate searches `html || contentJSON`, so **a one-surface edit leaves the finding standing whichever
  surface you picked.**
- ⚠ **psql does NOT interpolate `:vars` inside a dollar-quoted body** — `DO $$ … v := :x … $$` fails
  with `syntax error at or near ":"`. Pass it via a GUC set outside the quotes:
  `SET LOCAL app.expect_delta = :expect_delta;` then `current_setting('app.expect_delta')::int`.
- ⚠ **A short `matched` token disqualifies a target**, it is not merely awkward. `claimStillOnPage` is
  a case-insensitive substring over the slot, so `"3"` / `"11"` / `"100%"` will match unrelated prose
  for ever and pin the item at `gate_claims_still_present`.
- ⚠ **Some claims_unverified findings are genuine false positives on pages you must not edit** —
  an age threshold in a privacy policy read as an unregistered number. Read the `snippet`.
- ⚠ **A manual sweep is fleet-wide** (§2.5). Report what it closed; do not wind
  `scheduled_tasks.last_triggered_at` back to fire it — that moves the daily anchor permanently on a
  row this lane does not own. Mirror `cmd/scheduler/main.go` `fireTrigger()` and name the sender
  `from_agent_type=cli` with an `orchestration_name` beginning `manual-`.
- ⚠ **The `page-rerender` pod restarts often.** Check `.status.startTime` before dispatching —
  CLAUDE.md's ~300s post-restart spawn-drop rule bit this session (pod restarted 9s before dispatch).

**Carried forward, still live:**

- ⚠ **`result.revalidation.arm` is LAST-WRITE-WINS.** Drop the words *ever*, *how often*, *rate*. The
  structural tell: a `jsonb_build_object` assignment means no second row per item can exist.
- ⚠ **`resolved_all_gates_passed` carries NO `gate_` prefix**, so a prefix-only reach query counts only
  refusals and **misses every closure** — which now inverts the reading, because the one reach we have
  IS a closure. Use `arm LIKE 'gate\_%' OR arm = 'resolved_all_gates_passed'`; `_` is a LIKE wildcard.
  `arm IS NULL` is a **vintage** marker (pre-instrument closures), not a gap.
- ⚠ **An `archived` page can be SERVING.** Never conclude a page is dead from `pages.status`; curl it
  with a fabricated-URL control on the same domain.
- ⚠ **The council trigger refuses `plan.risks` as an ARRAY** — one prose string. (Moot until the LLM
  cap lifts on 2026-09-01.)
- ⚠ **`- **status-evidence:**` appears in EVERY register entry.** Target the LAST occurrence when
  appending; check `wc -l` against the expected delta.
- ⚠ **A bare `git stash` by another session deleted this lane's uncommitted tests on 08-12.** Commit
  before mutating; recover with `git checkout stash@{0} -- <path>`, never a bare `pop`.
- ⚠ **`LANDMINES.md` takes same-file passengers.** This session's commit carried the bugfix_213 lane's
  uncommitted LLM-cap entry; verified whole (8 bullets) first. `git show <sha> --numstat` distinguishes
  a modification from a deletion in one line.

**CLEARED — do not carry these forward:**

- ~~The shared package would not compile (another session's WIP in `palette_specialised_slots.go`)~~ —
  **re-checked 2026-08-14 17:04Z: `go build ./platform/orchestration/actions/...` returns rc=0.**

---

## 5. Commits, scripts and correlations

**This session: `1cbda2109`** — docs, landmine, the four worked scripts. **No platform code changed,
so no council submission is owed** (gate scope is `platform/`, `internal/`, `pkg/`). Earlier:
`92b59138b` the arm instrument · `bb05ce78a` gofmt · `ac6a86f58` the snapshot-not-history correction.

Worked scripts, committed into this directory and referenced by the RUNBOOK:
`SQL_2026-08-14_clean_case_studies_content_data.sql` · `SQL_2026-08-14_clean_case_studies_rendered_html.sql`
· `SCRIPT_2026-08-14_rerender_case_studies.sh` · `SCRIPT_2026-08-14_fire_revalidation_sweep.sh`.

| what | id |
|---|---|
| **the page-clean rerender** | orch `d150ca27-1ca9-4e04-8d5c-01f72248a6f0`, corr `d02c4452-6ccd-4532-bf57-8dfecaeefc85` |
| **the manual sweep that reached the gates** | orch `84db99fc-5ebd-4bd7-9d1e-0272c7fd7557`, name `manual-review-queue-revalidate-20260814-164838` |
| arm instrument council — APPROVED at r1 | `fe7dccb3-3038-4177-b77a-0cf620860556` |
| claims_unverified council — APPROVED at r7 | `b67eb26a-14ef-45d7-b755-3e489fd57ef0` |
| voice_tells council (APPROVED r1) | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |

Verdicts saved verbatim, newest first: `VERDICT_2026-08-13_APPROVED_revalidation_arm.json`,
`VERDICT_2026-08-12_round7_APPROVED_claims_unverified_council.json`, `VERDICT_2026-08-12_round6_*.json`,
`VERDICT_2026-08-11_round5_*.json`.
