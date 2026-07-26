# PLAN — bugs_open/079: a detected dead in-body link must not ship

**Opened:** 2026-07-26. **Bug:** `bugs_open/079_HANDOFF_2026-07-26_phantom_in_body_links_detected_but_never_blocked.md`

## The defect

`validate_page_content` extracts every in-body `href` and resolves it against the site's
real `pages.url` set — correctly and completely. A miss is filed at `Severity: "warning"`.
`valid := blockerCount == 0 && errorCount == 0` does not count warnings, so the page goes
on to `save_sections` → `update_status` → `deploy_page`. The platform sees the dead link,
writes down that it saw it, and publishes it.

## Decisions, and why

### D1 — Repair in-band at the gate. Do NOT promote the severity.

079 said "do not do this without measuring first". Measured, over the retained window
(`orchestration_states` reaches back to 2026-07-13 — 13 days, not the ~24h the bug file
assumed):

| | |
|---|---|
| builds carrying a `validation_result` | 16 |
| builds carrying phantom-link findings | **3 (19%)** |
| phantom instances / unique targets | 17 / 15 |
| of those 15, targets existing in ANY form (incl. `+.html`) | **0** |
| builds that deployed anyway | 3 of 3 |
| pages | oufe.com `/index.html`, webdesign.co.uk `/index.html` — **both homepages** |

Promoting to `error` returns `(nil, error)` from the action, which routes to
`mark_needs_review` → `fail_work_item`. The page never saves and never deploys. On this
sample that is two homepages that do not ship — "no page at all" instead of "a page with a
bad link", exactly what 079 warned was worse.

### D2 — Record durably, but do NOT file a work item.

`bugs_open/083`: `phantom_internal_link` has been **detected 22 times and fixed zero times,
ever**; 98 rows fleet-wide sit at `status='detected'` because the only promoter,
`TriageDetectedItemsAction`, is reachable only from the disabled `improvement-sweep`.
`bugs_open/077`: do not file items whose handler has no remit. So the finding goes to
`agent_error_log` — durable, countable, queryable, and promising nothing. This is
`071` candidate 1 applied to the link class.

### D3 — Repair, then unlink. Never construct a URL.

- Target resolves → untouched.
- Target resolves at its `.html` form → **rewrite the href to the stored `pages.url`**.
  `049` measured 8 fleet targets that exist and return 200 at `.html` while the writer
  emitted them bare; unlinking those would be a content loss. The emitted string is always
  one the database handed us — `bugs_closed/029` was a whole bug about an emitter that
  assembled plausible URLs instead of citing real ones.
- Otherwise → **unwrap**: drop the `<a>`, keep the inner markup. 079's own candidate 3.
  Owner confirmed 2026-07-26. Styling lives on the `<a>`, so a de-linked CTA also stops
  looking clickable.

Explicitly NOT done: loosening `NormalizePagePath` to tolerate a missing `.html`. `049`
recorded why — `/contact` genuinely 404s on these sites, so a tolerant matcher would
silence the detector on a live defect. The tolerance belongs in the repair, not the
comparison.

### D4 — `repair_internal_links`, default ON.

Default OFF would be inert and the bug would still be live. Default ON with a config
toggle gives a reversal lever that is **live-immediately** — DB config does not wait for
an image roll, which matters for a fleet-wide content change.

### D5 — The page-set load must stop failing open.

`loadValidPagePaths` swallowed a query error and returned an empty set, and never checked
`rows.Err()`. Harmless while the only consequence was a spurious warning; catastrophic once
findings drive a rewrite — an empty set means every link is a phantom and the pass would
strip the lot. It now returns `(index, ok)` and both detection and repair are skipped when
not ok. A NULL `pages.url` is skipped rather than treated as failure, so one malformed row
cannot disable link checking for a whole site.

## Scope — decided with the owner, 2026-07-26

**In:** the guard. **Out, and filed separately:** the writer never receives its link
constraints (see NOTES — 20 of 20 writer runs recorded `page_count: 0`). Different
mechanism, different file, different agent's path; it belongs with `071` candidate 4 and
must not ride along in this commit.

## Verification

Unit, plus the live induced branch. See `RUNBOOK_phantom_link_gate.md`.

## Status

- 2026-07-26 — code + tests written, all green, induced-fault probed. Council pending.
