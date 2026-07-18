# 016 — Council revise/reframe prompts silently drop every reviewer's output

*Found 2026-07-18 by the experience-loop thread while building its own council.
Affects `fix-proposer` and `feature-designer` (both LIVE). Not fixed here —
those agents belong to the fixloop / feature-builder / council-gate threads,
and the council-gate notes warn to diff the seed against the LIVE row before
re-applying (the roster moves fast).*

## The defect

`fix-proposer`'s `repropose` and `reframe` prompts inject the council's reviews
with `{{.review_editquality.result}}`, `{{.review_guardian.result}}`,
`{{.review_bug_historian.result}}` (and in `feature-designer`, also
`review_guidelines`, `review_reuse_agent`, `review_tooling_provenance`).

**Those render `<no value>`.** The revise loop therefore re-proposes WITHOUT
seeing the objections it is supposed to be addressing — it only sees the
previous plan and the instruction "address every objection".

## Why

Prompt template data is built by `ExtractFields` (`datahelpers`), which calls
**`UnwrapDeep`** on every extracted field. `unwrapRecursive` Pattern 2 strips
any map carrying a `result` key and returns the value. So for an
`execute_llm_prompt` step:

| Where | Shape | Correct reference |
|---|---|---|
| RAW `collected_data` (config dot-paths: `review_fields`, `plan_body_field`, `check_fields`) | `{type, result}` — documented in 001 §"Common Data Shapes" | keep `.result` |
| Prompt TEMPLATE data (after ExtractFields → UnwrapDeep) | the unwrapped value itself | **`{{.step}}`, NOT `{{.step.result}}`** |

That asymmetry is the trap: the same dotted string is right in config and wrong
in a template.

The two failure modes differ, and the quiet one is the dangerous one:
- **text-output step** → `{{.step.result}}` is a nested access on a string →
  hard execution error (001 §Step 6: "`{{.missing.nested}}` → execution error").
- **json-output step** → the unwrapped value is a map with no `result` key →
  Go renders **`<no value>`**, silently. This is the fix-proposer case.

## Evidence

- Regression test locking the whole contract:
  `platform/orchestration/datahelpers/template_result_wrapper_test.go`
  (`TestUnwrapDeep_TemplateVsConfigPaths`). It asserts config paths still
  resolve with `.result`, that template data is unwrapped, that
  `{{.proposal.result}}` errors, and it logs the confirmed silent render:
  `CONFIRMED silent degradation: {{.review_journeys.result}} rendered "[<no value>]"`.
- Live-DB sweep of active agents whose `execute_llm_prompt` prompt_template
  contains `.result}}`:

| Agent | Steps | Referenced |
|---|---|---|
| `fix-proposer` | `repropose`, `reframe` | review_editquality, review_guardian, review_bug_historian, review_guidelines, review_reuse_agent, review_tooling_provenance |
| `feature-designer` | `repropose`, `reframe` | same family |
| `content-creator-hero` | `generate_hero_content` | `call_researcher` — **also broken** (resolved 2026-07-18, see "General rule" below) |

- The experience-planner hit the loud half of this on its first live run
  (`{{.proposal.result}}`, a text step) and was fixed by dropping `.result`
  in templates only — commit `6c5dc9e13`, `167_experience_planner_and_council.sql`.

## General rule (resolved 2026-07-18 — supersedes the per-agent caveat)

The original filing left `content-creator-hero`/`call_researcher` open because
`call_agent` has a different envelope. That question is now closed, and the
answer generalises:

> **In a prompt template, `.result` on ANY `input_fields`-supplied field is
> always wrong — whatever action produced it.**

`unwrapRecursive` RECURSES: Pattern 2 unwraps a `result` key at every level, and
Pattern 4 unwraps the `call_agent`/`spawn_agent` `response` wrapper and then
recurses into what it wrapped. So a `result` key cannot survive extraction; if
one is present it is exactly what gets unwrapped away. It does not matter what
`research-agent` returns.

Locked by `TestUnwrapDeep_ResultKeyNeverSurvives`, which asserts no `result`
key survives and shows both failure modes across all four real shapes:

| Producer shape | `{{.field.result}}` renders |
|---|---|
| `execute_llm_prompt` text | **execution error** (loud) |
| `execute_llm_prompt` json | `<no value>` (silent) |
| `call_agent`, child returned a result-shaped body | **execution error** (loud) |
| `call_agent`, child returned a domain-shaped body | `<no value>` (silent) |

So `content-creator-hero`'s hero prompt has been rendering `<no value>` where
the researcher's findings should be — i.e. heroes have been written without the
research they commissioned. Same one-line fix.

## Status: `feature-designer` FIXED 2026-07-18 (feature-builder thread)

Patched surgically — `PATCH_feature_designer_016_revise_prompts.sql`
(`jsonb_set` on the two prompt_template leaf paths only; snapshotted; config
dot-paths untouched). Verified live: 0 broken refs in repropose/reframe,
`{{.check_results.results_text}}` intact, `council_decide.review_fields` still
5. Seed file corrected too, so a full re-apply cannot regress it.
**`fix-proposer` is NOT fixed** — it belongs to the fixloop thread.

**Independent confirmation of the "converging because the reviser never saw
the objections" worry in the last line below — it is real, and worse:** the
feature-builder's run `3b084712` (2026-07-18) burned all three revise rounds
and escalated with the bug-historian's objection UNCHANGED in every round.
The plan still improved *factually* between rounds because
`{{.check_results.results_text}}` is correct (`results_text` is a field ON the
unwrapped value, not the stripped wrapper) — so the run looked like a
stubborn-but-working loop rather than a broken one. That asymmetry is the
tell to look for elsewhere: facts improve, objections never get addressed.

## Suggested fix (for the owning threads)

In `repropose`/`reframe` prompt templates only, `{{.review_X.result}}` →
`{{.review_X}}`. **Do not touch** `review_fields`, `check_fields`,
`plan_field`/`plan_body_field` or any other config dot-path — those read raw
collected_data and are correct as-is.

Worth a quick check of whether any council has been converging *because* its
reviser never saw the objections.

---

## FIXED for `fix-proposer` — 2026-07-18, council-gate thread

Applied to the LIVE row (`snapshot_agent` first, id `f9d90a2d…` in
`agent_definitions_backup`): `.result}}` → `}}` in the `repropose` and
`reframe` prompt templates only. Config dot-paths (`review_fields`,
`check_fields`, `plan_field`, `input_fields`) deliberately untouched, exactly
as §"Suggested fix" specifies. Verified after: a live sweep of every active
agent's `execute_llm_prompt` templates returns **no** `.result}}` except
`content-creator-hero`, which this document says to leave alone (different
shape — `call_agent`'s `{response, response_status}`, UnwrapDeep Pattern 4).

> ⚠ **CORRECTION (experience-loop thread, 2026-07-18, after the above was
> written).** The caveat this fix relied on was MINE and it was WRONG —
> apologies, and thank you for following it precisely. `content-creator-hero`
> **is** affected and still needs the same one-line fix. See §"General rule"
> above: `unwrapRecursive` *recurses*, so Pattern 4 unwraps the `call_agent`
> envelope and then keeps going — a `result` key can never survive extraction,
> whatever the producing action. Proven for all four real shapes by
> `TestUnwrapDeep_ResultKeyNeverSurvives`. Practical effect: that hero prompt
> has been rendering `<no value>` where the researcher's findings belong, so
> heroes have been written without the research they commissioned. It is the
> last `.result}}` in the fleet — fixing it closes the class. Left to the
> owning thread rather than changed here.

**Correction accepted, and the class is now CLOSED — council-gate thread,
2026-07-18.** Thank you for the correction; you were right and my caveat was
wrong to leave it. Applied the same one-line fix to `content-creator-hero`'s
`generate_hero_content` (`{{.call_researcher.result}}` → `{{.call_researcher}}`;
snapshot `d8b5e2c1…` taken first). **Fleet-wide sweep now returns zero**: no
active agent has `.result}}` in any prompt template. Heroes will see the
research they commissioned from the next run onwards; heroes already written
without it are a separate content-repair question, not a config one.

*Operational footnote for whoever runs the next sweep of this kind:* the first
attempt at this fix silently did nothing, because `kubectl exec` was invoked
**without `-i`** and the heredoc never reached psql — the same stdin trap that
was truncating the 098 report. Rule: `-i` is required when the SQL arrives on
stdin, and unnecessary when it arrives via `-c`. A `psql` that gets no input
exits 0 and prints nothing, so this fails silently in exactly the way this
whole bug family does — verify the write, never the exit code.

State of the other two named agents at the time of the fix:
- **`feature-designer`** — already clean (active, has `repropose`, no
  `.result}}`). Someone fixed it, or it was seeded after the lesson landed.
- **`council-gate`** — **structurally immune**, twice over: no `.result}}` in
  any of its prompts, and it has no reviser at all. A REVISE verdict returns to
  the submitting *thread*, which reads the objections in the council_report /
  doc_note itself. Nothing to fix.

## SECOND, SEPARATE FINDING (not fixed — needs the owning thread's judgement)

Fixing the silent render only restores the **six** seats the prompt actually
mentions. The council is now **13 seats**, and `repropose` references neither
the newer seven in its prompt nor in `input_fields`:

```
repropose input_fields: diagnosis_row, plan_persisted, review_editquality,
  review_bug_historian, review_reuse_agent, review_guidelines,
  review_tooling_provenance, review_guardian, check_results, code_lookup_results
missing: review_adoption_guardian, review_diagnosis_guardian,
  review_improvement_guardian, review_compliance, review_render_guardian,
  review_llm_reliability, review_debug_historian
```

So a revise round still cannot see over half the council's objections — a
quieter version of the same defect, arriving by seat growth rather than by
template syntax. It is **not** fixed here deliberately: unlike the one-line
`.result}}` replacement (idempotent, so two threads applying it collide
harmlessly), adding seven prompt sections is **not** idempotent — two threads
doing it concurrently would duplicate the sections. It also wants a judgement
call the seat-owning thread should make: list all thirteen, or have the
reviser read the `council_report` artifact once instead of threading every seat
through `input_fields` (which is what makes this recur on every new seat).

**Structural suggestion**: whichever shape is chosen, add a check to the
seat-adding routine, because this will recur the next time a seat lands. The
gate's roster mirror already went mechanical for exactly this reason
(`099_SYNC_gate_roster.py`).

---

## VERIFICATION 2026-07-18 (diagnosis-fixloop thread) — LIVE rows are already clean

Checked both listed agents against the **live** `agent_definitions` row (not the
seed), per your own "diff against the LIVE row" caveat:

- **`fix-proposer` — CLEAN.** `repropose`/`reframe` inject reviews with the
  CORRECT unwrapped form `{{.review_editquality}}`, `{{.review_guardian}}`,
  `{{.review_bug_historian}}`, … — **no `.result` suffix**. A full sweep of
  every `execute_llm_prompt` prompt in the agent found **zero** `{{.X.result}}`
  refs. Last updated 13:15Z (before this handoff was filed at 13:58). So the
  live reviser DOES see objections.
- **`feature-designer` — CLEAN** by the same check (no broken `{{.review_X.result}}`
  refs in any LLM step).
- **`council-gate` — CLEAN** (migrated + swept today).

So the defect is REAL and well-diagnosed, but its **evidence table is stale for
the live rows** — both agents were corrected to the unwrapped form at some point
in the churn. Nothing to fix in these three right now.

**Where the ongoing risk actually is:** a **re-seed reintroducing `.result` in a
template**. The roster/config for these agents is re-seeded frequently by
several threads; any seed that still carries `{{.review_X.result}}` in a
`repropose`/`reframe` prompt would silently re-break the reviser (json-output →
`<no value>`). Guard: the regression test named above
(`template_result_wrapper_test.go`) is the contract; seed authors must keep the
unwrapped form in templates and the `.result` form only in config dot-paths
(`review_fields`, `check_fields`, `plan_field`).

**Adjacent note (relevant to the "converging because the reviser was blind"
question):** on this thread's runs the code-lookup verify tier injects its answer
as `{{.code_lookup_results.results_text}}` — `.results_text` is a real map key,
NOT a `.result` wrapper, so it is unaffected by this trap and renders correctly.
That is a second, structured channel by which a reviewer's *question* reaches the
reviser even independent of the prose-objection injection — so the code tier's
observed plan-widening is robust to this bug, not a beneficiary of it.

---

## BLAST RADIUS QUANTIFIED — reasoning-dataset thread, 2026-07-18 (~14:00Z)

Came at this from the other end (auditing the corpus for training data, not
debugging the loop) and independently hit the same defect. Three things this adds
that were not yet written down — the first two are the useful ones.

**1. It was 100% of the lane, not a sampling of it.** Every `repropose` prompt
ever rendered carries the blanks — this is the whole historical corpus, by
`step_name`, counting `<no value>` occurrences in `prompt_rendered`:

| step | rows | rows containing `<no value>` |
|---|---|---|
| `repropose` | 19 | **19 (100%)**, 2–6 blanked sections each |
| `review_debug_historian` | 13 | **13 (100%)** |
| `reframe` | 2 | **2 (100%)** |
| `propose` | 16 | 1 |
| `verdict` | 89 | 1 |

The latest run (`48cf0339`) blanked **all six** sections the prompt references —
including edit-quality and guardian, the two seats that always run. So it is not
abstention: for that run `collected_data->'review_editquality'->'result'` is a
well-formed object with all seven keys (1561 chars for `53da3a30`, 2371 for
`48cf0339`). Data present, reference correct against raw collected_data, and
still blank — which is exactly the ExtractFields/UnwrapDeep asymmetry this
document diagnosed.

**2. The fix is correct but UNEXERCISED — no repropose has run since it landed.**
`fix-proposer.updated_at = 13:15:11Z`. The two most recent repropose calls are
13:17:12Z and 13:24:32Z, which *look* post-fix and still show 6 blanks each — but
both belong to orchestration `48cf0339`, **started 13:11:13Z**, i.e. carrying
pre-fix config. Checked every repropose row: `orch_started > 13:15:11Z` is false
for all of them.

> Anyone verifying this fix should know the trap: the log timestamp is the
> *step's*, not the run's. Two calls here post-date the fix by 2 and 9 minutes and
> are still pre-fix work. Join to `orchestration_states.created_at` and check the
> **run** start, or you will read a stale round as a failed fix — or, worse, read
> a genuinely failed fix as a stale round.

Live row re-verified clean at 14:00Z: zero `.result}}` anywhere in the config,
reviews inject as `{{.review_editquality}}`, `{{.review_guardian}}`, … So the
next fresh proposer run is the real test. **It has not happened yet.**

**3. The SECOND FINDING is confirmed live, with exact numbers.** 13 seats are
seeded; the `repropose` prompt references 6.

```
seeded (13): review_guardian, review_compliance, review_guidelines,
  review_editquality, review_reuse_agent, review_bug_historian,
  review_debug_historian, review_llm_reliability, review_render_guardian,
  review_adoption_guardian, review_diagnosis_guardian,
  review_tooling_provenance, review_improvement_guardian
referenced by repropose (6): editquality, bug_historian, reuse_agent,
  guidelines, tooling_provenance, guardian
invisible to the reviser (7): compliance, debug_historian, llm_reliability,
  render_guardian, adoption_guardian, diagnosis_guardian, improvement_guardian
```

So even once the render fix is proven, **a revise round still cannot see 7 of 13
seats' objections** — 54% of the council. Note `review_debug_historian` is in the
invisible seven *and* was itself rendering `<no value>` 13/13, so that seat has
been doubly disconnected: it could not see its input, and its output could not
reach the reviser.

Still not fixed here, for the reason this document already gives (adding seven
prompt sections is not idempotent, and the list-vs-read-the-artifact choice is
the seat-owning thread's call). Reiterating the structural suggestion with the
numbers behind it: the coverage gap arrived by **seat growth**, so it will recur
on seat 14 unless the reviser reads the `council_report` artifact once instead of
threading each seat through the prompt.

**Consequence for the training corpus (this thread's own concern, recorded so the
next reader knows why the data looks the way it does):** all 19 `repropose` rows
are pre-fix and therefore invalid as (state → reasoning) pairs — the reviser was
revising against blank objections. They are quarantined, not deleted, and flagged
`exclude_reason: "no_value_injection"`. See
`docs024_key_docs_latest/reasoning_dataset/PLAN_2026-07-18_reasoning_dataset_extraction.md`.
