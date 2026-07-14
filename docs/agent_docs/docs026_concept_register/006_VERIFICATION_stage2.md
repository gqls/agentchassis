# Stage 2 — Code/DB Verification Log

Started 2026-07-14. Each entry cross-checks a stage-1 *documentary signal* against
code, config, and deployment reality. Stage-1 status is what the docs claimed;
**verified status** is ground truth.

Method: for each concept, follow its `verify-later` pointers into the repo. Record
the exact file/line evidence. Where ground truth differs from the documentary
signal, correct the register entry and note why the docs misled.

---

## Batch 1 — priority tensions + suspected duplicates (2026-07-14)

### MCL-002 — Adjacent-cluster Phase 4a rollout: va001 second cluster
- **stage-1 signal:** deployed
- **verified status:** **aspirational** ⚠️ *(status corrected)*
- **evidence:**
  - `va001` appears in **zero** code, config, kubeconfig, or manifest files. Every
    hit is under `docs/_archive/` (running_notes prose).
  - Kubeconfigs present: `config_production_uk001`, `gpu_config_sanjose001`.
    There is **no** `config_production_va001`.
  - `remote-job-spawner` has kustomize overlays and terraform for **`uk_001` only**
    (`deployments/kustomize/services/remote-job-spawner/overlays/production/uk_001/`,
    `deployments/terraform/environments/production/uk001/services/agents/2260-remote-job-spawner/`).
    No va001 overlay exists.
- **why the docs misled:** `PLAN_isolated_chat_environment(5)` describes va001 in the
  **present tense** ("A second K8s cluster (va001…) *runs* remote-job-spawner"). It is a
  plan document narrating its own design as if built. Consolidation reasonably preferred
  the "more recent, more specific" evidence — but recency does not equal reality.
  **This is the canonical stage-2 failure mode: aspirational prose written in present tense.**

### MCL-001 — Multi-cluster dispatch contract (dispatch_agent + remote-job-spawner)
- **stage-1 signal:** partial
- **verified status:** **partial** ✓ *(confirmed, and made precise)*
- **evidence:**
  - `platform/orchestration/actions/dispatch_actions.go` — exists, 278 lines.
  - `cmd/remote-job-spawner/main.go` — exists, 591 lines.
  - `platform/orchestration/actions/registry.go:95` — `"dispatch_agent"` **is** registered.
    The docs listed "the registry patch is outstanding" — that gap is **now closed**.
  - Deployed to the primary cluster (kustomize + terraform, uk_001).
  - **Zero** `.sql` / `.yaml` / `.yml` / `.json` files reference `dispatch_agent` —
    **no workflow or agent definition invokes it.**
- **precise state:** the machinery is built, registered, and deployed — but nothing calls
  it, and (per MCL-002) there is no second cluster to dispatch *to*. It is a complete,
  live, unused code path. The owner's read ("aspirational, not deployed") is correct at
  the **system** level; "partial" is correct at the **component** level.

### MCL-003 — Cluster-filter gap in remote-job-spawner (Gap A)
- **stage-1 signal:** partial
- **verified status:** **partial** ✓ *(confirmed exactly)*
- **evidence:** `cmd/remote-job-spawner/main.go:202` — the `target_cluster` filter **exists**
  (`if targetCluster != "" && targetCluster != clusterID && targetCluster != "any"`).
  Line 203 logs the skip path at **`logger.Debug`**, not `Info`. The one-line
  observability fix is still outstanding, precisely as the register says.

### MCL-004 — Dispatch confirmation observability gap (agent_dispatch_log)
- **stage-1 signal:** aspirational
- **verified status:** **aspirational** ✓ *(confirmed)*
- **evidence:** `agent_dispatch_log` — zero hits across all `.sql` and `.go`. Table never created.

### MDL-029 / FTW-003 — LoRA fine-tuning path and the iter0 adapter
- **stage-1 signals:** MDL-029 deployed; FTW-003 partial ("last-mile wiring outstanding")
- **verified status:** **both correct — they describe different things.** Tension dissolves.
- **evidence — the artefact is real and closed out (MDL-029 = deployed ✓):**
  - `iter0_adapter_out/adapter_model.safetensors` — 828 MB, present on disk.
  - `iter0_adapter_out/manifest.json`: base `unsloth/Llama-3.3-70B-Instruct-bnb-4bit`,
    1,958 examples, 3 epochs, lora_r 16, final loss **0.266**, 25.1 h runtime
    (90,337 s), peak 44.2 GB VRAM, completed **2026-06-04T20:33:11Z**. A genuine,
    successfully-trained adapter.
- **evidence — it never reached inference (FTW-003 = partial ✓):**
  - `deployments/kustomize/services/ollama-adapter/base/deployment.yaml:34` pulls only
    **`nomic-embed-text`** and **`mistral-small3.1`** — stock models. No Modelfile, no GGUF
    import, no adapter mount.
  - Zero references to `lora` / `adapter_model` / `iter0` anywhere in the Go source. The
    only hit is an aspirational comment in `vet_med_price_scrape_action.go:187`
    ("collects training data for *future* LoRA fine-tuning").
- **resolution:** the adapter was **trained but never served**. The `log→export→LoRA→GGUF→
  Ollama→swap` pipeline is complete through `LoRA` and stops dead there. Consistent with the
  owner's "late testing" read. **Recommend clarifying MDL-029's summary** to say the adapter
  is a closed-out *training artefact*, not a serving model — as written, "deployed" invites
  the misreading that it is in production.

### ADM-007 / ADM-008 / PUB-001 — public API + site_ownership *(duplicate cluster)*
- **stage-1 signals:** ADM-007 aspirational; ADM-008 abandoned; PUB-001 aspirational
- **verified status:** all three **confirmed unbuilt** ✓
- **evidence:** zero hits for `site_ownership` (any `.sql`/`.go`) and zero hits for
  `api/v1/sites` (any `.go`).
- **duplicate ruling:** **PUB-001 is a genuine duplicate** — it is exactly ADM-007
  (the `/api/v1/sites/*` endpoint plan) plus ADM-008 (the `site_ownership` junction table),
  merged into one entry by a different consolidator cluster. Recommend: keep ADM-007 and
  ADM-008 (the finer-grained pair, correctly split by status — the endpoints are
  *aspirational*, the table is *abandoned*), and retire PUB-001 to a pointer.
  The `public-api.md` category holds only this one concept and can be folded into
  `admin-dashboard-and-api.md`.

---

## Scope finding: the `deployed` bucket is not safe to trust

The single status error found in batch 1 (**MCL-002**) came from the **deployed** bucket —
not from the partial/unknown bucket that stage 2 was scoped to verify. Every partial,
aspirational, and abandoned signal checked in batch 1 held up exactly.

That is the opposite of the assumed risk profile. The mechanism is clear and general:
**a plan document that narrates its own design in the present tense reads as evidence of
deployment.** Consolidation's tie-break rule — prefer the most recent, most specific
evidence — actively *selects for* this failure, because the polished present-tense plan is
usually the newest document in the family.

Implication: the 871 `deployed` concepts are the bucket most likely to contain
false positives, and false positives there are the *expensive* kind — they are what a
stage-3 council agent would confidently assert as built. The partial/unknown bucket is
comparatively safe: it is already flagged as uncertain, so a wrong signal there is
self-limiting.

**Recommendation:** widen stage 2 from *"verify the 314 partial/unknown"* to
*"verify the 314 partial/unknown **and** sweep the 871 deployed for present-tense-plan
false positives."* The deployed sweep is cheap per concept — most have a named file, table,
or endpoint in `verify-later`, and existence is a one-grep check.
