# HANDOFF — 2026-05-08 — Flywheel D iter_0 evaluated; ready for C phase 2

## Context in one paragraph

Today's session ran flywheel D end-to-end on the iter_0 adapter. Held-out evaluation set built from production `llm_call_log` (50 cases, post-training-export, distinct orchestration_ids, defensively excluded from training set; 20 used for this round, 30 reserved for future iterations). 20 inferences on a fresh Thunder Compute A100 80GB SXM4 instance using the Unsloth template (~$0.50). Three-level analysis pipeline written and run: L1 structural metrics (free, local), L2 Claude-Opus-as-judge with anonymized A/B comparison (~$1), L3 auto-selected spot-checks. Honest findings: head-to-head 16-4 in Claude's favour but 4 of those went to Claude despite identical rubric scores (self-recognition signal), mean dimension gaps are small (≤0.25 on 5-point scale), both models fabricate occasionally in real usage, iter_0 won 4 cases for substantive reasons. iter_0 is shippable for low-stakes use; voice fidelity is the main quality lever for iter_1. Total session cost ~$1.50 — sustainable for many iteration cycles. C phase 2 is the next priority, with refined design informed by today's data.

## Current state

```
Flywheel A — training data export pipeline    ✓ operational
Flywheel B — RAG infrastructure                 ✓ done
Flywheel C phase 1 — manual training run        ✓ iter_0 adapter shipped
Flywheel C phase 2 — chassis automation         → next priority, design refined
Flywheel D — Claude vs adapter eval             ✓ iter_0 evaluated, pipeline reusable
```

## iter_0 evaluation results

```
inference_run:    20 held-out briefs, ~22s/case, A100 SXM4 + xformers
judge_model:      claude-opus-4-7 (different from training-label producer Sonnet 4.6)
total_cost:       ~$1.50 (inference $0.50, judge calls $1)

L1 structural:
  iter_0 valid JSON:   20/20    Claude valid JSON:   19/20  (1 db-truncation)
  schema match:        19/19
  length ratio:        median 1.01×, mean 1.11×, range 0.75-2.43
  forbidden phrases:   iter_0 0, Claude 1   [from 14 parseable avoid-lists]
  fabrication regex:   iter_0 0, Claude 0   (regex has poor recall — see L3)

L2 head-to-head (claude-opus-4-7 judge, anonymized A/B):
  Claude won:    16/20
  iter_0 won:    4/20
  TIE:           0/20
  position bias: A won 11/20 (55%) — no clear bias

  Mean dimension scores (iter_0 / Claude / Δ):
    relevance:  4.35 / 4.60 / -0.25
    voice:      4.35 / 4.55 / -0.20
    integrity:  4.65 / 4.75 / -0.10

L2 self-recognition signal:
  4 cases had identical R/V/I scores for both responses
  4/4 of those went to Claude — judge breaks ties on something the rubric
  doesn't measure, correlated with Claude-style. Adjusted reading: Claude 12,
  iter_0 4, rubric-ties-decided-by-judge-preference 4.

L3 substantive observations:
  - iter_0 wins (4) were not flukes: matching brief specifics better, executing
    rewrite guidance Claude ignored, being tighter on voice.
  - 1 decisive iter_0 loss: fabricated specifics in case-study brief
    ("under a week", "per-agent topic routing" — neither in the brief).
  - 1 Claude fabrication: invented "two business days" SLA in contact brief.
  - Both models fabricate at this scale; not an iter_0-specific problem.
```

## Tooling produced

| File | Where | Purpose |
|---|---|---|
| `04_eval_iter0.py` | flywheel_C/ | Inference runner — uses `peft.PeftModel.from_pretrained` (stable cross-version) |
| `level1.py` | flywheel_D/ | Structural metrics (local, no API) |
| `level2.py` | flywheel_D/ | Claude-as-judge with anonymised A/B |
| `build_report.py` | flywheel_D/ | L1 + L2 + L3 → markdown report |
| `held_out_cases_v1.sql` | flywheel_D/ | Reproducible eval-set query |
| `held_out_cases_v1.jsonl` | flywheel_D/ | Stable 50-case comparison set |
| `iter0_evaluation_report.md` | flywheel_D/ | The actual report |

The `held_out_cases_v1.jsonl` is the canonical comparison set — iter_1, iter_2 etc. should evaluate against the same cases so trends are meaningful. 30 cases unused; can extend to all 50 once we want statistical power, or pull a fresh `_v2` for novelty checks.

## Lessons captured today (for FOCUS § 14)

1. **`peft.PeftModel.from_pretrained` is the stable cross-version inference pattern.** Don't rely on `FastLanguageModel.from_pretrained(adapter_dir)` — it broke between unsloth_zoo versions because the newer loader requires `config.json` which `save_pretrained()` doesn't write.
2. **Thunder Compute Unsloth template ships nvcc + CUDA toolkit** (vs Ollama template's runtime-only). torch 2.10 / cu128 by default. The `transformers<5` and `torchao<0.17` pins from the Ollama-template stack may not be needed on the Unsloth template — worth testing on iter_1.
3. **FA2 absent on the Unsloth template.** xformers 0.0.35 is pre-installed and provides memory-efficient attention. ~5 tok/s on 70B inference at seq 4096; FA2 would roughly double that. For inference, xformers is acceptable; for training, install the prebuilt FA2 wheel.
4. **`tnr scp` of a directory creates `{dest}/{source_basename}/`** in BOTH directions. Workaround: scp individual files, or `mv` to flatten after. Phase 2's `model-trainer` should not use `tnr scp` for directories.
5. **Self-recognition bias is observable, not just theoretical.** When the rubric scored responses identically, the judge picked Claude 4/4 times. Using a different Claude model (Opus vs Sonnet) reduces this but doesn't eliminate.
6. **L1 regex fabrication detection has poor recall.** Contextual fabrications (made-up SLAs, invented case-study specifics) don't match patterns like "%" or "300%". L2 judge catches them; humans do better. Don't rely on L1 alone for integrity assessment.
7. **claude.ai markdown auto-linking corrupts code in chat.** `[setup.sh](http://setup.sh)`, `[torch.cuda.is](http://torch.cuda.is)_bf16_supported()`. Run a sanitize step on any code that came through chat before executing.
8. **Production-mode A100 vs Prototyping-mode A100** — Prototyping (TGV virtualised) worked fine for 70B inference at 80GB VRAM. ~$1.79/hr Production vs ~$0.78/hr Prototyping. Phase 2 should default to Prototyping for inference, Production for training (unverified that Prototyping handles long training runs well).
9. **iter_0 won 4 substantive cases.** The model has real strengths that don't aggregate into headline scores. Worth carrying into iter_1 design — preserving them while addressing the voice gap.

## Revised iter_1 priorities

The previous handoff's recommendations were based on smoke-test signals. With aggregate eval data, some need revision:

| Recommendation | Status |
|---|---|
| Filter `<no value>` rows from training data | **Still stands** — fix-deploy date is the filter floor, still need that from the team |
| Save adapters as fp16 (791MB → ~400MB) | **Still stands** — pure win, one-line script change |
| 2-epoch ablation | **Still stands** — epoch 3 loss curve suggests memorisation; cheaper to train |
| ~~Address verbosity~~ | **Dropped** — L2 showed length ratio median 1.01× and mean 1.11×, comparable to Claude. The smoke's "iter_0 longer than Claude" was one-row noise. |
| Improve voice fidelity (NEW) | **Add** — Δ −0.20 on voice is small but the largest dimension gap. Tractable via longer training (4 epochs?), larger LoRA rank (32?), or curating training rows with stricter voice compliance. |
| Fabrication awareness (NEW) | **Add to system design, not iter_1** — both models fabricate occasionally. Probably needs prompt-time guardrails or post-hoc verification, not adapter-level fixes. |

## Cost recap

| Phase | Cost |
|---|---|
| iter_0 phase 1 (training, prior session) | $20 |
| iter_0 flywheel D evaluation (today) | $1.50 |
| **Total per iteration cycle** | **~$22** |

Eval at ~7% of training cost is sustainable across many iterations. Most of the spend is GPU-hours during training; cutting to 2 epochs for iter_1 saves ~$5 (estimated $15-17 total).

## Resumption — Flywheel C phase 2 design (refined with today's data)

The previous handoff's design stands in shape; today's data changes a few specifics.

### Architecture decision: SSH-exec, not HTTP job server (initially)

Original design: HTTP server on the VM, chassis POSTs jobs. Today's data suggests starting simpler:

- Snapshots are uneconomic ($15/mo each) so every training run starts from a fresh instance with `00_vm_setup.sh`. ~25 min cold start.
- Direct SSH-exec via chassis (run setup, scp data, run training, scp adapter, delete instance) is simpler than building+deploying an HTTP server before phase 2's first run.
- HTTP server can be added later as an optimisation when run frequency justifies it.
- This means the `model-trainer` agent uses Thunder Compute API (`POST /instances/{id}/up`, `/down`) plus SSH/scp for data movement — no separate VM-side server to maintain.

### Adapter transport

Today's `tnr scp` of the 791MB adapter took 17 min at ~800KB/s upstream. Round-tripping through the laptop is the wrong shape for chassis-driven training:

- VM produces adapter → uploads directly to Backblaze S3 (~50× faster from a VM than tnr scp upstream)
- chassis reads from S3 to register `model_artefacts` row
- evaluator VM downloads from S3 (also fast)
- laptop never touches the binary

Halving adapter size via fp16 save (per iter_1 priority) makes this even cheaper.

### Tables needed (refined from previous handoff)

```sql
-- Training run lifecycle and metrics
CREATE TABLE model_training_runs (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    export_id               uuid NOT NULL REFERENCES training_exports.runs(id),
    status                  text NOT NULL CHECK (status IN ('pending','running','complete','failed')),
    hyperparameters         jsonb NOT NULL,    -- epochs, lora_r, etc.
    started_at              timestamptz,
    completed_at            timestamptz,
    train_runtime_s         numeric,
    final_loss              numeric,
    peak_vram_gb            numeric,
    cost_usd                numeric,
    error_message           text,
    thunder_instance_id     text,              -- for debugging post-mortem
    created_at              timestamptz NOT NULL DEFAULT now()
);

-- Model adapter artefacts (decoupled from runs to support re-quantization, etc.)
CREATE TABLE model_artefacts (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    training_run_id         uuid NOT NULL REFERENCES model_training_runs(id),
    storage_uri             text NOT NULL,     -- s3://bucket/path/adapter_model.safetensors
    sha256                  text NOT NULL,
    size_bytes              bigint NOT NULL,
    format                  text NOT NULL,     -- 'lora_safetensors_fp16'
    created_at              timestamptz NOT NULL DEFAULT now()
);

-- Evaluation results, one row per (artefact, eval_set, judge_model) triple
CREATE TABLE model_evaluations (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    artefact_id             uuid NOT NULL REFERENCES model_artefacts(id),
    eval_set_uri            text NOT NULL,     -- e.g. held_out_cases_v1
    judge_model             text NOT NULL,     -- e.g. claude-opus-4-7
    n_cases                 int NOT NULL,
    l1_metrics              jsonb NOT NULL,
    l2_metrics              jsonb NOT NULL,
    deployment_decision     text,              -- nullable — set after human review
    created_at              timestamptz NOT NULL DEFAULT now()
);
```

### Agents needed

Three new entries in `agent_definitions`, all category=`specialist` except the orchestrator:

1. **`model-trainer`** specialist
   - Inputs: `export_id`, `hyperparameters` JSONB
   - Behaviour: provisions Thunder instance via API, uploads dataset (fetched via export_id), runs training, uploads adapter to S3, records `model_training_runs` + `model_artefacts` rows, deletes instance
   - Outputs: `training_run_id`, `artefact_id`

2. **`model-evaluator`** specialist
   - Inputs: `artefact_id`, `eval_set_uri`, `judge_model`
   - Behaviour: provisions Thunder instance, downloads adapter from S3, runs `04_eval_iter0.py` against eval set, runs L1+L2 locally (or in-VM), records `model_evaluations` row
   - Outputs: `evaluation_id`

3. **`training-flywheel-orchestrator`** wrapper (orchestrator + coordinator)
   - Workflow: run flywheel A export → spawn `model-trainer` → spawn `model-evaluator` → emit summary report
   - Auto-deployment NOT included in v1 — human reviews and sets `deployment_decision` manually

### Phase 2 first concrete step

I'd start with the schema migrations, since:
- They're the contract everything else depends on
- They can be reviewed and committed independently of any agent code
- Existing migration patterns in the repo (we have files like `002_intake_orchestrator.sql`) are ready to reuse

Then `model-trainer` agent, since it's the longest-pole and exercises the most new infrastructure (Thunder API, S3 upload).

## Resumption checklist (when you come back)

1. **Get the `<no value>` fix deploy date** from whoever shipped it (probably Slack/git-log). Becomes the filter floor for iter_1's training export.
2. **Decide the storage location for adapters** — Backblaze bucket name, path scheme. Small but-blocking decision.
3. **Decide schema migration ordering** — these tables go where in the migrations list? Look at the next available number after `018_briefing_questionnaire.sql`.
4. **Agree on Thunder Compute API key handling** — env var? K8s secret? The chassis needs it to provision instances; same key handles all our flywheel work probably.

## Decisions locked today

- **iter_0 is shippable for low-stakes use** — internal tooling, prototypes, sites without elaborate voice prescriptions. Not for client-facing where Δ−0.20 on voice would be visible.
- **`held_out_cases_v1.jsonl` is the canonical eval set across iterations** — same 20 cases used for iter_0 will be used for iter_1, iter_2, etc., so trends are meaningful.
- **claude-opus-4-7 is the canonical judge model** — different from Sonnet 4.6 training-label producer to reduce self-recognition. Document in evaluation rows.
- **Phase 2 starts with SSH-exec from chassis, not HTTP job server.** HTTP can be added later when run frequency justifies it.
- **Adapter transport via S3, not chassis round-trip.** Phase 2's `model-trainer` uploads directly from VM.
- **Save adapters as fp16, not fp32, in iter_1.** One-line script change.
- **Drop "address verbosity" recommendation** — L2 data showed length is comparable.
- **Add "improve voice fidelity"** — Δ−0.20 on voice is the main quality lever for iter_1.

## Known issues / technical debt

- `<no value>` rendering bug in production prompt builder is fixed but training data inherits it. Filter required for iter_1.
- 1 of 50 held-out rows had Claude's response truncated at db level (`source_log_id 7d5f4ea0…`). Affects iter_0 evaluation by ~5%. Not blocking; flagged in PATCH_2026-05-06.
- `gpu-ollama` k8s endpoint still DOWN. Not used in flywheel D path A; would matter if we revisit Option B (Ollama serving the adapter).
- Snapshot economics ($15/mo for 100GB) ruled out at our run frequency. If iteration tempo increases substantially (>15 runs/month), revisit.
- L1 regex fabrication detection misses contextual cases. Don't tighten the regex — extend to L2 instead.
- 791MB adapters (fp32) cost 17 min on tnr scp. fp16 saves halve that.

## Files to commit

Today's deliverables that should land in the project repo (probably `docs/.../finetuning/` tree):

- `flywheel_C/04_eval_iter0.py` — inference runner
- `flywheel_D/level1.py` — structural analysis
- `flywheel_D/level2.py` — judge runner
- `flywheel_D/build_report.py` — report generator (with the three patches applied: colspan fix, empty-table guard, self-recognition signal)
- `flywheel_D/README.md` — pipeline docs
- `flywheel_D/held_out_cases_v1.sql` — eval-set reproducibility
- `flywheel_D/held_out_cases_v1.jsonl` — actual eval set (20 used, 30 reserved)
- `flywheel_D/iter0_eval_results_v1.jsonl` — raw inference outputs (20 rows)
- `flywheel_D/level1_metrics.json` — L1 output
- `flywheel_D/level2_judgments.jsonl` — L2 output
- `flywheel_D/iter0_evaluation_report.md` — the report
- This handoff doc.
