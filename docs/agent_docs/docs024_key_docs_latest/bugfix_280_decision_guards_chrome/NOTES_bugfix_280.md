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

## 2026-08-16, later — council round 1: REVISE, reviewers' checks answered

Committed (`2c75bb526`), submitted to the council gate
(`SUBMISSION_CORR=d37ef89e-1bfa-485a-aa97-e3b317de7901`). Verdict came back
**REVISE**, gating objection from `editquality`:

> "The fix's entire correctness rests on `site_components.slot_name`
> literally being 'header'/'footer'. This is asserted in the sketch but
> never verified — the landmine set references `content_components.function`
> values 'site-header'/'site-footer'/'head' and a distinct
> `ChromeSlotFunction` resolver in `component_library.go`, suggesting slot
> naming may follow a different vocabulary than the plain string…"

A legitimate question given only the sketch text — answered with live
evidence, not argued from precedent alone:

1. **Live query**: `SELECT slot_name, count(*), count(*) FILTER (WHERE
   coalesce(length(rendered_html),0)>0) FROM site_components GROUP BY 1` →
   `header 22/22`, `footer 22/22`, `head 22/22`. `site_components.slot_name`
   IS literally `'header'`/`'footer'`/`'head'`, fleet-wide, no other values
   exist.
2. **`component_library.go:300-317`**, `ChromeSlotFunction`'s own doc
   comment: *"maps a `site_components.slot_name` to the
   `content_components.function` that serves it"* — confirming
   `site_components.slot_name` IS the plain-vocabulary side (`'header'` in
   → `'site-header'` out); `content_components.function` is a DIFFERENT
   table's DIFFERENT column, used only to resolve which library component
   definition SERVES a slot. My fix reads `site_components.rendered_html`
   directly by `slot_name` and never touches `content_components` or the
   resolver at all — the resolver is real, it exists, and it is simply not
   on the code path this fix touches.
3. `check_missing_structure.go` (bug 270, **APPROVED, LIVE, pod-verified**)
   already reads the identical literals against the identical column and is
   confirmed correct in production — checked whether 270's own council round
   (`524ff897-b697-4c5c-a66f-8939b0457049`) had already surfaced this
   question, per the reviewer's own suggested check: **it had not** — the
   verdict note for that round contains no mention of `slot_name` or
   `ChromeSlotFunction`. So this is a genuinely new question, first answered
   here with direct evidence, not an inherited-and-rechecked one.

**Second thing surfaced by the reviewers' own read-only checks, and worth
recording as a correction even though it doesn't change the fix**: one
check asked to "verify the plan's claim of 5 live decision-record rows,
none chrome-scoped" and returned **8**, not 5. Re-queried by hand:
`categories ? 'decision-record'` now returns 8 rows (up from 5 at 280's
filing on 2026-08-15 — 3 new ones landed same-day, 2026-08-15 14:18 and
18:03, from an unrelated workstream, `bugs_open/279`). **But row COUNT is
the wrong denominator**: `DecisionGuardsCheck.Run()` skips any row whose
body has no parseable ` ```guard ` fence
(`guardFenceRe.FindStringSubmatch`, `m == nil → continue`). Checked which of
the 8 actually carry one:
```sql
SELECT subject_key, (body LIKE '%```guard%') FROM doc_notes
WHERE categories ? 'decision-record' ORDER BY created_at;
```
Only **2 of 8** do — `D-001-free-beside-paid` and
`D-002-no-tools-directory-on-index`. (The bug file's original "5, none
chrome-scoped" was not wrong in its conclusion, but was looser than it read:
it treated all 5 then-existing rows as examined for chrome-scoping without
distinguishing which actually carry an enforceable guard at all; D-003,
D-004 and `write_site_plan` turn out to have no fence and are inert to this
check regardless of content.) Read both fenced guards in full: D-001 asserts
an `href` inside the `brief-explanation` body slot; D-002's own decision
text is explicit that it covers "no section OUTSIDE header/footer" — i.e.
deliberately excludes chrome from its own claim. **Corrected conclusion,
stronger than the original**: exactly 2 rows are live, enforceable guards
today, both are page-body scoped by their own stated design, and the 3 new
rows since filing carry no guard fence at all — so the "currently
symptom-free" claim holds, for a more precise reason than originally stated.

Also checked the reviewers' own flagged landmine, "`categories ? 'decision'`
… means TWO different things" (LANDMINES.md) — it recommends filtering on
`'decision-record'`, not bare `'decision'`; the live `Run()` query already
does exactly that (`WHERE site_id = $1 AND categories ? 'decision-record'`),
so this landmine's outstanding "code change owed" note is stale (already
shipped, unrelated to this fix) and does not touch anything this fix edits.

**Resubmitting on `RESUBMIT_CORR=d37ef89e-1bfa-485a-aa97-e3b317de7901`** with
the evidence above folded into `grounded_in` and `risks` — no change to the
`edits` themselves, since the objection was answerable with evidence, not a
defect in the plan.

## 2026-08-16, later still — council round 2: APPROVED

3 advisory objections, none high-severity, 13 reviewers (4 abstained).
Verdict note (`doc_notes` id `edd6fac7-64ef-44b8-b926-a27f39026cf0`) does not
print the advisory objections' text (unlike a REVISE verdict, which dumps
the full reviewer JSON) — checked `diagnosis_artifacts.metadata` for a
richer record and found only the summary counts, same shape as bug 270's own
approved verdict recorded them. Nothing further to extract; advisory and
none high-severity is the operative fact.

Commit `2c75bb526` already carries `Council-Submitted:
d37ef89e-1bfa-485a-aa97-e3b317de7901` — **not amending it** per forward-only;
`098` resolves and credits it automatically now the correlation is approved.
Updated `bugs_open/280` with the full round-1/round-2 trail. Shipping
(image build + fleet roll) remains the owner's call, unchanged from the
PLAN's original sequencing decision — this session is not building or
rolling unless explicitly asked.

## 2026-08-17 — owner shipped it; verified live; closed

Owner: "A fresh chassis build has been deployed." Before acting on that,
re-checked the tree state rather than assuming continuity — real time had
passed (HEAD had advanced 135 commits since the fix, all normal concurrent
traffic; confirmed via `git merge-base --is-ancestor` + commit dates that
nothing regressed, and that nobody else had touched
`check_decision_guards.go` in the interim).

**Verification, not assumption.** `agent-chassis`'s startup `"build
provenance"` line is a known-rotated-away signal on this busy service —
confirmed absent (`--tail=3000`, `--since=13h`, ~3,100 lines/replica, zero
hits: genuinely rotated, not unstamped). Used the sanctioned binary-probe
fallback correctly — the landmine here is subtle and easy to get backwards:
**do not `grep -aq` your own commit's sha directly** (a binary is stamped
with the ONE commit it was built from, never its ancestors — testing an
ancestor sha reads "absent" even on a binary that genuinely contains your
change). Instead: extract every 40-hex candidate from `/proc/1/exe` (79 of
them — Go's internal digit tables are noisy, as documented), filter to only
ones that are real commit objects in this repo (`git cat-file -t`), and
trust the ONE survivor as the actual stamp:
`6a782274b626c9f4977c9246d905deebb097cb1f`, dated 2026-08-16T18:43, ~2.5h
after the fix. `git merge-base --is-ancestor 2c75bb526 <that stamp>` →
**YES**. Negative control: current HEAD (135 commits ahead, definitely not
built) correctly comes back absent from the same binary — proves the
extract-and-filter method discriminates rather than matching everything.
Repeated stamp-present / HEAD-absent on the second replica: identical.

No behavioural fleet check was expected or needed (all live decision-record
guards are body-scoped, so no check output was ever going to move) —
artefact verification alone is what "shipped" means for this bug, and it is
now positive on both replicas.

Closed: `git mv bugs_open/280_… bugs_closed/280_…`, with the closing
evidence appended to the bug file itself before the move (so the file's own
history shows the reasoning, not just the destination). Updated
`README_where_we_are.md`. This is a real milestone per the standing-five
convention (fixed AND live, not just approved) — writing a fresh
`SUMMARY_2026-08-17_shipped_and_verified_live.md` rather than editing the
2026-08-16 one.

**One thing worth carrying forward, not specific to this bug**: the "ask the
service what it is running" recipe genuinely fails on `agent-chassis`
specifically (documented landmine, confirmed again here) — every future
verification on this service should go straight to the binary probe with
the extract-and-filter method, not waste a round on the log line first
unless the pod is known to be under ~10 minutes old.
