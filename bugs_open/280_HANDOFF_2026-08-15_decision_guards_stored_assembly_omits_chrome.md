# 280 — `check_decision_guards.go`'s "stored assembly" silently omits chrome: any guard testing header/footer/nav content is unenforceable

**Filed 2026-08-15** by the `bugfix_270_missing_structure` session, found while
reading `bugs_open/232` for cross-references during the 270 fix (232, dated
2026-08-09, had already spotted this exact caller in passing while diagnosing
a different bug — "read only by two `discovery_checks` —
`check_missing_structure.go:96` and `check_decision_guards.go:95`" — and
nobody followed up on the second one).

> **On the 2026-07-31 ruling (a cross-cutting root-cause claim goes through
> `090`, or the filer states why first-hand verification substitutes):
> substituted.** The mechanism is seven lines of quoted SQL, read directly
> from the file. The false premise (the columns are always empty) is the
> SAME fleet-wide measurement already established for `bugs_open/270` and
> re-confirmed live the same day this file was written. The "no observed
> wrong verdict yet" claim is a full census of all 5 live decision-record
> rows, read by hand rather than sampled, since 5 is cheap to do
> exhaustively. There is no not-where-you-are-looking cause left for a loop
> to find.

## The defect

`platform/orchestration/actions/discovery_checks/check_decision_guards.go:72-78`,
`storedPageAssemblySQL` — the single definition of "the page, as stored" used
both by the check and by its own completion verifier
(`VerifyDecisionRegressionResolved`, deliberately shared per that file's own
comment, "so the two cannot drift"):

```sql
SELECT COALESCE(pg.rendered_header,'') || COALESCE(pg.rendered_footer,'') ||
       COALESCE((SELECT string_agg(COALESCE(pc.rendered_html,''), '' ORDER BY pc.position)
                 FROM page_components pc WHERE pc.page_id = pg.id), '')
FROM pages pg
WHERE pg.site_id = $1 AND pg.name = $2
```

`pg.rendered_header` and `pg.rendered_footer` are the same vestigial columns
`bugs_open/270` documents: empty on all 694 pages fleet-wide (LANDMINES.md,
"`pages.rendered_header` / `rendered_footer` / `rendered_head` are
VESTIGIAL", 2026-08-03; re-confirmed live 2026-08-15 as part of 270's own
verification pass). Chrome actually lives in `site_components`, not these
columns.

So the "stored assembly" this check evaluates every decision guard against is
**silently missing all chrome/nav content, always** — it is really just
`page_components.rendered_html`, concatenated. A `contains` guard asserting
something that lives in the header or footer would ALWAYS report a
violation (false positive: the decision reads as broken when it isn't). A
`not_contains` guard on chrome/nav content would ALWAYS report clean (false
negative: a real regression in the header/footer would never be caught).

This is a different failure SHAPE from `bugs_open/270`, not the same bug:
270's check fired unconditionally and dispatched real (wasted) work; this
check's predicate quietly evaluates against an incomplete document and would
silently mis-report specifically the guards nobody has written yet, or that
happen to touch chrome. It is filed separately for that reason, per the
270 fix's own scope decision (see that bug's fix commit and
`docs/agent_docs/docs024_key_docs_latest/bugfix_270_missing_structure/PLAN_2026-08-15_missing_structure_check.md`
§5).

## Why no wrong verdict has been observed (yet)

```sql
SELECT count(*) FROM doc_notes WHERE categories ? 'decision-record';
-- 5
```
All 5 were read in full, not sampled — cheap at this count. None currently
assert anything about header/footer/nav content:

- `D-001-free-beside-paid` — asserts `href="/tools.html#audience-check"` and
  a link to `/report.html`, both from a page-body CTA section (`covers:
  {"pages":["index"],"slots":["brief-explanation"]}` — explicitly a
  page_components slot, not chrome).
- The other 4 were not chrome-scoped either (`D-002` no-tools-directory,
  `D-003` logo-reads-idea-on-banana, `D-004` guide-copy-hand-authored,
  `write_site_plan`).

So the defect is real and structural, but currently inert — this is
precisely the "silent, no symptom yet" case the standing debugging guide
asks to be filed rather than left for whoever writes the first chrome-scoped
decision to discover the hard way.

## Fix candidates

1. **Retype `storedPageAssemblySQL` to read chrome from `site_components`**,
   the same store `bugs_open/270`'s fix points at — concatenate the site's
   `header`/`footer` slot `rendered_html` (by `site_id`, not `page_id` —
   chrome is site-level) ahead of the existing `page_components` aggregation.
   Must update the check AND its verifier in lockstep, since they
   deliberately share this one SQL constant — that sharing is exactly what
   makes this fix safe to do once rather than twice inconsistently.
2. **Or: explicitly redefine and document "stored assembly" as body-only**,
   if chrome genuinely should be out of scope for decision guards (e.g. if
   guards are meant to police page-body content only, by design). This is a
   real option, not a fallback — but it must be a stated decision, not the
   silent accident it is today, and the file's own header comment
   ("Case-insensitive substring over the page's STORED assembly (chrome +
   page_components...)" — see line ~15) currently claims chrome IS included,
   so leaving it as body-only requires correcting that comment too, not just
   leaving the code be.

## How to verify a fix

- Unit: a guard pattern known to exist only in chrome (e.g. a fixture site's
  `site_components` header slot) must evaluate as present once the fix
  lands; today it evaluates as absent regardless of what the header actually
  contains.
- The check and its verifier (`VerifyDecisionRegressionResolved`) must agree
  — since they share `storedPageAssemblySQL`, this should be automatic, but
  confirm the verifier wasn't given its own copy anywhere.

## Relations

- `bugs_open/270` — the sibling instance (a firing predicate, not a stored-
  assembly definition) of the same root defect: a second, previously
  undocumented reader of the vestigial `pages.rendered_header/footer`
  columns, alongside `check_missing_structure.go`.
- `bugs_open/232` (2026-08-09) — first identified this exact caller in
  passing, filed as a cross-reference, never followed up.
- LANDMINES.md, "`pages.rendered_header` … are VESTIGIAL" — once this and
  270 both ship, that entry's "read by exactly one caller left in the tree"
  line is stale in the other direction (zero callers) and its pointer should
  be updated.
