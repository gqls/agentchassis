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

---

## Progress (2026-07-22, bugfix-056 session)

**The suggested 090 is now IN FLIGHT: corr `b361298a-e030-456e-956f-adf1e05503b1`**
(item_key `needs_diagnosis:regen-drops-blocked-content`, seed scope
`validate_page_content.go` + `rerender_single_page_action.go`, runtime site
fundamentallyai.com, pages model-fine-tuning + contact).

It could not be fired earlier: **the diagnosis loop itself was broken** for any
run reaching the code tier — the OTHER bug numbered 056
(`056_HANDOFF_2026-07-21_diagnose_route_step_nul_byte_kills_jsonb_persist.md`,
the ambiguous-number trap in person). That bug is now FIXED & LIVE (commit
`7a9f5f652`, prod v1.0.1149 pod-verified 2026-07-22); this diagnosis run doubles
as its end-to-end verification (≥2 iterations must advance past `route`).

Dispatch notes: the 090 coverage probe refused first (open items touching the
target pages) — read and overridden with FORCE=1 because none is a competing
fix: `needs_page:model-fine-tuning` at `needs_human_review` is this incident's
own residue (evidence, not work), and the `contact` hits are parked design-audit
items on other sites matched by page name. The 055 session's in-flight
allowlist fix narrows the TRIGGER, not this loss mechanism, and the evidence
(agent_error_log rows, deployed page_components) is historical and stable.

**Next:** read the verdict (`diagnosis_artifacts` on the corr above), then
implement whatever fix it confirms — candidate direction unchanged (fail-loud
guard comparing a newly-passing generation against a previously blocked one for
the same page, raising a work item instead of silently deploying).

---

> **CORRECTED 2026-07-22 — the inferred mechanism was REFUTED by the diagnosis
> loop (corr `b361298a`, 5 iterations, verdict REFUTED, stopped
> scope-not-narrowing). Caught by: the 090 run this file itself demanded.**
> Two parts of the inference were wrong, on citation:
> 1. **"No record of the loss" is false.** Every blocker is captured verbatim in
>    `agent_error_log` (`CONTENT_VALIDATION_BLOCKER_DETAIL`, naming the exact
>    flagged value `leopardessconsulting.co.uk`, category, description), and
>    validate_content's `error_step: mark_needs_review` parks the work item at
>    `needs_human_review`. The record exists; what is missing is reconciliation.
> 2. **The rerender path was the wrong suspect.** `RerenderSinglePageAction` /
>    `assemblePage` / `rerenderLoadSections` never generate or validate content —
>    pure reassembly of stored `page_components` ("Simple concatenation - no
>    template re-rendering", the file's own header). The clean deploy came from
>    the page-build pipeline itself: `page_component_history` row sourced
>    `save_page_sections_overwrite` at 2026-07-21 03:40:44 → `pages.deployed_at`
>    03:41:45.

## Corrected mechanism (grounded 2026-07-22)

The gap is **review-bypass-by-sibling-item**, in the build pipeline:

1. Item A (`needs_page:model-fine-tuning`) hits the blocker →
   `mark_needs_review` parks A at `needs_human_review` (**still parked, live DB:
   updated 2026-07-20 21:27, never completed**). Blocker detail recorded. ✓
2. Item B (`page_rerender:model-fine-tuning`, fired when an image asset landed)
   runs the SAME pipeline for the same page; generation is non-deterministic;
   the fresh copy omits the flagged element; validation passes; `save_sections`
   → **deployed 2026-07-21 03:41:45**. B completes.
3. Nothing connects B's success to A's parked review: `save_page_sections_action.go`
   has zero references to blocker records or review state (grep verified
   2026-07-22); `complete_work_item`'s preserve-the-flag guard
   (`load_work_item_actions.go:792-808`) protects only the flagged item ITSELF.
   The human review A demanded never happened; the page shipped without the
   contested element; A dangles as queue noise.

## Fix direction (for council review — NOT yet implemented)

Fail-loud reconciliation at the point the superseding save lands
(`save_page_sections_action.go`), NOT a dispatch-time block: blocking dispatch
behind a parked review would wedge pages indefinitely while the review queue has
no working drain (bugs_open/033), the 029-class risk. Shape: after a successful
save for page P, if P has a sibling item parked `needs_human_review` or an
unresolved `CONTENT_VALIDATION_BLOCKER_DETAIL` newer than P's last deploy,
check the newly-saved content for each previously-flagged value; write a loud
`agent_error_log` record (present→absent = "the element that tripped review was
dropped, not resolved") and annotate the parked item's `result` (never its
status — a human still owns the flag). Whether a parked review should HOLD
deploys outright is an owner policy call (033 interacts); the reconciliation is
safe either way.
