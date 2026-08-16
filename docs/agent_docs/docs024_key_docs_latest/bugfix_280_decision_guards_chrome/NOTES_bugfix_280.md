# NOTES — bugfix 280 (append-only, newest at the bottom)

## 2026-08-16

Owner asked to pick up "bug 180." `bugs_closed/180_...md` shows that bug
fixed and pod-verified live since 2026-08-02/03 (chassis v1.0.1233) — nothing
to do there. The session that had just handed off (`bugfix_270_missing_structure`'s
`HANDOFF_2026-08-15_continue_here.md`) explicitly names `bugs_open/280` as
"the sibling bug found while fixing this one — STILL OPEN, may be worth
picking up next," and its own text anticipates exactly this confusion
("check first whether you actually meant bugs_open/280"). Asked the owner via
AskUserQuestion rather than guessing; confirmed 280.

Ran `scripts/who-owns.py 280`: VERDICT said "OWNED or recently active,"
naming `bugfix_270_missing_structure` as the owning workstream — but per the
known false-positive shape in that tool (a filing session reads as owning a
bug it only filed and moved on from), checked the actual signal: zero
commits touch `check_decision_guards.go` since 226c1a723 (2026-08-10),
`git status` on the file is clean, mtime is 2026-08-10. 270's own workstream
docs explicitly hand 280 off rather than claim it. Read as unowned.

Grepped live session transcripts (36 peer sessions running) for the
mechanism name (`storedPageAssemblySQL|check_decision_guards|VerifyDecisionRegressionResolved`)
in the last 45 minutes, excluding this session's own transcript: 4 hits, all
inside broader `discovery_checks` registry sweeps (item-type-to-check-file
listings), zero mentions of the load-bearing symbol `storedPageAssemblySQL`
alone. No workstream directory named `*280*` existed yet either. Read as
genuinely free to pick up — recorded per the "leading signal is the other
sessions' live transcripts, grep symbols not the bug number" practice.

Read `check_decision_guards.go` in full and `check_missing_structure.go` +
its test file as the sibling precedent to follow. Confirmed via
`\d site_components` that `(site_id, slot_name)` is UNIQUE, so a scalar
subquery per slot is correct (no `string_agg` needed for chrome, unlike the
`page_components` aggregation which genuinely can have many rows per page).

**One thing checked and deliberately avoided**: my first draft of the fixed
query considered dropping the `FROM pages pg WHERE ...` structure entirely in
favour of a bare `SELECT <three subqueries>` with no FROM clause, since the
header/footer terms don't need a `pages` row at all. Caught before writing
it: a bare `SELECT` with no FROM always returns exactly one row, so
`QueryRowContext` would stop erroring when the named page doesn't exist —
silently breaking the "missing page ⇒ zero rows ⇒ skip / fail-closed" contract
both the check and its verifier rely on (the verifier's own doc comment says
a missing page must be a hard error, RFC_017). Kept the original `FROM pages
pg WHERE pg.site_id = $1 AND pg.name = $2` structure and correlated the new
subqueries to `$1` directly instead. Recording this because it is exactly the
kind of "looks like a strict improvement, quietly changes a different
contract" mistake this file's own review culture exists to catch, and it
would have shipped clean through every test I'd already written before I
thought to check it.

Wrote `check_decision_guards_test.go` following
`check_missing_structure_test.go`'s pattern: one SQL-text anti-regression
test, two behavioural demand controls (contains-sees-chrome,
not_contains-catches-chrome-regression). Mutation-tested the anti-regression
test by reverting the constant to the pre-fix SQL, running the suite, and
confirming it — and only it — fails (the two behavioural tests pass either
way, since they mock the DB's *output* directly rather than deriving it from
the real query; that is by design, they test the caller's handling of chrome
content, not the SQL itself). Restored the fix; full package suite green
(`go test ./platform/orchestration/actions/discovery_checks/...`); whole
platform builds (`go build ./...`).

Not yet committed as of writing this entry — commit, council submission and
the standing-five completion (README, SUMMARY) are the immediate next steps
in this same session.
