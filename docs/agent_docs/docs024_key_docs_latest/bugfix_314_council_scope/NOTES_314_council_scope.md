# NOTES — bugfix 314 (council gate scope)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-19 — the shape of the fix changed the moment I counted the copies

The bug names one file (`097:87`). Grepping for the rule found **three** hand-maintained copies:
`097:87` (admission), `scripts/council-coverage-nudge.sh:58` (commit-msg nudge), `098:76` (the
coverage report's pathspec). That turned a one-line regex widening into a single-sourcing job —
and it is the difference between fixing the bug and fixing the class. Widening 097 alone would
have produced a gate that admits migrations while the nudge stays silent on them and the report
mis-buckets them: three components disagreeing about one rule, which is worse than all three
being wrong together.

Deliberate non-move: **not** sourcing the vocabulary from `run-migrations.sh`. That would couple
the review gate to the apply path — a syntax error in a review script would stop migrations
applying. Verbatim copy + a `grep -qF` anchor instead.

`097` had no dry-run mode, so §6's POSITIVE control ("a config-only submission is ACCEPTED")
could only be obtained by spending a real council round. The negatives were always free (the
refusal is client-side, `exit 2`, pre-dispatch). **A filter that can only be half-tested is how
this bug survived** — so `DRY_RUN=1` went in with the fix.

## 2026-08-20 — the council found a real defect, and it was this bug's own defect

Verdict on `85fac99c`: **APPROVED round 1**, 12 reviewers, none high-severity. It was still
right about something that mattered.

**`editquality` (medium): excluding `_HOLD.sql` was wrong.** I had written the exclusion as "the
hand-run sidecars", citing `run-migrations.sh:33`, and implemented it by reusing the runner's
`SIDECAR_RE`. But that regex answers *"will `--apply` run this?"* and council scope needs *"is
this the change?"*. `_HOLD` is precisely where the two diverge: a `_HOLD.sql` is a real migration
held back from the runner **for ordering**, with hand-apply commands in its own header. Live
config reaching production, unreviewed — the exact risk the fix exists to catch.

**`bugs_open/314` IS the bug "a PATH filter stood in for an INTENT".** I read that sentence,
wrote it into my rationale, and quoted it in the fix's header comment — then reached for the
nearest available path proxy anyway, because it was authoritative, adjacent and already written.
Proximity to the lesson is not protection from it. `WRONG_CALLS.md`.

Verified before acting rather than taking the report at face value: 157 `_ROLLBACK`, 12 `_HOLD`,
7 `_VERIFY`, 2 `_SUPERSEDED`; the sampled `_HOLD` files carry real `UPDATE`/`jsonb_set` writes
and hand-apply instructions.

**The fix is a better shape than the original.** The exclusion is now an ENUMERATION of
not-the-change suffixes, so a suffix must be *shown* to be excludable and anything unrecognised
defaults to in-scope — a wasted credit, never an unreviewed change. And the drift guard changed
job: it used to assert "my copy matches the runner's", which **passed happily while the rule was
wrong**. It now censuses the directory for unclassified suffixes — a check that can fail for the
right reason.

### The second defect, which the control matrix could not see because the matrix was also broken

Fixing the first, I added the census pipeline and omitted `|| true`. When every suffix IS
classified, the final `grep -v` matches nothing, exits 1, the command substitution inherits it,
and the callers' `set -euo pipefail` kills the script — **no output, before the scope check, so
every submission fails**. It fires *only* when the census is clean: the gate breaks when
everything is healthy.

The file's own comment, eleven lines above and written by me the day before, says every grep here
needs `|| true`. I applied it to `in_council_scope()` and not to the function I added after.

**What did not catch it: the control matrix.** At that moment its harness read
`printf '… exit=%s' "$label" "$(basename "$file")" "$?"` — the command substitution runs before
`$?` expands and resets it, so **every case reported exit=0**, including the negatives that must
return 2. Two instrument failures stacked and masked each other. The tell was not any individual
result; it was that *nothing failed*, which no honest matrix does. Fixed with `local got=$?` on
the very next line.

### Answered on the record

- `prior_art_librarian` (medium) — "the three-copies count is your own grep". Re-swept
  independently: exactly one executable definition (the fragment) + three consumers reading its
  variables; `pattern-check.py:1251` is a docstring quoting the old literal, not code.
- `prior_art_librarian` (low) — "'already drifted' is stated, not verified". Now concrete:
  `pattern-check.py:385`'s `^\d{3}_[a-z0-9_]+\.sql$` is lowercase-only and cannot see
  `482_ROLLBACK_claim_timeout_exclusion.sql` — a live, runner-appliable migration whose name
  merely *begins* with ROLLBACK — nor any `_HOLD` file. Out of remit; a candidate for its own ticket.
- `tooling_provenance` (medium) — "document in `doc_plans`/`doc_notes`, not only markdown".
  Not done, and stated rather than silently dropped: no `doc_plans` subject for the gate's scope
  was found, the council's own verdict notes are per-run artefacts, and LANDMINES is this
  estate's prescribed channel for exactly this class. Recorded as an open question.
- `guardian`/`reuse_agent`/`architecture` (low, all the same point) — the jq and shell arms are
  one rule in two dialects. Accepted and disclosed; the control matrix exercises both paths, and
  the alternative (piping shell output through jq) trades one duplication for a process boundary.

## Lane state

**CLOSED.** `bugs_open/314` → `bugs_closed/`. Registered FIX-061 (+ index row). Residuals named
in the close-out banner: tooling still out of scope (which is why the LANDMINES entry was amended,
not retired), and the drifted fourth copy in `pattern-check.py`.
