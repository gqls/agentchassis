# 455 — the Gemini image provider's prepayment credits are depleted; ALL image generation fleet-wide is blocked, not just logos

**Filed 2026-09-03 by the `bugs_open/424` lane, found incidentally while watching a logo
retry.** Not run through `090` — the root cause is asserted by the provider itself in the error
body, the same shape `bugs_open/202`/`243` were filed on the same grounds. What's open is the
owner-level decision (below), not a mechanism to diagnose.

## Symptom, verified at the adapter log, not inferred

`internal/adapters/imagegenerator/banana` (the Gemini image provider — `model=gemini-3-pro-image-preview`):

```
banana api: POST /models/gemini-3-pro-image-preview:generateContent returned 429:
{ "error": { "code": 429,
  "message": "Your prepayment credits are depleted. Please go to AI Studio at
  https://ai.studio/projects to manage your project and billing..." } }
```

First (only, as of filing) occurrence: `2026-09-03T10:31:01–02.819/830Z`, on
`designblog.co.uk`'s `needs_imagery:site:-:logo` retry (`bugs_open/424`'s remediation), one pod
(`image-generator-adapter-985f5f66d-2tqfn`), 4 log lines within ~1 second (the client's own
internal retry, not 4 separate attempts). The sibling pod shows nothing — consistent with one
request landing on the exhausted account regardless of which pod sent it (shared credential, not
a per-pod state).

## This is NOT `bugs_open/202` or `bugs_open/243`

- **202** (2026-08-05, still open): `page-content-writer`'s **text** model
  (`gemini-pro-latest`/`gemini-3.1-pro`) hitting a **daily RPD quota cap** (Tier 1, 250/day) —
  resets, or fixed by Tier 2 / a rate-limit-increase request. A quota, not a payment failure.
- **243** (2026-08-10, closed): the **Anthropic** (Claude) account's usage limit, resolved same day
  by the owner adding credit.
- **This**: the **image** model (`banana` provider, Gemini image family), and the error is
  explicitly a **prepayment/billing** message — "credits are depleted" — not a rate/quota message.
  Different product, different account state, different remedy shape (top up, not wait-for-reset).

## Scope `[MEASURED 2026-09-03 10:34 UTC]`

```sql
SELECT count(*), domain, max(occurred_at) FROM agent_error_log
WHERE error_message ILIKE '%prepayment credits%' GROUP BY domain;
```
**UPDATED 10:41 UTC — CONFIRMED ONGOING, not a blip.** `gamedesign.uk` (a second, fully
independent site — different work item, different request, ten minutes after `designblog.co.uk`'s
first hit) failed with the byte-identical error at `10:41:27Z`. Two independent sites, ten minutes
apart, same wall:
```sql
SELECT count(*), domain, max(occurred_at) FROM agent_error_log
WHERE error_message ILIKE '%prepayment credits%' GROUP BY domain;
--   1 | gamedesign.uk    | 2026-09-03 10:41:27Z
--   4 | (unattributed)   | 2026-09-03 10:41:27Z
--   1 | designblog.co.uk | 2026-09-03 10:31:04Z
```
**This is not resetting on its own between attempts. Treat as a live outage of Gemini image
generation fleet-wide, needing owner action (add credit at AI Studio), not a transient blip to
wait out.** Both of `bugs_open/424`'s remaining remediation retries (`designblog.co.uk`,
`gamedesign.uk`) are currently blocked by this, not by anything in that fix — they will keep
retrying and keep hitting this same wall until credit is restored.

**UPDATED 10:43 UTC — THIRD independent site, third confirmation.** `boxingonline.com`'s own logo
run (`d71b7877`, `site_delivery_and_editor`'s lane, unrelated to `bugs_open/424`'s three) was
picked up 10:43Z and failed with the byte-identical error BEFORE reaching the matte at all.
`site_delivery_and_editor` is escalating to the owner directly. **Any further failures on
`designblog.co.uk`/`gamedesign.uk`'s remaining attempts must not be read as `bugs_open/424`'s
guard refusing them** — until this is resolved, assume every image-generation failure fleet-wide
is this, not a content problem, and check the error text before attributing a failure to any
other cause.

## RESOLVED same day, ~11:08–11:41 UTC — matches the `bugs_open/243` pattern exactly

`[MEASURED 2026-09-03 11:45 UTC]` `agent_error_log` now shows **30** `prepayment credits` rows
total (up from 3 at first measurement), **last occurrence `11:08:06Z`, none since.**
`gamedesign.uk`'s `bugs_open/424` remediation retry reached the model and genuinely SUCCEEDED at
`11:41:02Z` — a real generation, verified independently at the served bytes (fresh key date,
colour type 6, 100% of the border ring transparent) — which could not have happened while credits
were still depleted. **Outage window: `10:31:01Z` (first observed) → `11:08:06Z` (last observed) →
confirmed cleared by `11:41:02Z`, roughly 37–70 minutes**, resolved without this session's
intervention (no billing action available from here) — matches `bugs_open/243`'s own pattern
exactly: the owner (or someone with account access) topped up credit rather than it resetting on
a quota clock, most likely prompted by `site_delivery_and_editor`'s direct escalation at ~10:43Z.
Leaving this file OPEN rather than closing it: unlike `243`, no one has yet confirmed IN THIS FILE
that a deliberate top-up happened (only inferred from the traffic gap) — close once that's
confirmed, and consider whether this warrants the same prevention discussion `243`'s own residual
raised (recurring provider-credit exhaustion, third instance counting `202` and `243`).

## What this means for `bugs_open/424`'s own remediation

`designblog.co.uk`'s retry hit this, not a matting failure — its `error` column now reads the 429
message, not a `border_keyed` refusal. `gamedesign.uk`'s next attempt (due ~10:35Z) has not fired
yet as of filing; it may hit the same wall. **Do not read a 429 here as evidence against the 424
fix** — it is a different, unrelated failure mode at the provider/billing layer, upstream of
anything `bugs_open/424`'s code touches.

## The decision needed (owner-level, per the 202/243 precedent — not taken unilaterally)

1. **Add credit to the Gemini/AI Studio project** — restores service, matches how `243` was
   resolved for Anthropic same-day.
2. **Wait and see if it self-resolves** — "prepayment" (as opposed to a quota) does not obviously
   reset on its own the way a daily RPD cap does; treat this as needing action, not time, unless
   proven otherwise.
3. **Re-drive affected items after credit is restored** — `bugs_open/424`'s two pending retries
   (`designblog.co.uk`, `gamedesign.uk`) will retry on their own via `retry_after`; anything else
   fleet-wide that failed the same way and is not on an automatic retry ladder needs to be found
   and re-triaged by hand.

## Verify

```sql
-- fleet-wide scope, not just the two pods this session read
SELECT count(*), max(created_at) FROM agent_error_log
WHERE error_message ILIKE '%prepayment credits%' OR error_message ILIKE '%code.:.429%';
```
Re-run the adapter-log grep across a longer window and all pods once credit is restored, to date
the recovery boundary the way `243` did.

## Related
- `bugs_open/202` (Gemini text quota — different product, different remedy).
- `bugs_open/243` (Anthropic usage limit — same PATTERN, resolved by owner adding credit same day).
- `bugs_open/424` (the lane that surfaced this; not the cause, and not blocked in its own code by it).
