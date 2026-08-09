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
