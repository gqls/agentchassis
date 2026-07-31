# NOTES — finetuning.uk service (append-only, newest at the bottom)

Running record: what was tried, what the system actually said, and every misstep.

---

## 2026-07-31 — workstream opened; design session

**What happened.** Owner asked for finetuning.uk to offer a real, paid, demo
fine-tuning service (a few pounds, payment link). Full exploration + plan written:
`PLAN_2026-07-31_finetuning_uk_service.md` (approved by owner in-session).

**Key verifications this session** (evidence inline in the PLAN):
- finetuning.uk: live Class A site, Cloudflare zone ACTIVE, Worker route bound
  (`/worker-health` → "Worker is running!") [verified-live].
- Cluster has **no ingress and no LoadBalancer at all** (`kubectl get ingress -A`,
  `get svc --field-selector spec.type=LoadBalancer` → none) [verified-db].
- `thunder_config`: $1.80/hr a100xl, $30/day cap, 2 concurrent, gate currently
  OPEN ($0 spent, 0 active) [verified-db].
- `training_exports.runs`: 3 rows, all 2026-04-23, page-content-writer. The
  a8484922 export **reports 1957 rows and holds 0** — known trap, re-confirmed.
- `model_lifecycle.training_runs`: 1 complete (iter_0, 2026-06-03), 5 pending,
  4 failed. Instance history: last real boxes 2026-06-18 (~$32 each).
- tools-api: guarded public surface on the island (httpguard, clientip, CORS
  from own DB), behind an outbound-only Cloudflare tunnel [verified-source].
- The cluster→VM pull seam (`pull_report_requests` + `emit_report_status_files`
  + `ndjson_feed.go`) is BUILT; scheduled tasks seeded disabled; the island-side
  `/requests` endpoint is NOT built [verified-source].
- idea.uk `billing.go`: complete Stripe integration (checkout + HMAC webhook,
  no SDK), proven by a real card payment 2026-07-27 [verified-source].

**Owner directions (this session):** sample dataset OR upload; playground
GPU-served in bounded named hours (not 24h; not CPU-Ollama first — "ollama will
be more trouble"); protect the chassis using tools-api with guards; concierge
first; NEW Thunder token required before any GPU use; Thunder is stop/start with
all artefacts persisted by us (B2/Postgres); price for GPU, adjust later.

**Design decisions of record:**
- Front door = `finetune` route group in tools-api on the island;
  `tune.finetuning.uk` CNAME → existing Cloudflare tunnel. Chassis protected
  structurally (pull-only, owner ruling 2026-07-25).
- **Posture change, recorded honestly:** Stripe keys will live on the island,
  which was designed to hold no production credentials. Money credentials, not
  cluster credentials — but a deliberate widening, to be registered.
- Playground = cheap GPU (l40s/a6000 class) provisioned per booked window,
  model loaded from B2, decommissioned at window end; reaper is the backstop.
- Phase −1 (owner): mint new Thunder token → RUNBOOK §1. Phase 0: one measured
  small-model run + playground rehearsal; it also closes the flywheel's
  `RUN_SH_DONE ⟹ durable` gate (FTW-032/035) for ~$1 instead of ~$20.

**Work done today:**
- Standing five created (this directory).
- `run.sh` parameterised for small models (`BASE_MODEL`, `SAVE_STEPS` env vars,
  defaults preserve 70B behaviour exactly). Bundle deliberately NOT re-uploaded
  to B2 — the deploy happens when Phase 0 runs (RUNBOOK §2).
- Token rotation procedure verified against the live cluster: key confirmed in
  `personae-default-secrets` (11 other keys beside it — patch, don't recreate);
  adapter pod running; auth shape `Authorization: Bearer` from client.go:235.

**Licence research (web, 2026-07-31) — shapes the model menu:**
- **Mistral 7B v0.3 — Apache 2.0.** No flow-down, no naming requirement,
  derivatives may be closed. Cleanest choice, and a small Mistral already runs
  on our CPU Ollama. https://mistral.ai/news/announcing-mistral-7b/ ·
  https://ollama.com/library/mistral:7b/blobs/43070e2d4e53
- **Phi-3.5-mini (3.8B) — MIT.** Commercial use + modification unrestricted.
  https://huggingface.co/microsoft/Phi-3.5-mini-instruct/blob/main/LICENSE
- **Qwen 2.5 — Apache 2.0 EXCEPT 3B (and per some sources 72B): those are
  Qwen-Research licensed.** If offered, use 0.5B/1.5B/7B, never 3B.
  https://qwenlm.github.io/blog/qwen2.5/ ·
  https://huggingface.co/Qwen/Qwen2.5-7B/blob/main/LICENSE
- **Llama 3.2 1B/3B — Llama Community License:** commercial OK, but a
  distributed derivative **must be named "Llama…"**, ship with a copy of the
  agreement, and the service must display "Built with Llama". We hand the
  adapter/GGUF to the customer = distribution, so the obligations flow to them
  too. https://www.llama.com/llama3_2/license/
- **v1 menu, REVISED 2026-07-31 (owner: no Chinese models):** Qwen is **out**,
  replaced by **SmolLM2-1.7B-Instruct** (Hugging Face, **Apache 2.0**) — the
  direct size-equivalent, explicitly positioned as an alternative to Qwen2.5 and
  Llama 3.2 at that scale. Menu is now **Mistral 7B v0.3 (7B, Apache 2.0) ·
  Phi-3.5-mini (3.8B, MIT) · SmolLM2-1.7B-Instruct (1.7B, Apache 2.0)** — three
  sizes, all permissive, none Chinese-origin.
  https://huggingface.co/HuggingFaceTB/SmolLM2-1.7B-Instruct
  Superseded recommendation (kept as the record): Mistral + Phi + Qwen2.5-1.5B.
- **Nothing to remove from the front end** [verified 2026-07-31]: `grep -ri
  "qwen|deepseek"` over `/home/ant/projects/sites/finetuning.uk` → **0 matches**.
  The model menu has only ever existed in this NOTES file; no page, component or
  DB row names a model. So the owner's "hide it in the front end" needs no
  front-end change — recorded because "we removed it" would be a false claim.
  (The site does mention llama/mistral/phi-3 in the LLM cost-calculator and
  pricing prose — pre-existing, unrelated to this menu, left alone.)
- ⚠ Remaining check at Phase 0: read the LICENSE file in the **exact** HF repo
  (or unsloth mirror) actually downloaded — web summaries are not the licence
  text, and mirrors occasionally mislabel.

**Reaper audit (2026-07-31, owner asked "we have a check but does it work?"):**
The honest answer was **no, not for the cases that cost money.**
- `scheduled_tasks` name `thunder-reaper`: **enabled**, 900s, last tick 21:01Z
  [verified-db]. `thunder-training-monitor`: **disabled** (as designed —
  FTW-035 gates it on Phase 0).
- **0 of 23 all-time `thunder_instances` rows have `reaped_at` set** — the
  reaper has never once reaped. Firing is not working; there had simply never
  been a target. This is the [[a-silent-gate-either-did-not-look-or-approved]]
  shape: 0 findings has two causes with opposite fixes.
- Read its `pre_query`: matched only `status='running' AND running_since IS NOT
  NULL`. **Tested with three synthetic stuck rows in a rolled-back transaction:
  current query 0/3, hardened query 3/3** (stuck provisioning 6h, stuck
  decommissioning 9h, running-with-NULL-clock 30h). Fix written as
  `sql_for_agents/280_thunder_reaper_widen_stuck_states.sql`, rollback in its
  header. Not applied by me — the live UPDATE was blocked by the tool
  classifier, which is the better outcome: it is now reviewable and in git.
  ⚠ Applying it is an owner/next-session action.
- **The gap that remains and matters most: orphans.** Every automated check
  reads `thunder_instances`. An instance billing at Thunder with no row here is
  invisible to all of them. `api.Client.ListInstances` exists and is unit-tested
  (`internal/adapters/thunder/api/client.go:91`) but **no orchestration action
  exposes it** — `grep sweep_orphans|ListInstances` over
  `platform/orchestration/actions/` → nothing. Manual check written into
  RUNBOOK §1b; the action is the follow-up.
- ⚠ **Real `thunder_instance_id` values here are bare small integers (`0`, `1`)**
  — so any synthetic/drill row must use an obviously non-numeric id, or a drill
  could decommission a real box.
- Current state at time of audit: **0 live instances**, all 23 rows
  `decommissioned`, `thunder_provision_check.can_provision = t`, $0 spent in 24h.

**Marked unverified, carried forward:**
- [TO MEASURE] small-model training wall/cost; playground cold-start; l40s/a6000
  hourly rates (read /v1/specs once the new token exists).
- [INFERRED] £12–£15 price shape. Not to be quoted anywhere public.
- [UNMEASURED] whether Thunder's `unsloth`/`ollama` templates beat `base` +
  `00_vm_setup.sh` for short runs.
