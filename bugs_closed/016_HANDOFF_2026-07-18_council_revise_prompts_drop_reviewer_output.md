# 016 — Council revise/reframe prompts silently drop every reviewer's output

> ## ✅ CLOSED 2026-07-21 (bugfix-016 thread, owner-directed) — fixed AND live
>
> All fixes are DB config (live immediately, no image roll), and the core
> machinery is now **verified against the running system**, not just committed:
> - **Finding 1** (`.result}}` → `<no value>` render): fixed on every affected
>   agent — `feature-designer` (proven in the wild, run `a8b66dee`), `fix-proposer`
>   and `content-creator-hero` (per-seat refs removed / corrected), and the
>   2026-07-21 **recurrence** in `domain-research-classifier`'s new
>   `review_mission_alignment` seat (fixed live + seed corrected). Fleet sweep for
>   `.result}}` in any active LLM prompt template returns **zero**.
> - **Finding 2** (reviser blind to seats added after the prompt): fixed on
>   `fix-proposer` (`load_council_reviews`) and `feature-designer`
>   (`load_council_report`); N/A on `council-gate` (no reviser loop). The
>   `fix-proposer` step is now **exercised live** — run
>   `177b9fb1` ran `council_decide → load_council_reviews → check_approved →
>   complete` and populated `council_reviews.body` with the full 5994-char council
>   report. Per this document's own bar, *an approval proves the step*.
>
> **Documented residual, deliberately NOT blocking the close** (owner ruling): the
> only post-fix verdict was **APPROVE round 1**, so `repropose`/`reframe` has not
> rendered `{{.council_reviews.body}}` on a live REVISE round. It is *strongly
> evidenced* — the field is populated and the identical `.body` template form
> renders in the same run — but not directly observed. It will be confirmed
> naturally the next time any real `fix-proposer`/`feature-designer` run draws a
> REVISE; no separate work is scheduled. See the 2026-07-21 "LIVE EXERCISE"
> section at the bottom for the full evidence.
>
> **Adjacent finding, NOT part of 016, left for the fixloop thread:** `propose`
> has its own `max_tokens=8000` and no `tolerate_truncation`, so a complex-plan
> bug (the pilot `e08c5b01`) refuses the whole fix run with `0 chars recovered`
> before any council runs. Cap not raised (019 ruling).
>
> *(Numbering trap: `bugs_closed/016` is a DIFFERENT case — ssh/`$HOME`. Resolve
> 016 by slug, never by number.)*

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

> **Still unexercised at 2026-07-19 12:41Z**, and worth recording *how* that was
> established, because two obvious checks both lie:
> - `llm_call_log.agent_type='fix-proposer'` returns **nothing ever** — the
>   chassis stamps these rows `generic` (the same reason `council_report`'s
>   `source_agent` is `generic` fleet-wide). Filtering on it reads as "never
>   ran".
> - `persist_plan` is **not** fix-proposer-only — `feature-designer` and
>   `experience-planner` have a step of that name too. Eleven `persist_plan`
>   executions on the 19th (10:44–11:48) looked like fix-proposer activity and
>   were the experience-planner's council (fingerprinted by its distinctive
>   steps: `review_journeys`, `review_feasibility`, `review_mvp`,
>   `review_honesty`, `compose`/`recompose`).
>
> The reliable probe is `load_council_reviews`, which **only** `fix-proposer`
> has: `SELECT count(*) FROM orchestration_state_audit WHERE
> new_current_step='load_council_reviews';` → **0**. Cross-checked: the last
> `propose`/`repropose`/`load_diagnosis` of any kind was 2026-07-18 15:39:29Z.
> So fix-proposer has not run since before the patch — the fix is untested
> because the loop has been idle, not because anything is wrong with it.

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

---

## RE-CHECK 2026-07-20 (bugfix-016 thread) — a NEW recurrence, fixed; and the "PROVEN IN THE WILD" section is MISATTRIBUTED

Three things, in order of how much they change the picture.

### 1. The class re-opened, in a seat added the same day — found and fixed

The fleet sweep this document closed at "zero" on 2026-07-18 returned **one** hit
today:

| agent | step | reference | producer shape | failure mode |
|---|---|---|---|---|
| `domain-research-classifier` | `review_mission_alignment` | `{{.analysis.result}}` | `analysis` = `{type: json, result: {...}}`, confirmed live in `collected_data` | silent `<no value>` |

Seeded today by the R0/R1 mission lane
(`SEED_classifier_mission_R0_R1_2026-07-20.sql`, commit `6a1d5b8f3`) — i.e. the
predicted regression vector in §"Where the ongoing risk actually is" fired, via a
**new seat**, not a re-seed of an old one. The seat is observe-only and had
**never executed** (`llm_call_log` rows for `review_mission_alignment` = 0), so
nothing was damaged; it was caught before its first run.

**Fixed** (snapshot `e6ca8cca…` taken first): `{{.analysis.result}}` →
`{{.analysis}}` in the prompt template only. Verified against the row, not the
exit code — template now matches `{{.analysis}}`, no `.result}}` remains, and the
fleet sweep is **back to zero**. The seed file was corrected too, plus a comment
naming the asymmetry, so a fresh-environment apply cannot reintroduce it. (The
seed's own `UPDATE` is guarded by `NOT (steps ? 'review_mission_alignment')` and
matches 0 rows today, so the LIVE fix was the load-bearing one.)

This seat is a good teaching case for the asymmetry, because it contains **both
forms, both correct**: the template needs `{{.analysis}}`, while the sibling
`gate_mission_note` condition `mission_review.result.objection_found` is a CONFIG
dot-path reading raw `collected_data` and **must keep** `.result`. Left untouched.

### 2. CORRECTION — run `a8b66dee` is the FEATURE-DESIGNER, not fix-proposer

§"✅ THE `.result` FIX IS NOW PROVEN IN THE WILD" attributes run `a8b66dee`
(started 15:27:33Z, `<no value>`: false) to `fix-proposer`. It is not. Its
workflow steps are:

```
load_spec, design, check_spec_approved, persist_plan, run_checks,
council_decide, repropose, reframe + 5 seats (editquality, guardian,
bug_historian, guidelines, reuse_agent)
```

That is `feature-designer`. `fix-proposer`'s shape — `load_diagnosis`, `propose`,
`code_lookup`, `select_panel`, 13 seats + 12 `gate_*` steps — belongs to
`48cf0339`, whose run started **13:11:13Z**, four minutes BEFORE fix-proposer's
13:15:11Z fix. Every fix-proposer repropose on record is pre-fix.

**This is the `persist_plan` trap that §"Still unexercised at 2026-07-19" already
warns about, one level down: `repropose` is not fix-proposer-only either.** Three
agents have a step of that name, and the earlier section's own join filters on
none of them. The tell that caught it was reading the rendered prompt rather than
the timestamp: `a8b66dee`'s repropose opens "REVISE the staged build plan …
stages are commits … capabilities listed missing", which is designer language —
fix-proposer revises an *edit plan* against a *diagnosis*.

So finding 1's true state is per-agent, not fleet-wide:

| agent | `.result` fix | exercised post-fix? |
|---|---|---|
| `feature-designer` | applied | **YES — proven by `a8b66dee`** |
| `fix-proposer` | applied | **NO** — last repropose run started 13:11:13Z, fix at 13:15:11Z |
| `content-creator-hero` | applied | [UNVERIFIED] — `generate_hero_content` has **zero** rows in `llm_call_log` ever, so this table cannot answer it either way |

**The closing line of the previous section is therefore CORRECT** ("016 stays OPEN
on finding 1 … no fix-proposer repropose has STARTED post-fix"). Recording that
explicitly because this thread first concluded the opposite — that the closing
line contradicted the proof section — and was one query away from writing "016
finding 1 is proven, close it". The proof section is the wrong one, not the
closing line. Logged in `WRONG_CALLS.md`.

### 3. Finding 2 is still unexercised on BOTH agents — and today's re-seed did not regress it

```
orchestration_state_audit, new_current_step='load_council_reviews' (fix-proposer)  -> 0 rows
orchestration_state_audit, new_current_step='load_council_report'  (feature-designer) -> 0 rows
```

Neither reviser-reads-the-artifact patch has ever executed. `fix-proposer` has
not run since 2026-07-18 15:39Z; the loop is idle, not broken.

**Re-seed check (the risk this document names):** 162 active agents share
`updated_at = 2026-07-20 17:57:45Z`, so a fleet-wide re-seed ran today. It did
**not** regress either patch — verified against the live rows:
`council_decide → load_council_reviews → check_approved` and `council_decide →
load_council_report → check_approved` both intact, per-seat `review_*` entries
still absent from every reviser's `input_fields`, no `.result}}` in either agent.

### What can be said about the unexercised path WITHOUT spending a council round

Rather than fire a run, the three fragile parts were tested separately. The
first two are live checks; only the wiring is left.

- **The query works.** `load_council_reviews`'s SQL, run verbatim against a live
  correlation (`0a07f5ed…`), returns a **16,294-char** `council_report` body
  carrying every seat verbatim (`{"abstained":4,"decided_by":"objection from
  bug_historian","decision":"revise","reviews":[…`). The documented type worry is
  a non-issue: `diagnosis_artifacts.correlation_id` is **`text`**, so the `$1`
  param needs no cast (unlike `load_diagnosis`, which casts `$1::uuid`).
- **`{{.council_reviews.body}}` is the right form, not a guess.** A
  `query_database` step with `output_format: object` flattens the selected
  columns onto the result map alongside the envelope. Proven from a step that HAS
  run — `diagnosis_row`'s live keys are `rows, count, status, columns,
  conclusion`, where `conclusion` is a selected column. So `council_reviews` will
  carry `body` the same way `diagnosis_row` carries `conclusion`, and
  `{{.council_reviews.body}}` is structurally identical to
  `{{.diagnosis_row.conclusion}}` — which rendered real content, no `<no value>`,
  in a live prompt.
- **[UNPROVEN] the wiring itself**: that `council_decide` actually routes into the
  load step, that `output_field: council_reviews` lands in `collected_data`, and
  that the reviser prompt renders it. Only a run proves that, and per this
  document an APPROVED round proves the step while a REVISE round proves the
  whole path.

**016 stays OPEN**, on the same finding 1 it has been open on — now stated
per-agent: `fix-proposer` unexercised, `feature-designer` proven. Finding 2
unexercised on both. The one-line class fix in §1 is closed and live.

---

## LIVE EXERCISE 2026-07-21 (bugfix-016 thread) — `load_council_reviews` FINALLY RAN; step PROVEN, revise-render path still owed

Fired `fix-proposer` twice to close the "0 audit rows, never executed" state that
has kept finding 2 open since 2026-07-18. Result: **the step is proven live**; one
detour and one honest gap remain.

### Run 1 — REFUSED at `propose` (a separate, real defect surfaced)

`091_TRIGGER` on the pilot `e08c5b01` (dartsonline guides). It never reached the
council: `propose` hit `stop_reason=max_tokens (output_tokens=8000, 0 chars
recovered)`, **failed after 4 retries**, and routed to `complete_refused`. The
pilot's staged plan is too large to emit under propose's 8000 cap (adaptive
thinking on sonnet-5 spends from the same budget → 0 recoverable chars).

Two things worth recording, neither of them 016:
- **`propose` has its OWN `ai_service.max_tokens=8000`** (no root block, so no
  shadowing — checked). It carries **no `tolerate_truncation`**; that protection
  (019) was armed on council *review* seats, not on `propose`. So a complex-enough
  bug refuses the whole fix run with no review at all. That is a genuine fix-loop
  limitation — surfaced here, left for the fixloop thread; **cap NOT raised**, per
  019's standing ruling that raising just moves the cliff.
- This is why the pilot is a poor vehicle to exercise the council. Switched to the
  minimal seeded benchmark `11111111` (the PR-#1 one-line debug-log fix), whose
  plan is small.

### Run 2 — COMPLETED end-to-end; `load_council_reviews` executed and populated correctly

Run `177b9fb1-428a-40e7-b084-bdb48547822c` on `11111111`. (Dispatched ~11h after
firing — queued behind the backlog, bugs_open/030; a reminder that a missing row
is latency, not a drop.) Fingerprinted as genuine fix-proposer by its steps
(`load_diagnosis`/`propose`/`code_lookup`/`select_panel` + 13 seats + 12 `gate_*`),
per the §9 lesson two entries above — not by a step name three agents share.

Audit trail, the part that matters:

```
council_decide        11:02:38
load_council_reviews   11:02:42 → 11:02:51   <-- FIRST execution in 016's history
check_approved         11:02:52
complete               11:02:53
```

`SELECT count(*) … new_current_step='load_council_reviews'` had been **0** every
time this document checked it since 16:21:44Z on 2026-07-18. It is no longer 0.

**The step did its job, verified against `collected_data` (not the audit alone):**
`council_reviews` landed as an `object` with keys `body, rows, count, columns`,
`count=1`, and **`body` = 5994 chars** — byte-identical in length to this run's
`council_report` artifact, holding every seat's verdict verbatim
(`{"abstained":6,"decided_by":"all reviewers approve","decision":"approved",
"reviews":[{"reviewer":"editquality",…`). So the documented structural prediction
is now confirmed with live data: `query_database` + `output_format:object` flattens
the selected `body` column onto the result map, and **`{{.council_reviews.body}}`
resolves to the full report** — the exact shape argued from `diagnosis_row`.

**Render corroboration from the SAME run's `propose` prompt** (the template engine
that a repropose would use):
- `{{.diagnosis_row.conclusion}}` → rendered (both distinctive diagnosis phrases
  present; **no** `<no value>` on the conclusion).
- `{{.last_bundle.body}}` → rendered `<no value>`, and it is **benign**: `11111111`
  has no `kind='bundle'` artifact, so `load_last_bundle`'s `string_agg(body…)` over
  zero rows is NULL. Correct `.body` form, genuinely-absent data — the *opposite*
  of the 016 defect (a wrong reference blanking data that is present). Recorded so
  the next reader does not misread it as a recurrence.

### What this does and does NOT settle

- **SETTLED:** `load_council_reviews` executes, completes, routes
  `council_decide → load_council_reviews → check_approved`, and **populates
  `council_reviews.body` with the whole council verbatim.** Finding 2's "the step
  is untested, not merely unrevised" is now false. Per this document's own bar —
  *"an approval proves the step"* — the step is proven.
- **STILL OWED:** the verdict was **APPROVE round 1**, so `repropose`/`reframe`
  never executed — *"a revise proves the whole path"* has not happened. The
  reviser rendering `{{.council_reviews.body}}` into its prompt is now
  **strongly evidenced** (the field is populated; the identical `.body` form
  demonstrably renders in the same run) but **not directly observed**. Marking it
  as exactly that, not as closed — this bug's whole history is inferences stated
  as findings.

**016 stays OPEN**, but its remaining gap has shrunk from "the step has never run"
to "no *revise* round has run since the fix." Closing it needs a fix-proposer run
whose council **objects** — the minimal benchmark approves, the pilot truncates
`propose`, so forcing a revise deterministically is a credits-costing exercise and
is flagged as the owner's call rather than spent here unasked.
