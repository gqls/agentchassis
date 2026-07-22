# BUG 056 — page regeneration silently drops the content that tripped a validation blocker, with no record

**Filed:** 2026-07-21 · **Status:** OPEN · **Severity:** medium (silent loss of
owner-intended content) · **Class:** cross-cutting / build-loop integrity
**Found by:** council-gate review of bugs_open/055 (bug_historian + editquality
seats, corr `03908b72-2471-474e-baaf-7952d1903460`, round 1) + this workstream.
**Mechanism is [INFERRED] — needs diagnosis (090) before any fix.**

## Verified effect

On fundamentallyai.com, `validate_page_content`'s contamination check blocked
every page that named `leopardessconsulting.co.uk` (the owner-approved
self-correction story — see bugs_open/055). Two of those pages
(`model-fine-tuning`, `contact`) nonetheless reached `build_status='deployed'`.

**[VERIFIED, live DB 2026-07-21]:**
- `model-fine-tuning` has an `agent_error_log` blocker row at 2026-07-20 23:49
  (`cross_site_domain` / leopardessconsulting.co.uk) — it *did* generate the
  story once — yet later reached `deployed` (03:41).
- `count(*) FILTER (WHERE rendered_html ILIKE '%leopardess%')` = **0** across all
  9 pages' saved `page_components`. The flagship narrative is absent from the
  entire site — not merely blocked.

So a page can generate owner-required content, have it blocked, then deploy a
*later* generation that silently omits that content, and **nothing records that a
required element was present-then-lost**. A degraded page passes looking "done".
This is the "trust the artefact, not the status" failure at the build-loop level.

## Inferred mechanism (NOT yet confirmed — do NOT assert without diagnosis)

Pages are re-rendered by multiple triggers (e.g. the `needs_page` "Re-render X
after its image asset landed" items observed here). Each rebuild generates fresh
content non-deterministically. The likely shape: when a generation trips a
blocker the page is held at `needs_human_review`, but a subsequent rerender
generates a version that happens to omit the flagged element and passes — because
omitting it removes the trigger. There is no diff/guard comparing a newly-passing
generation against a previously-blocked one to notice "the thing that was blocked
is now simply gone", and no work item recording the loss.

This is distinct from bugs_open/055: 055 is the *over-broad trigger* (a false
positive on a legitimate reference); 056 is the *downstream loss* (regeneration
discards whatever tripped any blocker — false positive OR a true positive a human
would have wanted resolved differently — with no record). 055's fix removes one
trigger; it does **not** touch this mechanism, which stays generic and will fire
for the next site that hits any content-tripping blocker.

## Why file separately / do not fix in 055

Fixing 055 (+ seeding the allowlist + regenerating the pages) resolves *this
site's* immediate incident, because with the blocker gone the story-bearing
generation will pass and be saved. But the silent-loss mechanism is a separate,
platform-wide defect in the regeneration path, in different code, and its precise
mechanism is inferred, not read. Per CLAUDE.md ("confident structural claim from
grep hits" is the classic trap), this needs the diagnosis loop, not a hand-fix.

## Suggested next step

Fire `090_TRIGGER_needs_diagnosis_v1.sh` stating the mechanism (regeneration
path discards content that tripped a validate_page_content blocker without
recording the loss; evidence: 2 pages deployed missing content an earlier
blocked generation contained) and pointing at the page-build / rerender
orchestration + `validate_page_content` failure handling. Candidate fix
direction (for the diagnosis to confirm/refute): a fail-loud guard that, when a
newly-passing generation is missing an element a prior blocked generation for the
same page contained, raises a work item instead of silently deploying.

## How verified / queries

```sql
-- the story-bearing generation existed then vanished:
SELECT context->>'page_name', occurred_at FROM agent_error_log
WHERE site_id=(SELECT id FROM sites WHERE domain='fundamentallyai.com')
  AND error_code='CONTENT_VALIDATION_BLOCKER_DETAIL' ORDER BY occurred_at;
-- yet nothing deployed mentions it:
SELECT p.name, p.build_status,
       count(*) FILTER (WHERE pc.rendered_html ILIKE '%leopardess%') AS mentions
FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id=(SELECT id FROM sites WHERE domain='fundamentallyai.com')
GROUP BY p.name, p.build_status;
```
