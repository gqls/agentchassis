# SUMMARY — 2026-08-04 — where we are on finetuning.uk actually offering a service

Written in answer to a direct question. **Current state only, measured today
against live code, the live cluster and the live database** — not read off the
concept register, and not carried forward from the older docs in this directory,
two of which turn out to disagree with the question as asked.

---

## The short answer

The website is close. The product is not started, and the older planning in this
folder says it deliberately should not be fine-tuning.

> **UPDATED 2026-08-05 — the "one thing to decide" below has been DECIDED.**
> Owner: **fine-tuning first, RAG second; both get built.** The offer is a
> bounded diagnostic pilot — small data subsets, a choice of small models, modest
> cost, answering "enough data? bigger model? what's missing?" as the scoping
> instrument for a full corporate solution. Recorded at the top of
> `BUSINESS_PLAN_finetuning_uk.md` with corrections marked in place; pricing
> deferred; service setup proceeds first. The section below stands as the state
> of the question when this summary was written.

Three things are true at once and they need separating before any of it is built:

1. **The delivery machinery for a dynamic site now exists and is proven in
   production** — measured today, by the webdesign.uk lane, not by me. This is the
   good news and it is bigger than expected.
2. **The GPU/training half has been idle for seven weeks** and its last recorded
   state is a bug, not a success.
3. **The recorded product decision was RAG, not fine-tuning**, and it was taken
   deliberately, in writing, with a stated reason. That may have changed — but it
   should change on purpose, not by accident.

---

## The one thing to decide before anything is built

`BUSINESS_PLAN_finetuning_uk.md` (last touched 2026-04-21) does not describe a
fine-tuning service. It describes **a RAG platform for SMEs**, and it rules
fine-tuning out explicitly and with reasoning:

> *"Not 'I need to fine-tune a model' — they don't know what that means and mostly
> don't need it."*
>
> *"Fine-tuning only makes sense when tone, format, or offline deployment matters.
> For 'make AI know our stuff', RAG is the honest answer."*
>
> *"Flagship first, bespoke later: RAG platform is the product. Custom AI
> assistants, text LoRA, image LoRA, multi-agent workflows are later tiers added
> as product matures and customer demand surfaces."*

And `FOCUS_finetuning_flywheel_and_service(25).md` (2026-04-23) splits the work in
two and puts the thing being asked for squarely in the second half:

| Concern | What it is | State (as recorded) |
|---|---|---|
| **A. Internal flywheel** | our own pipeline produces training data; we fine-tune local models on it to cut API cost | A (export) + B (RAG) done · C (training) scripted, awaiting first run · D (eval) paused |
| **B. finetuning.uk product** | external users upload data and fine-tune a model through a UI | **"Not started. Questions to answer before scoping."** |

**This is not an objection to building it** — the domain is literally
finetuning.uk, three and a half months have passed, and the owner is entitled to
change direction. It is a flag that the change would be a reversal of a written,
reasoned decision, and the doc that records it has not been updated. **Whichever
way it goes, that plan should be edited to say so**, because otherwise the next
session reads April's conclusion as current and builds toward RAG.

The cheapest version of the question: *is the offer "fine-tune a model on your
data", or "AI that knows your documents" (RAG) — with fine-tuning as a later
tier?* The plumbing overlaps heavily; the front end, the pricing and the sales
copy do not.

---

## The Thunder token

**The path in the request does not exist.** There is no
`~/.config/thunderadapter/token`. The token is at:

```
~/.config/thundercompute/token     65 bytes, modified 2026-08-03 09:52  [MEASURED]
```

which matches "fresh" — it was written yesterday morning.

**The cluster has a key of its own, and I could not verify it is the same one.**
`THUNDER_COMPUTE_API_KEY` is present in the `personae-default-secrets` secret and
reaches `thunder-adapter` via `envFrom` [MEASURED — key name confirmed present].
Comparing it to the local file was **blocked by the tool sandbox**, deliberately
and correctly, so:

> **[UNVERIFIED] Whether the cluster's key is the fresh one is unknown.** This is
> the single most likely reason for a "the token is fine, why isn't it working"
> confusion later. It needs one deliberate step, by you or with your say-so.

What *is* known: the adapter came up clean this morning (08:08), read its config,
connected to the DB, initialised object storage against the
`personae-model-training` bucket, and reported no auth failures. But it has also
made no Thunder API calls since restarting, so **a clean log is not evidence the
key works** — nothing has exercised it.

The honest test is one API call, not a log read.

> ### Incidental, unrelated, and worth someone's attention
> The thunder-adapter logs its B2 credentials **in plaintext at INFO level** on
> every startup — `DEBUGaa: B2_APPLICATION_KEY from env` followed by the actual
> secret, in `storage/s3.go:32`. Anyone with log access has the storage keys.
> Nothing to do with this task; found while reading for the token. Worth a
> separate fix.

---

## The GPU / training half

| fact | value | source |
|---|---|---|
| Thunder instances ever created | 23 | `thunder_instances` [MEASURED] |
| Currently running | **0** — all 23 `decommissioned` | same |
| Last instance created | **2026-06-18** — seven weeks ago | same |
| Daily spend cap | $30, max 2 concurrent, not paused | `thunder_config` |
| Training bucket | `personae-model-training`, wired and reachable | adapter startup log |
| `thunder-training-monitor` scheduled task | **disabled**, never triggered | `scheduled_tasks` [MEASURED] |
| `thunder-reaper` | enabled, every 900s, ran 09:52 today | same |

So the machinery provisions and decommissions GPU boxes and has done so 23 times.
What it has not done recently is anything at all. The last substantive work in this
folder is `phase5`, whose newest documents are about a **checkpoint upload race**
(`HANDOFF_2026-06-06_checkpoint_upload_loop_await_race`,
`CONTEXT_PACK_thunder_checkpoint_race`) — i.e. the lane stopped on a bug, not on a
finish line.

**[UNVERIFIED] Whether a training run has ever completed end to end and produced a
usable adapter.** I did not find a table recording training runs or their outputs,
and I have not gone through the phase5 notes far enough to answer it from prose.
That is the first thing to establish before promising anyone a fine-tune, and it is
a contained piece of work.

---

## The front end — and this is the genuinely good news

**Measured today by the webdesign.uk lane, independently of this question.** Read
`webdesign_uk_build_service/SUMMARY_2026-08-04b_dynamic_site_capability.md` — it is
the current authority and I am not restating it here beyond the load-bearing part:

> *"The framework already builds and deploys websites onto real servers, in
> production, today. What it does not do is write the server-side code. That is the
> actual line, and it is much further along than 'static sites only'."*

Concretely, against the request "designed, prepared, hosted, maintained, deployed
through the framework and not manually through this CLI":

| piece | through the framework? | evidence |
|---|---|---|
| **Design** the site | **yes** | the whole build pipeline; and per the 2026-08-04 owner ruling, this is now mandatory — no hand-built sites |
| **Prepare / build** pages | **yes** | proven fleet-wide |
| **Deploy to a real server** (not a bucket) | **yes** | per-site switch: `sites.github_repo='vm-sites'` routes deploys to a box. relojistas.com, 20 pages, deployed 2026-08-04 |
| **Maintain** — detect and repair drift | **yes**, with the caveat below | the improvement loop, now understood and drivable per-site |
| **Monitor** the box | built, **not switched on** | `check_backend_unreachable` probes `/health` on `deploy_config.target='vm'` sites; enablement tracked in `bugs_open/149` B1/B1a |
| **Generate the backend service** (upload, auth, job control, billing) | **no — and nobody does** | `site-engine` is one hand-written Go binary, same on every box; register DYN-001 tier 2 is aspirational |
| **Provision the box** | **no — by ruling, not by gap** | the estate plan's owner ruling is "merge the generator, not the trust boundary"; estate P1–P5 all pending |

**So the answer to "can the framework host it dynamically" is yes for everything
except the part that makes it a product.** The pages, the design, the deploys and
the upkeep are all framework work. The service behind `/api/…` — the thing that
takes a customer's upload, starts a Thunder box, tracks a job, returns an adapter,
and charges for it — is hand-written Go today, once, and there is no mechanism that
writes it per site.

That is not a blocker so much as a scoping fact: **one backend service, written
once, living in this repo, deployed to one box by the same machinery relojistas
already uses.** It is the same shape webdesign.uk is doing this week for its chat
service, which means the second one is much cheaper than the first — and it argues
strongly for *following* that lane rather than starting a parallel one.

---

## Where we are, in one table

| layer | state |
|---|---|
| **Website (marketing)** | live, and as of yesterday no longer visibly broken — 0 broken images, icons render, first dispatchable work queue since April |
| **Website (content quality)** | poor and known: no case studies with real outcomes, no named people, one high-commitment CTA, 17 phantom links, 28 misdirected CTAs — all filed, all queued |
| **Product front end** | **not started**; the delivery path for it is proven and reusable |
| **Backend service** | **not started**; must be hand-written once; a sibling lane is writing an analogous one this week |
| **GPU / training** | built, exercised 23 times, **idle 7 weeks**, last state a checkpoint-race bug, end-to-end success **unverified** |
| **Product decision** | **contested** — the written plan says RAG, the request says fine-tuning |
| **Thunder token** | fresh locally at a different path than given; cluster key present but **unverified as matching** |

---

## Where we're going — ordered, and each step cheap enough to stop after

1. **Settle the offer** (RAG vs fine-tune vs both) and edit
   `BUSINESS_PLAN_finetuning_uk.md` to say what was decided. Everything else
   branches off this and it costs a conversation, not a build.
2. **Prove the Thunder key works** — one authenticated API call, and reconcile the
   cluster secret with the fresh local token if they differ.
3. **Establish whether a training run has ever finished.** If phase5 stopped on the
   checkpoint race, that bug is the first engineering task, and it is a
   prerequisite for promising the service to anyone.
4. **Do not start a parallel VM lane.** webdesign.uk is standing up the identical
   pattern this week (framework-built pages + one hand-written service on one box,
   behind a Cloudflare tunnel). Follow it, reuse its scripts and its estate
   profile, and let it take the first-time costs.
5. **Then** scope the product front end as a framework build like any other site
   section, with the backend as a named, hand-written service — sized honestly, in
   the open, rather than assumed to be generated.

**Not recommended:** building the front end first because it is the visible part.
It is the one piece the framework can already do well, which makes it the cheapest
to do last, once there is a service for it to talk to.
