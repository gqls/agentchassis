# HANDOFF — 2026-08-09 — `bugfix_136_config_key_aliases`: clear the deferred items, cold start

**Owner instruction 2026-08-09: "we can fix those deferred items now."** That is the mandate —
the items below are no longer deferred; they are the task.

**Read first:** this file, then `bugs_open/136` **§6, §10, §11** (§1's harm claim is superseded
by §6 — never quote it; §11 is the runtime witness). The lane RUNBOOK
(`RUNBOOK_config_key_aliases.md`) has every command with its gotcha. The previous handoff
(`HANDOFF_2026-08-08_continue_here.md`) is history — its one owed item was delivered (§11).

## Where the lane stands (one paragraph)

The bug's mechanism is fixed at the framework level and proven at every layer: the alias seam
(`ActionInputSpec.DeprecatedConfigKeys` + `ResolveConfigSetting`, SCR-006) is in the live
binary (pod-grep, both controls), joins the live definitions (audit UNKNOWN KEYS 4→1), and is
**witnessed working at runtime** (§11: a throwaway agent filed `item_domain: "content"` and
the row landed `pipeline='content'` on v1.0.1268). `summary_template` was resolved by owner
decision (§10, mig 343). What remains is the residue: mislabelled rows not yet repaired, dead
keys not yet removed, live definitions still spelling keys the old way, and the opt-in not yet
completed. All data below re-measured **2026-08-09 ~09:30Z** — re-run every query before
acting on it; figures here go stale in days.

## The items, in recommended order

Data-side items (B, C, A) are live immediately, no build, no roll. Item D is platform code —
council gate, image build, roll. Do B/C/A in one session comfortably; D can follow in the
same session or another.

### A. Repair the mislabelled rows — now THREE, not four

`bugs_open/136`'s update says four rows (2 `complete`, 2 `detected`). **Live today: 3.** The
2026-08-03 `capability_gap` (`detected`) no longer exists — it was not repaired, it is simply
gone (row deleted by something; not investigated). Re-derive, never trust this count either:

```sql
SELECT id, item_type, pipeline, status FROM site_work_items
WHERE created_by='completeness-discovery-agent' AND pipeline='design';
--  b6edce72… page_canonical_collision complete   (2026-08-04)
--  a833555d… page_canonical_collision complete   (2026-08-04)
--  74bb48ff… capability_gap           detected   (2026-08-04)
```

Note it is **`created_by`**, not `source` (`source` says `discovery`, which also matches the
*legitimate* design-discovery agent's rows — filtering on `source`+`pipeline` sweeps in ~90
genuinely-design rows; `created_by` is the discriminator).

**The fix** is a keyed data update, mirroring §10's pattern (key on what you measured, so a
row something else has since changed is left alone):

```sql
UPDATE site_work_items SET pipeline='content', updated_at=now()
WHERE created_by='completeness-discovery-agent' AND pipeline='design'
RETURNING id, item_type, status;   -- expect exactly the ids you enumerated first
```

- §6 proved **no live consumer distinguishes `design` from `content`** — this is a record
  correction, safe both ways. It belongs in a migration file (next free number **347**;
  sequence lives in `docs/agent_docs/sql_for_agents/`) so it is replayable and reviewed.
- **Do not scope-creep**: `74bb48ff` (`detected`) stays undispatched after the repair —
  `detected` is a dead queue (`bugs_open/083`, open). Its finding (11 near-duplicate section
  pairs on site `00ff3af5`) may be worth routing, but that is an owner call on 083's queue,
  not part of this label fix. Say so in the bug file rather than silently fixing the label
  and implying the item is now live.
- **Verify the no-op case too**: post-roll mislabels are **0** (query above with
  `created_at > '2026-08-08 17:00Z'`) — the code fix is holding; the repair should not
  change that number.

### B. `plan_sections.domain` — the last UNKNOWN key in the audit

`page-build-handler`, step `plan_sections`, carries `domain: "site_record.domain"`. The spec
declares `pipeline`; **no live step sets `pipeline`**, and `"domain"` occurs 0 times in
`plan_sections_action.go` (§2c). It was left in place only because `page-build-handler` is
hot. The audit (`./scripts/audit-config-keys.sh`) reports it as the sole UNKNOWN key.

**Before touching it, read the action** (`platform/orchestration/actions/plan_sections_action.go`)
and answer one question: does anything on the live path read the spec's `pipeline` key, and
would the site's domain flowing into it change behaviour? §2c concluded genuinely dead —
re-confirm, then **delete the key** from the live definition (migration, same file as A if
you like) **and from the seed in the same commit** (the `bugs_open/134` lesson — §10 did
exactly this for `summary_template`; grep `sql_for_agents/` for the page-build-handler seed).
Do NOT "fix" it by renaming to `pipeline` without reading the action first — wiring a
dot-path string into a key the code reads as a literal is this bug's own §2 mistake inverted.

- `page-build-handler` is **hot** (builds run all day). The edit is one `jsonb` key removal —
  definition config is live immediately. Make the migration's UPDATE surgical
  (`default_config #- '{workflow,steps,plan_sections,config,domain}'`) and re-read the live
  row after. Leave snapshots (`is_snapshot=true`) alone — they are history.
- **Acceptance**: `audit-config-keys.sh` UNKNOWN KEYS goes **1 → 0**. That is the headline
  number for the whole lane.

### C. Rename the deprecated keys in the live definitions — drain the DEPRECATED list

The aliases keep the old spellings *working*; the standing migration list is meant to be
drained. Live carriers today (RUNBOOK query, re-run it):

| old key | steps | agents |
|---|---|---|
| `item_domain` | 9 | build-briefing-agent, deduplicate-sections, domain-research-classifier, domain-strategist, domain-submitter, improvement-loop (×2 steps), tool-improver, vertical-exemplar-researcher |
| `check_domain` | 3 | completeness-discovery-agent, design-discovery-agent, quality-discovery-agent |
| `target_domain` | 1 | improvement-loop |

(`item_pipeline` — the NEW name — already has 2 carriers: claims-auditor,
site-work-orchestrator. They need nothing.)

**The fix**: one migration renaming the key in each step config
(`jsonb_set` the new key with the old key's value, then `#-` the old key), **plus the seed
files in the same commit** (grep `sql_for_agents/` for each key: `049_domain_research_classifier`,
`050_build_briefing_agent`, `060_domain_strategist`, `068_domain_submitter_agent`,
`054_improvement_loop`, `269_deduplicate_sections_handler`, `074_completeness_discovery_agent`,
`059_quality_discovery_agent`, `048_discovery_agents`, … — enumerate by grep, don't trust
this list). Safe in any order relative to anything: the binary honours both names and **the
new name wins when both are set**, so even a partially-applied migration is correct at every
intermediate state.

- **These are other lanes' agents.** The rename is behaviour-preserving by construction, but
  check `scripts/who-owns.py` / `MEMORY_workstreams.md` for the domain-pipeline and
  improvement-loop lanes before editing, and grep live transcripts (`who-owns` reads commits
  only, so a session mid-edit is invisible). Contribute a note into their lane dirs or the
  bug file — per the 2026-07-29 owner ruling, consumers of a shared seam get *told*.
- **Acceptance**: audit DEPRECATED section shows the aliases still *declared* but the
  **in-use count 0**; the RUNBOOK census query returns zero rows for all three old keys; and
  the §11 warn line stops being emitted (unobservable in logs anyway — <1s retention — so
  use the census, not logs).
- Aliases stay declared in the specs afterwards — they are the net for snapshot replays and
  stragglers. Do not remove them.

### D. `create_work_item` full opt-in (`CheckConfig: true`) — the platform-code item

Blocked previously on three unadjudicated keys; two remain (`summary_template` was resolved
in §10):

1. **`spec_fields` — 1 carrier: `grounded-explainer`, step `create_review_item`.** Dead (read
   by nothing; §3, reaffirmed §10 — it is why both review items had empty specs). Remove from
   the live definition + seed `224_grounded_explainer_agent.sql` (same file §10 already
   corrected once — the pattern is there).
2. **`domain` on a `create_work_item` step — 1 carrier: `claims-auditor`, step
   `request_claims_review`, value `"site_record.domain"`.** DO NOT convict it dead by grep:
   the action calls `inputs.Get("domain")` when building `item_key` from `item_key_prefix`,
   and `ExtractActionInputs` **Strategy 1 resolves whatever `config["input_fields"]` names,
   outside the declared spec** (the §"LANDMINE" in the bug file — this exact key nearly got
   filed as dead once already). Read the claims-auditor step's full config: if it sets
   `item_key_prefix` and/or `input_fields` naming `domain`, the key is LIVE and belongs in
   the spec's `Optional`/`ConfigKeys`, not deleted. Adjudicate with evidence, then either
   remove it (data) or declare it (code).

   Note: 18 steps across 13 agents carry a config key literally named `domain` — **only the
   claims-auditor one is on `create_work_item`**. The rest belong to other actions where
   `domain` may be a real key. Out of scope; do not sweep them.

Then the code change: fill in `CreateWorkItemInputSpec.ConfigKeys` with the full literal-key
contract (from the action body: `item_type`, `handler_agent`, `item_pipeline`, `severity`,
`source`, `summary`, `status`, `priority`, `item_key_prefix`, `item_key_suffix_field`,
`spec_paths`, `spec_literal`, `recurrence_expected`, `input_fields`, plus whatever the
adjudication of `domain` decides) and set `CheckConfig: true`. **`priority` IS read** — via
`GetIntField` (§5 note; the §3 read-list was wrong on this) — the RUNBOOK's "grep the key
name, never the access pattern" gotcha exists because of exactly this.

- **This is `platform/` code → council gate** (one run for the coherent task; budget ~30 min;
  `Council-Submitted:` trailer if committing before the verdict). It is an opt-in to existing
  detection, not a new seam — no RFC, same scope argument the `architecture` seat already
  confirmed for SCR-006 (§7).
- Inert until an image rolls (`make build-agent-chassis` from committed HEAD, bump
  `IMAGE_TAG`, verify at the pod per RUNBOOK §"Proving the fix" — enumerate by IMAGE, both
  controls).
- **Order matters within D**: remove/declare the dead keys in data BEFORE the code ships,
  or the new `CheckConfig` warns on every `grounded-explainer` run about a key you already
  know is dead. Data first, then code.

### E. Optional — `resolveAgentTypeForSpawn` convergence (`spawn_actions.go:3154-3163`)

The fourth hand-rolled literal-setting alias (`group_type` → `agent_type`), measured at
**zero live carriers** (§7.3). Converging it onto `DeprecatedConfigKeys` is a small code
change with no behavioural exposure; it can ride D's council round as a second edit, or stay
recorded. If you skip it, say so in the bug file — it is currently listed as "recorded, not
pulled in", which remains honest.

## Bookkeeping (non-optional)

- **Migrations**: `docs/agent_docs/sql_for_agents/`, next free number **347** (346 is taken).
  Dry-run the runner per session; `--apply` takes EVERY pending file — scope it
  ([migration-runner-practice] memory). Verify blocks that must abort need `DO`/`RAISE`,
  not bare `SELECT`s.
- **Commits**: pathspec per task, forward-only, no amends. A same-file passenger is possible
  on any shared doc (it happened to this lane's LANDMINES entry on 08-08 — recorded in
  NOTES); check the commit-scope block.
- **When each item ships, update**: `bugs_open/136` (a §12), the lane NOTES (append, with
  evidence inline), `README_where_we_are.md` (plain prose, append-only — it is the owner's
  document), **and this handoff** (a handoff outlives its work; the next reader cannot tell
  it shipped unless the asking file says so). The audit output before/after is the evidence
  to bank for A/B/C; the pod-grep recipe for D.
- **The acceptance instrument for the whole lane** is `./scripts/audit-config-keys.sh`:
  target state = UNKNOWN KEYS **0**, DEPRECATED declared but **in-use 0**, and (after D)
  `create_work_item` present in the opted-in set. Bank the full output in NOTES.
- If everything lands: update the concept-register SCR-006 entry (visible dated correction —
  the standing migration list is drained), and the memory topic file
  `bugfix-136-domain-pipeline-rename.md`. The bug stays in `bugs_open/` (owner ruling
  2026-08-06) — mark it finished *in the file*, do not move it.

## Traps you will otherwise rediscover (pointers, not restatements)

- `ActionInputSpec.Deprecated` CANNOT carry a literal setting — it path-resolves the value
  and silences the detector. `DeprecatedConfigKeys` is the literal-setting field. (LANDMINE,
  bug §6-landmines, RUNBOOK "Which alias field".)
- Grep the **key name**, never `config["` — helper reads (`GetIntField`, …) are invisible to
  the access-pattern grep. (RUNBOOK, WRONG_CALLS 08-08.)
- Definition edits are live immediately; snapshots are history — filter
  `deleted_at IS NULL AND NOT COALESCE(is_snapshot,false) AND is_active` in every census,
  and leave snapshot rows unedited.
- Pod logs cannot witness anything here — an active chassis pod retains **<1 second** of log
  (LANDMINES "retrievable log holds less than a second"; §11). If you need a runtime witness
  for anything, use the §11 throwaway-agent recipe (RUNBOOK, `witness_136_fire.sh` /
  `witness_136_poll.sh` in this directory).
- `kubectl -l app=agent-chassis` sees 2 of ~25 pods carrying the image. Enumerate by IMAGE
  for any deploy proof (RUNBOOK).
- The 090 coverage check: before dispatching anything at these agents, check open
  `site_work_items` on the target — and none of these items needs a 090 run (mechanisms all
  diagnosed in this bug already; the work is repair, rename, declare).

---

## STATUS 2026-08-09 (afternoon) — THIS HANDOFF IS DISCHARGED. Do not start from it.

Everything asked for above is done or explicitly declined with a reason. A handoff outlives
the work it asked for and the next reader cannot tell it shipped unless the asking file says
so — so it says so here.

| item | outcome |
|---|---|
| A — repair the mislabelled rows | **DONE**, migration 349, verified by id |
| B — `plan_sections.domain` | **DONE**, 349 (def + seed 065). `UNKNOWN KEYS` 1 → **none** |
| C — drain the deprecated spellings | **DONE**, 349 — but **19 carriers, not 13** (see below) |
| D — `create_work_item` opt-in | **DONE**, data 350 + code `ee07e3d86`, council `98d0ef43` |
| E — `resolveAgentTypeForSpawn` | **SKIPPED**; the reason below supersedes this file's |

Acceptance banked: `./scripts/audit-config-keys.sh` → `UNKNOWN KEYS: none`,
`DEPRECATED KEYS: none`, exit 0.

**Three corrections to the file above, so nobody re-derives them:**

1. **§C's table is wrong: there are 19 live carriers, not 13.** Six live inside a loop
   step's `sub_workflow.steps` (component-quality-auditor, internal-linker, tool-auditor ×2,
   tool-suggester ×2). The RUNBOOK census this table was built from walks
   `->'workflow'->'steps'` and cannot see them — a 32% undercount that reads as complete.
   RUNBOOK is corrected; use its recursive query or the text scan, or call
   `validation.WalkSteps` from Go.

2. **§Bookkeeping's "next free number 347" was stale by the time work started** — 347 and
   348 were taken by other threads. Used 349 and 350. Re-derive it; never carry it forward.

3. **§E's premise is wrong.** It is not "a small code change with no behavioural exposure":
   `spawn_agent` has **no `ActionInputSpec` at all**, `agent_type` is a framework key, and
   the literal (`group_type`) and path-valued (`group_type_field`) halves belong in
   *different* alias fields. Zero live carriers re-verified, so nothing is at risk from
   leaving it.

**What is now open, and it is NEW rather than left over:** `create_work_item` steps carrying
a config key named `spec` — read by nothing, three live steps, and in improvement-loop's
case it means `refresh_site_components` never reaches the rerender gate (16/16 rows with an
empty spec). Fixing it is a behaviour change that interacts with `bugs_open/226`, so it is an
owner call, not a tidy-up. Under diagnosis: **090 `be967639-d195-444a-b9c3-ef1445ff7ae1`**.
Full account in `bugs_open/136` §12 and the lane NOTES.

**Cold start for whoever picks this lane up next:** `bugs_open/136` §12 first, then the lane
NOTES tail. Not this file.

---

**UPDATE 2026-08-10 (bugfix_234 lane):** the open item above is **CLOSED** — owner call
taken (RESTORE the flag), migration 364 translated all three `spec` carriers, and the
class fix landed as `ActionInputSpec.RemovedConfigKeys` (SCR-007) + `StrictConfig: true`
on `create_work_item` (commit `d278d7b25`, rides v1.0.1278). Case file `bugs_open/234`.

**And one NEW tracked item lands in its place, so it does not become folklore
(council round `3eb0d1f1`, bug_historian objection):**
`mark_page_needs_attention.notes_field` and `.validation_issues_field` — adjudicated dead
by migration 356 (left standing deliberately: they encode an author's intent the action
never had, and pages has no column for it). They are the next **`RemovedConfigKeys`
candidates**: whoever adjudicates implement-vs-delete should either implement them or
declare them removed — the mechanism now exists, one line each. **Blocked on RFC_021's
question 1** (the adoption protocol for live hard-fail on the shared validator).
