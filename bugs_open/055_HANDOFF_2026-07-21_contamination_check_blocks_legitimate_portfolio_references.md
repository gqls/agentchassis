# BUG 055 — content-validation contamination check blocks legitimate cross-site references

**Filed:** 2026-07-21 · **Status:** OPEN · **Severity:** high (blocks a whole
site from going live) · **Class:** cross-cutting / platform-wide validation gate
**Found by:** brochure_component_library workstream (fundamentallyai.com go-live)

## Symptom

Every content page of a new site that is *about our own platform*
(fundamentallyai.com) is held at `needs_human_review` with:

```
step validate_content failed: failed to execute action validate_page_content:
content validation failed: 1 blockers, 0 errors
```

5 content pages blocked (`index`, `about`, `capabilities`,
`multi-agent-review-council`, and earlier `model-fine-tuning`); `contact` and
`model-fine-tuning` eventually reached `deployed` — but only by *dropping the
blocked content* on a retry (see "Second-order damage" below).

## Root cause — CONFIRMED, self-evidencing

`checkDomainContamination` in
`platform/orchestration/actions/validate_page_content.go:481-534` flags any
occurrence of a **known site's domain** appearing in another site's copy as a
`blocker` (severity set at line 508). The known-site list is **hardcoded**
(lines 484-493) and includes `leopardessconsulting.co.uk`.

fundamentallyai.com's mission brief
(`docs/.../brochure_component_library/MISSION_BRIEF_fundamentallyai_2026-07-20.md`
lines 46-49) **deliberately and by owner approval (2026-07-20)** names
`leopardessconsulting.co.uk` as the worked example of the platform's
self-correction / claims-verification story. So the anti-leakage check fires on
a *legitimate, intentional* reference. The check has **no per-site allowlist** —
it assumes no site ever legitimately references another of our sites. That
assumption is false for a **portfolio / meta site**, which fundamentallyai is by
design (the same brief also cites idea.uk and relojistas.com as case studies).

### Evidence (live DB, 2026-07-21, chassis v1.0.1144)

The blocker detail **is persisted** — in `agent_error_log`, not in
`site_work_items.result` (which is `{}`). `ValidatePageContentAction` writes a
structured sibling row (`writeValidationFailureLog`, lines 344-420) with
`error_code = 'CONTENT_VALIDATION_BLOCKER_DETAIL'` precisely so post-mortems
don't need pod logs. All 9 such rows for this site carry exactly one issue:

```
type: cross_site_domain
value: leopardessconsulting.co.uk
category: contamination · severity: blocker
description: Found domain 'leopardessconsulting.co.uk' in content for 'fundamentallyai.com'
location (snippet): "invented details on leopardessconsulting.co.uk. Our
                     verification system flagged it; we corrected it."  ← the story, as designed
```

Query to reproduce:
```sql
SELECT occurred_at, context->>'page_name' AS page, context->'issues' AS issues
FROM agent_error_log
WHERE site_id = (SELECT id FROM sites WHERE domain='fundamentallyai.com')
  AND error_code = 'CONTENT_VALIDATION_BLOCKER_DETAIL'
ORDER BY occurred_at DESC;
```

Each row has **exactly one** issue and the action writes *every* blocker/error
to that row — so leopardess is the sole blocker on every page. Unblocking the
leopardess reference unblocks all 5 content pages.

## Second-order damage — the gate silently strips owner-approved content

`contact` and `model-fine-tuning` reached `deployed`, but **no saved
`page_components.rendered_html` on ANY of the site's 9 pages mentions
leopardess** (verified: `count(*) FILTER (WHERE rendered_html ILIKE '%leopardess%')`
= 0 across all pages). `model-fine-tuning` was blocked once for the leopardess
reference (row at 2026-07-20 23:49) then re-generated *without* it and passed.
So the flagship self-correction narrative — the whole reason the site names
leopardess — is **absent from the live site**, not merely blocked. A page that
keeps the story stays blocked; a page that drops it deploys generic. This is
worse than a hard block: it lets a degraded page through looking "done."

## Fix candidates

1. **Per-site allowlist of referenced domains (recommended).**
   `checkDomainContamination` consults a per-site allowlist; a known-site domain
   on the allowlist is skipped, not flagged. Absent/empty allowlist → current
   behaviour unchanged, so **zero regression for every existing site** (opt-in).
   Storage options (decide at build; reuse existing machinery per CLAUDE.md):
   - a `site_specs` row (e.g. a `content_policy` aspect, loaded like
     `loadEvidenceBase` at line 686), or
   - a well-known key on `sites.content_data` / `identity` spec.
   Live-editable data either way — only the *reading* code needs an image roll.
   Seed fundamentallyai's allowlist with its portfolio (leopardessconsulting.co.uk
   now; extensible to idea.uk / relojistas.com / finetuning.uk as case studies
   land).
2. **Downgrade `cross_site_domain` from blocker → error/warning.** Cheaper but
   weakens genuine leakage detection fleet-wide and still routes to human review;
   rejected as primary — it removes the guard for the accidental-leak case the
   check exists for.
3. **Source `knownSites` from the DB (all 11 sites) instead of the hardcoded 5.**
   Orthogonal improvement (the detector is currently blind to 6 of our sites),
   but *expands* what's flagged → bigger blast radius. Do as a follow-up, paired
   with candidate 1, not instead of it.

The fix touches `platform/` → **council-review before commit** (CLAUDE.md).
Go change → inert until an image rebuild + roll.

## How to verify the fix

1. After building the fix and seeding fundamentallyai's allowlist, re-fire one
   blocked page (e.g. `about`) and confirm it passes `validate_page_content`
   with the leopardess reference **present** in the saved `rendered_html`.
2. Confirm no new `CONTENT_VALIDATION_BLOCKER_DETAIL` row appears for it.
3. Behavioural regression check: a *different* site with an accidental
   cross-site domain must still be flagged (the allowlist is opt-in per site).
4. Re-build the 2 generic "deployed" pages so the story is restored fleet-wide
   on this site.

## Notes / landmines

- The handoff (`HANDOFF_2026-07-21_start_here.md`) said the blocker reason was
  "NOT recoverable from the DB" and prescribed a live re-fire to capture it. That
  premise was **wrong** — it checked only `site_work_items.result`; the detail
  was in `agent_error_log` all along (see WRONG_CALLS 2026-07-21). No re-fire /
  queue-wait was needed to diagnose.
- Blocker 2 (nothing serves) is **separate** — `page_components.deploy_commit`
  is empty even on the "deployed" pages, so the git-adapter never pushed
  rendered pages to the portfolio deploy repo for this new domain. Not this bug.
