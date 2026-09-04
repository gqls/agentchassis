# HANDOFF 2026-09-04b — continue here (bugs_open/257 round 3: SHIPPED, migration APPLIED, two things outstanding)

**Supersedes** `HANDOFF_2026-09-04_continue_here.md`, which asked for the owner's ruling on four
decisions. **All four were answered on 2026-09-04 and all four are acted on.** Read that file only for
its analysis of rounds 1–2.

**Status:** the ladder is committed (inert until the next chassis roll); migration `769` is **applied and
live**; `bugs_open/480` is filed for decision 3. **Two things are genuinely outstanding and neither is
urgent** — §4.

---

## 1. What the owner decided, and what each turned into

| # | ruling | what happened |
|---|---|---|
| 1 | *"go ahead with the is-it-live test"* | **two probes ARMED** — see §4a. Neither has fired: both callers are bursty and went quiet. |
| 2 | *"leave it"* (merging the two precedence copies) | Honoured as **option (c)**: the two ladders differ permanently, the difference is a written contract, `TestDirectCallerLadderStaysTwoLevels` pins it. They share only the walker. |
| 3 | *"own piece of work, let me know the bug number"* | **`bugs_open/480`**, with a lane at `docs024_key_docs_latest/direct_caller_llm_observability/`. 257 no longer carries it. |
| 4 | *"the limits are set in each individuals config, sometimes it has been set in the wrong place and sometimes the agent reads the wrong place, please fix it properly"* | **Done at both ends** — §2. |

## 2. Decision 4, and why it was bigger than the brief

The brief was "four `site-adoption-agent` steps declare a budget where nothing reads it". A **recursive**
jsonb walk — instead of the fixed-path query every earlier census used — found that is half of it.

**[MEASURED 2026-09-04, 208 active non-snapshot agents] 171 declarations in FIVE shapes**, of which the
reader looked at two, least specific first:

| shape | n | read? |
|---|---:|---|
| `workflow.steps.<step>.config.ai_service.max_tokens` | 149 | yes — canonical |
| `max_tokens` (agent root, bare) | 10 | **yes, FIRST** |
| `workflow.steps.<step>.config.max_tokens` | 7 | **no — by nobody** |
| `ai_service.max_tokens` (agent root) | 3 | yes |
| `<nested loop step>.config.ai_service.max_tokens` | 2 | yes, via runtime StepConfig |

**Failure (1) SHADOWED** — ten agents' bare root key beat their own step's canonical 8000, pinning it to
500–2000. ⚠ **None of the ten has ever appeared in `llm_call_log`, so this was latent, not live damage.**
**Failure (2) UNREAD** — seven step-level bare keys, including `site-adoption-agent`'s four asking for
32000/6000/4000/4000 and running at the root 16000 across 8 live calls.

**Fixed in two independent halves.** Code (`d88afbf84`, `f18704b9c`): one ladder, most specific first —
`step_config.ai_service` → `step_config` → `workflow_step.ai_service` → `workflow_step` →
`root.ai_service` → `root` — used by `ExecuteLLMPromptAction`, `llmOptionsFromConfig` and `getMaxTokens`,
for `max_tokens` and `budget_tokens`. Config (`763c8002f`, migration `769`, **APPLIED**): 14 declarations
moved into the canonical place, **live on apply**, so both failures were fixed under the current binary
without waiting for a roll.

Verified after: **0** bare root declarations (was 10), **13** canonical root (was 3), **153** canonical
step (was 149), `site-adoption-agent` sending 32000/6000/4000/4000.

**The detector:** `scripts/audit-budget-placement.sh` — calls production's own ladder
(`actions.ResolveStepBudget`), carries no copy of the rule. Register entry **MDL-045**.

## 3. Two mistakes this round made and caught — do not undo either fix

- **A precedence WARNING fires on the healthy majority.** The first cut of the detector, and a runtime
  `Warn` already committed in `ai_actions.go`, both called "two levels declare different numbers" a
  defect. **First live run: 18 findings, every one healthy** — a root `ai_service` default overridden by
  a step is the documented overlay design and `feed-triage` does it on purpose. Both removed; both
  non-findings pinned by name in tests. **Every Go test passed and the warning was correct by its own
  description — only live data disagreed.**
- **`s.value->'config' - 'max_tokens'` is not what it looks like.** PostgreSQL binds subtraction tighter
  than `->`. Caught by running the migration with `COMMIT` replaced by `ROLLBACK`; the runner's own probe
  had **skipped the file** because a header comment contains the word `Rollback`.

Both are in `LANDMINES.md`, with two more from this round.

## 4. WHAT IS OUTSTANDING

### 4a. ~~TWO LIVE CONFIG PROBES ARE ARMED AND MUST BE REVERTED~~ ✅ DONE — both fired, both reverted

> **RESOLVED 2026-09-04, same session.** Both probes fired and both are disarmed. Nothing outstanding here.
>
> | probe | post-arm calls | sent | control (same step, same day, pre-arm) |
> |---|---|---|---|
> | `page-content-writer` `…rewrite_negations` | **50** across 6 loop iterations | **all 15999** | **182 calls, all 16000** |
> | `offer-analyser` `repair_ordering_register` | **2** | **all 1999** | configured 2000 |
>
> **Round 2 is live on `v1.0.1360`, proven first-hand rather than by ancestry.** The `offer-analyser`
> probe closes §2c's hole: 2000 was both its configured value and the old Go literal, so that call site
> could never be checked before; 1999 broke the tie and it read 1999.
>
> Reverted and verified TWICE — the config no longer carries the key, and
> `scripts/audit-budget-placement.sh` no longer reports the two `ambiguous` findings the probes created,
> which is an independent instrument confirming the disarm rather than a re-read of my own write.
>
> The original text is kept below because the METHOD is reusable — read it if you need to re-arm.

#### The method, kept for re-use

They are harmless (one token of difference each) but they are live production config and they will sit
there until someone acts.

| agent | step | armed at | configured | probe |
|---|---|---|---|---|
| `page-content-writer` | `…sub_workflow.steps.rewrite_negations` | 2026-09-04 10:57:51Z | 16000 | **15999** |
| `offer-analyser` | `repair_ordering_register` | 2026-09-04 11:38:21Z | 2000 | **1999** |

Each puts a number where **only round-2 code looks** (`llmOptionsFromConfig` reads the step's BARE key
before the merged `ai_service` block). `offer-analyser`'s is the valuable one: **2000 was BOTH its
configured value and the old Go literal**, which is exactly why that call site could never be checked
before (bug §2c).

```
read:   15999 / 1999 -> round-2 code IS live      16000 / 2000 -> it is NOT
```

**Both callers are bursty.** `rewrite_negations` ran 4–38 times an hour overnight and then nothing for
over two hours. **Do not read silence as a result.** Full arm/read/revert SQL is in the lane RUNBOOK
under "The induced 'is it live' check".

⚠ If nothing has fired and you do not want to babysit it: **revert both and say so** — the ancestry proof
in `bugs_open/257` already establishes that round 2 is live; this probe upgrades an inference to a
first-hand measurement, which is worth having and is not worth leaving live config modified for days.

⚠ The probe works because the direct-caller ladder reads the bare key FIRST. Round 3 deliberately left
that ordering alone (decision 2). **If a later round unifies the ladders, this probe silently stops
discriminating.**

### 4b. The council verdict — read `decided_by` BEFORE `decision`

`SUBMISSION_CORR = 5de01fd3-e683-4425-991c-9cd9fd4d6025`.

**Round 1 came back REVISE and no seat objected to the change.** `decided_by` read
`"unreadable reviewer(s): …"` — **8 of 12 seats returned unparseable output, 4 abstained**, leaving four
readable verdicts of which **two were APPROVE** (constitution, mission). Round 2 was resubmitted on the
same trail at 11:34Z with every readable objection answered. ⚠ **CAUSE FOUND — it did not stall, it DIED.** A fleet-wide Anthropic credit outage returned HTTP 400
"credit balance is too low" from **11:21:11 to 11:56:49Z**, killing **146** calls of which **92 were
council-gate** (measured here; flagged by the `inter thread comms` lane). Round 2 was dispatched at
**11:34:20Z — inside that window.** CLAUDE.md's standing rule *"a missing orchestration row is latency,
not a dropped dispatch — do not retry"* is exactly the WRONG read for this case, and it has been
resubmitted. ⚠ Keep the two 400s distinct while reading today's error rows: a rejected manual thinking
budget and an empty credit balance are both HTTP 400. Earlier note, now superseded: **as of ~14:40Z its report had NOT
landed** — the seats ran (46 `council-gate` calls after dispatch) but no `council_report` row exists,
and no council report has been written fleet-wide since 11:28Z. Treat round 2 as PENDING or possibly
STALLED; re-run the query below before assuming either.

```sql
SELECT metadata->>'decision', body::jsonb->>'decided_by', body::jsonb->>'unreadable'
FROM diagnosis_artifacts
WHERE correlation_id='5de01fd3-e683-4425-991c-9cd9fd4d6025' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```

If round 2 approved: nothing to do — `098` credits the `Council-Submitted:` commits automatically, with
no amend. If it revised again: the objections are in the same row.

[MEASURED 2026-09-04] a round decided by unreadable seats is **3 of 225** councils over 7 days. It is
`bugs_open/138`'s territory and is written up in `LANDMINES.md`.

### 4c. ~~After the next chassis roll — the two steps that close this bug~~ ✅ BOTH DONE

> **RESOLVED 2026-09-04 16:0xZ.** The roll landed as `v1.0.1361`. The pod states
> `git_commit: 06c0b18f233bc600918ef481d32b40f29535f78f`; `d88afbf84`, `f18704b9c` and `22ba668dd` are
> all ancestors, and `34fae21e4` (committed after the cut) is absent, so the control passes and the
> probe discriminates. ⚠ **Ancestry was the only honest instrument** — `AcceptsThinkingBudget` has no
> caller inside the chassis binary, so it is linker-dropped and a literal probe would have said ABSENT
> with clean controls.
>
> **Migration `770_HOLD` applied by hand.** `--record-only` refuses a sidecar, so `bugs_open/257`
> §2026-09-04c is the record, not `schema_migrations`.
>
> **End state:** five declaration shapes are three, all canonical; zero bare-spelling declarations in
> the fleet. `audit-budget-placement.sh` reports `non_canonical` none, `ambiguous` none,
> `thinking_unsupported` none. The one finding left is `provocation-generator-manual.gate`, unconfigured
> since before this lane and now visible for the first time — `bugs_open/205`'s shape, not this bug's.
>
> **`bugs_open/257` MEETS ITS BAR (fixed AND live) and can be moved to `/bugs_closed/`.** Not moved
> here only because two council rounds are in flight on this code (§4b) — the gate cannot block a
> close, but a REVISE landing on a closed file is the messier direction. Read those verdicts, then move it.

#### The original steps, kept for the method

1. **Prove the ladder shipped**, per service, not per fleet:
   ```bash
   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=-1 | grep -m1 'build provenance'
   git merge-base --is-ancestor d88afbf84 <the sha that line reports> && echo SHIPPED
   ```
   ⚠ It is a STARTUP line, so it scrolls; empty means "not in range", not "unstamped". Pass `--tail=-1`
   with `-l` (a fresh landmine from another lane: `-l` defaults `--tail` to 10 per pod).
2. **Apply migration `770_HOLD` by hand** — `docs/agent_docs/sql_for_agents/770_html_developer_chunked_budgets_join_the_canonical_place_HOLD.sql`.
   The apply condition is written into the file as the `merge-base` command above. Then re-run
   `scripts/audit-budget-placement.sh` and expect the **3** `non_canonical` findings to become **0**.

**Then `bugs_open/257` can close**: its bar is *fixed AND live*, and at that point both halves are.

## 4d. One open question a peer handed us — the client-side half of the `budget_tokens` 400

`anthropic.go` emits `thinking:{type:enabled,budget_tokens:N}` on the presence of a positive value, and
round 3 wired that key to the full six-level ladder. That shape is a **400 on `claude-sonnet-5` and
`claude-opus-4-8`** — 88 of 176 model declarations across 22 agents.

**Guarded, not fixed** (`22ba668dd`): `aiservice.AcceptsThinkingBudget` + a parity test + a
`thinking_unsupported` arm in the report. **Live exposure is ZERO** (two independent encodings), so
nothing is broken today — this is a trap armed for the first operator to declare the key.

⚠ **The rule is NOT blanket, and that is why the reader was not changed**: `claude-sonnet-4-6` /
`claude-opus-4-6` still accept it (deprecated), and **`claude-haiku-4-5` REQUIRES it** — 32 model
declarations across 24 live agents. A guard that simply dropped the key would break those.

**The open question:** should the client refuse, drop-with-warning, or keep returning the 400? Dropping
silently would be this bug's own defect class — a configured key nothing reads. Not decided; the
predicate is in place so it can be, without a second model list. Council `47ea9498`.

## 5. What is deliberately NOT done, and should stay that way unless the owner says otherwise

- **72 step-level `temperature` declarations across 12 agents are dead config** —
  `ai_actions.go` reads only `agentConfig["temperature"]` and none of the 12 has a root temperature. Every
  council seat asks for 0.0 and none of it reaches the reader. **Do not "fix" this**: `anthropic.go:300`
  deliberately never sends temperature at all ("Claude Opus 4.7+ returns a 400 for any non-default
  temperature"), so honouring them is a no-op on every one of these agents today and a live behaviour
  change on any future non-Anthropic one. It needs a ruling, not a patch. The detector reports it.
- **`budget_tokens` has ZERO declarations fleet-wide.** The ladder covers the key so the first operator to
  declare one is not caught by a second rule, but that arm is unexercised.
- **`suggest_related_pages@300` truncates 40 of 217 calls** (`tool-deployer`, `tool-generator`). That is a
  budget declared correctly, in the right place, and simply too small — a sizing decision for that lane,
  and the live `fleet-step-token-pressure` check already flags it. Not this bug's.
- **Two ACTIVE `agent_definitions` rows both claim type `content-creator-contact`**, with different
  budgets. A `SingleOwner`-shaped finding for another lane.

## 6. Files

- **Bug:** `bugs_open/257_HANDOFF_2026-08-12_…md` — read §2026-09-04 (round 3).
- **New bug:** `bugs_open/480_HANDOFF_2026-09-04_a_model_call_outside_the_orchestration_layer_is_invisible…md`
- **Code:** `platform/orchestration/actions/llm_options.go` (the ladder + the contract) ·
  `ai_actions.go` (`workflowStepConfig`, the budget block) · `html_actions.go` (`getMaxTokens`) ·
  `llm_budget_ladder_test.go` · `cmd/config-key-audit/budgetplacement.go` (+`_test.go`) ·
  `scripts/audit-budget-placement.sh`
- **SQL:** `769_token_budgets_move_to_the_one_place_every_reader_looks.sql` (**applied**) + `_ROLLBACK` ·
  `770_html_developer_chunked_budgets_join_the_canonical_place_HOLD.sql` (**held**) + `_ROLLBACK`
- **Register:** **MDL-045** (new) and **MDL-039** (corrected — its "root wins" rule is refuted) in
  `docs026_concept_register/register/model-infrastructure.md`; MDL-041 updated.
- **Fleet-wide:** `LANDMINES.md` (four new entries) · `WRONG_CALLS.md` (three).
- **Council:** `COUNCIL_SUBMISSION_2026-09-04_257_round3_one_budget_ladder.json` and the `…-04b_…r2.json`
  revision.

## 7. Traps this round hit

- **A fixed-path jsonb census can only find what you already believe.** Four censuses of this key over
  three weeks; the fifth correction was of METHOD, not of a number. `WITH RECURSIVE … jsonb_each`, then
  `GROUP BY` the path shape — RUNBOOK has it.
- **Migration numbers collide on this tree.** 765 and then 767 were both taken by another session while
  these files were being written. `git add` on arrival so the next session's `ls` sees yours.
- **`HEAD` does not pass its own tests in either package touched here, and it is not this lane's.**
  `platform/orchestration/actions`: `TestFindingCodeScanEveryWriteIsRegistered` /
  `TestTemplateExecutorsAreDeclared`. `cmd/config-key-audit`: `TestShippedRegistryIsSelfConsistent`.
  All three fail on HEAD alone, identically. Verify with
  `scripts/verify-head-builds.sh --with <file> …` and compare against a bare `--test` run before spending
  an hour on them.
