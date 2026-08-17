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

## UPDATE 2026-08-16 — fix candidate 1 implemented, committed, council-submitted; NOT YET SHIPPED

Picked up by the `bugfix_280_decision_guards_chrome` session (started after
the owner asked for "bug 180," which turned out closed-and-live and
unrelated — the still-open sibling this file names was the intended target,
confirmed with the owner directly). Full workstream docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_280_decision_guards_chrome/`
(PLAN has the design reasoning, NOTES has the full evidence/session trail).

**Fix candidate 1 implemented** (retype `storedPageAssemblySQL` to read
`site_components`), not candidate 2 — same reasoning 270 used to reject its
own "delete the check" alternative: candidate 1 makes the code match the
file's header comment, which already describes the post-fix behaviour
correctly ("chrome + page_components.rendered_html"), rather than requiring
the comment to be corrected to describe a body-only design that was never
the stated intent.

The `FROM pages pg WHERE pg.site_id = $1 AND pg.name = $2` gate was
deliberately kept unchanged (not simplified to a bare `SELECT <subqueries>`)
specifically so a nonexistent page still yields zero rows — both the check
and the fail-closed verifier (`VerifyDecisionRegressionResolved`, RFC_017)
depend on that to tell "page missing" apart from "page exists with empty
chrome." Checked before writing the query, not discovered after.

`check_decision_guards_test.go` added: one SQL-text anti-regression test
(mutation-tested — reverted to the pre-fix SQL, confirmed only that test
fails, restored the fix), plus two behavioural demand controls for the
false-positive (`contains` guard sees chrome) and false-negative
(`not_contains` guard catches a chrome regression) shapes described above.
Whole-package suite green; `go build ./...` clean.

**Committed** `2c75bb526` (pathspec: the two files only), trailer
`Council-Submitted: d37ef89e-1bfa-485a-aa97-e3b317de7901` — verdict not yet
read at commit time, per the documented pre-verdict flow (098 resolves and
credits automatically once approved; forward-only forbids an amend to
`Council-Reviewed:` later, so do not attempt one — re-check the verdict and
record it in this file's next update instead).

**Shipping (image build + fleet roll) is NOT this session's call** — same
owner-decision shape as 270 one day earlier ("leave it, I'll ship it").
Assume nothing; check whether it has already shipped before offering,
per-service artefact probe (`kubectl logs -l app=agent-chassis --tail=300 |
grep -m1 'build provenance'`, then `git merge-base --is-ancestor 2c75bb526
<the stamp>`), not by inference from a roll having happened.

No behavioural fleet verification is expected once live: all 5 live
decision-record rows are body-scoped (census in the original filing,
re-confirmed here), so no guard's verdict should visibly change. "Shipped"
is confirmed at the artefact, not by a change in check output.

## UPDATE 2026-08-16, later — council APPROVED on round 2 (round 1 REVISE, answered with evidence)

Round 1 verdict was **REVISE**, gating objection from `editquality`: is
`site_components.slot_name` really the literal `'header'`/`'footer'`, or
could it follow a different vocabulary (the landmine set separately
mentions `content_components.function` values `'site-header'`/`'site-footer'`
and a `ChromeSlotFunction` resolver)? A fair question from the sketch text
alone — answered with direct evidence, not argued from precedent:

- Live query confirms `site_components.slot_name` IS exactly
  `'header'`/`'footer'`/`'head'` fleet-wide (22/22/22, no other value exists).
- `ChromeSlotFunction`'s own doc comment (`component_library.go:300-317`)
  states it "maps a `site_components.slot_name` to the
  `content_components.function` that serves it" — confirming `slot_name` is
  the plain-vocabulary side; `content_components.function` is a different
  table's different column, used only to resolve which component definition
  serves a slot. This fix never touches `content_components` or the
  resolver.
- Traced both live writers of `site_components.slot_name`
  (`render_site_components_action.go:70`, `link_site_components_action.go:171-176`)
  — both hardcode the plain literal set at the source; `ChromeSlotFunction`
  is used only to look up which component fills the slot, never as the
  value written back.
- Checked whether bug 270's own approved council round
  (`524ff897-b697-4c5c-a66f-8939b0457049`, same literals, same column,
  live and pod-verified) had already surfaced this question, per the
  reviewer's own suggested check: it had not. Genuinely new, now closed
  with evidence.

Round 1's checks also surfaced a stale figure worth correcting even though
it changes nothing about the fix: the "5 live decision-record rows, none
chrome-scoped" census had moved to 8 (3 new, unrelated rows landed
2026-08-15 from `bugs_open/279`). Re-checked properly: only 2 of the 8 carry
a parseable ` ```guard ` fence at all (`D-001-free-beside-paid`,
`D-002-no-tools-directory-on-index`) — the check skips everything else
regardless of content — and both are page-body-scoped by their own stated
design. The "currently symptom-free" conclusion holds, more precisely
stated than the original filing.

Resubmitted on the same trail (`RESUBMIT_CORR=d37ef89e-1bfa-485a-aa97-e3b317de7901`)
with this evidence folded into the submission. **Round 2: APPROVED, 3
advisory objections, none high-severity.** No code change between rounds —
the objection was answerable with evidence, not a defect in the plan.

The commit (`2c75bb526`) already carries `Council-Submitted:
d37ef89e-1bfa-485a-aa97-e3b317de7901` — per CLAUDE.md, **not amending it**;
the `098` coverage report resolves and credits it automatically now that
this correlation shows approved.

**Still NOT SHIPPED.** Image build + fleet roll remain the owner's call, per
270's precedent. This bug stays open until build+roll+artefact-verification
are all actually true, not merely committed and approved.

## UPDATE 2026-08-17 — SHIPPED, VERIFIED LIVE (both replicas). CLOSED.

Owner reported a fresh chassis build deployed. Before trusting that, checked
the actual state of the shared tree first: real time had passed since the
approval above (HEAD had advanced 135 commits since `2c75bb526`, all from
other concurrent sessions — confirmed via `git merge-base --is-ancestor` and
commit dates, nothing anomalous; `check_decision_guards.go` itself had zero
further commits in that window).

**Verified at the artefact, not by inference.** `agent-chassis`'s startup
`"build provenance"` line is rotated away within minutes on this busy
service (documented landmine) — confirmed absent from `--tail=3000` and even
`--since=13h` on both replicas (pods 12h old, ~3,100 log lines each, zero
`build provenance` hits: genuinely rotated, not unstamped). Fell back to the
sanctioned binary probe, done the safe way (per the "grep -aq <ancestor-sha>
reads absent even when it shipped" landmine — **never** grep for your own
commit directly; obtain the actual stamp first, THEN check ancestry):

1. Extracted every 40-hex substring from `/proc/1/exe` on `q7b82`
   (79 candidates — Go's internal digit tables produce plenty of noise, as
   documented) and filtered to only those that are real commit objects in
   this repo (`git cat-file -t`): **exactly one** —
   `6a782274b626c9f4977c9246d905deebb097cb1f` ("readme(257): the
   owner-facing account of the approval…"), dated 2026-08-16T18:43:53+01:00
   — ~2.5h after this fix.
2. `git merge-base --is-ancestor 2c75bb526 6a782274b…` → **YES**. Fix is an
   ancestor of the actual build point.
3. **Negative control**: current HEAD (896c5aeeb, 135 commits ahead of the
   stamp, definitely never built) is correctly **absent** from the binary —
   proves the extraction+filter methodology discriminates, not just matches
   everything.
4. Repeated the stamp-presence and HEAD-absence checks directly (`grep -aq`)
   on the second replica (`r6sf2`): both confirmed identical to replica 1.

No behavioural fleet check performed or expected — per this file's own
earlier note, all enforceable decision-record guards today are body-scoped,
so no check output changes. Artefact verification is the whole of "shipped"
here, and it is now positive on both replicas.

**Closed.** Moved to `bugs_closed/`. Full workstream docs (final SUMMARY,
NOTES, README) at
`docs/agent_docs/docs024_key_docs_latest/bugfix_280_decision_guards_chrome/`.
