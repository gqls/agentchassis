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
**3 rows total, fleet-wide, all within `10:31:02.871Z`–`10:31:04.255Z`, all `designblog.co.uk`**
(one row unattributed to a domain). So as of this measurement it is a single incident on a single
site, not (yet) proven to be blocking the rest of the fleet — nothing else has hit the same wall
in the ~3 minutes since. **Whether it is a momentary blip or an ongoing depletion is still open**:
`gamedesign.uk`'s own next retry is due `~10:35:34Z` (`bugs_open/424`'s other pending remediation)
and will be a second, independent data point within minutes — check `agent_error_log` again then
before concluding either way. If it recurs there or anywhere else, treat as ongoing and escalate
to the owner immediately rather than waiting for more data.

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
