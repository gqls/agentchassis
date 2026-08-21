# HANDOFF — 2026-08-20 (rev. 2026-08-21 ~11:3xZ), fresh chat starts here: **steps 1–4 LIVE + PROVEN. Tier A is EMPTY. The `?` enabler is LIVE on `v1.0.1321`. Two live classes left (`bdl/commit_sha`, `tg/related_pages`), one migration in review (515, `pbh/page_type`), and one applied-but-unverifiable (512).**

> **⚠ THIS FILE NOW CONSOLIDATES TWO.** `HANDOFF_2026-08-18b_continue_here.md` was still being
> updated by a parallel session of this lane until ~10:17Z today (its audit results are folded in
> below and credited). **That file is now bannered SUPERSEDED — read only this one.** Two files
> both saying "fresh chat starts here" existed for ~9 h; that was my doing, and §7 records it.

**Read in this order:** this file → NOTES `## 2026-08-20 (evening)`, `(morning)`, `## 2026-08-19 (night)`
→ `bugs_open/330` (090 CONFIRMED; **§9 carries the sizing audit + its two dated corrections**) →
`bugs_open/334` (bdl/`commit_sha`) → the CONTRIB at
`docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/CONTRIB_2026-08-20_from_staged_component_build_commit_sha_resolves_by_guess.md`.

## 1. What is true now (measured 2026-08-20 ~17:00Z)

- **Live build `v1.0.1320`**, both pods up 16:09Z. **Nothing of this lane is in it** — steps 1–4
  all shipped in `v1.0.1317` and step 5 is unbuilt. The 1320 roll is a **non-event here**; do not
  manufacture a verification task for it.
- **Steps 1, 2, 3, 4 are all LIVE and PROVEN.** Step 4 verified twice independently: at the
  artefact on both pods (capability probe `current_page_name` + present- and absent-controls) and
  by the parallel session (stamp `2d13d530d`, 3/3 trees filing `current_page_name`). pcw/
  `current_page`'s last-ever conflict row is **2026-08-19 22:24:39Z — two minutes before the roll.**
- **Step 5 is NOT built.** `findFieldRecursive` still carries only the Phase-1 warn
  (`grep -c "conflicting = true"` → 1; no Phase-2 arm). **`bugs_open/334`'s candidate 1 is NOT
  built either** — no live step config wires `commit_sha` (only `code-indexer/index_symbols` and
  `site-work-orchestrator/build_items_loop` mention it, neither being the caller).
- **Nothing owed on any council review.** Prune `ae0dfb93`, tie-break `96ac93e6`, gate `07468ec0`,
  step 4 `f3716ebe` — all APPROVED, all shipped.
- **`bugs_open/330`** — 090 CONFIRMED. **Its owed sizing audit is DONE** (parallel session, 10:12–10:17Z):
  451 plain Strategy-0 wires fleet-wide / 309 pairs / 83 agents; stripped-probe runtime sample of
  the 8 high-demand agents → **4 genuinely rescue-prone wires** (the naive `LIKE` probe said 10 and
  over-counted by matching `agent_config`/`__raw_message__`/`retry_payload`, which the search skips
  via `isInfrastructureKey`). Of the 4, **pbh `page_id`+`page_name` produce AGREEING candidates**
  (0 genuine disagreements / 40 runs) so they are **invisible to step 5's conflict flip** and matter
  only to 330 candidate 2. Wrong-value population on the sampled slice = **exactly 330's own wire**.
  Unsampled remainder: 269 pairs / 75 agents; the corrected probe is reusable (RUNBOOK).
- **`bugs_open/334`** — filed for bdl/`commit_sha`. Its 090 came back **UNVERIFIABLE at the
  iteration cap** = no information (016b §9's standing reading), so the mechanism rests on two
  sessions' first-hand verification. Two corrections landed there: the onset is the **486/487
  traffic batch**, not an adapter roll; and `Deprecated: commit_sha_field` is a **LIVE Strategy-3
  bridge** which **cannot** stop the conflict (it runs *after* the search and is gated on a missing
  value) — so candidate 1 must be the Strategy-0 `commit_sha?` form.

## 2. THE KEY ADVANCE (2026-08-20 evening): step 5's list is TIERED, not flat

Every conflicting field was checked against the `ActionInputSpec` that declares it. **All but two
are `Optional`.** That splits the ~13 pairs into three tiers with completely different costs, and
it is the design step 5 needed:

> ### ⚠ 2.0 — SECOND AXIS, added ~17:5xZ by the other session of this lane: cross the tiers with **WIRED?** and **WHICH SIDE OF THE PRUNE ROLL?**, and Tier A empties
>
> The tiering asks *what does the spec say*. The complementary question is *does the step's config
> WIRE the field, and did the row arrive before or after the step-1 prune shipped*. Crossing the two
> **dissolves both Tier A entries and six of Tier C**, and it answers the trigger §2's Tier A note
> calls "unidentified".
>
> **The mechanism, and it is in the prune's own comment** (`action_inputs.go:767-781`): the prune
> (step 1, live `v1.0.1310`) removes a Strategy-0-resolved field from what Strategy 1/2 request.
> **Before it shipped, an already-resolved field was STILL handed to `ExtractFields`** — the search
> ran, `findFieldRecursive` wrote its conflict row, and then the merge threw the answer away
> (`if _, alreadyResolved := result.Values[k]; alreadyResolved { continue }`, both Strategy 1 and 2).
> The comment quantifies exactly this population: *"~28% of all conflict rows, measured 2026-08-18"*.
> So a pre-prune row on a WIRED field is **noise that never affected a value** — the winner never won.
>
> **Measured `2026-08-20 17:5xZ`, and the boundary is sharp.** The prune's own target class
> (`bdl`/`work_item_id`) stops at **18:02:24Z** on 08-18. Every row of all three wired "blockers"
> precedes it:
>
> | pair | rows | first row | last row | after 18:02:24Z | wired to |
> |---|---|---|---|---|---|
> | tool-generator / `function` (Tier A) | 11 | 08-16 15:48 | **08-18 16:33** | **0** | `input_data.spec.function` |
> | tool-generator / `description` (Tier C) | 11 | 08-16 15:48 | **08-18 16:33** | **0** | `input_data.spec.description` |
> | site-review-agent / `audit_source` (Tier A) | 2 | 08-17 12:20 | **08-17 12:37** | **0** | `audit_source_literal.audit_source` |
>
> **So `tg/function`'s trigger is a DATE, not config drift or a rebuild path.** It is also exactly
> what §2's own durable baseline predicts: `function` present **46/46** ⇒ Strategy 0 always resolves
> ⇒ post-prune the search is never asked ⇒ no row can be written. The two facts were already in this
> file and only needed joining. **`tg/function` is not a hard blocker; it is closed by step 1**, and
> the demand control agrees (tool-generator 16 runs/24h, class silent since 08-18 16:33).
> `audit_source` gets the same explanation but keeps a weaker warrant — 0 runs in 24 h, so the wire
> is unconfirmed rather than confirmed; treat it as *explained, not demonstrated*.
>
> **Consequence for the "binding constraint".** §2's call to add step/orchestration attribution to
> the `resolver_findings` bridge stands on its own merits, but it is **no longer needed for
> `tg/function`** — the join that answers that case is a timestamp against a roll, which the
> instrument already carries. Do not spend step 5's first commit on the bridge believing Tier A is
> waiting on it.
>
> **The same cross also fixes the wired/unwired split for the live classes**, which is what the
> remaining work actually turns on:
>
> | live class (rows since step-4 roll) | wired? | disposition |
> |---|---|---|
> | bdl / `commit_sha` — 348 | **no** | 315 lane owns the path (Tier B, unchanged) |
> | tg / `related_pages` — 9 | **yes, and it MISSES** | = `bugs_open/330`; refusal is the DESIRED outcome |
> | **pbh / `page_type` — 3** | **no** | ⚠ **NEW/REAWAKENED, see §2.1** |
> | tg / `reason` — 9 | no | ✅ **CLOSED: migration 512, APPROVED + APPLIED 17:38Z** |
>
> ### 2.1 pbh / `page_type` — live again, and it is NOT a record-a-decision item
>
> It read as quiet-since-08-18 12:07 at 14:30Z and had fired three more times by 15:11Z
> (15:03 / 15:07 / 15:11Z). Candidates: `load_page_record.page_type`, `page_record.page_type`, and
> 28 × `{ensure_site_record,site_record}.content_data.pages[N].page_type`. **The winner is
> `load_page_record.page_type` — the page's OWN record, i.e. almost certainly the RIGHT value** — so
> under the flip this LOSES a good value. The asker is `plan_sections` (`plan_sections_action.go:54`
> declares `page_type` Optional; pbh's `plan_sections` step wires nothing), and the explicit mapping
> is `"page_type": "load_page_record.page_type"` on that step. One key, one migration, **not built**
> (this session's one migration went to `reason`).
>
> ### 2.3 STATE AS OF 2026-08-21 ~11:3xZ (third session) — what is live, what is in flight, what is stuck

**The `?` OPTIONAL-EXPLICIT parser is LIVE.** `v1.0.1321` (both pods, up 2026-08-20T19:51Z) was
built from **`0483e7f4e`**, and `git merge-base --is-ancestor ecc419bd1 0483e7f4e` is **true**.
**So `?` adopter migrations no longer need `_HOLD`** — that precondition is discharged. (There were
never any queued: the enabler shipped ahead of every adopter, so 515 below is its first production use.)

⚠ **Proving that needed a method this estate did not have, and the standard one LIED.** The
capability probe returned ABSENT with both controls behaving correctly, because **both of
`ecc419bd1`'s quotable phrases are inside COMMENTS** (Go strips them) and it adds **no named
function** — only locals and closures. The `build provenance` line had scrolled at 14 h. What works:
**test candidate stamps one FIXED STRING at a time** (`grep -aqF "<sha>" /proc/1/exe`, with a
`deadbeef…` control), then ask `merge-base` for **ancestry**, not equality. A 60-way alternation
(`grep -aoE`) **times out past 2 minutes**. Full entry + the forward fix (add one probeable marker
on purpose when a change is destined for a `_HOLD`) is in `LANDMINES.md`.

**Migration 515 — `pbh`/`page_type` — BUILT, submitted, NOT YET APPLIED.**
`Council-Submitted: a452fc2a-160f-485c-949c-367c34c65df2` (commit `cc798cb34`). Adds
`"page_type?": "page_record.page_type"` to pbh's `plan_sections` step. §2.1 called this "one key,
one migration, not built"; it is now built, and §2.1's framing understates it — **[MEASURED] on 31
live pbh orchestrations the page's own record is PRESENT on only 13 and ABSENT on 18**, and on
those 18 the only candidates are the 28 sibling-page entries, so where siblings agree **no conflict
row is written and the substitution is SILENT**. The 40 logged rows understate the pair. Hence `?`
rather than a plain wire (a plain wire falls through to the search precisely on the 18). Absence is
safe, read at the consumer: `plan_sections_action.go:972-975` falls back to `pageName`. Dry-run
proven against the live DB in a transaction ending in `ROLLBACK`, and the idempotence guard proved
to fire by applying twice in one transaction. **Next step: read the verdict, then apply, then
verify with a demand control.**

🔴 **512 CANNOT BE VERIFIED BY WAITING — the queue is DRAINED, not quiet.** §2.2 left the test
unrun expecting "hours, not minutes". Run 17 h later: **tool-generator runs since the 17:38:34Z
boundary = 0**, conflict rows = 0. All **44** `add_tool` items are `complete`; the only survivors
are 2 `deferred` from 08-05; last tool-generator run was **17:00Z on 08-20**. So the zero means
nothing and no amount of patience changes that. Options: (a) wait for another lane to queue a tool
build and read it then — recommended; (b) dispatch one deliberately, which builds a real component
on a real site and is **not ours to fire unasked**; (c) record 512 as *applied and explained, not
demonstrated*. Do (a) with (c) written down meanwhile. **Do not let "applied, no rows since" harden
into "verified".**

**So the live-class ledger now reads:**

| class | state |
|---|---|
| `bdl` / `commit_sha` | 🔴 **the real gate.** ~~blocked on the 315 lane's answer~~ **ANSWERED** (2026-08-21, `34cff7080`): no single path works — 19 live `git_commit` steps use 9 distinct `output_field` names. Split agreed between the two sessions: the **306 session** converts each handler's `complete` step to `result_mapping` so `commit_sha` surfaces at a uniform `handler_result.response.commit_sha` (bigger than "16 configs": list-mode `output_fields` cannot rename a nested path to a top-level key at all, and report-builder alone has three completion paths); **then** one `"commit_sha?"` wire from the marker session. ⚠ **ordering is the OPPOSITE of 516's** — the wire must land AFTER the standardisation, or it resolves for the conforming handlers and silently drops the sha for the rest |
| `tg` / `related_pages` | 🟢 **516 BUILT + HELD** (`4c2169831`, Council-Submitted `101ed0c6`) — `related_pages?` on BOTH tool build steps. ~~needs a recorded decision, not a wire~~ **superseded: the wire IS the recorded decision.** Unlike `reason` there IS a path we want whenever the spec carries it, so `input_fields` would lose the legitimate case and a plain wire is what is deployed and broken; `?` is both halves. ⚠ **held on 512, not on a roll** — see the correction below |
| `pbh` / `page_type` | 🟡 **515 in review** — apply + verify. (Author is a THIRD session, not either correspondent above; `cc798cb34`.) Its `?` adopter entry is already written in `optional_explicit_wire_acks.json` so it does not trip the new gate — its author should check it says what they would have said |
| `tg` / `reason` | 🟢 512 applied; **verification STILL unrunnable** — re-measured 2026-08-21 ~11:4xZ: tool-generator runs since the 17:38:34Z boundary = **0**, eighteen hours. Option (a) is still the plan |

> **CORRECTED 2026-08-21 ~12:0xZ by the `?`-marker session**, at the 306 session's invitation
> ("I'll leave that correction to whoever next touches the ledger"). Two things had gone stale and
> one was actively misleading:
> - **`related_pages` needed a wire after all**, and the reason the old line was wrong is worth
>   keeping: it generalised from `reason`, where `input_fields` was right *because there is no path
>   we want*. `related_pages` has one whenever the spec carries it (1 of 5 recent runs did), so
>   "record a decision and leave it unwired" would have left the search armed for the present case
>   and lost the value for it. The `?` marker did not exist when the line was written.
> - **516 is HELD, and the condition is NOT a roll** — the parser is live on `v1.0.1321`. It is held
>   because applying it CONSUMES this table's own bottom row: 512's test is *"reason 0 **while
>   related_pages keeps firing**"*, so silencing `related_pages` first would make 512 permanently
>   unverifiable rather than merely unverified. Apply conditions are in the migration header; after
>   516 lands, the instrument-alive control must come from another agent's class (`bdl`/`commit_sha`).
> - **Adopters of `?` now pass a gate**: `config-key-audit --optional-explicit-wires` exits 1 on any
>   live `?` wire not acknowledged in `architecture_review/optional_explicit_wire_acks.json` with a
>   statement of what was checked DOWNSTREAM. It came out of the council's gating objection to the
>   marker (round 1 REVISE, `5f82423b`).

**Remaining after those: the Tier C decisions** (§2's list minus what §2.0 dissolved) — paragraphs,
not migrations. Then the flip, using §4's reasons for the tolerance retirement, not the retention one.

### 2.2 tg / `reason` — CLOSED
>
> `create_rerender_items` declares `reason` Optional, `enqueue_rerender` wired nothing, and the
> search handed it `load_brand_context.specs.classification.content_features.news_feed.reason` out of
> **42** candidates. Damage nil and the reason is what matters: `reason` acts only when it equals
> `section_data_resolved` / `image_landed` / `cta_links_stale`
> (`create_rerender_items_action.go:216-231`), so the substituted prose is inert **by luck of its
> value, not by design**. Fixed with `input_fields: ["site_id","domain"]` (483's shape — `?` is the
> wrong tool here: there is no path we want, we want absence). Council **APPROVED r1**, corr
> `2bd7fb37-cac2-409a-8452-50a7ed933467`; applied 17:38Z, `UPDATE 1`, live row read back.
> **⚠ VERIFICATION IS NOT RUN — DO NOT READ "APPLIED" AS "PROVEN".** Baseline banked at the apply
> boundary 17:38:34Z: reason **16** rows/24h (last 17:08:01Z), `related_pages` **12** (last 17:07:43Z),
> against **16** tg runs/24h (last run 17:06:09Z). Read at **17:44:15Z**: reason rows since the
> boundary = 0 — **and tool-generator runs since the boundary = 0.** So the zero is *no demand*, not
> *no defect*; at ~16 runs/24h (one per ~90 min) this needs hours, not minutes. The test, for
> whoever reads it next:
> ```sql
> SELECT count(*) FROM orchestration_states            -- the DEMAND CONTROL: must be > 0 first
>  WHERE owner_agent_type='tool-generator' AND created_at > '2026-08-20 17:38:34Z';
> SELECT context->>'field', count(*), max(occurred_at)  -- then: reason 0, related_pages STILL FIRING
>   FROM agent_error_log WHERE error_code='RESOLVER_CONFLICTING_CANDIDATES'
>    AND agent_type='tool-generator' AND occurred_at > '2026-08-20 17:38:34Z' GROUP BY 1;
> ```
> Pass = `reason` 0 **while `related_pages` keeps firing** (if both go quiet the instrument died and
> the zero means nothing) and no `component_id` class appears. State *n* runs observed and the floor
> it buys — *n* runs cannot detect a residual rarer than ~1 in *n* — never the word "fixed".
>
> Full working: NOTES `## 2026-08-20 (~17:4xZ)`. The `?` marker's two-surface trap and its shelf
> life (inert on `v1.0.1320`; `ecc419bd1` is 17:20Z) are in `LANDMINES.md`, and my own wrong call
> about it — reading an absence census as proof of a typo when it was an unbuilt feature — is in
> `WRONG_CALLS.md`.

### Tier A — HARD BLOCKERS: the field is `Required`, so the flip turns a guess into a FAILURE (2)

⚠ **BOTH ENTRIES IN THIS TIER ARE DISSOLVED BY §2.0 — read it before acting on this table.** Every
row of both classes pre-dates the step-1 prune roll (18:02:24Z on 08-18), where a wired field's
search ran and its answer was DISCARDED at the merge. `tg/function` is closed by step 1 with a
live demand control; `audit_source` is explained on the same grounds but has 0 runs in 24 h, so it
is *explained, not demonstrated*. The "unidentified trigger" note below is **answered: it is a
date.**

| pair | action | spec | note |
|---|---|---|---|
| **tool-generator / `function`** | `create_tool_component` | **Required** (`create_tool_component_action.go:44`) | Confirmed by elimination — `create_tool_component` is the only action in tool-generator's 10 steps that takes `function`. Wired by migration 211 (`config.function = input_data.spec.function`); the conflict fires only when that path is empty, and then the search rescues it. **Post-flip the action fails hard.** Arguably correct (better than building a tool under another tool's function name) but it WILL surface as failures — decide deliberately, do not discover it |
| **site-review-agent / `audit_source`** | `write_audit_findings` | **Required** (`write_audit_findings_action.go:43`) | Agent has **0 runs in 24 h** and no trace in `orchestration_states` at all, so low risk today — but see the §6 caveat: that is not the same as retired |

**MEASURED 2026-08-20 evening, on a DURABLE 15-day baseline (46 `add_tool` items, 08-05→08-20 — the
work-item specs persist, unlike `orchestration_states`, which keeps only ~24 h and gave a
misleading 16-run sample):**

| wired field | present | missing | reading |
|---|---|---|---|
| `function` | **46/46** | 0 | the known producer NEVER omits it |
| `description` | **46/46** | 0 | never omits it |
| `related_pages` | 18/46 | **28/46 (61%)** | **330's population, now on a durable baseline** |

⚠ **This does NOT clear `tg/function` — it deepens the question, so do not wave it through.** If
`input_data.spec.function` is always populated, Strategy 0 resolves it and the whole-tree search
should never run for that field — yet there are 11 conflict rows with winner
`~unwrap.current_page.function`. And no OTHER action in tool-generator's 10 steps declares
`function` (checked `create_rerender_items`, `read_site_spec`, `write_doc_plan`, `rag_index`,
`ensure_site_record`, `execute_llm_prompt`, `complete_workflow`). So **the trigger is unidentified**
— candidates: config drift at the time (the wire is newer than the rows), a rebuild path with a
different `input_data.spec`, or a step whose config lacked the wire on 08-16→08-18. **Identify it
before deciding; a Required field is exactly where a wrong assumption becomes a hard failure.**

🔴 **THE BINDING CONSTRAINT ON FINISHING STEP 5 IS NOW THE INSTRUMENT ITSELF.** The
`resolver_findings` bridge writes `agent_type` but leaves **`step_name`, `action` and
`orchestration_id` empty** (`action` is the literal `"input-resolver"`), so a conflict row cannot be
joined to the run or the step that caused it — which is exactly the join `tg/function` needs, and
the join every remaining Tier A/C decision would need if it turns out to be non-obvious. This is
the observability point the prune's council round recorded for "step 5's design"; it has now become
load-bearing rather than nice-to-have. **Consider adding step/orchestration attribution to the
bridge as step 5's FIRST commit** — it is small, it is inside our own footprint, and every
remaining decision gets cheaper with it.

### Tier B — SILENT-LOSS BLOCKER: `Optional`, but something downstream depends on the value (1)

**bdl / `commit_sha`** — `Optional` in `CompleteWorkItemInputSpec`, written to `result.commit_sha`
(`load_work_item_actions.go:937`). Absence is "handled" in the sense that Optional means no error —
**and that is exactly the danger**: the field silently stops being recorded, and `bugs_open/315`'s
page stamping depends on it. **This is the one pair that genuinely needs a config wire before the
flip.** 442 conflict rows and growing (last 15:23Z). `bugs_open/334` candidate 1. A CONTRIB asks
the 315 lane for the path that is correct *by their lights* — **do not pick one from the shape.**

### Tier C — RECORD A DECISION, no migration needed (~10)

`tool-generator`/`reason`,`related_pages`,`description` · `component-creator`/`description`,`site_type` ·
`page-build-handler`/`page_type`,`sections` · `generic`/`summary`,`page_id` · `page-rerender`/`current_page` ·
`rerender-pages`/`reason`

All `Optional`, and for the spec-array shape **absence IS the correct answer** — the audit's own
words for 330: *"absence is the fix"*. So the precondition's "give every pair an explicit mapping"
is satisfied here by a **recorded judgement that nothing is the right value**, not by config. That
is a paragraph per pair, not a migration per pair. **This is why the remaining work is much smaller
than the flat count of 13 suggests.**

⚠ The one in Tier C to actually look at is **`page-rerender`/`current_page`** (78 rows, **627
runs/24 h** and quiet): step 4 deliberately did not rename the stored template key, so this route
survives by design. Confirm absence is right there rather than assuming it.

## 2.4 THE CENSUS IS NOW FULLY DISPOSITIONED — 19 pairs → 4 live, 4 quiet-unwired, 11 closed (2026-08-21 ~11:4xZ)

§2.0's two axes (WIRED? × WHICH SIDE OF THE PRUNE?) generalise to the whole census, and applied
mechanically they close it out. Two queries did it.

**Axis 1 — how many rows does each pair have AFTER the prune boundary** (`2026-08-18 18:02:24Z`, the
last row of `bdl`/`work_item_id`, the class step 1 was built to kill)? **Only EIGHT of the 19 pairs
have any at all.** The other **eleven are pre-prune only**, i.e. entirely inside the window where a
resolved field's search still ran and its answer was then discarded at the merge.

**Axis 2 — of those eleven, which are WIRED?** Seven are, and the wires are real dotted paths, not
prose (checked the VALUES, because a config key called `description` could easily have been a
step's own description — it is not):

| pair | step | wired to |
|---|---|---|
| component-creator / `description` | `store_component` | `input_data.spec.description` |
| component-creator / `site_type` | `store_component` | `input_data.spec.site_type` |
| page-build-handler / `sections` | `plan_sections` | `spec_sections.sections` |
| rerender-pages / `reason` | `create_rerender_items` | `input_data.spec.reason` |
| site-review-agent / `audit_source` | `write_strategic_findings` | `audit_source_literal.audit_source` |
| tool-generator / `description` | `save_tool` | `input_data.spec.description` |
| tool-generator / `function` | `save_tool` | `input_data.spec.function` |

**WIRED + zero post-prune rows ⇒ CLOSED BY STEP 1.** The prune removed Strategy-0-resolved fields
from what Strategy 1/2 request, so these cannot write another row; and their historic rows never
affected a value, because the merge discarded the search's answer. **Seven pairs need no decision,
no migration and no paragraph.** Together with the four already killed by steps 1–4
(`bdl`/`work_item_id`, `bdl`/`current_page`, `pcw`/`current_page`, `pbh`/`current_page`) that is
**eleven of nineteen closed.**

### The eight that remain, and each one's actual state

| pair | rows post-prune | wired? | state |
|---|---|---|---|
| **bdl / `commit_sha`** | **640, last 2026-08-21 10:28** | no | 🔴 **THE GATE. Firing right now.** Blocked on the 315 lane's answer for the correct path — `bugs_open/334` |
| **tg / `related_pages`** | 17, last 08-20 17:07 | **yes, and it MISSES** | 🟠 `bugs_open/330`. Refusal is the DESIRED outcome ⇒ **recorded decision**, not a wire |
| **pbh / `page_type`** | 3, last 08-20 15:11 | no | 🟡 **515 in review** (`a452fc2a`) — apply, then verify with a demand control |
| **tg / `reason`** | 22, last 08-20 17:08 | no → wired by 512 | 🟢 512 applied; **verification unrunnable, queue drained (§2.3)** |
| bdl / `result` | 0 (last 08-17 16:29) | **no** | ⚪ quiet, UNWIRED ⇒ not closed. `452_..._goes_strict_HOLD.sql` is its intended `!` fix and is **NOT applied** (checked `result`/`result?`/`result!` — none present). Why the rows stopped is unattributed |
| page-rerender / `current_page` | 0 (last 08-18 13:07) | **no** | ⚪ quiet, UNWIRED, and the agent runs **~600×/24 h** ⇒ condition-dependent, NOT fixed. Step 4's rename plausibly removed a candidate, but the rows stopped BEFORE step 4 rolled, so do not credit it |
| generic / `page_id` | 0 (last 08-17 12:40) | **no** | ⚪ quiet, UNWIRED. 2 rows |
| generic / `summary` | 0 (last 08-17 18:29) | **no** | ⚪ quiet, UNWIRED. 3 rows; resolves into the agent's OWN workflow config |

**So the remaining work is exactly: one external answer (`commit_sha`), one migration to land (515),
one decision to record (`related_pages`), one verification with no path (512), and FOUR
quiet-unwired pairs needing a recorded decision each.** Not thirteen unknowns — eight named items,
half of them already in motion.

### 2.5 One of the four ⚪ pairs is now CLOSED ON MECHANISM: page-rerender / `current_page`

The most worrying of the four (78 rows, and the agent runs **~600×/24 h**) is disposed, and not by
its quietness. **No production `ActionInputSpec` declares `current_page` at all** — a fleet grep of
`Required:`/`Optional:` lists finds exactly one hit and it is a *test fixture*
(`action_inputs_optional_explicit_test.go:249`). So `current_page` never arrives as a requested
action input; it arrives only through **`ensureCoreFields`**, which is precisely what step 3 gated
(`unified_extractor.go:827`, `requested := func(name string) bool { return contains(fieldNames, name) }`).

And no page-rerender step requests it. Only one of its ten steps declares `input_fields` at all:

| step | action | input_fields |
|---|---|---|
| `render_page` | `rerender_single_page` | `["page_id","site_id","domain"]` |
| `rerender_sections` | `rerender_page_sections` | none (spec Optional: `page_name`, `page_id`, `reason`) |
| `save_sections` | `save_page_sections` | none |

`current_page` appears in none of them ⇒ the gate skips the injection ⇒ the whole-tree search is
never asked ⇒ **the class cannot fire.** That is a mechanism, so it holds regardless of traffic —
which is what makes it a real disposition rather than the "quiet ≠ fixed" mistake this lane keeps
warning about. It also explains the pre-step-4 silence honestly: **step 3's gate closed it, not step
4's rename**, and the rows stopping on 08-18 (before either rolled) is a separate, still
unattributed fact — do not claim step 3 as the *cause of the silence*, only as the *reason it cannot
recur*.

**Three ⚪ pairs remain**: `bdl`/`result`, `generic`/`page_id`, `generic`/`summary` — 5 rows between
them, all pre-prune, all unwired. Worth doing the same read-at-the-consumer for each, but they are
the smallest items on the board.

⚠ **The remaining ⚪ rows are the ones to be careful with.** Unwired means the prune does not protect
them: the search still runs whenever the shape appears, so they are *armed*, not fixed — and
page-rerender at 600 runs/day is the clearest case. Their disposition is a judgement (is absence
correct for this consumer?) which must be read AT the consumer, as 515 did, not inferred from the
row count.

## 3. Recommended order of work

1. **Tier C first** — it is ~10 paragraphs of recorded judgement and it shrinks the list fastest.
   Write them into `bugs_open/330` §9 or a new step-5 design doc, one per pair, each naming the
   action and why absence is correct.
2. **Tier A** — two decisions. For `tool-generator/function`, measure how often
   `input_data.spec.function` is actually empty (if never, the flip is free; if sometimes, it is a
   deliberate new failure mode and needs the owner or the council). For `audit_source`, the agent
   is dormant — a safe-by-inspection note is probably enough, with §6's caveat stated.
3. **Tier B last, and it is the true gate** — blocked on the 315 lane's answer. Chase it, or take
   their instruction. This is the only item that can silently lose data.
4. **Then** flip, council-gated, and retire the read-side tolerance in the same commit — using the
   §4 reasons below, **not** the retention argument.

## 4. Retiring the read-side tolerance: the plan's REASON was wrong, the conclusion holds

Do **not** repeat "the step-4 roll has outlived `orchestration_states`' ~24 h retention". Rows from
**2026-07-19** are still in the table, and the tolerance's second call site is
`mergeIntoRenderContext` — the RE-RENDER restore — where stored component `content_data` **never
expires** (20 live `page_components` rows across 12 sites hold `current_page` as a string; 17 on
`deployed` pages). Cite these instead, both one query:
1. **Zero NON-TERMINAL pre-roll orchestrations** — all 2,476 pre-roll rows are
   COMPLETED/CANCELLED/FAILED, so none can be resumed into the build-side call site.
2. **`buildRerenderBaseData` writes the NEW key fresh** from its `pageName` argument, and the
   tolerance's first branch `continue`s whenever `current_page_name` is present — so those 20 stored
   rows never reach the second branch.
```sql
SELECT count(*) FROM orchestration_states WHERE created_at < '2026-08-19 22:26:25Z'
  AND status NOT IN ('COMPLETED','FAILED','CANCELLED');            -- must be 0
SELECT jsonb_typeof(content_data->'current_page'), count(*) FROM page_components
 WHERE content_data ? 'current_page' GROUP BY 1;                    -- know this number
```

## 5. The instrument's permanent blind spot (unchanged, and it bounds what the flip can promise)

A conflict row requires the candidates to **differ** (`reflect.DeepEqual`). A tree with ONE match —
or several that agree — substitutes silently: no WARN, no row, and the value can still be wrong.
**So "zero conflict WARNs" can never establish the search is safe**, only that the conflicting
subset is empty. The audit above put a *measured floor* under this (4 rescue-prone wires on the
sampled slice, 1 carrying a wrong value) but 269 pairs / 75 agents are unsampled. Step 5's design
must state this rather than inherit the precondition's "or" branch as if it were sufficient.

## 6. ⚠ Correction to this lane's own demand control (new, 2026-08-20 evening)

The demand-control join I introduced (`orchestration_states.owner_agent_type`, RUNBOOK) is sound
for **"is this agent running NOW?"** and nothing more. It cannot size a class historically:
`site-review-agent` has **58 rows in `agent_error_log`** and **zero trace in `orchestration_states`
by any column** (`owner_agent_type`, `workflow_plan`, `execution_metadata`) — its runs aged out.
`agent_error_log` retains from 07-20; `orchestration_states` keeps ~24 h of COMPLETED rows. So:
- `runs_24h = 0` **does** license "cannot fire right now".
- `runs_24h = 0` does **NOT** license "retired", "never ran", or any statement about the period when
  the class was firing. For that, use the error log's own history.
My earlier table marked two rows "agent idle" — the conclusion held, the stated reason was too
strong. The LANDMINES entry has been amended.

## 7. Traps carried forward

- **Two "continue here" handoffs coexisted for ~9 h** because I wrote a new dated file at 07:59
  while a parallel session kept updating the old one. **Before creating a dated successor, grep the
  lane for other `*_continue_here.md` and check `git log` on them for edits newer than your read.**
  Consolidated here; 08-18b bannered.
- **097's `plan` is an OBJECT** (`summary`/`edits`/`grounded_in`/`risks`), not an array. A schema
  refusal is CLIENT-side — no round spent. A *published* run must never be re-triggered.
- **Council scope widened 2026-08-19** to appliable migrations under `docs/agent_docs/sql_for_agents/`
  (`bugs_open/314`), `_HOLD.sql` included — so a Tier A/B migration is now council-gateable.
  Scope is single-sourced in `scripts/council-scope.sh`; `DRY_RUN=1` tests admission free.
- **A mutation that breaks the BUILD proves nothing.** Mutate to a no-op (`if true { return fields }`).
- **The instrument stores candidate PATHS, never VALUES.** Judge a class at
  `orchestration_states.collected_data` (RUNBOOK four-step method).
- **Config wiring is at `config.<field>`, NOT `config.params.<field>`** — probing the wrong key
  returns NULL and reads as "not configured" (cost a false root cause, `WRONG_CALLS.md` 08-19).
  If config has nothing, the asker is an **action input spec**.
- **A top-level `jsonb_each(steps)` census misses sub-workflow/loop substeps** — bit this lane
  three times now, incl. 334's `mark_complete`. Use the whole-config text search as the ceiling.
- **`grep -aq` exits 1 on no match**, so `&& echo` prints nothing and the shell says
  `command terminated with exit code 1` — that IS the absent-control passing.

## 8. Session-start checklist

1. `git log --oneline -10`; re-read this file from disk. **~270 commits landed here in one day** —
   assume anything you remember is stale.
2. Nothing owed on reviews; nothing owed on rolls.
3. Re-run the census (RUNBOOK) **with the demand control read per §6's limits**, before trusting
   any row of §2.
4. Start at §3 step 1 (Tier C). Tier B is the real gate and is blocked externally.
5. **Do not flip anything** until every §2 row is killed, wired, or has a recorded decision.
