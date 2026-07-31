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

**Marked unverified, carried forward:**
- [TO MEASURE] small-model training wall/cost; playground cold-start; l40s/a6000
  hourly rates (read /v1/specs once the new token exists).
- [INFERRED] £12–£15 price shape. Not to be quoted anywhere public.
- [UNMEASURED] whether Thunder's `unsloth`/`ollama` templates beat `base` +
  `00_vm_setup.sh` for short runs.
