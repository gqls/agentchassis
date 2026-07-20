# 008 FEATURE — halve image-generation cost via the Gemini Batch API

**Filed:** 2026-07-20, from the `bugs_open/011` R1 provider-routing thread, at the owner's
instruction to park the decision rather than take it now.
**Status:** OPEN — **deliberately deferred, not rejected.** The owner approved the provider
change and read the costing; this is the lever we chose not to pull yet.

## The owner's framing

> "we can leave the half price batch api decision"

Deferred because the sums are currently small enough that it does not matter. It becomes
worth doing when volume climbs — see the trigger below. This file exists so that when
someone asks "can we cut the image bill?", the answer and its unknowns are already here
instead of being re-derived.

## The opportunity

Gemini 3 Pro Image is billed at **$0.134 per 1K/2K image** on the standard API and
**$0.067 via the Batch API — exactly half** ([Google's pricing
page](https://ai.google.dev/gemini-api/docs/pricing)). The batch tradeoff is asynchronous
submission with completion typically within 24 hours.

**Why that tradeoff looks nearly free here, and this is the non-obvious part:** our image
pipeline is *already entirely asynchronous*. Generation is fired by Kafka messages from
discovery sweeps and work items; nothing and nobody blocks on an image while it renders.
A page that lacks its hero simply renders without it and re-renders when the asset lands —
that flow already exists and is exercised routinely (`page_rerender` with
`reason='image_landed'`). So a ≤24h turnaround costs us a property we are not currently
using.

Since `bugs_open/011` R1, **every declared kind routes to Banana**, so this discount would
apply to essentially all imagery the platform generates, not a subset.

## What it would be worth (2026-07-20 figures)

| | now | with batch |
|---|---|---|
| Fleet image bill | ~**$14.50**/month | ~**$7.25**/month |
| 89 undrained planned heroes | ~$11.93 one-off | ~$5.96 one-off |

Volume it is measured against — from `assets`, excluding `derived-*`:

| month | heroes | all images |
|---|---|---|
| 2026-05 | 8 | 22 |
| 2026-06 | 15 | 46 |
| 2026-07 (to 20th) | **40** | **108** |

**Saving today: about $7 a month. That is the honest reason this is parked.**

## The trigger — when to stop deferring

**Watch the slope, not the total.** Generations roughly doubled each month (22 → 46 → 108).
At **10× July's volume the bill is ~$145/month** and batch saves ~$72/month, at which point
the engineering below pays for itself quickly. Suggested review point: **when monthly
generations exceed ~500**, or when any single planned sweep exceeds ~200 images.

Cheap standing check:

```sql
SELECT date_trunc('month', created_at)::date AS mth, count(*)
  FROM assets
 WHERE origin_model IS NOT NULL AND origin_model NOT LIKE 'derived%' AND origin_model <> ''
 GROUP BY 1 ORDER BY 1 DESC;
```

## Open questions — the real work is here, and none of it is a setting

1. **Can our code even submit batch jobs?** **UNVERIFIED, and this is the gating question.**
   `internal/adapters/imagegenerator/banana/api` speaks the synchronous
   `contents/parts` wire format. Batch is a different submission/collection shape. Nobody has
   looked at whether our client can express it.
2. **The adapter is request/response over Kafka.** `generateImage` blocks on the provider call
   and replies on the reply-to topic; the whole saga step awaits that response. Batch breaks
   that shape — you submit, get a handle, and collect later. That is **not** a provider-level
   change; it needs a submit-then-poll/callback path and a place to park the handle. This is
   the bulk of the work.
3. **What happens to `store_*_asset`?** Today the orchestration step that persists the asset
   runs off the adapter's synchronous reply. Under batch it needs to run on collection
   instead.
4. **Do we want it per-kind?** A logo (generated once, human-approved, then locked) is a fine
   batch candidate. A hero someone is waiting to review during an active site build may not
   be. The per-site/per-kind precedence machinery from 011 R1
   (`imagery_style_guide` → `provider_hint`) is the obvious place to express such a choice,
   and would extend naturally to a `batch: true` flag.
5. **Does batch change the model's output?** Assume not, but it is worth one A/B before
   committing the fleet — the D13/D14 history says style consistency is fragile and
   expensive to rediscover.

## Related

- `bugs_open/011` §6 — the full costing, sources and caveats this file summarises, plus the
  note that **the platform records no cost data at all** (`llm_call_log` covers text calls
  only). Fixing that is a prerequisite for ever *proving* a saving rather than projecting it.
- `docs/agent_docs/docs024_key_docs_latest/imagery/HANDOFF_2026-07-20_provider_routing_011.md`
  — resume point for the routing thread.
- `bugs_open/028` — Banana discards negative prompts; unrelated to batch but the same
  provider path, so anyone opening `banana/` should read both.
