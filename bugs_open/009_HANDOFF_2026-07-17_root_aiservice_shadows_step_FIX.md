# HANDOFF — FIX: root `ai_service` SHADOWS the step-level block (dead per-step config, fleet at 2048)

> **STATUS 2026-07-20 — code fix WRITTEN, COMMITTED, and INERT.** The step-wins
> overlay is in `4b11f223e` (`resolveAIServiceConfig` in `ai_actions.go` +
> `ai_service_overlay_test.go`, 11 cases green). The inverted runbook gotcha is
> corrected in `99dc0f95d`. **Not deployed** — inert until the chassis image is
> rebuilt and rolled, so the defect is still live and this case stays OPEN.
> Council submission `581754c8-390c-4f91-a3e6-43cd08a34e99` in flight (queued at
> 20:14Z; no verdict yet, hence no `Council-Reviewed` trailer on the commit).
> §4's fleet sweep is NOT done. See §7 for the audit that sized the change.

**Filed:** 2026-07-17, from the "diagnosis fixloop 3" thread. Cold-start for a fixing
thread. Mechanism is fully established (loop-cited + proven by direct experiment);
this handoff is about the FIX and the FLEET SWEEP.

**Severity:** High, fleet-wide, and it has ALREADY misled documentation: the fixloop
runbook's gotcha *"max_tokens lives INSIDE a step's ai_service block; root is dead
config"* is **INVERTED** — correct rule: **the ROOT block wins; the step block is
dead whenever a root block exists.** (Corrected in NOTES turn 34 + auto-memory;
`RUNBOOK_diagnosis_fix_loop(10).md` still needs the correction — do it in this
thread's first commit.)

## Working rules
Same as 008 §Working rules (Go; schema first; deploy from committed ref via
`make build-agent-chassis-ref`; bump IMAGE_TAG; verify the POD binary; commit per
task with explicit paths and READ `git diff --cached --name-only` first).
DB: `PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`.

## 1. The mechanism

`platform/orchestration/actions/ai_actions.go` → `ExecuteLLMPromptAction`
(~lines 146–188 as of 2026-07-17):

1. If `agentConfig["ai_service"]` exists (the ROOT block) → `aiServiceConfig` = it.
2. The step-level lookup (`workflow.steps[<current>].config.ai_service`) runs ONLY
   `if aiServiceConfig == nil` — i.e. never, when a root block exists. The step's
   entire block (model, max_tokens, everything) is dead config.
3. `options["max_tokens"]` is then read from `agentConfig["max_tokens"]` (top-level
   key) or `aiServiceConfig["max_tokens"]` (the winning ROOT block). If neither is
   set, `GenerateText` falls back to its hardcoded `"max_tokens": 2048`.

**Proof by experiment (2026-07-16):** diagnose-agent had step-level
`verdict.config.ai_service.max_tokens: 8000` — every verdict call since 2026-07-10
logged `max_tokens=2048`. Setting `max_tokens: 32000` on the ROOT block made the
next call log 32000. (That root-block value is LIVE now — an interim mitigation for
diagnose-agent only; def backed up to `bak_agentdef_diagnose_20260716`.)

Why the docs got it backwards: `page-content-writer` has NO root block, so its
step-level fix (2000→8000, the article-body incident) worked — and was generalised
into a rule that only holds for agents without a root block.

## 2. Diagnosis provenance

- Loop runs (intakes closed, terminals in their `error` fields):
  - `b606dbf6-…` honest UNVERIFIABLE — symptom embedded empirical claims not in
    the bundle (correct refusal).
  - `af19fa62-…` FAILED, API 529 (transient).
  - `80c35dea-…` spawn lost — parent wedged at `spawn_diagnoser` 13.7h, child never
    created. **Preserved deliberately as a live instance of case 003** — do not
    tidy that orchestration row.
  - `960b554d-…` COMPLETED — raw verdict CONFIRMED (5 static citations: the
    root-first assignment, the `if aiServiceConfig == nil` gate, the step-map
    assignment, the max_tokens if/else, the 2048 literal; symptom_check 3/3) but
    **gated UNVERIFIABLE by the two-evidence-family guard** (a CONFIRM needs a
    state/runtime citation alongside static; a mechanism-only symptom gave the
    verdicter nothing to data_request). Graded PARTIAL vs the pre-registered
    rubric. The mechanism is not in doubt — the experiment settles it.
- If this thread wants a GATED CONFIRMED (needed only if you want the loop's
  fix-proposer to consume it): re-dispatch per **authoring rule 4** — state the
  mechanism, then POINT at the evidence tables (`llm_call_log` for the 2048 verdict
  calls; `agent_definitions` for the dual blocks) without asserting rows or counts;
  the verdicter will data_request them (that is exactly how 008's case passed).

## 3. The fix (sketch)

**Recommended: merge with step-wins overlay.** Build the effective config as a COPY
of the root block overlaid by the step block key-by-key (step wins per key). This
preserves root as the fleet default (provider, api_key_env_var, model) while making
per-step overrides (max_tokens, model) actually work — which is plainly what config
authors intended (diagnose-agent's dead 8000; the content-creator family's dead
declarations).

Alternatives considered: step-wins-wholesale (loses root defaults the step omits —
worse); root-wins-status-quo + config sweep only (leaves the trap armed for every
future author).

**Before changing precedence, audit who has BOTH blocks** (behaviour change scope):
```sql
SELECT type,
       (default_config ? 'ai_service')                                   AS has_root,
       jsonb_path_exists(default_config, '$.workflow.steps.*.config.ai_service') AS has_step
FROM agent_definitions
WHERE deleted_at IS NULL
  AND (default_config ? 'ai_service')
  AND jsonb_path_exists(default_config, '$.workflow.steps.*.config.ai_service');
```
For each hit, diff the two blocks' fields — any agent where the STEP block carries a
DIFFERENT model/provider than root will change behaviour under the overlay. Confirm
each is intended (they almost certainly are — that's why the author wrote them).

**Unit test:** table-driven over the three shapes (root only / step only / both) —
`json_envelope_test.go` shows the house fixture style. Also note `ai_actions.go:181`
has a THIRD source (`params.StepConfig.Config["ai_service"]`) — fold it into the
overlay order deliberately and test it.

## 4. The fleet sweep (after — or instead of waiting for — the code fix)

17 agent defs have a root `ai_service` with NO max_tokens → every call at 2048.
10 of them ALSO declare max_tokens elsewhere (dead). Enumerate where:
```sql
SELECT type, path
FROM agent_definitions,
     LATERAL jsonb_paths_of(default_config) -- pseudo: use jsonb_path_query with '$.**.max_tokens' and record paths via a plpgsql helper, or eyeball each def
WHERE ...;
-- practical version: for each of the 17, SELECT jsonb_pretty(default_config) and locate the max_tokens declarations by hand — there are only 17.
```
If the overlay fix ships FIRST, the 10 step-level declarations start working on
their own (configs self-heal — this is the argument for code-first). The remaining 7
(no declaration anywhere) need a deliberate per-agent max_tokens decision — 2048 may
even be RIGHT for tiny content steps; don't blanket-raise. **Back up
`agent_definitions` before any sweep** (pattern: `bak_agentdef_<slug>_<date>`).

Sequencing recommendation: **code fix → deploy → verify with diagnose-agent's own
step 8000 vs root 32000 (root should now LOSE per-key) → then sweep the 7.**

## 5. Verification

- Unit tests green; deploy from ref; grep pod for a log line you add at the overlay
  point (e.g. "ai_service: step overlay applied").
- Live check: `llm_call_log` — fire the diagnose digest or any cheap step-configured
  agent and confirm the logged max_tokens matches the STEP value.
- Interaction with 008: once BOTH ship, a capped call fails loud AND per-step caps
  are honoured — the 2048 silent-truncation class is closed end to end.

## 6. Do-not-relearn

- diagnose-agent root block currently carries the interim 32000 — after the overlay
  fix, its step-level 8000 would WIN for the verdict step. Decide which value stays
  (32000 was sized for Sonnet 5 adaptive thinking; 8000 is too tight — probably
  move 32000 INTO the step block and drop it from root).
- Sonnet 5: omitting `thinking` runs ADAPTIVE and thinking spends from max_tokens —
  at 2048 the verdict produced ZERO text blocks (hard fail). Any agent moved to
  Sonnet 5 needs its cap re-sized, not just carried over.
- Deploys landed 5× during the diagnosing session (1123→1128). Re-verify §1 line
  numbers against the code you check out, not against this doc.

## 7. Behaviour-change audit (run 2026-07-20, live `agent_definitions`)

§3 asks for this before changing precedence. Done — the blast radius is far
smaller than the handoff assumed, because **no step block anywhere overrides a
provider or a model**; every real difference is a `max_tokens` raise.

Query used (dual-block agents):
```sql
SELECT ad.type, s.key AS step, jsonb_pretty(s.value->'config'->'ai_service')
FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.deleted_at IS NULL AND s.value->'config' ? 'ai_service';
```

Exactly **3** agents declare both blocks (not a long tail):

| agent | step(s) | root | step | delta under the overlay |
|---|---|---|---|---|
| `diagnose-agent` | `verdict` | sonnet-5, 32000 | sonnet-5, 32000 | **none** — blocks are byte-identical, so the §6 worry ("step 8000 would WIN") is already moot: the interim 32000 was copied into the step block at some point. Nothing to decide. |
| `feed-triage` | `score_relevance` | sonnet-4-6, 4000 | sonnet-4-6, **8192** | **the only live change:** 4000 → 8192. The author's intended cap, never once applied. |
| `site-adoption-agent` | `analyze_site`, `classify_archetype`, `derive_content_direction`, `generate_design_intent` | sonnet-4-6, 16000 | sonnet-4-6, **no max_tokens** | **none** — the steps restate provider/model and omit the cap, so root's 16000 survives. This shape is *why* the overlay must be per-key: step-wins-wholesale would have dropped these four steps to the hardcoded 2048. |

Two further checks, both clean:
- **Zero** agents declare both a top-level `max_tokens` and a step-level
  `ai_service.max_tokens`, so the untouched top-level branch (`agentConfig
  ["max_tokens"]` wins over the block) cannot newly shadow a step value.
- The **16** active non-snapshot agents with a root block and no `max_tokens`
  (§4 said 17; live count is 16 — `content-creator-contact` appears twice)
  declare no step-level `ai_service.max_tokens` either, so **none of them
  changes behaviour under the overlay**. §4's claim that "the 10 step-level
  declarations start working on their own" does **not** hold: there are no such
  declarations among them. They still run at 2048 and still need the deliberate
  per-agent decision. **The code fix does not self-heal the sweep.**

> **CORRECTED 2026-07-20:** §4's "configs self-heal — this is the argument for
> code-first" is wrong as written. It is right that code-first is the safer
> order, but the reason is not self-healing; it is that the overlay stops new
> per-step config from being silently inert. The 16 capped agents are unaffected
> either way. Caught by running §4's own audit query before trusting its prose.

Sweep list (all `is_active`, root block, no cap anywhere — each needs a
deliberate decision; 2048 may be right for short steps):
`content-creator-about`, `content-creator-contact`, `content-creator-cta`,
`content-creator-features`, `content-creator-hero`,
`content-creator-hero-without-research`, `content-creator-testimonials`,
`content-researcher`, `content_researcher`, `copywriter`, `reasoning`,
`researcher`, `simple-content-writer-with-approval`, `vet-batch-processor`,
`website-builder`. (Note `content-researcher` and `content_researcher` are two
distinct rows — hyphen vs underscore.)

## 8. What the fix actually does (as committed, `4b11f223e`)

`resolveAIServiceConfig(agentConfig, runtimeStepConfig, currentStep)` builds the
effective block by overlay, least- to most-specific, **later wins per key**:
root → `workflow.steps[<current>].config.ai_service` → `params.StepConfig.Config
["ai_service"]`. A block contributes only the keys it declares. The merged map
is a fresh copy, so a runtime override cannot poison the cached agent
definition (there is a test for this). The not-found error path is unchanged.

§3's note about `ai_actions.go:181` being "a THIRD source" is folded in: it is
now the last overlay layer rather than a third fallback.

**Pod-verify literal** (created by this change, not merely used by it — a
pre-existing symbol would give a false pass):
`ai_service: step overlay applied`. Positive control: `ai_service: single source`
(also new) or any known-present symbol.

**Live check after the roll:** fire `feed-triage`'s `score_relevance` and
confirm `llm_call_log` shows **8192**, not 4000. That is the one call in the
fleet whose logged cap changes, which makes it the discriminating test — a
green run of any *other* agent proves only that nothing broke.
