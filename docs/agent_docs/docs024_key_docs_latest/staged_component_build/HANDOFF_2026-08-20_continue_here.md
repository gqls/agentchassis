# HANDOFF — 2026-08-20 (rev. 2026-08-21 ~17:1xZ): **STEP 5 IS BUILT. THE FLIP IS COMMITTED** (`5fe010ada`, council `26186633`) — steps 1–4 live+proven, the census fully dispositioned, and the resolver now REFUSES a conflict. **Inert until a chassis roll carries it.** Remaining: read the verdict, verify post-roll, retire the read tolerance.

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
| `bdl` / `commit_sha` | 🟢 **HANDLER SIDE COMPLETE, 2026-08-21 ~15:3xZ.** All **10 real handlers** standardised, applied, verified live: 519 page-rerender (dominant, 4,810/7d — verified end-to-end on real traffic), 521 rerender-pages, 522 css-patch-agent (REVISE r1 → resubmitted → APPROVED), 523 section-editor, 527 webdesign-agent, 528 nav-updater, 534 page-build-handler (dependent on 519), 535 asset-deployer, 536 image-build-handler (dependent on 535), **540 content-feed-orchestrator** (found by the WIRE's own apply-time guard, not either census — see below). Found across **three distinct discovery methods, each blind to what the others caught**: a structural census (action = `git_commit`) found 7; an empirical presence check (`site_work_items.result ? commit_sha`) found 2 more but also flagged a **false positive** — `tool-generator`'s one hit traced to a sibling loop iteration's `section-editor` commit, cross-contamination from the very defect this bug is about, not a real gap, not built; the wire's own property-of-the-handler guard (does THIS handler's tree, over 30 days, ever carry a `commit_sha` — not "did an item record one") found the 10th, `content-feed-orchestrator`, which one census had seen-and-deferred (two independently-conditional commits, `commit_news`/`commit_rss` — mapped to news, present 17/17 vs rss 2/17, loss disclosed not hidden) and the other missed cold. **The population is dynamic, not a fixed set — do not assume 10 is final if picked up again; re-derive.** **Remaining: the wire itself (migration 537, `staged-component-build`'s), which should now pass its guard.** Full path-per-agent table + every lesson: NOTES `## 2026-08-21`, entries from `~13:5xZ` through `~15:3xZ`. |
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

## 2.6 515's council trail (live as of 2026-08-21 ~12:1xZ) — REVISE answered, round 2 in flight

`SUBMISSION_CORR=a452fc2a-160f-485c-949c-367c34c65df2`. **Round 1 = REVISE**, `decided_by: gating
objection from guardian`; 7 of 9 seats approved. Both objections were real; the file was revised
rather than defended (`b9d62857e`).

- **guardian (HIGH, gating) — right about the reasoning.** *"that provenance check proves the
  ancestor commit is present, not that this specific code path … was the thing `ecc419bd1` added."*
  My chain read the config peel **at HEAD** and proved ancestry **at the stamp**, and those do not
  join — HEAD is ahead of the build, so the peel could have arrived later and the binary would carry
  the commit without the behaviour. Closed at the built commit:
  `git show 0483e7f4e:…/action_inputs.go | grep -n 'TrimSuffix(k, "?")'` → 694, 708, both in the
  CONFIG loop; `git log -L` → `ecc419bd1` introduced it. **Carry this: "the commit is aboard" and
  "the behaviour is aboard" are different claims, and `merge-base` only establishes the first.**
  Logged in `WRONG_CALLS.md`; my own morning LANDMINE (which recommended the ancestry fallback) is
  corrected to say so.
- **The seat cited a LANDMINE against me, and it was STALE.** The entry titled *"…`?` … NOT YET in a
  step's action `config`"* already carried two corrections beneath it, the second pinned to
  `v1.0.1320`, and the fleet had since rolled to `v1.0.1321` which parses it. So a high-severity
  objection landed against a correct migration on the strength of a heading its own body retracted.
  Corrected (`32ca8ebf0`). **Generalisable: a reader matches the heading and footprint, not the
  third paragraph — put a shelf life in the TITLE or retire the entry.**
- **guardian (low) — found a third consumer I had missed.** THREE agent types carry an unwired
  `plan_sections`/`page_type`: pbh, `page-content-writer`, **`page-rebuild`** — and page-rebuild's
  step is NESTED, invisible to a top-level `jsonb_each` census (use
  `jsonb_path_query_array(default_config,'$.**.steps.plan_sections.action')`). **Scope still stands,
  for a better reason than isolation: only pbh has a `load_page_record` step**, so `page_record` is
  absent from the other two trees entirely (pcw carries it on 0/15 runs). Wiring it there would name
  a path that never exists. Recorded as named follow-ups; neither has produced a conflict row.
- **editquality (medium)** — added a single-active-row assertion (pbh is not one of the four
  duplicate-row types; measured 1 row, version 1, but the assertion is free).
- **guidelines (nit)** — `snapshot_agent()` moved INSIDE the guard, **proved**: a double run now
  emits ONE snapshot notice where it emitted two, guard still raising.
- **debug_historian (gap)** — the file now says its evidence was a REHEARSAL and the real apply owes
  its own logged verify plus a demand control.

⚠ **Round 2's FIRST attempt died at `complete_invalid` (10:40:27Z) — the Anthropic account hit its
API usage limit, not a fault in the submission.** The 400 says *"you will regain access on
2026-09-01"*; **that is the billing reset, not when access returns** — the burst ran 10:34→10:41:29Z
and the estate completed 200+ orchestrations in the next 87 minutes. A `complete_invalid` run
produced **no verdict**, so resubmitting on the same trail is correct and precedented (step 3 hit
this on 08-19 and was approved on resubmission). Re-run in flight, `RUN_ORCH_ID=cd19b246-…`, past
the seat that died last time. Full trap in `LANDMINES.md`, including that
*"a missing row is latency"* stops applying once `current_step` reads `complete_invalid`.

**ROUND 2 = APPROVED** (2026-08-21 12:22:34Z, `decided_by: all reviewers approve`; 5 advisories, all
LOW). **APPLIED 13:19:19Z** by hand with `psql` — deliberately NOT via `run-migrations.sh --apply`,
which takes every pending file while other sessions have WIP there — then recorded with
`--record-only`. Live row read back independently: `"page_type?": "page_record.page_type"`, snapshot
present in `agent_definitions_backup`. Corroboration found afterwards: the step **already** wired
`page_name: "page_record.name"`, so `page_record` was its own established prefix.

### ✅ 515 IS DONE — approved, applied, and BEHAVIOURALLY PROVEN at runtime (2026-08-21 14:24Z)

**The evidence, from a real post-apply `plan_sections` extraction:**
```
PASS 515 | 2026-08-21T14:24:28.457Z | requested_fields: ['section_facts', 'pipeline', 'site_type']
```
`page_type` is **absent** from what `ExtractFields` is asked for ⇒ **the `?` marker parsed.** The
control is inside the same line: **`site_type` is still present** — same `Optional` list, same
action, also unwired — so the exclusion is specific to the marked field, not a general
disappearance or a truncated list. **This is the first production use of `?` on the step-config
surface, so this line is also the fleet-level proof the marker works — it unblocks every other
adopter, including the held `516` for `tg/related_pages`.**

Negative half, with its weakness stated: 0 conflict rows against **8** pbh runs since the apply; at
~0.12 rows/run that predicts about **one**, so the zero is unremarkable alone. **Cite the positive
test, not the zero.**

⚠ **The instrument took three attempts and the first two failed SILENTLY** — kept below because the
next runtime check on this estate will face the same thing.

<details><summary>(historical) how this was armed, and the two ways it failed</summary>



**Two seats independently raised this and the approval did not close it** (guardian,
prior_art_librarian): the `?`-parses-on-step-config claim rests on a **source grep at the built
commit**, and *no orchestration has ever exercised that parse path*. Round 1 taught that ancestry ≠
behaviour; this is the next rung — **source-at-the-stamp ≠ observed.**

**The zero-test is weak here, so do not lean on it.** Baseline at the apply boundary: 3 rows against
**26 pbh runs** in 24 h ⇒ **~0.12 rows/run** (step 4 was 3.1/run). 26 runs cannot detect a residual
rarer than ~1 in 26.

**Use the POSITIVE test instead — it is decisive in a single log line.** `withoutStrict`
(`action_inputs.go:802-813`) removes `explicitOnly` fields from what is handed to `ExtractFields`,
and `ExtractFields` logs `requested_fields` on its `=== MASTER EXTRACTOR START ===` line. So on any
post-apply `plan_sections` extraction, read that line. **Filter on `step_name`, not on `section_facts`** — that was
my first draft and it selects the step by a field that happens to be unique to it rather than by the
step itself. The line is JSON; its full key set is `action, agent_id, agent_type, available_keys,
caller, level, msg, orchestration_id, pod_name, requested_fields, stateless, step_id, step_name, ts`.
- **`page_type` ABSENT from `requested_fields` ⇒ the marker PARSED.** ✅
- **`page_type` PRESENT ⇒ it did NOT parse**, and the config key is an inert literal. ❌
```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis \
               -o jsonpath='{range .items[*]}{.metadata.name} {end}'); do
  kubectl -n ai-persona-system logs "$POD" --since=60s \
    | grep -F 'MASTER EXTRACTOR START' \
    | python3 -c 'import json,sys
for ln in sys.stdin:
    try: d = json.loads(ln)
    except Exception: continue
    if d.get("step_name") != "plan_sections": continue
    rf = d.get("requested_fields") or []
    print("FAIL: marker did NOT parse" if "page_type" in rf else "PASS: marker PARSED", d.get("ts"), rf)'
done
```
Dedupe on `orchestration_id`+`step_id` if you poll with overlapping `--since` windows.
⚠ **AND WATCH THE RIGHT PODS — this cost an hour on 2026-08-21.** `page-build-handler` does **NOT**
run in the `agent-chassis` deployment. Every run gets its **own ephemeral pod**,
`agent-page-build-handler-<hash>-<suffix>`, with a **fresh hash each time**. Two watchers pinned to
`-l app=agent-chassis` returned nothing while **six** pbh orchestrations ran and **three** reached
`plan_sections`. Confirm the producer from the DB before arming anything —
`SELECT processing_node FROM orchestration_states WHERE owner_agent_type='page-build-handler'
ORDER BY created_at DESC LIMIT 3` — and poll the **pod list**, not a fixed selector, because the pod
does not exist when you arm the watcher:
```bash
kubectl -n ai-persona-system get pods --no-headers -o custom-columns='N:.metadata.name' \
  | grep '^agent-page-build-handler-'
```
Retention in those pods is brutal: the one survivor held **180 lines / 3.4 minutes** and had already
rotated past `plan_sections` when read 5 minutes after the run. **Poll every ~20 s.** Full trap in
`LANDMINES.md`; my wrong call in `WRONG_CALLS.md`.
⚠ **There is no durable fallback for this one** — checked: `section_plan` records counts and names
but **not** the `page_type` it was handed, so `collected_data` cannot answer it. The log is the only
route, which is exactly why the watcher has to be right.

⚠ **Poll it; do not tail once.** These lines churn out of a chassis pod in minutes, and
**page-build-handler is BURSTY** — 26 runs/24 h on average, but its last run was **11:57Z** and
there were **zero in the first 6 minutes after the 13:19:19Z apply** (the window was only 6 minutes
long at the time of writing; an earlier draft of this line said "75 minutes", which was my
arithmetic error — I had subtracted from the last RUN rather than from the APPLY. The honest
statement is that the post-apply window was far too short to conclude anything, not that pbh had
gone quiet for over an hour after it). A watcher was armed at 13:2xZ for ~30 minutes; if it expired
without a hit, **re-arm rather than concluding anything**. **Zero pbh runs ⇒ the zero conflict rows
carry no information** (demand control = 0 at the time of writing).

**Whoever next sees a pbh run: record the result of that one line.** It is the last open evidence
on this migration.

</details>

## 2.7 bdl/`commit_sha` — UNBLOCKED, and use `?` NOT `!` (2026-08-21 ~15:0xZ)

**The 315 lane answered our CONTRIB in full and the gate is open.** They laid out three routes,
chose **(b) — make the path stable at the SOURCE** — and then *built* it: migrations 519/521/522/523/
527/528/534/535/536 convert each handler's `complete` step from `output_fields` list mode to
`result_mapping`, hoisting `commit_sha` to a canonical top-level key. Their handoff:
*"handler side COMPLETE — all 9 real handlers standardised, applied, verified live. The wire is
staged-component-build's remaining piece."* **The migration is explicitly ours; they said they are
not taking it.** Their full answer is at the bottom of
`docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/CONTRIB_2026-08-20_from_staged_component_build_commit_sha_resolves_by_guess.md`.

⚠ **BUT: their §3 suggests `commit_sha!` (strict) and their own §4 refutes it. Use `?`.**
`!` means *explicit path or FAIL* (`action_inputs.go:750` — "strict enforcement at the bottom fails
`!` fields"). Their §4 warns that **absence is CORRECT** for items whose handler contains no
`git_commit` at all. I measured it independently and it is bigger than their figure:

```sql
SELECT count(*) AS completed,
       count(*) FILTER (WHERE result ? 'commit_sha')       AS has,
       count(*) FILTER (WHERE NOT (result ? 'commit_sha')) AS lacks
FROM site_work_items WHERE status='complete' AND jsonb_typeof(result)='object'
  AND completed_at >= now() - interval '3 days';
```
→ **2195 completed, 1115 with, 1080 WITHOUT — 49%.** So `commit_sha!` would **hard-fail about half
of all work-item completions.** `?` gives exactly what is wanted: resolve from the canonical path,
otherwise absent, no error. Same shape as 515, and the marker is now proven in production by 515.

**The wire to build:** `"commit_sha?": "handler_result.response.commit_sha"` on `build-dispatch-loop`'s
`complete_work_item` step. **Verify the path at runtime first** — the 315 lane's §4 also warns the
envelope is doubly nested (`response.<field>.response.data.`) on `call_agent`-reached handlers, and
their `result_mapping` work is meant to have flattened that. **Read a live post-535 bdl tree before
writing the path**; do not take `handler_result.response.commit_sha` from this paragraph.

## 2.10 ✅ THE FLIP IS BUILT AND COMMITTED — `5fe010ada` (2026-08-21 ~17:1xZ)

`findFieldRecursive` now returns **nil** on a conflict. Council submitted,
`SUBMISSION_CORR=26186633-a9a2-4425-bbb4-a9e58c418c66`. **Go code ⇒ INERT until a chassis build
carrying it rolls.** `v1.0.1322` (up 16:54Z) predates the commit.

**The precondition was met on the SECOND branch, deliberately — say so, do not claim the zero.**
The comment allowed "zero conflict WARNs **or** every observed field/caller pair given an explicit
mapping first". The zero branch **can never be sufficient** (a row needs the candidates to DIFFER;
a lone wrong candidate substitutes silently — `bugs_open/330` §4, `bugs_open/350`). All 19 pairs are
dispositioned: §2.4–§2.9.

**What changed, precisely:**
- conflict ⇒ `return nil`. Unique-value resolution, the collector, the depth/rank sort, the
  infrastructure-key skip list and the no-candidates path are **untouched**.
- `phase` now reads **`2-refuse`** in both the WARN and the persisted finding. Message and all other
  fields unchanged, so the window's existing queries keep working. **This is how a post-roll reader
  tells which build produced a row** — necessary because a run in flight keeps the behaviour it
  started with (measured at **8.5 minutes** during 537's verification).
- `winner_path` is still reported. Nothing resolves from it; it names the candidate the ranking
  **would** have picked, which is the first thing anyone tracing an absent field needs. **The
  ranking is therefore not dead code** — `bugs_closed/306` is why it is declared.

**The tests, and a warning for the next flip of this kind.** The test-file header named **one** file
as the flip site. The real blast radius was **four files, 13 tests** — `unified_extractor_search_test.go`,
`unified_extractor_tiebreak_test.go` (306's six), `resolver_findings_test.go` (three recorder tests),
`action_inputs_prune_test.go` (step 1's overreach guard), plus step 4's own control in
`render_context_step_boundary_resolver_test.go`. Each now asserts **nil AND the reported
`winner_path`**, so none passes vacuously — and each asserts `phase == "2-refuse"`, so a build that
silently reverted to Phase 1 **fails** rather than warning quietly.
> **CORRECTED 2026-08-21 ~19:3xZ (council round 2, editquality — I had the count wrong):** it is
> **13 tests across FIVE files, not four.** The fifth is
> `platform/orchestration/actions/render_context_step_boundary_resolver_test.go` — step 4's own
> control, in a *different package*. The four in `datahelpers` are search, tiebreak,
> resolver_findings and action_inputs_prune. The commit message and the first NOTES entry both say
> "four"; forward-only, so they stand with this correction rather than being rewritten. The
> under-count came from mentally filing the actions test as "step 4's, not the flip's" — it is both.

**MUTATION-PROVED**: reverting only the `return` (file still compiles) fails all 13; restoring it
returns both packages to green. `./platform/...` fully green — the `internal/adapters/thunder` build
failure is **pre-existing at HEAD** and another lane's.

### What is still owed on step 5

1. **Read the verdict** (`26186633`). **R1 REVISE, R2 REVISE, R3 submitted ~19:2xZ** — both gatings
   from `bug_historian`. R2's was that *"this council's own case index lists 085 under OPEN"*.
   **⚠ THE COUNCIL'S CASE INDEX IS STALE — expect this again on any recently-moved bug.** 085 is
   closed: `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 085_` returns one line,
   in `bugs_closed/`, moved by `916c8b22b` — **which is this lane's own step-4 round-1 answer to the
   same seat on 08-19.** The seat is reading a pre-08-19 snapshot. **Pre-empt it with the
   `git ls-tree HEAD` proof in the rationale; asserting "it's closed" will not clear it.**

1b. **A POST-ROLL MONITORING GATE IS NOW AN OBLIGATION OF THIS LANE, not a risk** (council round 2,
   guardian, accepted). Step 5 is not closed until it has run, and its terms are fixed in advance so
   it cannot be graded loosely: **(a)** demand control first — orchestrations created after the roll
   must be > 0, per agent, before any zero is read; **(b)** attribute rows by joining
   `orchestration_states.created_at` against the roll boundary, **never by wall clock** (a pre-change
   run was still emitting old behaviour **8.5 minutes** after 537 went live); **(c)** the
   discriminator is `context->>'phase' = '2-refuse'`; **(d)** any NEW `(agent,field)` pair appearing
   with that phase — one not among the 19 — is a finding to **trace to its consumer** and give an
   explicit mapping, never a reason to revert; **(e)** ≥48 h, and the result recorded here whether
   clean or not.
2. **After a roll that carries `5fe010ada`** — verify, and note the flip's evidence is *unit-level
   plus the three per-pair runtime proofs*; the flip itself has none yet by construction. Read the
   window with a **demand control**, and attribute rows to orchestrations by `created_at`, **never by
   wall clock**. `phase='2-refuse'` is the discriminator:
   ```sql
   SELECT context->>'phase' AS phase, agent_type, context->>'field' AS field, count(*)
     FROM agent_error_log WHERE error_code='RESOLVER_CONFLICTING_CANDIDATES'
      AND occurred_at > '<roll>' GROUP BY 1,2,3 ORDER BY 4 DESC;
   ```
   A `2-refuse` row is **not a regression** — it is the instrument working. What matters is whether
   any *consumer* broke, and the remedy for one that did is an explicit mapping
   (`<field>?: <path>`), **never a return to picking**.
3. **Retire the read-side tolerance** in `setRenderContextScalarsFromData` — **deliberately NOT in
   the flip commit**, because it is a separable cleanup and bundling puts two changes under one
   review. Use §4's two reasons (zero non-terminal pre-roll orchestrations; `buildRerenderBaseData`
   writes the new key fresh so the tolerance's second branch is unreachable) — **NOT** the retention
   argument the plan originally gave, which was unsound.

## 2.9 `bdl`/`commit_sha` IS WIRED AND LIVE — via **537**, not 539. Verification is a NAMED PREDICTION, not yet met (2026-08-21 15:3xZ)

**A parallel session built the identical wire as `537` and applied it at 15:33:39Z**, about an hour
before my `539` returned APPROVED at council round 2. Neither of us saw the other's WIP — **the
second identical-migration collision on this lane in two days** (514/515 was the first, and that
author retired theirs in my favour; this one goes the other way).

- **539 is RETIRED `_SUPERSEDED`, never applied** — proof: `agent_definitions_backup` holds **zero**
  `539_%` snapshots, and 539 takes its snapshot immediately before the UPDATE.
- **537 is better in one respect worth stealing:** it **discovers** the nested step path at run time
  instead of hardcoding `workflow.steps.process_item.config.sub_workflow.steps.mark_complete`, so it
  survives a restructured loop where 539 would refuse.
- **Three analyses 537 lacked are contributed into its file** (`c1530eed0`): the `bugs_closed/287`
  spawn-record argument for why `handler_result` is trustworthy at all; the 49% measurement that
  will stop anyone later "tightening" `?` to `!`; and the premise-guard idea.
- **537 was APPLIED but MISSING from `schema_migrations`.** That is a live hazard — the next
  `--apply` sweep re-offers it and its own idempotence guard RAISEs, aborting a batch that may hold
  other sessions' files. Recorded with `--record-only` and a note naming the applier.

### ✅ PROVEN 2026-08-21 15:4xZ — PERFECT SEPARATION on the orchestration join, 4/4, no exceptions

**The prediction was stated before the evidence and it held.** Every `mark_complete` extraction
found in the live `agent-build-dispatch-loop-*` pods, joined to its own orchestration's `created_at`
against the wire boundary (15:33:39Z):

| orchestration | created | side | `commit_sha` in `requested_fields`? |
|---|---|---|---|
| `0dd63a2f` | 15:28:58 | PRE-wire | **FAIL** — still requested |
| `2c3b482e` | 15:31:49 | PRE-wire | **FAIL** — still requested (iter_0 **and** iter_1) |
| `110e7e15` | 15:34:48 | **POST-wire** | ✅ **PASS** — absent, `requested_fields: []` |
| `f16cc519` | 15:42:11 | **POST-wire** | ✅ **PASS** — absent, `requested_fields: []` |

**Every pre-wire run fails and every post-wire run passes.** The pre-wire rows are the built-in
control: a test that passed everything would prove nothing, and these do not. And the negative half
agrees — **0 conflict rows after the pre-wire tail drained at 15:42:25Z**, against a demand control
of 1 completed post-wire multi-iteration run (the shape that must conflict).

**Two things this run taught that are worth more than the result:**

1. **The step is named `process_item_iter_N_mark_complete`, not `mark_complete`** — the loop prefixes
   the iteration onto the nested step name. My first filter matched `step_name == "mark_complete"`
   and found **nothing**, which looks exactly like "the line was not emitted". Match on
   `.endswith("mark_complete")`. That is **the third filter this lane has armed that certified the
   wrong proposition** (515: `section_facts` as a proxy; 515: the wrong pods; here: the exact name).
2. **The LOG LINE carries `orchestration_id`, even though the `agent_error_log` bridge does not.**
   §2.9 warned attribution had to be by TIME — that is true of the conflict rows and **false of the
   positive test**, which can be joined exactly. That matters here: `2c3b482e` is a **pre-wire** run
   that was still emitting FAILs at **15:42:14**, eight and a half minutes after the wire. Time-based
   attribution would have called that a failure of the fix. **Prefer the positive test precisely
   because it can be attributed.**

### (historical) why the first 4 rows after the wire were NOT evidence it failed

Live config: `commit_sha?` = `handler_result.response.commit_sha`, `updated_at` **15:33:39Z**.
In the **3 minutes** after that, **4 more conflict rows appeared** (15:34:19, 15:34:43, 15:35:06,
15:35:36). **Do not read those as the wire failing.** Their candidate-path counts run
**10 → 14 → 18 → 22**, growing by 4 — the signature of ONE loop accumulating iterations — and the
only bdl orchestration that could have produced them was created at **15:33:22, seventeen seconds
BEFORE the wire**. A run already in flight carries the config it started with.

**THE PREDICTION, stated before the evidence so it could fail — and it held (see above):**
> An orchestration **created after 15:33:39Z** produces **ZERO** `commit_sha` conflict rows when it
> reaches `mark_complete`.

At the time of writing, orchestration `110e7e15-12c3-445c-bf02-297798dfdcc1` (created **15:34:48Z**,
post-wire) was still `AWAITING_RESPONSES` at `process_item_iter_1_call_handler` — **it had not yet
reached the step**. That is the run to read.

```sql
-- the demand control FIRST: post-wire runs that actually reached completion
SELECT count(*) FROM orchestration_states
 WHERE owner_agent_type='build-dispatch-loop' AND created_at > '2026-08-21 15:33:39Z'
   AND status='COMPLETED';
-- then: rows must be 0 once every PRE-wire run has finished
SELECT count(*), max(occurred_at) FROM agent_error_log
 WHERE error_code='RESOLVER_CONFLICTING_CANDIDATES' AND agent_type='build-dispatch-loop'
   AND context->>'field'='commit_sha' AND occurred_at > '<last pre-wire run completed_at>';
```
⚠ **The instrument records no `orchestration_id`** (the known bridge gap), so rows cannot be joined
to runs — attribution is by TIME, which is why the pre-wire tail must be drained before a zero
counts. **A zero read too early is the pre-wire tail, and a non-zero read too early is also the
pre-wire tail.** Wait for the drain.

**Positive test, if you want the stronger one** (same shape as 515's): on a post-wire
`mark_complete` extraction, `commit_sha` must be **ABSENT** from `requested_fields` on the
`=== MASTER EXTRACTOR START ===` line. ⚠ `build-dispatch-loop` runs in its **own ephemeral per-run
pods** (`agent-build-dispatch-loop-<hash>`), NOT in `agent-chassis` — see §2.3's warning, which cost
an hour on 515.

## 2.8 TIER C IS CLOSED OUT — every remaining pair now has a recorded decision (2026-08-21 ~15:3xZ)

⚠ **First, a CORRECTION to §2.4's own table.** It lists `bdl`/`result` as *"quiet, UNWIRED ⇒ not
closed"*. **That is wrong — it IS wired**, `result!: handler_result`, strict. I missed it because I
probed `config ? 'result'` at the TOP level and `mark_complete` is **nested** inside the
`process_item` loop's `sub_workflow`. **That is the fifth time this lane has been bitten by the
top-level census.** Use the recursive form and it is unambiguous:
```sql
SELECT jsonb_path_query_array(default_config, '$.**.config."result!"')::text
  FROM agent_definitions WHERE type='build-dispatch-loop' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;   -- ["handler_result"]
```
**`bdl`/`result` is CLOSED by its existing strict wire.** So the ⚪ list is two, not four.

**The remaining decisions, each read AT THE CONSUMER rather than inferred from a row count:**

| pair | rows | decision | why |
|---|---|---|---|
| `bdl` / `result` | 326, last 08-17 | ✅ **CLOSED — already wired** `result!: handler_result` | §2.4's entry was my error; see above |
| `page-rerender` / `current_page` | 78, last 08-18 | ✅ **CLOSED ON MECHANISM** | §2.5 — no production spec declares `current_page`, so it only ever arrives via the call step 3 gated, and no page-rerender step requests it |
| `generic` / `page_id` | 2, last 08-17 | 📝 **ABSENCE IS CORRECT — flip is the fix, no migration** | Winner was `call_completeness_discovery.response.discovery_result.findings[0].page_id` out of **86–90** candidates: an arbitrary finding's page. Every consumer declares `page_id` Optional (`create_work_item`, `load_current_section_content`, `load_page_record`). A random finding's page is strictly worse than nothing |
| `generic` / `summary` | 3, last 08-17 | 📝 **ABSENCE IS CORRECT — flip is the fix, no migration.** ⚠ **AND IT EXPOSED A NEW BUG** | Winner was `config.workflow.steps.probe_control.config.summary` — the resolver descended into the orchestration's **own workflow definition**. `create_work_item` declares `summary` Optional, so absence is fine and the flip handles it. **But the surface is not: `isInfrastructureKey` skips `agent_config` and `retry_payload` and NOT plain `config`, and 1941 of 3107 orchestrations carry a `config` key — 208 of them the whole workflow. Filed as `bugs_open/350`.** The flip closes the conflicting half of that and **not** the silent half |
| `tg` / `related_pages` | 17, last 08-20 | 📝 **ABSENCE IS CORRECT — flip is the fix** | = `bugs_open/330`; the audit's own words are *"absence is the fix"* |

**So the step-5 census is fully closed:** 11 pairs closed by steps 1–4, 2 more closed here
(`bdl`/`result`, `page-rerender`/`current_page`), 3 dispositioned as *absence is correct* (needing
no migration, only this record), `pbh`/`page_type` fixed and proven by **515**, `tg`/`reason` fixed
by **512** (verification blocked on demand), and `bdl`/`commit_sha` — the last live blocker — fixed
by **539**, in council review now.

**After 539 lands and is verified, the precondition is MET and the flip itself is the only work
left.** Its design must still say out loud what §5 says: "zero conflict WARNs" was never sufficient,
because the instrument cannot see agreeing candidates — and `bugs_open/350` is now a second worked
example of that, on a different surface.

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
| **bdl / `commit_sha`** | **640+, still firing (387 in 24 h to 15:01Z)** | no | 🟡 **NO LONGER BLOCKED — see §2.7. The 315 lane answered, chose option (b), applied it across 9 handlers and handed the wire back. Ours to build.** |
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
