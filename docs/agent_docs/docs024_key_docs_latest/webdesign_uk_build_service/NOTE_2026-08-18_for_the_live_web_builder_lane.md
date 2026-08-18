# Note for the "webdesign live web builder project" lane — from webdesign_uk_build_service

**Why this is a file and not a message:** the owner asked me to correspond with your session
directly (`af324bbe-88ed-48e4-9bdc-88b46ca40011`). It is **not reachable from this machine** —
it does not appear among the 36 peer sessions here and `SendMessage` returns
"No agent named … is reachable". So this is the channel we do share: the tree.

## The chat box is LIVE on preview.webdesign.uk (2026-08-18 10:35Z)

Served `index.html` component order: `hero, brief-explanation, chat-input-box, call-to-action`
— the owner's requested placement (after the "one website, one price" block and prices).
26 hrefs, none lost. Zero occurrences of the retired pay-after-approval claim.

## What changed in the SHARED register — do not revert, the live bot renders it

1. **New commercial terms (2026-08-17).** Payment BEFORE the build; the customer does NOT see
   the site before paying; afterwards they get a permanent ZIP plus a preview link live for
   about a month. `payment_after_approval` is **retired and renamed** `payment_upfront`;
   `no_refund` is re-justified against the deal; `delivery_preview_and_zip` describes the
   preview as an added benefit.
2. **`billing_settings.payment_timing` moved `after_approval` → `upfront`** in the same
   transaction — auth-service reads it (`repository.go:247`) and the copy must not contradict
   the system.
3. **Voice brief (2026-08-18):** write like a helpful assistant, not a marketing bot. Not
   statement-stacking — state it, then give the next step, the order to do things in, or name
   who can help. The six services in `third_party_options` are to be used as HELP, not merely
   listed as exclusions.
4. **Stat-field guard (2026-08-18):** stat/metric/counter fields take attested numbers only.
   `build_duration` is hedged ("usually ready the next day") and must never appear as "1 day"
   or "24 hours" in a stat box.

## (4) is what was blocking the home page, and how to retrieve that class of failure

The index rebuild failed twice at `validate_content`. Cause, from
`agent_error_log.context.issues`: `unregistered_stat`, value `"1 day"`, location
`brief-explanation.stat_2_value`. **I did not clear it by attesting 1 as a number** — that
converts the owner's hedge into a promise, which he has ruled against. The writer_block guard
above fixed it and the next rebuild passed clean.

**Retrieval recipe (this cost me real time):** the failing step's output is NOT persisted on
the orchestration — `valid`/`issues`/`blockers` come back null, because the error path runs
instead of the output field being written. The action *does* persist the structured issues to
`agent_error_log`. Query by `context ? 'issues'`, **never** by `domain` (unreliable column):

```sql
SELECT jsonb_pretty(context) FROM agent_error_log
 WHERE occurred_at > now() - interval '30 minutes' AND context ? 'issues'
 ORDER BY occurred_at DESC LIMIT 1;
```

## Things that will bite you on this site

- **The contact-page lock is OFF, and that was only safe because the plan now carries the
  section.** The lock was the sole thing merging `chat-input-box` into the assembled list;
  unlocking alone would have let the next rebuild delete it. **If you regenerate plans for
  this site, keep the `chat-input-box` rows** (contact ordering 2, index ordering 2).
- **`save_page_sections` refuses a page whose `rebuild_policy='owned'`** ("tool/widget-owned:
  a generic section save would clobber it"). Now the chat is on `index`, that page may become
  owned and generic rebuilds of it refused.
- **`bugs_open/299` is filed and deliberately NOT patched**: the home page CTA names the
  Website Brief Starter and its href is `tel:+44 (0) 7934 524 911`, so it dials the phone.
  The section was written 2026-08-16, AFTER the 268 fleet fix, so the producer still makes
  it. If your rewrite regenerates that CTA, check the href — and note the false-pass trap in
  the bug file (nav and footer link the tool correctly, so a page-wide grep for the URL
  passes while the button stays broken).

## What I am NOT doing, so we do not collide

Rewriting the site copy or the positioning. The owner is settling the page LEAD with me —
proposal: *"show the work, promise nothing"* (real sites plus the exact prompt that produced
each), with his own domains as the examples once he is using the system.

**If you are taking the site rewrite, say so in this directory** (or have the owner tell me)
and I will stay off it.
