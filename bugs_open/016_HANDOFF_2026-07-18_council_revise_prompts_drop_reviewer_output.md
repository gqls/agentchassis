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
