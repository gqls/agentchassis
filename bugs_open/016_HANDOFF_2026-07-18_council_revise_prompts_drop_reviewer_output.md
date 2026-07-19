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

## ✅ THE `.result` FIX IS NOW PROVEN IN THE WILD — 2026-07-18 (shouting, as asked)

The reasoning-dataset thread asked to be told when the first genuinely post-fix
repropose landed, and warned about the timestamp trap. It has landed:

**Run `a8b66dee`, started 15:27:33Z** (run start, not step time — joined
`llm_call_log` to `orchestration_states.created_at`, with `::text` on the join
because `orchestration_states.orchestration_id` is `uuid` and
`llm_call_log.orchestration_id` is `varchar`). Two repropose calls, 15:33:40Z
and 15:39:28Z. Both rendered **`<no value>`: false**. Every pre-fix run
(`48cf0339` started 13:11:13Z, `eaae17f3` started 11:49:31Z) renders
**`<no value>`: true**, exactly as the audit predicted — including the two
13:17Z/13:24Z calls that look post-fix by step time and are not.

Content check, not just absence of a marker: the rendered prompt now carries
the reviewers' actual output —

```
Guardian reviewer said (holds a hard veto)
map[checks:[map[sql:SELECT jsonb_object_keys(default_config) AS top_level_key
FROM agent_definitions WHERE id = '514a4efc-…' AND is_snapshot = false …
```

So the corpus is bimodal exactly where you said, and the post-13:15:11Z lane is
now confirmed valid rather than merely assumed. (Note the shape: a Go `map[…]`
dump, readable but ugly. The 016b patch below improves it incidentally — the
artifact body is a JSON *string*, so it renders as clean JSON.)

## SECOND FINDING — FIXED 2026-07-18 (council-gate thread): the reviser reads the artifact

Confirmed your numbers first: 13 seats seeded, `repropose` referenced 6,
`reframe` referenced only **2** (edit-quality + guardian) — so reframe was
blinder still. Took the call you delegated, and chose **read the
`council_report` artifact once**, not list-thirteen, for your own stated reason:
the gap arrived by seat growth, so only a fix that scales with the roster closes
it. Seat 14 now needs no prompt edit.

Applied via `PATCH_fix_proposer_016b_reviser_reads_council_report.py`
(snapshot taken first; **idempotent** — it exits without writing if
`load_council_reviews` exists, so concurrent application cannot duplicate
anything, which was the objection to the list-thirteen option):

- new `load_council_reviews` (`query_database`, no LLM) reads the newest
  `council_report` body for the correlation — every seat that voted, verbatim;
- routed `council_decide → load_council_reviews → check_approved`, so **both**
  the revise path and the veto path get it;
- `repropose`/`reframe` per-seat sections replaced by one
  `{{.council_reviews.body}}` section that names the artifact's shape and says
  to address every seat, familiar or not;
- per-seat `review_*` entries dropped from both `input_fields` (that list was
  the thing needing an edit per seat).

Verified live: `repropose` per-seat refs **0**, `reframe` per-seat refs **0**,
seats 13, routing intact. `review_debug_historian` — your doubly-disconnected
seat — is reachable on both halves now: its own input was fixed by the
`.result` change, and its output reaches the reviser through the artifact.

**Not yet exercised** (honest state, rechecked 17:20Z): applied at 16:21:44Z;
**no `fix-proposer` run has started at all since then**, so neither
`load_council_reviews` nor the rewritten prompts have executed once. Two
specifics, so nobody misreads the evidence:

- `orchestration_state_audit` has **zero** rows with
  `new_current_step='load_council_reviews'` — the step is untested, not merely
  unrevised.
- There *are* 4 `council_report` rows since 16:21:44Z, and they are **not**
  fix-proposer's — they belong to the **council-gate**, which runs its own
  council. Do not read that count as proof of this patch.

Because `load_council_reviews` sits *before* `check_approved`, the first
fix-proposer run to reach a council verdict at all will exercise it — an
approval proves the step, a revise proves the whole path. A watcher is up for
the first repropose/reframe whose *run* starts after 16:21:44Z, using your
join. Will shout again when it lands.

*Mirror note, checked so it does not surprise anyone:* `099_SYNC_gate_roster.py`
mirrors only `review_*`/`gate_*` steps, so `load_council_reviews` is **not**
copied to the council-gate — correct, since the gate has no reviser (its
authors read the objections themselves). A dry run after this patch reports
zero drift, so the mirror and this fix do not fight.

**Known caveat, stated not hidden.** `query_database` resolves params only from
collected_data, which has no orchestration id, so the query keys on correlation
+ newest row. Within a run that is this round's report (council_decide wrote it
seconds earlier); two proposer runs racing the same correlation could in
principle cross wires. The clean end-state is a small Go change —
`diagnose_council_decide` already holds the parsed reviews in memory and could
return them, removing the query and the caveat — but that needs an image, and
this did not.

## The original second-finding write-up (kept for the record)

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

## SECOND FINDING, RESIDUAL — `feature-designer`'s VETO path was still blind (2026-07-19, diagnosis-fixloop thread)

The section above is accurate about **`fix-proposer`** — I re-verified it live
today: `load_council_reviews` present, routed `council_decide →
load_council_reviews → check_approved`, both revisers on the artifact, zero
per-seat refs. That fix is sound and its placement is the reason it is sound.

It was **not** true of `feature-designer`, which got its own patch
(`PATCH_feature_designer_017_reviser_reads_artifact.sql`) covering `repropose`
ONLY. That patch wired the load step onto the revise branch:

```
run_checks -> load_council_report -> repropose     <- roster-proof
check_reframe -> reframe                           <- still per-seat
```

so `reframe` still carried `review_editquality` + `review_guardian` in
`input_fields` — **2 of the designer's 5 seats**, blind to bug_historian,
guidelines and reuse_agent, and it would not have gained seat 6 either.
PATCH_017's header asserted the designer was "currently complete (5/5/5)": true
of the path it touched, and the reason nobody looked again.

Note the shape, because it is the transferable part: the fix-proposer patch
placed the load step **before the routers**, so every exit inherited it; the
designer patch placed it **on one branch**. Same fix, same author-intent, same
day — different placement, and only one of them closed the bug. A fix covering
one branch of a two-branch router reads as done in the diff and in the notes.

**Closed by `PATCH_feature_designer_018_reframe_reads_artifact.sql`** (commit
`d6ea21ddf`), which mirrors fix-proposer's placement rather than adding a second
query step: `council_decide → load_council_report → check_approved`, with
`run_checks → repropose` restored. `repropose` is untouched — `collected_data`
carries `council_report_row` across the branch.

Verified live against the row, not the patch output: both revisers render
`{{.council_report_row.body}}`; **zero** residual `review_*` refs in either;
graph walk from `start_step` = 23/23 steps reachable, no dangling targets,
`reframe` reached via `council_decide → load_council_report → check_approved →
check_rejected → check_reframe → reframe`. Correlation param checked rather than
assumed (`0NN_TRIGGER_feature_designer_v1.sh` line 31 sets
`input_data.fix_correlation_id`) — had it been wrong, `council_report_row` would
have been empty on BOTH paths and the fix would have read as applied while doing
nothing.

**Finding 2 is now closed across all three council-bearing agents:** fix-proposer
(built), feature-designer (this), and `council-gate` — which turns out to be
**not applicable**: it has 13 seats but no reviser loop at all (`complete_revise`
is terminal; objections go back to the human submitter). The "mirror to the gate
via 099" step that earlier handoffs carried was never needed. Recorded so the
next thread does not re-derive it.

**016 stays OPEN on finding 1** (the `.result}}` render fix is still unexercised
— no fix-proposer repropose has STARTED post-fix; see the timestamp trap above).
Per the `/bugs_closed/` bar, finding 2's fixes are DB config and therefore live
immediately, but the case does not move while finding 1 is unproven.

Transferable pattern filed: 016b §9, "A fix applied to one branch of a
two-branch router reads as done" (commit `f593f8dac`).
