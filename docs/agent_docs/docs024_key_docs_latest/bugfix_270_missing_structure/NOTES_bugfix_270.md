# NOTES — bugfix 270 (append-only, newest at the bottom)

## 2026-08-15, session 1

**Ownership + validity check before starting.**
```
python3 scripts/who-owns.py 270
```
→ VERDICT "OWNED or recently active" — but the matched workstream is
`portfolio_positioning`, which is the SAME session that filed 270 and whose
own `HANDOFF_2026-08-13_continue_here.md` and `README_where_we_are.md`
explicitly say "Unowned, not on Phase B's critical path". Read both docs to
confirm before proceeding — `who-owns.py` matches on mention, not on active
work, and this is exactly the false-positive shape that produces: a filer
mentioning the bug they just wrote up reads as "recently active" forever
after, even once that thread has moved on. Confirmed no competing fix by also
checking `git log`/`git status` on the target file — clean, last touch was
the filing commit `fe211e277`.

**Re-verified the bug live, 2026-08-15** (2 days after filing):
```sql
SELECT count(*), count(*) FILTER (WHERE coalesce(length(rendered_header),0)>0),
       count(*) FILTER (WHERE coalesce(length(rendered_footer),0)>0),
       count(*) FILTER (WHERE coalesce(length(rendered_head),0)>0) FROM pages;
-- 694 | 0 | 0 | 0   (was 683/0/0 at filing — grew, still zero)

SELECT status, count(*) FROM pages GROUP BY 1;
-- active 658, archived 36  (confirms 'deployed' still never occurs)

SELECT status, count(*) FROM site_work_items WHERE item_key='missing_structure:rerender' GROUP BY 1;
-- cancelled 2, complete 31, deferred 2, detected 1, unresolved 14  (was 43 total/25 complete at filing; now 50/31)

SELECT min(created_at), max(created_at), count(*) FROM site_work_items WHERE item_key='missing_structure:rerender';
-- 2026-04-24 -> 2026-08-14 16:35, count 50
```
Bug confirmed still live and getting worse, not stale.

**Checked `site_components` shape** to ground the fix candidates against
reality rather than the bug file's 2-day-old numbers:
```sql
\d site_components   -- slot_name varchar(100), unique(site_id, slot_name), build_status text default 'pending'
SELECT slot_name, build_status, count(*) FROM site_components GROUP BY 1,2;
-- (header|footer|head, rendered) x22 sites, 66 rows total — every row 'rendered', none 'pending'

SELECT count(*) FROM sites;                                                            -- 40
SELECT count(*) FROM sites s WHERE NOT EXISTS(SELECT 1 FROM site_components sc WHERE sc.site_id=s.id); -- 18
```
All 18 zero-`site_components` sites are `sites.status IN ('pool','system')` —
internal placeholder/template sites. 2 of them (`pool-ai-agents.internal`,
`pool-energy-utilities.internal`) have 5 active pages each but still zero
`site_components` rows. Checked whether they've ever produced a false
positive under the OLD predicate:
```sql
SELECT s.id, s.domain FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.item_key='missing_structure:rerender' AND s.status='pool';
-- 0 rows
```
So `pool`-status sites never reach discovery at all — the exclusion lives
upstream of `discovery_checks.go` (not traced further, out of scope). This
matters for the fix design: a predicate keyed on "zero `site_components`
rows" would be safe against these two sites specifically, because they're
never checked — but that is a fact about the CURRENT dispatch scope, not a
property of the predicate itself, so the plan still treats "zero rows" as an
unsafe general proxy (see below) rather than leaning on this exclusion.

**The landmine that changed the fix's shape.** LANDMINES.md's vestigial-columns
entry records that on 2026-08-03, `loanandmortgagecalculator.co.uk` served
full chrome on 41 pages with ZERO `site_components` rows — chrome had been
baked directly into deployed page artefacts by an older pipeline. Re-checked
that specific site live today:
```sql
SELECT sc.slot_name, sc.build_status, length(sc.rendered_html)
FROM site_components sc JOIN sites s ON s.id=sc.site_id WHERE s.domain ILIKE '%loanandmortgage%';
-- head/header/footer, all rendered, 1857/1511/2228 bytes
```
It's since been backfilled (by `nav-updater`, per the landmine's own pointer)
— so the "baked chrome, zero rows" state doesn't exist anywhere live today,
but it's why the fix design treats "site_components existence" as necessary
but checks `rendered_html` content directly rather than treating an empty
table as an unconditional "broken" signal on its own. This directly ruled out
the naive "flag on zero rows" reading of the bug file's own Candidate 1
sketch.

**A second, previously-unflagged caller of the same vestigial columns**,
found by grepping `rendered_head` while reading `bugs_open/232` for
cross-references: `check_decision_guards.go:73` concatenates
`pages.rendered_header || pages.rendered_footer` (always empty) into what it
calls the page's "stored assembly" for evaluating decision guards. `232`
(2026-08-09) had already spotted this exact caller in passing while
diagnosing something else ("read only by two discovery_checks —
check_missing_structure.go:96 and check_decision_guards.go:95") but nobody
followed up on the second one. Checked live whether it's actually produced a
wrong verdict yet:
```sql
SELECT count(*) FROM doc_notes WHERE categories ? 'decision-record';  -- 5
SELECT subject_key, body FROM doc_notes WHERE categories ? 'decision-record' AND subject_key='D-001-free-beside-paid';
```
Read all 5 guard patterns by hand (only 5, cheap to do exhaustively rather
than sample) — none currently assert anything about header/footer/nav
content, so no observed wrong verdict yet, but the defect is real and
structural. **Decided: out of scope for 270, filed separately as its own bug
rather than folded in** — different check, different failure shape (wrong
STORED-ASSEMBLY definition vs. an always-true FIRING predicate), and 270 was
getting large enough already. Assigned number 279 (278 was taken by two OTHER
sessions between when this was first scoped and when it was actually filed —
checked `ls bugs_open/ bugs_closed/` fresh immediately before filing, not
reused from an earlier count).

**Plan authored via a `model: fable` agent** (per explicit instruction to use
Fable for the planning step), given the full evidence above as a brief. Fable
independently reached the same Candidate-1-over-Candidate-2 conclusion,
caught that the bug file's own sketch of Candidate 1 ("or when every row is
build_status='pending'") would manufacture a NEW false-positive class
(`chrome_link_policy.go` uses `build_status='pending'` as "re-render
scheduled", not "content missing" — a slot can be pending while still serving
valid old content) and dropped it, marking this as an explicit CORRECTION to
the bug file per the PLAN doc convention. Verified Fable's cited facts myself
before trusting them: `workItemClosedStatuses` (5 statuses, matches exactly),
`check_componentless_pages_test.go` exists and is the pattern it modelled the
test file on, and the `missing_structure` consumer grep it ran matches what I
got re-running it independently.

**Implemented and tested.** Rewrote `check_missing_structure.go`: one query
against `site_components` (EXISTS per slot, `coalesce(length(rendered_html),0)
> 0`), three branches (all healthy → `Resolved` retraction under the
UNCHANGED `item_key`; any slot missing + has active pages → the work item,
same shape, corrected reason text; any slot missing + no active pages →
silent, no claim made). Added `check_missing_structure_test.go` modelled on
`check_componentless_pages_test.go`'s pattern (assert on the SQL string
itself, not just on sqlmock-fed rows, so a regression back to `pages.
rendered_*` or to `build_status` can't hide behind a green happy-path test).
`go build ./platform/orchestration/actions/discovery_checks/...` clean;
`go test -run TestMissingStructure` — 5/5 pass.

**Deliberately did NOT touch** `check_site_structural_validity.go`, even
though its `head_essentials_missing` comment (lines 999-1025) now quotes a
stale count ("43 needs_rerender items ... 25 completed") that this fix will
make more stale still. That file is under active council review by
`portfolio_positioning` (round 2/3, per its own commit history) — editing
someone else's in-flight, council-reviewed prose is not this fix's business.
Noted here so it isn't lost: whoever closes 270 should flag that comment as
due for a refresh, to `portfolio_positioning` or in passing.

**Not yet done, in priority order** (see HANDOFF for the authoritative
version of this list): commit narrowly, submit to council gate, file 279,
update `bugs_open/270` itself with the fix pointer, wait for an image
build+roll, verify the 17 stale items actually close and the check goes
quiet on serving sites over one full discovery rotation, then move 270 to
`bugs_closed/`.

## 2026-08-15, session 1 (continued) — commit, council submission, sibling bug filed

**Number races on this tree, confirmed twice in ~20 minutes.** Scoped the
sibling finding as "279" while drafting the PLAN; by the time it was actually
ready to file, TWO other sessions had already taken 278, and by the time this
note was written, ANOTHER session had taken 279 too. Filed as **`bugs_open/280`**
instead — re-checked `ls bugs_open/ bugs_closed/ | grep '^280'` immediately
before writing the file, not reused from an earlier count. Lesson for next
time: don't pick the number until the moment of `git add`, not when the
content is drafted.

**Committed the fix**, narrowly, two files only (commit-scope report
confirmed 2 files / 1 area): `fdc5daec1`. Council submission went out first
(so the commit could carry `Council-Submitted:` rather than committing blind
with no trailer at all) — `SUBMISSION_CORR=524ff897-b697-4c5c-a66f-8939b0457049`,
`RUN_ORCH_ID=f4a43b01-849a-4686-a1a3-35fccb8bc6e7`. The submission JSON is
saved alongside this file: `COUNCIL_SUBMISSION_2026-08-15.json`.

**Filed `bugs_open/280`** for the `check_decision_guards.go` sibling finding,
committed separately (`0aec56e01`), per one-commit-per-task.

**Updated `bugs_open/270` itself** with a pointer to the fix commit, the
council correlation, and this workstream's HANDOFF (`036d9a7e9`) — appended,
did not rewrite the filing session's original text, per the "corrections are
visible, never silent" convention.

**Council verdict check** (~15 min after submission): still `EXECUTING_STEP`
at `review_editquality` — no verdict artifact yet. This is normal per the
~30-minute budget CLAUDE.md sets; not a stall. Re-check via:
```sql
SELECT current_step, status FROM orchestration_states WHERE
collected_data->'input_data'->>'fix_correlation_id' = '524ff897-b697-4c5c-a66f-8939b0457049';
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='524ff897-b697-4c5c-a66f-8939b0457049' AND kind='council_report' ORDER BY created_at;
```

**Still open, in order**: read the council verdict once it lands and act on
it (REVISE/REJECTED both require a response even though the code is already
on the shared branch); wait for an image build + fleet roll (not this
session's call to trigger — owner-run); verify the fleet-level checks in
PLAN.md §7; close out 270 to `bugs_closed/` only once fixed AND live AND the
stale items have actually closed.

**Also corrected the LANDMINES.md entry itself** (`017aae6e4`), independent
of whether 270's fix has shipped: it claimed "read by exactly one caller
left in the tree", which went stale the moment `bugs_open/280` was filed
(two callers now), and its "zero `site_components` rows is the real
'no chrome' signature" guidance has a counter-example named in its OWN
"confirmed" sentence — the reason this fix keys on `rendered_html` content,
not row existence. Ran `scripts/landmines-sync.py --apply` after — succeeded
on the second attempt; the postgres-clients-0 exec pipe was flaky for large
payloads this session (multiple plain `timeout 60s` queries against
`orchestration_states` hung past 60s and had to be run via
`run_in_background` instead; a trivial `SELECT 1` always returned instantly,
so this reads as write-lock contention from the ~7+ concurrent orchestrations
in flight at the time, not a broken connection — note for next session if it
recurs). Sync flagged the entry `NEEDS_VERIFICATION` (RFC_005's
`landmines-verify-dispatch.sh` — routine for any new/changed entry, not an
error); not run this session, low priority given the correction is backed by
direct measurement already quoted inline.
