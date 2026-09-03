# PLAN — finetuning.uk: a small, real, paid fine-tuning demo

*Status markers: **[verified-db]** queried production this session · **[verified-source]** read the
code/doc this session · **[verified-live]** curled/dug the live internet this session ·
**[INFERRED]** reasoned, not measured · **[TO MEASURE]** must be measured before it is asserted
or priced. Revised after owner direction (31 July): protect the chassis via tools-api + guards;
new Thunder token needed first; GPU stop/start with artefacts persisted by us; GPU-first
playground with named hours; price for GPU, adjust later.*

---

## Context

finetuning.uk should offer a *working* fine-tuning service — a small choice of base models, a real
training run, a real result — behind a payment link of a few pounds covering compute plus a little
of the owner's time. Its purpose is **demonstration**: proof that the thing the site talks about
actually happens here.

This is mostly assembly. Three things already exist and need joining:

1. **finetuning.uk is a live deployed site** — 25 pages, 5 client-side tools, Class A hosting
   (chassis → `gqls/sites` repo → Action → B2 `portfolio-sites` → Cloudflare Worker keyed by
   hostname). Cloudflare zone **ACTIVE**, Worker route **bound** (`curl
   https://finetuning.uk/worker-health` → `Worker is running!`) [verified-live]. Legacy site:
   `sites.content_data` but **no `site_specs` rows**.
2. **A real GPU fine-tuning pipeline exists and has produced a real model.** `thunder-adapter`
   (deployed `v1.0.1215`) is the credential boundary to Thunder Compute; `model-trainer` →
   `training-data-preparer` → `gpu-provisioner` → `training-launcher` is built; cost gate,
   instance table and reaper are live. One 70B QLoRA adapter trained for real: 9h12m, final loss
   0.2669, ~$20 [verified-source]. Dormant since ~21 June 2026.
3. **A complete paid-product pattern exists and has taken real money.** idea.uk's £29 report:
   Stripe Checkout via `net/http` + HMAC webhook verification (no SDK), order lifecycle, capacity
   slots, order expiry, per-order cost recording. First real card payment 2026-07-27
   [verified-source]. Its `billing.go` is the code to copy.

**Owner decisions (this session):** sample dataset *or* customer upload · deliverable includes a
hosted playground, **GPU-served in bounded named hours** (not 24h, not CPU-Ollama to start —
"ollama will be more trouble") · front door protects the chassis, **using tools-api with its
guards** · concierge first, then automate · **a new Thunder Compute token is required before any
GPU use** · Thunder is stop/start — **we persist all artefacts (datasets, adapters, GGUFs, eval
outputs, embeddings) ourselves in B2/Postgres** and reload on demand · price for GPU from the
start; adjust later.

---

## The facts that shape the design

1. **The cluster has no public front door, and must not get one for this.** `kubectl get ingress
   -A` → *No resources found*; no LoadBalancer service exists anywhere [verified-db]. The
   customer-facing service must live outside the cluster.
2. **tools-api is the purpose-built guarded public surface.** It already runs on the island VM
   (Mythic Beasts, `176.126.243.183`) behind a **Cloudflare tunnel that dials out** — zero inbound
   ports — with Caddy forwarding only `/api/v1/tools/*`, its own local Postgres, a spend-capped
   Anthropic key, and `httpguard` inbound protection (token-bucket rate limit, body cap, CORS
   allowlist read from its own DB, `clientip` handling) [verified-source]. Image ships by
   `docker save | ssh load` — no registry credentials on the box.
3. **The box→cluster seam is already built, and pull-only is an owner-ruled design constraint**
   (2026-07-25: *"the cluster never holds a credential to it"*). `pull_report_requests`
   (`platform/orchestration/actions/report_request_pull_action.go`) is a complete idempotent
   checkpointed loop: reads `sites.deploy_config.…{base_url, pull_key}`, `GET
   {base_url}/requests?since=…` with `X-Internal-Key`, INSERTs deduped `site_work_items` rows.
   Results return as a **static sidecar** (`emit_report_status_files`) committed *after* the
   artefact; the island polls it. No callback into the cluster, ever [verified-source].
4. **The training script is already parameterised by base model** (`--base-model`, `--epochs`,
   `--lora-r`, `--max-seq`, `--limit` via argparse) and the scripts bundle in B2 is the deploy
   unit — re-uploading `finetuning/scripts/bundle.tar.gz` **is** the deploy (FTW-031)
   [verified-source].
5. **The durability design already assumes stop/start.** The launcher pre-mints presigned PUT URLs
   (manifest) so the VM streams checkpoints + final adapter to B2; the resume path picks the
   highest checkpoint under `finetuning/checkpoints/<run_id>/` on relaunch (FTW-032/034). What has
   **never been proven in production** is one run reaching `RUN_SH_DONE` with the adapter durable
   in B2 — the exact gate keeping `thunder-training-monitor` disabled (FTW-035). Phase 0 closes it
   cheaply.
6. **finetuning.uk's tool pages are chassis-owned rows, not files.** Every commit touching them is
   a chassis `Rerender:`; hand-written HTML **will be clobbered**. New page content must be
   authored as chassis sections (idea.uk's `p2_01`/`p2_02` pattern) [verified-source].

---

## Architecture

**Front door: a `finetune` route group added to tools-api, served from the island, reached via
`tune.finetuning.uk` as a new hostname on the existing Cloudflare tunnel.** The chassis is
protected structurally: the public internet reaches only the island (which holds no cluster
credentials), and the cluster only ever pulls from it.

```
finetuning.uk         (unchanged) Cloudflare → Worker → B2 static site ── links to ──┐
                                                                                     │
tune.finetuning.uk    CNAME → <tunnel>.cfargotunnel.com  (zone already on Cloudflare)│
   └─ Cloudflare tunnel (outbound-only) → Caddy on island → tools-api container ◄────┘
        ├─ /api/v1/tools/finetune/*   public, httpguard-limited (upload, validate,
        │                             order status, playground chat)
        ├─ /stripe/webhook            signature-verified (billing.go copied from idea.uk)
        └─ /requests  (X-Internal-Key) ← cluster pulls pending jobs; island polls the
                                          published status sidecar for results

cluster:  scheduled task → puller (pull_report_requests pattern) → site_work_items
          → model-trainer → training-data-preparer → gpu-provisioner → training-launcher
          → thunder-adapter → Thunder GPU (stop/start; artefacts → B2) → sidecar committed

playground: on demand, thunder-adapter provisions a CHEAP GPU (l40s/a6000 class, not A100),
          pulls the merged model/GGUF from B2, serves for the named hours, decommissions.
```

Why this beats the earlier "new binary on the idea.uk box" shape (recorded so nobody relitigates):
no nginx vhost, no certbot, no inbound port anywhere; the guards are already written and tested;
idea.uk's live money path is untouched; and the finetuning.uk zone being on Cloudflare makes the
CNAME trivial. **Posture change to record honestly:** Stripe keys (checkout secret + webhook
secret) will live on the island, which was designed to hold *no* production credentials. They are
money credentials, not cluster credentials — the cluster remains unreachable — but this is a
deliberate widening of what the island is, and it goes in the concept register entry.

**State:** orders and playground bookings live in the island's own Postgres (it already runs one
for the gauntlet), not files — one step up from idea.uk's `orders.json`, zero new infrastructure.

---

## The product

**Sold:** one small-model fine-tune, fixed price, a few pounds.

**Customer supplies** either a prepared sample dataset (3–4 curated options) or their own upload
(JSONL message pairs or two-column CSV), capped ~500 rows / 2MB. **Validation is free and runs
before payment** — pure parsing, no LLM call (idea.uk's lesson: the free unauthenticated endpoint
is the only unbounded-spend risk).

**Delivered:**
- **Before/after comparison** — held-out prompts through base vs tuned model, side by side. The
  three-level eval pipeline that produces exactly this is built (FTW-019), ~$1 / ~5 min.
- **The LoRA adapter + GGUF** to download, with hyperparameters and the loss curve.
- **A GPU playground in named hours** — a bounded session (e.g. 2 hours) on a cheap GPU instance,
  at a time the customer names (or we name — offered slots). The instance is provisioned on
  demand, loads the model from B2, serves chat through the tools-api route, and is decommissioned
  at the end of the window. Nothing runs overnight. The page states the limitations plainly: small
  models, a demo, not production inference.

**Explicitly not offered, and said so on the page:** production hosting, large models, quality
guarantees, retention beyond the stated window (`bugs_open/043` — no overstating).

---

## Economics — measure, then price (price must cover the GPU window)

| Fact | Value | Status |
|---|---|---|
| Thunder A100XL 80GB | **$1.80/hr** | [verified-db] `thunder_config` |
| Daily cap / max concurrent / hard uptime | **$30/day · 2 · 18h** | [verified-db] |
| Cheaper GPUs for inference | `l40`, `l40s` 48GB, `a6000` (prototyping mode) | [verified-source] |
| Hourly rate of those | **unknown — read `GET /v1/specs` once the new token exists** | **[TO MEASURE]** |
| Cold-start environment build | ~25 min / ~$0.85 | [verified-source] FTW-015 |
| The one real (70B) run | ~$20, 9h12m | [verified-source] FTW-013 |
| A small-model training run | **unknown** | **[TO MEASURE]** |
| Playground cold-start (provision → model loaded → first token) | **unknown — sets how "named hours" must be booked** | **[TO MEASURE]** |

Plausible shape [INFERRED]: training 15–35 min ⇒ £0.35–£0.90 GPU; eval ~£0.80; a 2-hour
playground window on a cheaper GPU perhaps £1–£3. A £12–£15 price would carry all of it with
margin; £5 would not carry the playground. **Set the number after Phase 0's measurements** —
"adjust prices later" is the owner's stated posture, so start honest and roomy rather than tight.

Cost levers worth testing in Phase 0: Thunder's pre-built `unsloth` and `ollama` templates
(the eval work used the unsloth template successfully; for short runs the ~25 min environment
build is the dominant cost, and for the playground the `ollama` template may make "load GGUF and
serve" nearly instant). Snapshots break even ~18 runs/month — revisit only if volume warrants.

---

## Phasing

### Phase −1 — human action: the Thunder token (blocks everything)

The owner mints a **new Thunder Compute API token**. Then: update the secret the thunder-adapter
deployment reads (`THUNDER_COMPUTE_API_KEY`), restart the adapter pod, and verify with a
read-only call — list instances or `GET /v1/specs` — before anything provisions. While there,
capture the **current hourly rates per GPU type from `/specs`** and correct `thunder_config`
(`default_hourly_rate_usd`, `estimated_new_run_cost_usd` — the $20 estimate is sized for 70B runs
and will over-refuse small ones via the cost gate if left).

### Phase 0 — one small-model run + one playground rehearsal, measured

Also the cheapest way to close the flywheel's own blocking gate (`RUN_SH_DONE ⟹ durable`,
FTW-032/035): proving it with a ~20-min 1B run costs ~$1 instead of $20.

1. Parameterise `run.sh` (it hardcodes the 70B default; add `--base-model` pass-through;
   `--save-steps 0` for short runs — the **manifest is still required**, it carries the final
   adapter's presigned PUT URL). Re-tar and re-upload `finetuning/scripts/bundle.tar.gz`.
   ⚠ Editing a script without re-tarring deploys nothing (byte-identical md5 trap).
2. Pick candidate small models and **verify each licence in writing** — e.g. Llama 3.2 1B/3B
   (community terms, naming obligations), Qwen 2.5 (Apache 2.0 for some sizes only), **Mistral 7B
   (a natural fit — a small Mistral already runs on our CPU Ollama today)**, Phi-3.5-mini. Nothing
   below 70B has ever been trained here. Do not assert a licence from memory.
3. Run one small fine-tune end to end. Test the `unsloth` template vs `base`+`00_vm_setup.sh`.
   Record wall clock per stage, $ cost, `RUN_SH_DONE` reached, `adapter.tar.gz` genuinely in B2.
4. Merge + convert to GGUF **on the GPU box**, upload to B2 (this is the "we store everything
   ourselves" half of stop/start).
5. **Playground rehearsal:** provision a cheap GPU (`l40s` or `a6000`), pull the GGUF from B2,
   serve it (Thunder's `ollama` template is the first thing to try), measure provision→first-token
   time and tokens/sec, decommission. This decides how playground bookings work.
6. **Then** enable `thunder-training-monitor` (RUNBOOK step 6) — its DONE_OK path decommissions
   the box, so it stays disabled until step 3 proves durability.

**Critical files:** `docs/.../finetuning/working/scripts/run.sh`, `02_train_llama_3_3_70b.py`,
`internal/adapters/thunder/api/types.go` (template/GPU constants),
`internal/adapters/thunder/provision_action.go`.

### Phase 1 — the page and the payment link, live in days (concierge)

1. **Author the offer page as chassis sections** on finetuning.uk — follow
   `idea_uk_vm_site/sql/p2_01_tool_entry_forms.sql`: `content_components` INSERT
   (`component_level='section'`, `render_mode='template'`, `template_closed=true`), place via
   `site_plan_sections` onto an **existing** page (no new pages — re-plan landmine,
   `bugs_open/001`), direct-fire the rerender, confirm sections **rendered** not carried, backfill
   `pages.sections`, **lock** the authored sections.
2. Read `p2_02_fix_form_component_shape.sql` before authoring anything with JS: set `function` =
   component name (it names `/tools/assets/{function}.js`), and know that raw SQL bypasses
   `separateInlineJS` so `js_content` stays empty and no JS asset ships. Target shape:
   `js_len>0, src_ref=t, raw_inline=f`.
3. **Stripe Payment Link** (no code) at a provisional GPU-covering price. Orders arrive by
   email/form; the owner runs them by hand with the Phase 0 scripts; playground hours arranged by
   reply. Cap at ~10 orders; state turnaround honestly.
4. Legal: extend terms/privacy for customer training data — retention, deletion, base-model
   licence per model offered, playground hours.

### Phase 2 — automate: tools-api grows the finetune surface, the cluster pulls

1. **New route group in `internal/tools-api`**: upload + validate (free, no LLM), order create,
   order status, playground booking. Reuse `httpguard` limiter/body-cap/CORS and `clientip`
   exactly as the gauntlet routes do. Orders table in the island's Postgres.
2. **Billing:** copy idea.uk's `billing.go` into the service (do not import — do not touch the
   live money path; register the duplication in the concept register with extraction to
   `platform/billing` as the named follow-up). Keep the two proven guards: the `FakeProvider`
   type-check on any bypass path (`bugs_closed/089`) and real-peer client identity
   (`bugs_closed/090`) — noting that behind the tunnel, client IP **must** come from
   `CF-Connecting-IP`, and `count(DISTINCT)` is the check that catches a constant identity.
   **Stripe verifies a signature over the raw body — Caddy must not rewrite it. Ever.**
3. **Edge:** add `tune.finetuning.uk` to the tunnel config + a CNAME in the (already-Cloudflare)
   finetuning.uk zone; extend Caddy to forward the new route group and `/stripe/webhook`. Deploy
   the image by the existing `docker save | ssh load` path.
4. **Wire the pull:** a `finetune_request` item type and a puller modelled line-for-line on
   `pull_report_requests` + `ndjson_feed.go` (checkpoint, `item_key` dedup, no PII in the payload
   — dataset reference only, never the customer's email). Config on the finetuning.uk `sites` row
   (`deploy_config`, pull_key ≥24 chars as a psql var, never committed — follow
   `sql_for_agents/208`). Seed the scheduled task **disabled**; enable deliberately.
5. **Results path:** reuse `emit_report_status_files` — status sidecar committed *after* the
   artefact; the island polls it and flips the order to ready/failed.
6. **Dataset transport:** island uploads the validated dataset to B2 under a per-order prefix;
   the cluster presigns it for the GPU box exactly as `training-data-preparer` does. Bytes never
   traverse Kafka; one order per GPU box = physical tenant isolation.
7. **Capacity:** gate on `thunder_provision_check.can_provision` (do not invent a second cost
   gate); cap concurrent runs at `max_concurrent_instances` (2); queue with an honest estimate.
   Playground bookings must respect the same gate — a booked window is a reserved instance-hour.

### Phase 3 — the playground, automated

Booking flow on the app (customer names hours from offered slots) → at slot start the cluster
(pulled work item, `playground_session` type) provisions the cheap GPU, loads the order's GGUF
from B2, records the endpoint → the island proxies chat for the window through the existing
guarded route → at slot end, decommission; the reaper's hard-uptime cap is the backstop against a
forgotten instance. If Phase 0's rehearsal shows cold-start is awkward (say >10 min), offer
"we-name-the-hours" slots only, batched back-to-back on one instance. Later cost optimisation,
explicitly deferred: CPU Ollama serving with stated limitations.

---

## Landmines to obey (each already cost someone real time)

- **Never edit finetuning.uk's `<head>`/`header` chrome** — `Document Head` is shared by 9 of 14
  live domains; a per-site value ships to eight other sites. Use `site_specs` + a gated template.
- **A roll is not evidence your fix shipped.** Grep a string your change *added* plus a positive
  control in the same exec — for the chassis image, the adapter image, and the island container
  alike.
- **`b2 sync --skip-newer` silently skips a file whose bucket copy is newer** — a revert can fail
  to propagate while the Action reports green. Verify at the origin with a cache-buster.
- **Commit by explicit pathspec** — this repo and `gqls/sites` are shared with concurrent
  sessions and automated rerenders. Never `git add -A`.
- **`/tools/ai-agent-roi-estimator.html` is `needs_rebuild` in the DB but live** — a pre-existing
  divergence on this site. Don't trigger a blanket rebuild as a side effect.
- **Do not use export `a8484922…`** — records `rows_exported=1957`, holds zero rows. `count(*)`
  before launching anything against an export.
- **`kubectl scale`-style imperative fixes are undone by the next `apply -k`** — capacity changes
  for any new Ollama/eval pod go in the overlay.

## Process obligations

- **Council gate** for the puller action, the tools-api route group, and any thunder-adapter
  change (`097_TRIGger` script; budget ~30 min; `Council-Submitted: <corr>` if committing before
  the verdict; never `Council-Reviewed:` unread).
- **Concept register, same commit as each seam:** the `finetune_request` (+`playground_session`)
  item types, the puller, the copied billing lineage, **and the island's widened posture (Stripe
  keys)** — each with its landmine and open review question.
- **The standing five** in `docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/`,
  created at the start. The RUNBOOK earns its keep immediately — every tunnel, Caddy, B2 and psql
  command here has a gotcha attached.

## Verification — end to end, at the artefact

1. **Phase −1:** a read-only Thunder API call succeeds with the new token *from the adapter pod*,
   not a laptop.
2. **Phase 0:** `RUN_SH_DONE` in `train.log` **and** `b2 ls` shows `adapter.tar.gz` non-zero;
   `model_lifecycle.training_runs` → `complete` with real cost/runtime; `thunder_instances` →
   `decommissioned`. Playground rehearsal: measured provision→first-token and tok/s written into
   the RUNBOOK.
3. **Phase 1:** served page (no `-L`, status code recorded per URL) greps for the new section's
   marker.
4. **Phase 2:** one real order traced through: work item with expected `item_key` → training run
   complete → sidecar committed after the artefact → island flips the order → **a real card
   payment** with Stripe's signed webhook landing. Then attack it: a bypass attempt on any free
   path, and a forged `CF-Connecting-IP`/`X-Forwarded-For` to confirm rate-limit identity is the
   real peer.
5. **The idea.uk question:** it is not done until someone who is not the owner has paid and
   received a fine-tune.

## Open owner calls

1. **Confirm the island placement** — "maybe using the tools-api" is read here as: extend
   tools-api on the island, `tune.finetuning.uk` through the tunnel. The alternative (same code as
   a second container on the idea.uk box) exists if you'd rather keep payments off the island.
2. **Price** — set after Phase 0; must cover the GPU playground window. £12–£15 is the shape I'd
   expect if the playground is included; adjust later as you said.
3. **Playground booking shape** — customer-named hours vs offered slots; Phase 0's cold-start
   measurement decides.
4. **Sample datasets** — which 3–4, and are they ours to publish? They are the shop window.

## Phase P — the playground TOOL (owner decision 2026-09-03: BOTH a public demo and booked hours)

**Owner, 2026-09-03:** *"both public demo and booked hours - what is that public demo going to
cost me in llm fees?"* Decision recorded; the cost basis is below and in NOTES.

**Shape (from § "A GPU playground in named hours", now split in two):**
- ~~**Public demo, always on, CPU, in-cluster.**~~ **CORRECTED 2026-09-03 12:40Z: NOT in-cluster.** tools-api runs on the island VM (`toolsapisuk.vs.mythic-beasts.com`, 1 vCPU / 1 GB, no cluster callback by design), which cannot reach the in-cluster Ollama and cannot host the model itself. **Placement is an OWNER DECISION** (NOTES 2026-09-03 12:40Z): (a) a small dedicated CPU VPS running Ollama, reached from the island with a shared key — lane's recommendation, ~£10–20/month `[UNVERIFIED price]`; (b) resize the island and run Ollama beside tools-api; (c) expose the cluster's Ollama to the island — contradicts the isolation posture, needs a ruling. Prerequisite either way: finetuning.uk added to the island's `sites` allowlist. The paragraph below describes the model and speed, which stand.
- **Public demo, always on, CPU.** The Phase 0 fine-tune
  (`SmolLM2-1.7B-Instruct`, Apache 2.0, GGUF q4_k_m 1.06 GB in B2) loaded into the existing
  `ollama-adapter` pod (requests 2 CPU / 20 Gi, limits 8 CPU; idle today at 1m CPU / 327 Mi).
  **LLM API fees: £0** — no Anthropic/OpenAI call anywhere in the demo path; the marginal cost
  is CPU time on a node already paid for, bounded by the pod's limit. `[UNMEASURED]` tokens/s of
  a 1.7B q4 model on this node's CPU; measure before deciding the demo's reply-length cap.
- **Booked hours, GPU, on demand.** The Phase 0 path (a6000 at **$0.35/hr real, billed per
  minute**; ≈$0.75 per 2-hour window incl. warm-up; dispatch→first token 3 m 23 s), provisioned
  per booking through the thunder actions that already exist for training runs and
  decommissioned after. Never used for the public demo: an always-on a6000 is ≈$8.40/day.
- **One route for both:** `tools-api` `/api/v1/tools/playground/chat` behind `httpguard`
  (per-IP token bucket, body cap, CORS), proxying to the demo model by default and to the booked
  box during a session. **One widget:** fork the library's `chat-input-box` (requires-backend) onto
  `/playground.html`; `deploy_config.capabilities += backend` on the site row so TL-043's gate
  admits it.

**Order:** (1) demo model into the in-cluster ollama, measure tok/s — **DONE 2026-09-03 12:08Z: 14.3 tok/s CPU, 2.4 s cold load (NOTES)**; (2) tools-api route +
tests, council; (3) capability on the site row, fork the widget through the tool pipeline, rebuild
`/playground.html` through the framework (never hand-authored); (4) booked-hour provisioning as
a workflow, reusing the training-run actions; (5) the booking → session handoff. Each step
verified at the artefact (curl the route; chat on the served page).

## DIRECTION (owner, 2026-09-03, 11:45Z) — the site is the tool; details later

Owner, verbatim: *"As a whole, I'd like the finetuning site to be very much focused around this
tool. We can still have the other tools, but much of the "what else we do as a company" should
now move to leopardess consulting or other "me" sites. For finetuning.uk I'd like this tool
shown prominently on the home page and I want in the future example after real example of what
we've done and before and after examples. And I'd like to host those same models so they can try
them (at maybe a couple of pounds for an hour or something that covers our costs say 5x) We can
talk details later."*

What this changes in the plan, as decisions (not yet scheduled; "details later"):
- **Phase P's scope grows from "a chat on /playground.html" to "the site's centre".** The
  homepage hero and first sections present the playground; company-general sections
  (departments-grid, the consultancy-style case-studies-grid) are candidates to move to
  leopardessconsulting.co.uk. Moving copy is a re-plan of both sites through the framework, and
  the leopardess lane must be told before anything moves (a live session exists).
- **A catalogue of real examples, each with a before/after pair and a hostable model.** Data
  model to design: an example = the case (what the business was, what changed), the before/after
  text, the model artefact in B2, and a bookable-hour price. This is the booked-hours mechanism
  generalised from "the customer's own model" to "any model in the catalogue".
- **Pricing posture: cover cost ×5.** Measured cost is a6000 $0.35/hr real + warm-up (≈$0.75
  per 2-hour window, Phase 0 invoice). ×5 ≈ $1.75–$1.90/hr ≈ **£1.35–£1.45/hr**, so "a couple of
  pounds an hour" clears the posture with room. `[NOT DECIDED]` the number; set it in the details
  conversation, after the CPU-demo tok/s measurement says what the free tier can carry.
- Existing tools stay. Their pages are unaffected by this direction.

### Direction addendum (owner, 2026-09-03, 11:55Z) — a big GPU for the examples; third-party models

Owner, verbatim: *"the gpu for the examples could be a big one so it feels snappy - this might
change our pricing estimate. Eventually we may have third parties submitting their models and they
might have a page of their own and we'd show examples or their results similar to how we'd show
our own."*

- **GPU class for hosted examples: a big one.** Measured rates (vendor invoice 2026-08-18, per
  minute): a6000 **$0.35/hr**, a100xl **$1.09/hr**; `thunder_config` lists A100XL 80GB at $1.80/hr
  (the booked ESTIMATE that ran 5.1× over the invoice). What "snappy" needs depends on the model:
  Phase 0's 1.7B on an a6000 was already **0.36 s warm first token, 139 tok/s**; a big GPU buys
  **larger models (7B–70B) and concurrency**, not speed on a 1.7B. **Pricing at ×5 on an A100 ≈
  $5.45/hr ≈ £4/hr**, so "a couple of pounds an hour" no longer clears ×5 on that class; either
  the price rises, the multiple falls, or the class is chosen per model size. `[NOT DECIDED]`;
  details conversation. `[TO MEASURE]` A100 cold start (provision → first token) — Phase 0 only
  measured the a6000 (3 m 23 s).
- **Third-party models, each with a page of their own**, showing their examples/results the way
  we show ours. Implication for the catalogue design: an "example" is not ours by definition —
  it has an OWNER (us or a submitter), a model artefact, a page, before/after pairs, results, and
  a bookable hour. So the page type is "model page", ours are the first entries, and submission
  is a later flow (what a submitter provides, how the artefact is checked before hosting, licence
  and liability, who sets the price). Nothing designed yet; recorded so the catalogue is not
  built ours-only.

### Direction, DECIDED (owner, 2026-09-03, evening) — the four answers

Owner, verbatim: *"1: a with b as fallback, have you seen how webdesign.uk does it? 2: please confer with
the prompts lane 3: I'll leave the rotation until I next do it naturally. I am the only one reading this
transcript. 4: increase the price / give them the choice perhaps. It is, in part, an education to see the
differences in speed. examples catalogue, yes, let's think about the shape now. They need input
mechanisms, accounts, and so on, and the models will need removal mechanisms and terms and conditions
agreement clicks and ways to stop cheating etc. leopardess: tell leopardess as it already has alot of it.
we could move it to a new organisational site's brief instead - I am asking the domain thread"*

- **Phase P step 3, the widget: path (a) DECIDED** — generated through tool-generator from a brief that
  carries the route contract; hand-written (b, the gripper/gauntlet/webdesign.uk shape) only as the
  fallback if the generator cannot produce a working streaming client. webdesign.uk's own box is the
  library component placed by hand (`created_from='manual'`, no `add_tool` item) with an Option-A
  backend, so it is a precedent for (b), not (a).
- **Pricing: raise the price and/or offer the GPU class as a CHOICE.** The speed difference between
  classes is part of the education. So the booking shape gains a class selector (small GPU ~£1.40/hr at
  ×5, big GPU ~£4/hr at ×5), and the number is still `[NOT DECIDED]`.
- **Examples catalogue: design the shape NOW** (DESIGN doc in this dir). Named requirements: input
  mechanisms for submitters, accounts, a removal mechanism for a hosted model, a terms-and-conditions
  agreement click, ways to stop cheating.
- **Copy that leaves finetuning.uk:** leopardess told (`docs/leopardessconsulting/CONTRIB_2026-09-03_…`);
  the alternative, a NEW organisational site's brief, is with the domain thread (the owner's own ask).
  Nothing moves from this lane until one of those lands.
- SMTP password rotation: deferred by the owner to the next natural rotation.
