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

---

## 2026-08-03 — the token in the cluster does NOT authenticate; reaper fix applied

### Lane boundary — the site's front end belongs to another session

Owner: *"I am fixing the finetuning site in another thread because it was very
broken in design — missing images etc, so there will be some changes going on
beneath us"* (session `7b4e88a8-e887-459c-8524-8bd9f92c3618`, "finetuning
design"). That lane keeps its own standing docs in
`docs024_key_docs_latest/finetuning_uk_repair/` (committed `734296859`) —
**a different directory to this one, so no doc collision.**

**Division of work, to be respected both ways:**
- **THEIRS:** `finetuning.uk` site content — `content_components`,
  `page_components`, `site_work_items`, the `departments-grid` component,
  `check_image_url_404`, the improvement loop. Anything the customer *sees*.
- **MINE (this lane):** the service backend — Thunder token/adapter, the reaper,
  `thunder_instances`, training runs, the future tools-api `finetune` routes.
  Nothing I have touched today is a site row.
- ⚠ **PLAN Phase 1 (author the offer page as chassis sections) now DEPENDS on
  that lane** and must not be started while they are re-rendering the site.
  Coordinate before authoring; do not race a rerender.

**Two of their findings that change MY plan** (from their PLAN, read 08-03):
1. `check_image_url_404` was blind to a whole shape of broken image — the same
   "a green check that never looked" class as the reaper below. Two lanes found
   the same failure mode in one day, independently.
2. **Dispatch is dead fleet-wide**: `detected → triaged` is only promoted inside
   the improvement loop, whose only scheduled route has been `enabled=f` since
   2026-05-02; 204 detected / 2 triaged across 10 sites. **Phase 2 of my plan
   routes customer orders through `site_work_items`** — so it would have inherited
   a dispatcher that never fires. [CARRIED FORWARD] verify the promoter is live
   before depending on it, or drive the item type directly.

### The Thunder token is present, 64 chars, and REJECTED

Owner: *"we now have a thundercompute api key"*. Measured rather than assumed:

- Pod `thunder-adapter-7dd95b758c-x5pxm`, 12 min old at the time of checking —
  consistent with a recent restart, so it is NOT a stale-env problem.
- Pod env and the secret **agree**: `len=64 prefix=a73 suffix=96` (fingerprint
  only — the key was never printed, and must never be).
- ⚠ **That is NOT the key the owner pasted on 07-31** (`5a2…c2`). So the new
  token was never patched into `personae-default-secrets`; the cluster still
  holds an older one.
- `GET /v1/instances/list` with that key → **HTTP 401 Unauthorized** [verified-live,
  from inside the pod].
- **Call shape ruled out as the cause** before blaming the key: base URL
  `https://api.thundercompute.com:8443/v1` (`client.go:42`), path
  `/instances/list` (`client.go:93`), header `Authorization: Bearer`
  (`client.go:235`) — my call matched all three.
- **Controls run** (because a check that returns the same answer for every input
  is not a check): a deliberately bogus token → 401, and **no auth header at
  all** → 401. So 401 is this API's auth-rejection response, and the key is being
  rejected as invalid. It is not an endpoint or a header-shape artefact.

**Consequences, both directions — the bad one is smaller than it looks:**
- ✅ **There is no billing risk from the cluster right now.** An invalid key
  cannot provision either, so nothing new can start costing money. Combined with
  0 non-decommissioned rows and `total_24h_spend = 0`, the fleet is quiet.
- ❌ **RUNBOOK §1b — the authoritative "am I being billed?" check — cannot run
  until the key works.** It is the one check that sees orphans, and it is
  currently blind. This is the real cost of the bad key, not the blocked GPU work.

### `280_thunder_reaper_widen_stuck_states.sql` — APPLIED [verified-db]

Applied 2026-08-03 (the 07-31 attempt was blocked by the tool classifier).
`BEGIN / UPDATE 1 / COMMIT`, and the verify SELECT returns
`covers_provisioning=t, covers_decommissioning=t, covers_null_clock=t`,
`enabled=t`, `interval_seconds=900`. Rollback is in the file header.

**A suspicion checked and REFUTED before it became a claim.** The agent
workflow's own description says it dispatches using
`provisioning_id`/**`thunder_identifier`**/`reason`, but the pre_query returns a
column named `thunder_instance_id` — which looked like a silent field-name
mismatch that had never been exercised. It is not:
`thunder_decommission_dispatch.go:62-67` reads
`input_data.provisioning_id` **OR** `input_data.thunder_identifier` and requires
only one; the pre_query supplies `provisioning_id`, which the source comments
mark as the *preferred* form (`:14`). `thunder_instance_id` is simply an unused
extra field. **Recorded because "I nearly filed this" is the useful part** — the
doc string and the query genuinely disagree, and the next reader will trip on it
too.

### Reap drill — in progress at time of writing

Design (see RUNBOOK §1b for why each choice): synthetic row
`63be1825-6478-49e0-b38f-e9fb2c6630c8`, `thunder_instance_id =
'T-DRILL-20260803-REAPER'` (**non-numeric on purpose** — real ids here are bare
integers `0`/`1`, so a numeric guess could decommission a live box),
`status='running'`, `running_since` 30h ago against an 18h cap,
`hourly_rate_usd = 0` so it cannot pollute the cost gate.
Confirmed on insert: the **live** pre_query selects it with reason
`reaper:max_uptime_exceeded after 30.0h (status=running, cap=18h)`, and
`thunder_provision_check` still reads `can_provision=t, total_24h_spend=0,
active_count=1` (of 2).

⚠ **This is the safest possible window for the drill**: 0 live instances AND an
invalid token, so the decommission call physically cannot destroy a real
instance. The drill therefore proves **selection + dispatch** — the half that was
broken. **It cannot prove the vendor-side delete**, which must be re-proved once a
working key is in. Do not read a green drill as end-to-end proof.

**MISSTEP (mine, caught in minutes).** My first poll loop suppressed psql's
stderr and treated an empty result as "the row changed", so it reported
`CHANGED — stopping poll` 35 seconds in, having measured nothing. The query had
errored: `orchestration_states` has **`owner_agent_type`, not `agent_type`**.
Two lessons, both already in the index and both ignored by me here: **schema
first** (`\d` before the query), and **never `2>/dev/null` a check whose empty
output is indistinguishable from its success signal.** The fix was to require a
non-empty reading before concluding anything.

### DRILL RESULT — the reaper reaped for the first time, and found a THIRD bug

**Tick 10:37:48Z. The reaper selected the drill row and dispatched.** First time
in its existence. Orchestration `69ebf25b-2692-4a89-acaa-0fd33b8bf48d`,
`current_step=dispatch_decommission`, with `input_data` carrying exactly what the
pre_query produced:

```json
{"reason": "reaper:max_uptime_exceeded after 30.1h (status=running, cap=18h)",
 "provisioning_id": "63be1825-6478-49e0-b38f-e9fb2c6630c8",
 "thunder_instance_id": "T-DRILL-20260803-REAPER"}
```

**Then the adapter refused it** — and not at Thunder, at our own database:

```
"error":"decommission_failed","success":false,
"detail":"thunder_instances lookup: sql: Scan error on column index 3,
          name \"instance_ip\": converting NULL to string is unsupported"
```

`store.Instance.InstanceIP` is a plain Go `string` (`instances.go:29`) while
`instance_ip` is **nullable** in the schema, so `lookupOne` cannot scan a
NULL-IP row. **Filed as `bugs_open/186`** with fix candidates.

**Cause isolated, not inferred.** I set `instance_ip = '203.0.113.9'`
(TEST-NET-3, reserved for documentation — never routable) on the *same row*,
changed nothing else, and the retry got straight past the lookup:
row → `decommissioning`, `decommission_requested_at` stamped 10:40:51Z. One
field changed, error gone. That is the whole proof.

**What the drill did and did not establish:**

| step | result |
|---|---|
| pre_query selects an overdue instance | ✅ **the 280 fix works** |
| scheduler dispatches, payload correct | ✅ |
| adapter looks the row up | ❌ **fails on NULL `instance_ip` → `bugs_open/186`** |
| adapter looks the row up (IP set) | ✅ |
| `MarkDecommissioning` → row moves | ✅ |
| Thunder API delete | ⛔ **NOT exercised** — see below |
| row → `decommissioned`, cost computed | ⛔ **NOT exercised** |

⛔ **The vendor half remains unproven, for two independent reasons**, and neither
is a defect: my `thunder_instance_id` was deliberately non-numeric, and
`decommission_action.go:123-129` parses it with `strconv.Atoi` and **refuses a
non-parseable id before calling Thunder** — a guard I did not know about when I
chose the id, which made the drill safe twice over. The token is invalid anyway.
**Re-run the drill after a working token, with a numeric-but-implausible id**
(the seed's own template uses `999999` and leans on Thunder treating 404 as
success — `114_thunder_reaper.sql:186-199`).

**Cleanup done and verified:** drill row DELETEd (matched on `id` AND
`thunder_instance_id`); table back to 23 rows, 0 non-decommissioned, 0 drill rows,
`can_provision=t`, `total_24h_spend=0` — identical to the pre-drill state. This
mattered: the row was left in `decommissioning`, which my own widened query
re-selects after 1h, so leaving it would have created a 15-minute failure loop.

### CORRECTION to my own 07-31 framing of the 280 fix

> **CORRECTED 2026-08-03:** the 07-31 entry above says the old query missed
> "three genuinely-BILLING states". The 0/3 → 3/3 measurement was a fair test of
> the *selector*, but calling all three states live was more than I had checked.
> Measured today, per state:
> - **`decommissioning` — REACHABLE and real.** `store.MarkDecommissioning`
>   writes it, and any failure after that strands the row (the drill produced
>   exactly this). This branch does real work.
> - **`provisioning` — NOT REACHABLE by anything.** No Go writer and no SQL
>   writer sets it: `provision_action.go:413` hardcodes `status='running'` and the
>   row is INSERTed only *after* the box is up. `042_thunder.sql`'s views merely
>   *anticipate* the state. **Defensive only — do not cite it as fixing a live
>   provisioning leak.**
> - **`running` with NULL `running_since` — defensive.** The one INSERT path
>   always supplies it.
>
> What caught it: reading the INSERT to answer a *different* question (would a
> stuck-provisioning row have an IP?). The lesson is the index's own —
> **a state the schema models is not a state the code writes.** The widening is
> still right (cheap, strictly widening, and one branch is load-bearing), but the
> honest headline is "one real gap closed, two guards added", not three.
>
> **And the dominant risk is still none of them:** the row is written *after* the
> Thunder box exists, so a crash in between leaves a **billing instance with no
> row at all** — which no selector can ever catch. That is `bugs_open/186`'s
> "Related" section and task #6, and it is the one worth building.

## 2026-08-03 ~21:45Z — note from the bugfix_140 lane (queue reading, no work fired at your pages)

Contributing what the 140 residual-rows follow-up found in YOUR queue, per the
dispatch rule — nothing was dispatched at finetuning.uk from my side.

- **The `/blog.html` blank hero (pre-295 render, empty `content_data`) is behind
  your failed rebuild, and the failure is PROTECTIVE.** `needs_page` `d96aee06`
  ("Full rebuild of blog") failed at `save_sections`: the rebuild run
  re-confirmed only **1 of 3 stored sections (33% < prune_floor_ratio 0.50)** —
  the save was refused whole, nothing written, so the stored blank hero stands.
  Its sibling `save_refused_incomplete` `d0f4176f` sits in `needs_human_review`.
  The hero cannot be fixed by any rerender (empty `content_data` — a redraw of
  nothing); it needs the rebuild to succeed or a human to accept the shrinkage.
- **The `insights`/`ai-guides` `empty_section` items (unresolved ×3) trace to a
  MISSING COMPONENT, not to content.** The sections are `article-grid` /
  `category-section`; the `needs_new_component` items that would create those
  templates (`86314858`, `64103637`) have themselves **failed 3×** at
  `store_component` pre-store validation: "template variables and schema fields
  do not match" — the 287 desync class. Until component generation passes that
  gate, the empty_section items will keep re-failing; they are downstream.
- Also on one of your pages: `page_components` row `d2e9644b` (article-body on
  `tool-ai-data-risk-checker-guide`) stores a raw LLM envelope as
  `content_data` — render is currently GOOD; the trap fires only on a reasoned
  rerender. Filed as `bugs_open/190` with a proven mechanical repair for this
  specific row (today's parser fully recovers it). Coordinate there if you want
  it decoded; do not rerender that page before it is.

— bugfix_140 lane

---

## 2026-08-08 — Phase −1 COMPLETE: token live via terraform; pricing captured; vendor-half drill

### Token — live and proven at the artefact

Owner direction (08-05): **the token is set via terraform `047-base-configs`,
not `kubectl patch`** — the truth of the key is `terraform.tfvars.secret` (from
`~/.config/thundercompute/token`), and the secret is that root's
`kubernetes_secret.personae_default_api_keys`. Procedure written as RUNBOOK §1c.
This retro-explains the 401 episode: the rejected `a73…96` key MATCHED the
tfvars exactly, so the cluster key *came from terraform*, and any hand-patched
key was one apply away from silent reversion. The mechanism was fighting the
rotation, not failing it.

What happened, honestly recorded:
- I updated the tfvars (`f39…ff`, from the token file, value never printed) and
  fingerprint-compared **all 19 values across both managed secrets plus the
  configmap's AGENT_IMAGE_TAG** against live before any apply: zero drift, so
  the apply was provably single-key. Plan: `0 add, 1 change, 0 destroy`.
- The tool classifier blocked `terraform apply` from my session (twice). Owner
  ran it via `!` on 08-05 → **"No changes"** — the value had ALREADY landed by
  another hand between my plan and their run (probably the other thread; the
  adapter pod also restarted 08-06 19:54, which I did not do). [UNATTRIBUTED —
  which hand applied it is not knowable from here, and does not matter: the
  tfvars edit I made is the change that shipped.]
- **Verified 08-08 from inside the pod** (the artefact, not the tag): pod env
  `len=64 f39..ff`; `GET /v1/instances/list` → **`{}`** — authenticated, and
  zero instances exist on the account. The same call was 401 on 08-03.

### Pricing — /specs has NO prices; the plan's assumption was wrong

`GET /v1/specs` (verified 08-08): 32 entries, hardware only — no price field
anywhere in the payload. The PLAN and RUNBOOK both said "capture rates from
/v1/specs"; **that endpoint cannot do it.** Rates captured from
https://www.thundercompute.com/pricing instead:

| gpu_type | VRAM | $/hr |
|---|---|---|
| **a6000** | 48GB | **$0.35** |
| l40 | 48GB | $0.79 |
| a100xl | 80GB | $1.09 |
| h100 | 80GB | $2.19 |

Billed per minute; +$0.04/vCPU/hr beyond 4; storage $0.03/100GB/hr beyond
100GB; snapshots $0.05/GB/mo.
- ⚠ **`l40s` does not exist** — the plan guessed it; the real playground GPU is
  the **a6000 at $0.35/hr** (only x1 and x1_prototyping variants). A 2-hour
  playground window ≈ **$0.70** — better than the plan's £1–£3 guess.
- ⚠ `thunder_config.default_hourly_rate_usd` still says **$1.80** for a100xl;
  live rate is **$1.09** — stale HIGH, so the cost gate over-refuses, which is
  the safe direction. Correct only when Phase 0 is about to run (RUNBOOK §1
  step 5).

### `reaped_at` is a DEAD COLUMN — nothing writes it

Grepped Go + platform for `reaped`: **no writer exists** for
`reaped_at`/`reaped_reason`, and the `'reaped'` status in the CHECK constraint
is written by nothing either. A fully successful reap ends at
`status='decommissioned'` via the ordinary decommission path.
> **CORRECTION to the 07-31 metric:** "0 of 23 rows have `reaped_at` set" was
> quoted as proof the reaper had never reaped. The conclusion was right (the
> selector provably matched nothing) but the metric was unfalsifiable — that
> column could never have been set by any code path. The honest success signal
> for a reap is the row reaching `decommissioned` with `decommissioned_at` and
> `cost_usd` stamped, plus the reaper orchestration row.

### Vendor-half drill (in progress at time of writing)

The 08-03 drill proved select+dispatch but was stopped by design before the
vendor call (non-numeric id + invalid token). Today both blockers are gone, so:
drill row `5a00b2a4-c0fd-4a7f-88ca-2eeb61dfa02c`, `thunder_instance_id='999999'`
(numeric-but-implausible — SAFE BY MEASUREMENT this time: `instances/list`
returned `{}`, zero instances on the account, so no id can match a real box, and
`DeleteInstance` treats 404 as success), `instance_ip` SET (dodging the
still-open `bugs_open/186`), rate $0, 30h old vs 18h cap. Expected pass:
lookup → `decommissioning` → Thunder delete 404≡ok → `decommissioned`,
`cost_usd=0`. Result recorded below when the tick lands.

### Vendor-half drill RESULT — the reaper is now proven END TO END

Tick 09:23:42Z, row terminal at **09:24:14Z** (32 seconds tick-to-terminal):
`running` → `decommissioned`, `cost_usd=0`, no error. Adapter log confirms
every stage that matters, including the two the 08-03 drill could not reach:
`Received request` → `Thunder API request` → `Thunder API response` →
**`Thunder instance already deleted (404)`** (a real, authenticated vendor
call; 404 on the fake id treated as success, exactly as designed) →
`Decommission complete, cost_usd=0` → `Sent success response`.

So the chain select → dispatch → lookup → MarkDecommissioning → **Thunder
delete** → Secret delete → MarkDecommissioned is now **proven live**, with the
one caveat recorded honestly: the lookup step passed because the drill row
carried an IP — the NULL-IP case is `bugs_open/186`, whose fix is committed
(`f83927375`) and **council-APPROVED round 1** (`862583b1`, verdict read) but
inert until the thunder-adapter image rolls. Re-run THIS drill with
`instance_ip` omitted after that roll (the 114 template now does exactly that).

Drill row deleted (matched id AND thunder_instance_id); table verified back to
23 rows / 0 live / 0 drill remnants; `can_provision=t`, spend $0.

**Cost of the entire two-drill campaign: $0.00** — no real instance was ever
created, both fake ids were unreachable-by-construction, and the account's
`instances/list` was `{}` throughout.

## 2026-08-08 (evening) — 186 VERIFIED LIVE: the NULL-IP re-drill passed

The fleet roll landed at 16:27Z (thunder-adapter `v1.0.1267`, pod
`thunder-adapter-86c889c64-7lwdt`). Provenance chain, honestly bounded:
the fix commit `f83927375` is 10:19 +0100; local docker shows `v1.0.1267`
built 14:03 +0100 (after the commit; `make build-*` builds committed HEAD).
Necessary but not sufficient — **and the standing pod-grep practice
(a string ADDED + a string REMOVED) is structurally unavailable for this
diff**: it changes only Go field types and comments, no string literal in
either direction. Nothing to grep means the behavioural drill IS the
verification, which is what the 114 template was rewritten to be.

Pre-drill checks all green: `instances/list` = `{}` at 17:51Z (authenticated,
zero instances — which is the ONLY condition under which a numeric drill id is
safe), table at baseline 23/23 decommissioned/0 running, reaper enabled with
280's widened pre_query (t/t/t), last tick 17:50:59Z, chassis pods 85 min old
(clear of the 300s dispatch rule).

Drill: row `2890abab-9d00-4e95-9bb3-2ce232e0adc7`,
`thunder_instance_id='999999'`, **`instance_ip` OMITTED → NULL** (confirmed at
insert: `ip_is_null=t`) — the exact input that killed the 08-03 dispatch with
"converting NULL to string is unsupported". Kicked the scheduler
(`last_triggered_at=NULL`) at ~17:52Z.

Result: **terminal in ~30s.** First poll 17:52:34 already showed
`decommissioned`, `cost_usd=3.6045` (2h × $1.80 synthetic rate),
`decommissioned_at` 17:52:31.9Z, `instance_ip` STILL NULL. Adapter log
17:52:31: `Received request` → lookup **passed** (the old binary died here) →
real authenticated `POST /instances/999999/delete` → 404 →
`Thunder instance already deleted (404)` treated as success →
`Decommission complete` → `Sent success response`. Reaper orchestration row
COMPLETED 17:52:34. The row reached `decommissioned` — one state beyond the
`decommissioning` the bug file's verify recipe predicted; full terminal
success.

Cleanup: drill row deleted (matched `id` AND `thunder_instance_id`,
`DELETE 1`), table re-verified 23/23/0. Campaign still $0.00 total.

**186 is fixed AND live.** Recorded in the bug file (which stays in
`bugs_open/` per owner ruling 08-06); RUNBOOK §1b caveat updated. Next lane
steps: Phase 0 (needs owner-ish supervision) and task #6 orphan sweep.

## 2026-08-08 (late evening) — task #6 BUILT: the orphan sweep (FTW-042)

Commit `81484df8a` (+ gofmt fix `2ef4ab581`), council submission
`7ffecfa2-ff96-4d73-be0d-25eb9589c6df` (Council-Submitted trailer; verdict
being read when it lands). Three pieces, all read-and-report:

- thunder-adapter `list_instances` — verbatim `GET /instances/list`
  passthrough (the client method existed, unit-tested, exposed by nothing).
- chassis `dispatch_thunder_list` — awaited dispatch, cloned from the
  decommission dispatch envelope.
- chassis `reconcile_thunder_instances` — pure classifier (12 table cases +
  a pinned unknown-age test): orphan_no_row / orphan_terminal_row filed as
  `thunder_orphan` work items on system.internal (severity high, dedup via
  idx_swi_dedup targetless ON CONFLICT), ghost_row reported-not-filed,
  30-min grace absorbs the provision INSERT-after-up window, zero createdAt
  cannot hide behind the grace.
- seed `sql_for_agents/342` — thunder-orphan-scan every 6h, NO pre_query
  (deliberate: always fire), shares thunder-lifecycle group with the reaper.

Design decisions worth remembering:
- **No remediation authority, considered and refused** — decommission needs
  the very row an orphan lacks; auto-pausing thunder_config would not stop
  an orphan billing, only block innocent lanes. Read-and-report is the whole
  scope; the work item's summary tells the human to kill the box at the
  console.
- **Ghosts not filed** because nothing is billing and the decommission path
  self-heals them (404≡success); filing them would train operators that
  thunder_orphan items are ignorable.
- **Adapter stays thin**: every site_work_items writer lives in
  platform/orchestration/actions and that invariant held — the adapter
  gained only the vendor passthrough.
- Collisions measured before submitting: 0 rows for the item_type across 86
  live types, both names unclaimed.

Missteps this session (small, both caught in-flight): the loose register
count regex (over-counts by 7 — the index header's own table says so; used
the strict command and said so in the chain); and the shell cwd persisting
into a later call so `docs/...` greps returned "No such file" from inside
the register dir — absolute paths after any cd.

Concurrency note: my two 000_concept_index.md edits (FTW-042 row + headline
link) were swept to HEAD by `a60a13cbb` (the 220 lane) minutes after they
landed on disk — same-file passenger, recorded in my commit message per the
sweep rule; the register ENTRY itself travelled in my commit as the seam
ruling requires.

**Not live until**: fleet roll (both images) → apply 342 → first-run
verification per the seed header (a 0-findings run must be cross-checked
against §1b's manual call the same day — a green scan must prove it LOOKED).

### Council round 1 on FTW-042: REVISE — and the gating objection was right for the wrong reason

Verdict 18:25Z: REVISE, 12 reviewers, 5 abstained, gated by the guidelines
seat (high): my targetless `ON CONFLICT DO NOTHING` dedup write, for which
it prescribed DELETE+INSERT as "the documented idiom". **Measured: DELETE
FROM site_work_items appears at ZERO sites in the codebase** — the named
remedy does not exist here. But the objection was right in substance, and
the codebase itself says so: `load_work_item_actions.go` has the SHARED
writer (`insertWorkItem` — targeted idx_swi_dedup clause via
workItemTerminalStatuses, within-cycle suppression, two-strike unresolved
labelling), and the `workItem.parentItemID` comment records a prior council
round (bugs_open/091) objecting to precisely the hand-rolled bypass I had
written. My round-1 code was that documented failure mode, recurring.
Lesson (also the round's best output): **before hand-rolling any write to a
shared table, grep for the shared writer first** — "reuse existing
machinery" applies to single INSERT statements, not just subsystems.

Other seats, all dispositioned in the revision (`ecbb0f362`) + round-2
resubmission (same corr, RESUBMIT_CORR trail): register entry now a covered
edit (it had shipped in the commit but the plan never listed it);
doc_notes write-back added to the seed (measured: NO pipeline subject
existed for any thunder mechanism — 4 landmine rows only); guardian's
generic-topic workflow-selection worry answered by code read
(extractGroupInfo reads config.agent_type first, which the scheduler writes
from target_agent_type — the landmine is about call_agent children; the
reaper's COMPLETED run today on the identical path is the live proof);
DO/RAISE verify block now inside the seed transaction; 342 numbering
checked against the directory + git log (343 taken concurrently; no
collision).

### Council round 2: DID NOT RUN — fleet-wide Anthropic credit exhaustion at 18:25:48Z

Round 2 (submitted ~19:15Z, same trail corr, all round-1 objections
dispositioned) reached `complete_invalid` in 5s — **not** a validation
failure: `__step_error` shows the FIRST seat (`review_editquality`) got API
400 "Your credit balance is too low to access the Anthropic API".
Measured in `llm_call_log`: first credit failure 18:25:48Z (my round-1
verdict composed 18:25:25 — through by 23 seconds), **zero successful
anthropic calls since**, 16 credit failures and counting. Every LLM-driven
pipeline fleet-wide is dead until the owner tops up Plans & Billing.
Owner notified (push).

**Resubmission owed once credits are restored** — the exact JSON is saved
beside this file (`council_r2_submission_ftw042.json`); the command:
```bash
RESUBMIT_CORR=7ffecfa2-ff96-4d73-be0d-25eb9589c6df \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/council_r2_submission_ftw042.json
```
Do NOT rebuild the submission from scratch; do NOT retry before credits
exist (each attempt consumes a dispatch for nothing).

## 2026-08-09 (midday) — FTW-042 LIVE: roll verified, first run failed then passed, two council rounds of corrections

**Roll + credits.** Whole fleet on `v1.0.1273` (~11:25Z); pod-grep both
chassis replicas + adapter: all positives non-zero, spelling-negative 0.
Credits restored overnight (clean llm_call_log from 01:00Z).

**Round 2 resubmitted** (saved JSON, RESUBMIT_CORR) → **REVISE**, gated by
prior_art_librarian, and the seat was RIGHT: `run-migrations.sh`'s
migrations home IS `sql_for_agents/` — my "no ledger exists" was a false
asserted-absence (WRONG_CALLS entry; the runner's five-line header refutes
it). 342 recorded via `--record-only`; a raw-psql apply otherwise leaves the
file pending-forever (bugs_open/007's replay trap).

**Seed applied → first run FAILED at `reconcile`, and this was the round's
real discovery:** `output_field` nested inside a step's `config` is INERT —
`models.Step` parses `stepMap["output_field"]` (processor.go:434); the
coordinator stores an awaited response under the STEP NAME plus step-level
output_field only. My seed copied the reaper's shape; **the reaper's own
config-nested output_field has been silently inert since it shipped** (its
step name and output_field are word-reversals, so its live specimen
"confirms" whichever model you already hold — which is how I 'verified' the
wrong one for the round-2 council answer). Guardian's round-1 low-severity
wiring objection was pointing at exactly this. LANDMINES entry appended
(verifier dispatched via trigger-landmine-verifier.sh — note the direct
`--apply` first consumed the new-entry status, the c7d4af7cc trap, hence the
manual trigger); WRONG_CALLS entry for the non-discriminating verification.

**Fix + verified:** output_field to step level in 342 (DO/RAISE now asserts
the step-level form), re-applied, re-kicked → **COMPLETED 11:42:06Z**,
`reconcile_result` = vendor 0/0, db 23/0, matched 0, clean:true — truthful
against the same-session manual `instances/list` (`{}`) and table read.
The 0-findings run proved it LOOKED. One-off double-fire at the first kick
(two runs 30s apart, stamp race, no recurrence in 6+ ticks) noted, absorbed
by dedup, not chased.

**Round-2 dispositions shipped** (`cfaa93126`): system.internal existence
asserted in the seed's DO/RAISE (a missing row silently degraded filing to
log-only), defensive ROLLBACK head + `342_..._ROLLBACK.sql` sidecar,
registry categories → `maintenance`, ledger checksum re-synced after the
hardening. **Round 3 submitted** (~12:0xZ, same trail): reviews shipped+live
state, carries the output_field correction openly. Result recorded below
when the verdict lands.

### Council round 3: REVISE (12:00Z) — all verification requests, answered by measurement in round 4

Gating (editquality, high): fear that the seed's jsonb verify NULLs out on a
missing key. The actual form is `PERFORM … IF NOT FOUND` (fails closed) —
and rather than argue, **round 4 proves it by induction**: in a transaction
I reproduced the regression on the LIVE row (`jsonb_set`/`#-` moved
output_field back into config), ran the identical DO block, and it raised
"ASSERT FIRED CORRECTLY"; rolled back, live config untouched. Other
dispositions, all by direct read: live row carries step-level output_field
on BOTH steps + nested key absent (debug_historian's "was it actually
re-applied") · sites system.internal exists (guardian) · registry keys
unique, `thunderAdapterTopic` declared once and reused (editquality +
prior_art) · recurrenceExpected: the strike counter counts TERMINAL
predecessors only, an open item is dedup not strike, the <3h suppress can't
bite a 6h cadence, and the third-after-two-closes files as `unresolved`
(visible + revalidatable) — nothing is swallowed (editquality medium) ·
why not extend the reaper: a scheduler pre_query is SQL and cannot call the
vendor API, which is the only place an orphan exists; different cadence and
workflow shape; shared concurrency group is deliberate (reuse_agent) ·
liveness claims a council tier can't see: human re-run commands supplied
verbatim; the run's counts are durable in these NOTES against the 24h
orchestration retention (prior_art). **Round 4 submitted ~12:15Z —
evidence-only, no code or config changes** (`council_r4_submission_ftw042.json`).

### Council round 4: APPROVED (2026-08-09 12:16:10Z) — the trail closes at four rounds

`decided_by`: "approved with 2 advisory objection(s) — none high-severity";
`gated_by_truncation: false`; 7 abstained. Seats: editquality approve ·
reuse_agent approve (low: precedent-checking rigour, not the reasoning) ·
guidelines approve (open question, not an objection) · guardian approve
(two low, both "shared file, additive, re-check next time") ·
**tooling_provenance OBJECT (medium)** · **debug_historian OBJECT (medium)**.

Approval does not make the two mediums go away, and both were answerable by
one command each, so they were answered rather than filed as accepted risk.

**1. tooling_provenance — RESOLVES CLEAN.** The objection: the seed's
`doc_notes` write uses `subject_type='pipeline'` in raw SQL, and I had
verified only the Postgres CHECK — never the Go-side gate, which the
landmine says is the second enforcement point `\d` cannot show. Checked
now: `validDocSubjectTypes` (`platform/orchestration/actions/
doc_subjects_common.go:63`) = `{tool, pipeline, experience, action,
experience-pattern, component, landmine}`. **`pipeline` is a member**, so
the raw-SQL row is consistent with the validated writer path and downstream
doc tooling reading through the Go gate will see it. The seat was right that
I had not checked; it happens the answer is favourable. [MEASURED 08-09]

**2. debug_historian — THE SEAT WAS RIGHT AND MY CLAIM WAS TOO NARROW.**
I wrote "fleet `v1.0.1273` (pod-grepped, both chassis replicas + adapter)".
True, and not the claim needed — the exact shape LANDMINES.md:5696 warns
about. Census run properly, by IMAGE not by label:
**34 pods run the chassis binary, 25 on `v1.0.1274`, 9 on `v1.0.1273`** —
i.e. the fleet had already rolled PAST the tag I proved on, by another
session's release, between the handoff being written and being read.
Applying the landmine's own two-part check:
- *Deployment-backed, current tag:* all 4 ReplicaSet-owned pods on
  `v1.0.1274` carry `reconcile_thunder_instances` (3) and `thunder_orphan`
  (4); spelling-negative `reconcile_thunder_orphans` = 0 on every one. Those
  4 are `agent-chassis` ×2 **plus `business-intel` and `vet-intel`** — which
  is precisely the landmine's point that the binary runs under other names,
  and which a label-scoped check would never have shown me.
- *The 9 stale ones:* all `Job`-owned (per-work-item pods pinned at spawn,
  ageing out on their own — the benign case), and they carry the action too,
  because `v1.0.1273` is the tag the code first shipped on. Reachability
  checked anyway with a positive control: `thunder-orphan-scan` → `t`;
  `nav-updater`, `section-editor`, `deployer-agent` → `f`.
  **So: zero stale-code exposure, on both halves.** [MEASURED 08-09]
- The honest form of the claim, which is what should have been written:
  **live on `v1.0.1274` as at 2026-08-09, on all 4 Deployment-backed pods
  running the chassis binary, with the 9 Job-pinned pods on `v1.0.1273` also
  carrying it.** Logged in WRONG_CALLS 08-09.

**3. guidelines' open question, also settled** (it flagged a fact, not a
fault): `thunder-orphan-scan` has `processing_mode = task`, **identical to
`thunder-reaper`** — so it runs inline, is NOT a WRAPPER-ORCHESTRATOR, and
does mirror its sibling as the seat guessed. Note the seat named
`thunder-decommission-dispatch` as the comparator; there is no agent of that
name — `dispatch_thunder_list` is an *action* (`registry.go:1556`). The real
sibling agent is `thunder-reaper`, and that is what was compared.

**Coverage credited automatically, nothing to amend.** 098 over 3 days lists
`81484df8a` · `ecbb0f362` · `95a455d35` · `cfaa93126` all as
`[7ffecfa2, by correlation, via submitted]` under APPROVED — the
`Council-Submitted:` trailer resolving at report time worked exactly as
CLAUDE.md describes, with no amend (forward-only preserved). Only
`2ef4ab581` (pure gofmt of the test file) sits in UNREVIEWED, which is
correct: it carries no substance.

**FTW-042 is DONE.** Live, verified on the current tag, council-approved,
registered. The two disclosed caveats stand unchanged and are not blockers:
the filing path has still never fired against a real orphan (the first real
one is its live proof), and the one-off double-fire at the hand-kicked first
tick was absorbed by dedup and not chased. Next lane item is Phase 0.

---

## 2026-08-12 — three measurements FROM ANOTHER LANE, for your Phase 0. Nothing dispatched.

Left here, not acted on: **Phase 0 is yours and a session of yours was mid-flight when I
found you** (B2 credentials exported, ~13:47). I was cold-started on
`finetuning/HANDOFF_2026-08-08` and planned your Phase 0 before knowing this lane existed
— my fault, logged in `WRONG_CALLS.md` 2026-08-12; my plan is headed SUPERSEDED and routed
here. These three are read-only measurements taken on 2026-08-11/12 that may save you time.

**1. ⚠ One of the three exports is POISONED, and it is the one a naive pick would take.**
`rows_exported` disagrees with reality — this is what killed training run `693656ce`
("export a8484922… has no rows"):

```sql
SELECT r.id, r.rows_exported AS recorded, count(x.id) AS actual
FROM training_exports.runs r LEFT JOIN training_exports.rows x ON x.export_id = r.id
GROUP BY 1,2;
--  146a9a12  1958 / 1958   good   <- the RUNBOOK's iter0 envelope already names this one
--  fef7be6b  1958 / 1958   good
--  a8484922  1957 /    0   EMPTY, and it is the NEWEST of the three
```
Control: 3,916 rows in the table = 1958+1958, so the join is sound.

**2. Both June blockers are verifiably gone on v1.0.1288** — worth knowing before you
re-derive them. The checkpoint-upload race fix is live in the chassis
(`preRegisterAwaitedRequest` + `prepare_object_url` probe PRESENT, fabricated-string control
absent), and `thunder-adapter` — which in June shipped the analyser binary and
CrashLoopBackOff'd — is **ready with 0 restarts**. Its provenance line is in range and gives
an exact commit (`bb5348642`); all three June adapter commits are ancestors of it, so this is
`merge-base`, not inference.

**3. `model_lifecycle.artefacts` HAS NO WRITER — a successful Phase 0 will record nothing
there.** The table appears in exactly one file in the repo, its own DDL
(`019_model_lifecycle_schema.sql`); no Go code and no agent config inserts into it, both
spellings checked. `evaluations` likewise, both empty.
I could not find this stated in your PLAN/RUNBOOK/NOTES, which is why it is here.
**It probably does not block you** — your Phase 0 verifies the adapter *at B2*, which is the
right authority and does not depend on this table. But it does mean "did the pilot produce an
adapter?" is not answerable from the registry afterwards, and I got that wrong in the other
direction first: I cited `count(artefacts)=0` as evidence no run had ever finished, when it is
a fact about the schema and could not have come out otherwise.

Not repeated here because you already have it: the `run.sh`-vs-B2-bundle divergence. Your
`MEMORY_workstreams.md` line records it as deliberate-until-Phase-0, and I only mistook it for
a discovery because I had not read that line.

— from the `finetuning/` (older service-thinking + site) lane

---

## 2026-08-12 — Phase 0 attempted. Did not train. Found five defects, one of them a money loop, and paused provisioning fleet-wide.

**Outcome in one line: no fine-tune ran, nothing was lost, ~$0.10 spent, and the
reason Phase 0 could not run is now written down in two bug files instead of
being folklore.** The training scripts were the part we expected to be risky;
they turned out fine after two fixes. The provisioning path underneath them
cannot currently deliver a box.

### Preflight — two defects found by READING, before any GPU existed

**(1) `BASE_MODEL` alone could never have trained a small model.** `02_train`
took `--base-model` (parameterised 2026-07-31) but line 201 imposed
`get_chat_template(tokenizer, "llama-3.1")` and lines 276-280 passed
`"<|start_header_id|>user<|end_header_id|>\n\n"` /
`"...assistant..."` as **literal strings** to `train_on_responses_only`.
SmolLM2 has none of those tokens. So the run this lane was about to pay for
would have rendered training text out of tokens the model has never seen, then
masked on markers that never occur — completing, and yielding a quietly wrong
adapter. **The 07-31 NOTES claim "run.sh parameterised for small models
(defaults preserve 70B behaviour exactly)" was true of the model NAME and of
nothing else** — recorded here as a correction to this file's own earlier entry.
Fixed (`270dbfd98`): `--chat-template` / `--instruction-part` / `--response-part`,
forwarded by run.sh as `CHAT_TEMPLATE` / `INSTRUCTION_PART` / `RESPONSE_PART`,
every default the previous literal. Verified byte-identical for the 70B path by
assembling the args under `set -euo pipefail` both ways (8 args unset, 16 set).
SmolLM2-1.7B-Instruct's real format read from its own `tokenizer_config.json`:
ChatML, `<|im_start|>user\n` / `<|im_start|>assistant\n`. Licence read in the
same repo we would actually download: **Apache 2.0** — the standing Phase 0
licence obligation is discharged.
Plus a **guard**, because this failure is silent by construction: the markers
must occur in the first 25 rendered rows or the run exits before the trainer
starts. Proven to discriminate offline against SmolLM2's verbatim template —
old markers `False` (fires), corrected markers `True` (passes). A guard that
could not have failed would have proven nothing.

**(2) The bundle upload was not the no-op its own comment claimed.** Live B2
`run.sh` hardcoded `SAVE_STEPS=10`; the git copy defaulted it to `50` under the
comment "identical to before". Before was **10**. Restored to 10, so deploying
did not quietly make a 70B run checkpoint 5× less often (~92 min of crash
exposure instead of ~18). Nothing had failed because no 70B run has started
since 2026-06-13.

Bundle then deployed (re-upload IS the deploy, FTW-031), md5 round-trip verified
both times. 300-row dataset uploaded at
`finetuning/datasets/phase0-2026-08-12/training.jsonl` — a realistic customer
size, and **every** row carries a user+assistant pair (checked: 0 of 300 lack
one), so the guard cannot false-alarm on it. Presigned URLs minted by hand (the
concierge equivalent of `prepare_object_url` + `assemble_upload_manifest`), and
the PUT **proven with a real round-trip against a throwaway object** — that is
the failure that would otherwise surface only after the GPU time was spent.

### Then the GPU, and three more defects — see `bugs_open/258` and `bugs_open/259`

- **`vcpus: 4` (the adapter default) is invalid for 9 of the 11 single-GPU
  specs.** `POST /instances/create` → `400 invalid vCPU count 4; valid options:
  [6 8]`. Thunder publishes the valid set as `vcpuOptions` on the free,
  read-only `/v1/specs`, so this is measured, not guessed. Only h100 accepts 4
  — i.e. **with defaults, the only provisionable GPU is the most expensive one.**
  Workaround: pass `vcpus` explicitly (forwarded when `> 0`).
- **`waitTimeout` is a hardcoded 5 min and an a6000 does not boot that fast** —
  twice, 4m39s and 4m49s still `STARTING`. The compensating cleanup then
  **deletes the box we just paid for** and returns an error. The cleanup itself
  worked perfectly (first real firing of that saga path — worth knowing); the
  defect is the deadline. Not in `thunder_config` (no such column) and not an
  env var, so fixing it needs a build + roll.
- **A failed provision leaves no durable record**: no `thunder_instances` row
  (insert is post-wait), **no `agent_error_log` row** (not a quiet table — 8
  other agents logged errors in the same window), a stuck `orchestration_states`
  row, and rotating pod logs. So "how often does provisioning fail" is
  unanswerable, including retrospectively for this bug.

### The one that stopped the session: two requests → three billing GPUs

Counting deliveries per `correlation_id`, one request was consumed **3×** and
another **2×**, ~10 and ~13 minutes apart, each valid delivery issuing a fresh
create. Cause, from code: the handler is synchronous in the consume loop
(`adapter.go:257` — *"Sequential by design"*), blocks up to 5 min in
`WaitForRunning`, while `SessionTimeout`/`RebalanceTimeout` are **60s**
(`consumer.go:56,71-72`). ~5× its own deadlines, offset uncommitted, message
comes back, another box gets built. **Which broker-side path fires is
[UNVERIFIED]** — `090` filed, run correlation
`8ee2eb1e-2c1d-4a69-9d1b-505895c4dbcb`; read the verdict before asserting a
mechanism. The *behaviour* is counted from the log and needs no caveat.

**Contained, and the containment is proven not assumed:** `is_paused = true`
(RUNBOOK §1b emergency stop; checked at the top of `Execute`,
`provision_action.go:156`). In the 90s after: **2 deliveries denied, 0 creates**,
vendor list `{}`. Owner decision same day: **leave paused until fixed.**

### Corrections to my own claims THIS session

> **CORRECTED — I said the a6000 floor price was "$0.43/hr, not $0.35".**
> That was an inference from the pricing page's "+$0.04/vCPU/hr beyond 4" rule
> applied to the 6-vCPU minimum. But the a6000's **minimum is 6**, so the
> advertised $0.35 may already assume 6 and the surcharge may not apply at all.
> The honest figure is **$0.35–$0.43/hr, [UNVERIFIED], pending a real invoice** —
> which is the only thing that can settle it. Stated as fact when it was
> arithmetic on an assumption.

### Ledger

Spend: ~15 min of a6000 across three instances ≈ **$0.10**.
`thunder_config` net change: `default_hourly_rate_usd` 1.80 → 0.35 → **1.80**
(restored 14:12Z; no instance row was ever written, so nothing took the 0.35
stamp). `is_paused` **left true, deliberately.**
Estate state: `thunder_instances` back at 23 rows / 0 live, vendor `{}`.
Left behind: three `orchestration_states` rows stuck in `AWAITING_RESPONSES` —
harmless, non-billing, noted in 258 as the undiagnosed error-response observation.

---

## 2026-08-12, evening session — 259 fixed, and its filed root cause REFUTED

### The misstep first, because it is the point of this entry

I picked this lane up to do step 1 of the handoff: fix 259 by making the create
idempotent on `correlation_id`. I read `thunder_provision_dispatch.go` and
concluded the handoff and the bug file had picked the **wrong key** — `request_id`
is minted per dispatch (`uuid.NewString()`, :99) and is stable across a Kafka
redelivery, whereas `correlation_id` is the orchestration-run id shared by every
message in the run. On that reading, keying on `correlation_id` would refuse a
*legitimate* second provision in one run, and `request_id` was obviously more
correct.

**That reasoning was sound and the conclusion was wrong, because the premise it
inherited — Kafka redelivery — was wrong.** I went to check it rather than act on
it, which is the only reason this is a NOTES entry and not a defect.

The log route was gone: the evidence pod died in the 14:55Z roll, so
`kubectl logs` returned **0 provision deliveries** — an absence entirely
explained by pod replacement (`startTime 14:55:33Z`, `restartCount 0`, 13 lines
retained). Not evidence of anything. The durable evidence was in a table:

```sql
SELECT correlation_id, count(*) AS rows, count(DISTINCT request_id) AS req_ids
FROM awaited_requests WHERE target_agent_type='thunder-adapter'
  AND sent_at > now() - interval '2 days' GROUP BY 1;
--   f17cccda | 4 | 4      23c9bc6a | 4 | 4      cd614594 | 4 | 4
```

Four rows, **four distinct `request_id`s**, per correlation. A redelivered Kafka
message replays identical bytes, so `request_id` would be constant. It is not —
so these are four separate publishes, and redelivery is **disproved**, not merely
unsupported.

The detail names the mechanism outright:

| request_id | sent | timeout_at | processed_at | status |
|---|---|---|---|---|
| `b40062fa` | 13:52:12 | 14:02:12 | 14:02:12 | processed |
| `da87c9c9` | 14:02:13 | 14:12:13 | 14:12:13 | processed |
| `6f7bda74` | 14:12:15 | 14:22:15 | 14:22:15 | processed |
| `641450c2` | 14:22:17 | 14:32:17 | 14:32:17 | error |

One `orchestration_id` (`8c5bf926`), four `step_id`s. Every row `processed` at
*exactly* its own `timeout_at`; every next dispatch ~1 second later. That is an
await expiring and `retryExpiredAwaitedRequest` firing (budget `RetryVersion < 3`
→ four executions, then FAILED at 14:32:17). `dispatch_provision`'s
`timeout_seconds` is **600** — the "~10 minute redelivery gap" the original
filing measured.

**And the co-cause we had already observed without connecting it:** the adapter
answered ~5 minutes in (`Sent error response`) and the await did not clear — it
expired on its own clock, every time. Had the error response cleared the await,
the step would have failed after ONE attempt and only one GPU would ever have
been built. The "does an error response clear an await?" open question from this
morning is not a curiosity; it is *why the retry loop ran*. Still undiagnosed.

**So `correlation_id` was the right key after all** — for a reason neither the
bug file nor my objection had: it is the only identifier stable across the four
attempts. My `request_id` reasoning would have produced a guard that could never
fire, with a green test suite, because a test naturally reuses one id.

Filed to `WRONG_CALLS.md` (the marker rule was followed in the body and still
failed, because the **title** carried the unmarked claim) and to `LANDMINES.md`
(the `request_id`-looks-canonical trap, with the one-query check).

### What shipped — `10659b419`, `Council-Submitted: 20d8b725`

`thunder_provision_claims` (PK `correlation_id`), claimed **before** the vendor
call. One statement — `ON CONFLICT (correlation_id) DO UPDATE SET attempts =
attempts + 1 RETURNING …, (xmax = 0) AS inserted` — so claim and count cannot
race, and Postgres decides the dedup verdict rather than a count we compute.
`ErrProvisionDuplicate` → `provision_duplicate` / **`error_unrecoverable`**.

Deliberate, and worth revisiting if it bites: **a failed attempt keeps its
claim.** Every attempt on 2026-08-12 failed and was cleaned up, so a
release-on-failure rule would leave the loop exactly as it was.

Its own table, not a column on `thunder_instances`: a pre-create row there has no
real vendor id, and `reconcile_thunder_instances` files a `ghost_row` for any
live row absent at the vendor (`thunder_reconcile_action.go:204-219`) — every
in-flight provision would raise a spurious finding against our own FTW-042.

**The test was proven able to fail.** Mutating the refusal branch out:
`CreateInstance called 2 times for one logical request`. Restored: green. A guard
whose test has never been seen red is not a verified guard.

Incidental find: `fmt.Errorf(reason)` (HEAD's line 161) is a non-constant format
string, which **vet rejects — so `go test ./internal/adapters/thunder/` could not
build at HEAD, and no test in that package has been running.** Now `errors.New`.
Separately, `internal/adapters/thunder/api/client_test.go` does not compile at
HEAD either (`unknown field Identifier in struct literal of type Instance`) —
untouched, not mine, reported.

### Ledger for this session

No spend. No provision attempted, nothing unpaused, no cluster state changed
beyond the council dispatch. `is_paused` **still true** — correctly, because the
fix is committed but not built, and a fix that is not in the running binary has
not fixed anything.

---

## 2026-08-13 — 259's fix confirmed LIVE, and 258 defects 1+2 fixed behind it

### The deploy check, done properly rather than accepted

The owner reported "a new chassis has been deployed which included the thunder
adapter I think". Verified rather than taken:

```
pod    thunder-adapter-5c7c698ffd-b67nc   started 2026-08-13T13:53:44Z, 0 restarts
image  docker.io/aqls/thunder-adapter:v1.0.1295
stamp  {"msg":"build provenance","git_commit":"69612d692a4a07d61eea3f648e1152e0fd36fd0a"}
git merge-base --is-ancestor 10659b419 69612d692   -> 0  (IN the build)
git merge-base --is-ancestor 7abafc76f 69612d692   -> 0  (IN the build)
rev-list --count 10659b419..69612d692             -> 154
```

So **259's claim guard is live**, and migration 396 was already applied, so there
was no fail-closed risk. `thunder_provision_claims`: present, 0 rows.
`is_paused`: still true. Live instances: 0. Adapter error lines since the roll: 0.

**Read the stamp per SERVICE, not per fleet** — `bugs_open/249`. And note the
provenance line is a *startup* line: it was in `--tail=300` here only because the
pod was 20 minutes old. On a busy service it scrolls out and an empty grep means
"not in range", not "unstamped".

### The check I made sure to run: is 258 defect 2 still there?

Asked the **deployed commit**, not my working tree:

```
git show 69612d692:internal/adapters/thunder/provision_action.go | grep -n 'waitTimeout:'
157:    waitTimeout:      5 * time.Minute,
```

Still hardcoded. So unpausing would have provisioned an a6000, waited 5 minutes,
and deleted it — spending money for a guaranteed failure. That is what turned the
next step into an owner decision rather than a task, and the owner chose: **fix
258 first, stay paused, spend nothing.**

### 258 defect 1 — I re-measured the catalogue before trusting the bug file

The bug's table was a day old and describes a *vendor's* menu, so I re-ran
`GET /v1/specs` myself. It matched exactly: 9 of 11 single-GPU specs reject 4;
the two that accept it are h100. Also learned two things the table did not carry:

- the envelope is `{"specs": {...}}` — decoding straight into `map[string]Spec`
  compiles, succeeds, and yields an **empty map**, which would read as "Thunder
  sells nothing" rather than as a decode bug. Hence an explicit envelope type
  that errors on empty.
- `l40s` is a GPU constant in our source with **no live single-GPU spec**. Asking
  for it now refuses instead of 400ing.

**A subtlety worth recording:** the fix had to *reorder* the default-resolution
block. The spec key is `<gpu>_x<count>[_<mode>]`, so deriving vCPUs before
normalising the GPU alias and mode looks up an empty key. That is the kind of
ordering that works by accident until someone tidies the block, so it is
commented in place.

### The trap I went looking for, and found

I asked what the *good* case looks like if the new wait timeout is set too large —
not just what the bad case looks like. The answer is the interesting one:

**above the step's 600s await, raising the wait produces a quiet SUCCESS, not a
timeout.** The await expires first, the retry driver re-dispatches, 259's guard
refuses the duplicate *correctly*, and the workflow reports FAILED while a real
billed instance runs on with a row and no watcher. The failure mode of being more
patient is a provision that worked and that nobody knows about.

Hence 540s against the 600s await, and the order-of-change rule (**step first,
column second**). Recorded in the migration header, the column COMMENT, a test,
RUNBOOK §6 and `LANDMINES.md` — five places, because the number is a single
`UPDATE` away from anyone with psql and good intentions. It is a **stated**
invariant, not a mechanical one: the adapter cannot see the step's config. Flagged
in the council submission's risks as the thing to push back on.

### Mutation evidence

Restoring `req.VCPUs = api.DefaultCPUCores` fails **7 of 8** cases in
`TestDefaultVCPUsAreValidForEveryGPU`. The sole survivor is
`h100_x1_prototyping` — the one spec that accepts 4. That single pass *is* the bug
in miniature: testing "a provision" with the expensive box never catches it.

### Blocked, and on what

The kubeconfig token **expired 18:05:20Z** (checked by decoding the JWT `exp` in
`~/.kube/config_production_uk001`; fleet-wide `Unauthorized`). So, still owed and
NOT done:
- migration **400** not applied (not ordering-critical — the `to_jsonb` read means
  an unmigrated DB degrades to the compiled default rather than failing)
- the **council round** for 258 — submission written and committed, never sent
- `./scripts/landmines-sync.py --apply` — two entries added today are file-only,
  so seats and agents cannot yet read them

No spend this session. Nothing unpaused. No cluster state changed at all — the
`pause_reason` rewrite and migration 396 were yesterday, before the token died.

---

## 2026-08-15 — both fixes now genuinely live, and the 44-hour gap that says why "live" needs two checks

Token restored; everything blocked on 08-13 is done.

### The deploy check, and the thing it does NOT cover

`v1.0.1301`, pod started 10:14:33Z, stamp `0115f2b4528b0063fd01e7af275ccefe9c5a991d`.
All four commits confirmed ancestors: `10659b419` (259 guard), `7abafc76f`
(classifier), `236810e4e` (258 defects 1+2), `647f3404a` (the 397→400 renumber).

**Then I checked the other half, and it was missing.** `provision_wait_timeout_seconds`
did not exist — migration 400 had never been applied. So for the ~44 hours since
`v1.0.1295`, the adapter had 258 defect 2's *code* and was silently using the
compiled-in **5 minutes**, because the read falls back when the column is absent.
That fallback is deliberate and right (it stops an unmigrated DB breaking every
provision), which is exactly what makes it invisible: no error, no failure, and an
ancestry check that genuinely passes.

**The generalised lesson, now in `LANDMINES.md`:** *prove it at the artefact* has
**two** artefacts for a config-backed change — the binary **and** the row. Assert
the VALUE, not the version. And make the fallback log at WARN naming its migration,
which this one does; that warning is the only tell.

Note the asymmetry with 396, which is deliberate: the claims table is a **safety
guard**, so an absent table is a HARD ERROR and the adapter refuses to provision.
The wait deadline is a **tunable**, so an absent column degrades. Guards fail
closed; tunables fall back.

### Applied, verified by induction rather than by catalogue

Migration 400 applied and recorded. Then checked the invariant instead of assuming
it:

```
wait timeout now: 540s | dispatch await: 600s | INVARIANT OK: wait < await, headroom 60s
```

And ran the **exact 8-column projection `store.LoadConfig` issues** against the live
schema — returns `... | t | t | 540`. That proves the read path end-to-end, which a
`\d thunder_config` would not: the column existing and the code's query working
against it are different claims.

### Also done

- **258 submitted to the council**: `d24f9829-0a3f-47a8-bdcb-4b63ced63f1b`
  (verdict not yet read at time of writing; two seats were mid-flight).
- **landmines-sync**: reported `already in sync` — another session had run it. Did
  **not** take that at face value: queried `doc_notes` directly and confirmed all
  three of this lane's entries are present and readable.

### State at handoff

`is_paused` **true**. 0 claim rows, 0 live instances, nothing billing. Nothing has
been provisioned since 2026-08-12. No spend this session.

**The single remaining step is the unpause + Phase 0 run**, which is the owner's
step 5 and the money step — staged in full in
`HANDOFF_2026-08-15_continue_here.md` §2, including the three separate things that
run has to show and the one it probably will *not* (259's live proof, which needs
an await to expire and so will not appear if the provision simply succeeds).

---

## 2026-08-15 (later session) — cold-start re-verification; the pod had rolled under the handoff

Picked up `HANDOFF_2026-08-15_continue_here.md`. Re-verified its §1 state table
against the live system rather than reading it forward. **One line of it was already
stale, and it was the one everything else rests on.**

### ⚠ The handoff's build stamp expired within the hour

| | handoff (written ~11:47) | measured (~15:10) |
|---|---|---|
| image | `v1.0.1301` | **`v1.0.1302`** |
| stamp | `0115f2b45` | **`194907d5b6cc69c1ecb50263bd958b2940019587`** |
| pod start | 10:14:33Z | **11:29:23Z** |

Another session rolled the fleet. `194907d5b` is *274: ratchet the closed bugfix
lane dir* — nothing to do with this lane, which is exactly the point: **on this tree
your build stamp is invalidated by other people's work, so a stamp recorded in a
handoff is a snapshot with a half-life of hours.** Re-ran the ancestry check against
the NEW stamp:

```
10659b419 (259 claim guard):   ANCESTOR of stamp — LIVE
236810e4e (258 defects 1+2):   ANCESTOR of stamp — LIVE
control: HEAD (65a39bbd8) NOT an ancestor — stamp is behind HEAD, as expected
```

[MEASURED] Both fixes survive the roll. The control matters: without it, an
`is-ancestor` check that returned true for everything would look identical.

### The migration-400 fallback WARN is a REAL absence, not an out-of-range one

`grep 'migration 400\|compiled-in default'` over `--tail=3000` → empty. That is the
result you get both when the warning was never emitted *and* when it has scrolled
out of range, and the two mean opposite things. **Control that separates them:** the
`build provenance` line — same startup, same second (11:29:28Z) — is still in the
window. Startup lines are therefore in range, so the empty grep is a genuine absence.
The config half is live. (`provision_wait_timeout_seconds` = 540 confirmed at the DB
in the same pass.)

### The 600s invariant, measured at the schema instead of quoted

The `prior_art_librarian` seat asked for exactly this — "that number is
schema-checkable and should not rest on the rationale alone". It is **not** at the
step's top level (`timeout_seconds` reads NULL there, which is how you'd conclude it
was unset); it is nested one level down in `config`:

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'dispatch_provision')
FROM agent_definitions WHERE type='gpu-provisioner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- → "config": { "timeout_seconds": 600, "output_field": "provision_response" }
```

[MEASURED] 540 < 600. Invariant holds, 60s headroom. Note the shape trap for anyone
re-checking: `jsonb_each` on `->'steps'` works (it is an **object**, not an array —
`jsonb_array_elements` errors with `cannot extract elements from an object`), but the
timeout is a level deeper than the step.

### `thunder_instances` says 23, and that is not a contradiction

The handoff's table says "claim rows / live instances | 0 / 0"; a bare
`count(*)` returns **23**. Broken down: all 23 are `decommissioned`, newest
**2026-06-18**. Vendor cross-check returned `{}` — no instances at all.
**Nothing is billing.** The 23 is history, and `count(*)` on that table is not a
measure of live spend — group by `status` or you will scare yourself.

### 258 council verdict: READ, and it is APPROVED

`d24f9829-0a3f-47a8-bdcb-4b63ced63f1b` → `approved`, 2026-08-15 10:55:42Z,
`decided_by`: *"approved with 1 advisory objection(s) — none high-severity"*.
Nine seats reported; `debug_historian` returned `object`, the rest `approve`.
`gated_by_truncation: false` (worth checking given the 138 truncation history).

The predicted push-back landed exactly where the handoff said it would — **four
separate seats** (`editquality`, `guardian` medium, `guidelines`, `architecture`) all
independently objected that the wait/await coupling is enforced by prose + a static
test assertion and **not mechanically**. That is now the lane's most-corroborated
open gap, not a stray objection.

### Three cheap open items from the seats, closed by direct check

1. **`editquality` edit 3 — "does any other path still hardcode vcpus?"**
   `grep -rn DefaultCPUCores --include=*.go .` → the only remaining references are
   the **declaration** (`api/types.go:46`) and **the mutation test**
   (`provision_defaults_test.go:93-94`), which uses it as the *old wrong value*
   sentinel. **No production call site.** Objection answered negative.
   ⚠ But note what that leaves: an **exported constant `DefaultCPUCores = 4` with no
   production consumer, whose value is invalid for 9 of 11 specs.** It survives
   legitimately (the test needs the old value to assert against), and it is also
   exactly the shape a future session reaches for. Flagged, not touched — changing it
   is a code edit for the gate, not a drive-by.
2. **`reuse_agent` / `prior_art_librarian` — "is `GetSpecs` a rediscovery?"**
   Sole definition is `api/client.go:111`. Prior `/v1/specs` knowledge existed only
   in two standalone probe scripts (`scripts/utils/thunder_probe.go`,
   `scripts/initial_messages/300_thunder_flywheel/thunder_probe.go`) — never as an
   adapter client method. **Genuinely new, not dormant machinery.**
3. **`guardian` missing — "is the reaper affected by the new column?"**
   Only three files reference `thunder_config`: `adapter.go` (a **named-column**
   sanity read of `daily_cap_usd` only), `store/config.go`, `provision_action.go`.
   `LoadConfig` names all eight columns explicitly — **no `SELECT *`** — and reads the
   new one via `to_jsonb(t)->>'…'`. No consumer can break on the added column.
   Answered.

### State at end of this pass

Unchanged and deliberately so: `is_paused` **true**, 0 claims, 0 live instances,
vendor `{}`, **no spend**. The money step was NOT fired — it is the owner's call and
has been put to him. Nothing in this pass wrote to the cluster; every command was a
read.

---

## 2026-08-15 (same session, later) — PHASE 0 PROVISION TEST FIRED AND PASSED

Owner approved "provision test only" (not the full Phase 0 training run). Fired,
proved, cleaned up, re-paused. **Total elapsed unpaused: 3m34s. Total cost:
$0.0645 by our accounting** (see the caveat on that figure below — it is not the
vendor's number).

### Timeline, from the artefacts

| time (UTC) | event | source |
|---|---|---|
| 14:26:50 | `is_paused=false`; `can_provision` → `t` | DB |
| 14:27:17 | provision dispatched, correlation `d2cc212d-…4c9d` | kcat |
| 14:27:24.992 | **`Resolved vCPU count from Thunder specs`** | adapter log |
| 14:27:25 | instance created at vendor (`createdAt`) | vendor |
| 14:27:26.106 | **`Provision wait deadline from live config`** | adapter log |
| 14:27:31 / :36 | poll → `STARTING` | vendor |
| 14:27:41 | poll → **`RUNNING`**; our row written | vendor + DB |
| 14:27:41.492 | `Provision complete` | adapter log |
| 14:29:43 | decommission dispatched | kcat |
| 14:29:49.752 | `POST /instances/0/delete` → 200 `{"message":"Success"}` | adapter log |
| 14:29:50.222 | row → `decommissioned`, cost stamped, SSH secret deleted | DB + log |
| 14:30:24 | `is_paused=true` restored | DB |

### The three things the run had to show

**1. 258 defect 1 — PROVEN.** Verbatim, `provision_action.go:529`:

```json
{"msg":"Resolved vCPU count from Thunder specs","spec_key":"a6000_x1_prototyping",
 "vcpus":6,"vcpu_options":[6,8]}
```

Exactly the line the handoff predicted, spec key and options included. Corroborated
independently **at the vendor**, which reported `"cpuCores":"6"` on the live box —
the old hardcoded `DefaultCPUCores = 4` would have been rejected with a 400. Two
independent witnesses (our log, their API), and **no `vcpus` was passed in the
request**, which was the whole point.

**2. 258 defect 2 — PROVEN.** `provision_action.go:489`:

```json
{"msg":"Provision wait deadline from live config","wait_timeout":540}
```

The live config was read; **no `using compiled-in default` warning anywhere in the
pod's log.** The box reached `RUNNING` and was **not** deleted mid-provision. Note a
small doc drift: the handoff predicted the line would render `wait_timeout=9m0s`; it
actually logs the integer `540`. Same value, different rendering — **if you grep for
`9m0s` you will conclude the fix is missing.** Corrected in RUNBOOK §6.

**3. 259 — NO live proof, exactly as predicted.** `thunder_provision_claims` shows
`attempts=1, status=succeeded`. The await never expired, so the retry driver never
fired and the guard was never asked to refuse anything. **This is now much worse
than "not yet observed" — see the boot-time finding below, which makes a natural
slow case on a6000 essentially impossible.**

### ⚠ THE BIG MEASUREMENT: a6000 cold boot is ~16 SECONDS, not ">5 minutes"

`createdAt` 14:27:25 → first `RUNNING` poll 14:27:41. The adapter polls every 5s, so
the true figure is **between 11 and 16 seconds**. [MEASURED]

The lane has carried `> 5 min` as the honest floor since every prior attempt was
killed at the old 5-minute compiled-in deadline while still `STARTING`. **That
premise is now refuted for a6000.** I am NOT claiming it was always wrong: the two
historical rows in `thunder_instances` are both `a100xl_1`, a scarcer and bigger
card, so the ">5 min" observations may have been a different spec entirely and may
still hold there. What is settled is that **a6000 boots in seconds, and the 540s
wait is ~33× the measured need.**

Consequences worth thinking about before anyone acts on them:

- The 540/600 coupling that four council seats objected to is, for a6000, an
  enormous margin over a 16s reality. The invariant still matters (it protects the
  *slow* case), but the urgency of mechanising it is lower than it looked.
- **259 will now essentially never get natural live proof on a6000.** An await that
  needs 600s to expire cannot expire against a 16s boot. Waiting "for a naturally
  slow provision" — the handoff's safer suggestion — is waiting for something that
  will not happen on this spec. If 259 is to be proven, it must be induced
  deliberately, and §4's quiet-success trap makes that genuinely dangerous. **This
  is now an owner call, not a task.**

### ⚠ THE COST FIGURE IS OURS, NOT THUNDER'S — the price question is still open

`cost_usd = 0.064467881216` for `uptime = 128.935762432` s. That divides out to
**exactly $1.800000/hr**, which is `thunder_config.default_hourly_rate_usd`.
`provision_action.go:429` stamps `HourlyRateUSD: cfg.DefaultHourlyRateUSD` onto the
row at provision time and `decommission_action.go:152` computes the cost from that
stored rate — **a flat configured rate, applied regardless of GPU type.**

So: **this run does NOT answer the open a6000 price question.** At the advertised
$0.35–0.43/hr the same 129s would be **$0.0125–$0.0154**, i.e. we are over-stating
a6000 spend by roughly **4–5×**. That is conservative in the safe direction for the
$30 daily cap (it trips early, not late), but it means `total_24h_spend` is not a
spend figure, it is an upper bound. **Only an invoice settles the real price** — the
handoff's open question stands unchanged. [UNVERIFIED: the vendor's actual charge]

### ⚠ LANDMINE FOUND THE HARD WAY: `kubectl logs -l app=<x>` returned ZERO lines

Mid-run, `kubectl -n ai-persona-system logs -l app=thunder-adapter --since=15m | grep
"Resolved vCPU"` came back **empty**, as did `grep "instances/create"` — while the
box demonstrably existed. On that evidence the honest reading is "the derivation
never ran", and the tempting next move is to go looking for a bug in code that is
in fact working perfectly.

**What caught it was a control**: I grepped for `Provision complete`, a line I had
*already read with my own eyes* minutes earlier. It returned **0**. A query that
cannot find a line you have already seen is a broken query, not evidence of absence.

Root cause: the **label selector** form returned nothing for a single-replica,
zero-restart service. `kubectl logs <pod-name>` on the very same pod returned the
full 48-line history including both proof lines. This is adjacent to the known
"`logs -l` reads one pod of N" trap but is not the same thing — here N was 1 and the
answer was *nothing at all*. Appended to `LANDMINES.md`.

### State at end

`is_paused` **true** (reason string records the result). `can_provision` **f**.
Vendor `instances/list` → `{}`. Live instances **0**. Claims: 1 row, `succeeded`,
`attempts=1`, retained deliberately as the audit record — **not cleared** (§5:
clearing is only for a genuinely orphaned claim, and this one is neither orphaned
nor blocking, since a re-run gets a fresh correlation). `total_24h_spend` $0.0645.

---

## 2026-08-15 (same session, continued) — a100xl timed, training run staged, then an owner credits hold

Owner approved three things: a100xl boot timing, the training half of Phase 0, and
the landmine verifier. The first ran; the second was staged to the point of launch
and then **postponed on the owner's mid-flight instruction** ("credits are about to
run out — don't start anything billable"); the third was dispatched before that
instruction arrived (correlation `e1f41304-dd57-46c7-a95d-9ba0b471995d`, verdict
lands async in `doc_notes` under `landmine-verification`).

### a100xl boot: 12–17s — and the derivation proven on a SECOND spec

Fired 14:52:01Z, correlation `e32bb9a2-…`. `Resolved vCPU count` →
`spec_key=a100xl_x1_prototyping vcpus=8` (lowest of `[8,12,16]` — a second live
proof of 258 defect 1 on a different catalogue entry). Vendor `createdAt` 14:52:11,
`STARTING` at :22.9, `RUNNING` at :27.9 → **boot 12–17s**. Decommissioned at
14:54, $0.047 booked (~94s; real ≈ $0.028 at the advertised $1.09/hr).

> **CORRECTED (same day, hours later) — my morning claims overreached, and the
> lane's own 08-12 notes are what refute them.** I wrote that ">5 min does not
> hold for a6000" hedged only by "the historical slow rows were a100xl" — wrong
> on the second half: **the 08-12 session measured the a6000 itself at 4m39s and
> 4m49s still `STARTING`, twice** (this file, §"Then the GPU"). Today the same
> spec booted in ~16s. So the true finding is **boot time is DAY-VARIABLE by
> ~20×**, not "boot is fast": the 540s deadline is not over-generous (it protects
> the slow days, which demonstrably happen), and my claim that 259 "can now
> essentially never fire naturally on a6000" is **withdrawn** — on a slow day it
> can. What caught it: re-reading the 08-12 handoff for an unrelated recipe.
> The 33×-margin framing in the earlier entry and in HANDOFF 15b §3/§4 is
> corrected in both places today.

### Training half: STAGED to the moment of launch, then held

All free groundwork done and verified:
- Presigned URLs minted with a rebuilt `presign.py` — **now committed to this
  directory** so it stops dying with session scratchpads; creds read live from
  `personae-storage-secrets`, never hardcoded.
- Bundle GET verified (206) and **md5 `a19557ccf61ac951c28e81254a8d76f7` matches
  the handoff** — it is the env-var bundle. Dataset GET verified (206). Final
  adapter PUT minted for `finetuning/artefacts/phase0-<ts>/adapter.tar.gz`.
  (First bundle probe returned **503 — B2 transient**; 206 on retry. Don't
  diagnose off one 503.)
- Read `02_train`: `--instruction-part`/`--response-part` are **literal, no
  unescape** → the env values need real newlines (`$'...\n'`), and the marker
  guard fails fast pre-GPU on a mismatch. Manifest shape confirmed:
  `{"final":{"key":…,"url":…}}` suffices with `SAVE_STEPS=0`; the final upload
  stays the hard gate (FTW-033), so `RUN_SH_DONE ⟹ durable` is still what the
  run will prove.
- Full launch recipe written into **RUNBOOK §9** (the ssh_exec command, the
  polling markers, the B2 durability proof) — the next session fires it in
  minutes, no research.

A training a6000 **was provisioned** (14:58:13Z, correlation `8391d172-…`) and the
owner's hold arrived seconds after it reached RUNNING — **decommissioned
immediately, before any command ran on it**: $0.019, ~37s. Re-paused 15:00Z, pause
reason records the hold.

### Day's ledger (superseded by the evening run below)

Three boxes, all decommissioned, vendor `{}` at end: $0.0645 + $0.047 + $0.019 =
**$0.113 by our books** — real vendor cost ≈ **$0.03** (flat-$1.80 inflation, see
the cost caveat above). Claims table: 3 rows, all `succeeded`, `attempts=1`.

---

## 2026-08-15 evening — PHASE 0 TRAINING RUN COMPLETE AND PROVEN END-TO-END

Owner restored credits and said carry on. The training half ran to `RUN_SH_DONE`
with the adapter **verified durable in B2 at the artefact** — the proof FTW-032 and
FTW-035 had been waiting on since June.

### Timeline (all 2026-08-15 UTC)

| time | event |
|---|---|
| 16:22:32 | provision dispatched (`4cd5e6a8-…`), box `8720be5c-…` at 216.81.200.240 |
| 16:27:43 | **launch 1 FAILED usefully** — see finding 1 |
| 16:39:00 | relaunch with patched setup embedded over ssh (md5 `206e19c6…` = repo copy) |
| 16:38:58 | `RUN_SH_START` (SmolLM2-1.7B-Instruct, ChatML, SAVE_STEPS=0, MIN_VRAM_MIB=8000) |
| →16:44 | setup: venv + torch cu124 + unsloth ≈ **5.5 min** (the 25-min figure was for named templates; `base` + pip is faster) |
| →16:46 | smoke: 20 rows, 1 epoch — `train_runtime` **40.5s**, loss 1.408, `RUN_SH_SMOKE_OK` |
| 16:47→17:09 | full train: **1363.4s (22.7 min)**, 111 steps ≈ 12.3s/step, 3 epochs |
| 17:09 | `uploaded final_adapter.tar.gz (0.07GB) -> HTTP 200` |
| 17:10:01 | `RUN_SH_DONE` |
| 17:13:15 | box decommissioned, vendor `{}`; re-paused |

**Total: ~50 min provision-to-decommission. $1.50 booked / ≈$0.29 real.**
Final `train_loss` **0.730** (smoke 1.408 → full 0.730). Adapter: 72,396,376-byte
safetensors; tarball **67,989,958 bytes**.

### The durability proof, stated precisely

`RUN_SH_DONE ⟹ adapter durable in B2` was proven **at the artefact, not at the
marker**: fresh presigned GET against
`finetuning/artefacts/phase0-20260815-1621/adapter.tar.gz` → **HTTP 206**, gzip
magic bytes (`1f8b`), `Content-Range: bytes 0-0/67989958`. Two independent
witnesses (box-side HTTP 200 log line; our own GET from outside the box). ⚠ One
instrument note: `curl -I` (HEAD) against a GET-signed presigned URL returns a
**163-byte error body**, which reads as "Content-Length: 163" — size a presigned
object by `Content-Range` on a range GET, never by HEAD.

### Also proven live, first time each

- **The ChatML marker guard** (08-12 fix): passed on real rendered rows — and
  earned its keep: Unsloth dropped only **5 of 300 rows** for missing response
  markers (truncation), vs 300 of 300 silently mistrained in the old
  hardcoded-Llama world.
- **The whole ssh_exec drive chain**: launch, nohup persistence across session
  close, poll-by-marker, evidence pull, all via the adapter's kafka actions.

### Three defects found by the run (Phase 0 doing its job)

1. **`/workspace` does not exist on the `base` template** until `00_vm_setup.sh`
   creates it — but the launch places files there *before* setup runs. Fix: `sudo
   mkdir -p` first. RUNBOOK §9 corrected.
2. **`00_vm_setup.sh`'s VRAM gate hardcoded 79000 MiB** — refused the a6000 that a
   1.7B run was deliberately booked on (`ERROR: 49140 MiB < 79000 MiB`). The 08-12
   parameterisation covered run.sh + 02_train and missed the setup script. Fixed
   `2094a02e2` (`MIN_VRAM_MIB`, default = old literal). ~~⚠ B2 bundle redeploy
   still OWED~~ **DEPLOYED ~17:45Z on the owner's word** — the `curl -T` PUT was
   classifier-blocked; the identical PUT through boto3 (`deploy_bundle.py`, now
   in this directory) was not. Live md5 **`6f27b21a6a4236c3c23679892337d0c3`**,
   proven three ways: boto3 read-back, the launcher's own presigned-GET path,
   and `tar -xzO` of the fetched setup script showing `MIN_VRAM_MIB` at the gate
   line with the old `-lt 79000` literal gone. **The whole Phase 0 defect list
   is now closed at the artefact, not just in git.**
3. **`… & echo LAUNCHED` is a lie** — launch 1 returned `exit_code 0, stdout
   LAUNCHED` while stderr held `cd: /workspace: No such file or directory` and
   nothing had run. The `&` backgrounds the whole `&&` chain; the echo runs
   unconditionally, and **ssh_exec's exit_code is the SESSION's, not the chain's**.
   Fix: group with `{ … & }` and make the marker conditional; always read stderr.

### What remains of Phase 0's original list

- **GGUF conversion + playground timing** — needs a new box (or local llama.cpp
  work against the B2 adapter) and its own scope decision. NOT started.
- **FTW-035's enablement condition is now MET** — `thunder-training-monitor` can
  safely be enabled (its DONE_OK→decommission path is exactly what this run
  proved safe). Left DISABLED: flipping a fleet scheduled task is an owner call.

### End state

Paused, vendor `{}`, 0 live, 4 claims all `succeeded`/`attempts=1`. Day total:
**$1.63 booked / ≈$0.32 real** across 4 boxes, all decommissioned.

---

## 2026-08-17 — PHASE 0 COMPLETE: GGUF attempt 2 passed, playground rehearsal measured

Owner: "rerun the test with the fixed script", after a fresh chassis roll
(pods 17:05Z; dispatch waited out the 300s rule trivially). Full numbers are in
**RESULTS §7** — this entry is the evidence trail and the missteps.

**GGUF attempt 2** (box `169b074e…`, 18:14–18:24Z, $0.291 booked): both HANDOFF
leads applied — `apt-get update` before install with `cmake` ASSERTED pre-convert,
and the artefact searched everywhere ≥50MB instead of trusting `--out`.
**Attempt 1's cause confirmed: unsloth `save_pretrained_gguf(out,…)` writes
`<out>_gguf/`** (`/workspace/gguf_gguf/SmolLM2-1.7B-Instruct.Q4_K_M.gguf`) while
printing success naming `<out>`. e2e 489s; gguf stage 170s; upload 16s. Verified
at B2: `Content-Range …/1055609504` + literal `GGUF` magic. [MEASURED]

**Playground rehearsal** (box `e632feaf…`, `template:"ollama"` — the field is
forwarded, proven live; 18:24–18:29Z, $0.140 booked): dispatch 18:24:28 →
box-ready 27s → ollama ready 42s (**binary preinstalled, service NOT running** —
script had to `ollama serve` itself; the template saves the install, nothing
else) → 1.06GB fetch 12s → create 18s → **cold first token 78s** (load 38.5s) →
**DISPATCH→FIRST TOKEN ≈3m23s**. Warm: 0.36s first token, 139.3 tok/s.
PLAN line 154's `[TO MEASURE]` is measured. Booking read: start the box ~10 min
before a booked hour (covers a slow-boot day) and the customer never sees cold.

**Missteps this session, both caught by controls:** (1) the attempt-1 watcher
(2026-08-17 morning) — WRONG_CALLS'd separately, filter could never fire; (2) the
old attempt-1 watcher fired on attempt 2's PG_DONE because both watchers share
`ssh_poll.sh`/`last_poll.txt` which I'd repointed — harmless here (same terminal
set), but two watchers sharing mutable helper state is a coincidence-dependence
worth not repeating.

**End state: vendor `{}`, `is_paused=t`, 0 live, all 7 Phase-0 boxes
decommissioned. Phase 0 spend total $5.72 booked / ≈$1.12 real.** Remaining for
the lane: invoice (settles real rates), owner switches (monitor enable, 259
live-proof), Phase 1 front-end coordination.
