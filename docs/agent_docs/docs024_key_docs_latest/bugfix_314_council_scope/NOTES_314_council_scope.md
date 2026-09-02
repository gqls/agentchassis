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

---

## 2026-09-02 — lane RESUMED for the fourth-copy residual (owner-directed)

Lane had been idle since 2026-08-20 (`8d3415e66`). Checked ownership first
(`scripts/who-owns.py 314`): the only commits on the bug file or its subject are this lane's own
three, and the lane dir has had no commit in 13 days. Nothing in flight, so resuming here rather
than competing.

### Residual status, re-verified rather than assumed

The close-out banner named two residuals. **They have diverged, and one is now largely closed by
other lanes — so the banner is stale and must not be quoted as current.**

- **"Tooling is still out of scope"** — PARTLY CLOSED, not by us. Two owner-ruled widenings
  landed after this lane went quiet: `cmd/config-key-audit/` (2026-08-23, `eb5b0d5b8`) and
  `scripts/pattern-check.py` (2026-08-24, `66f1b3339`). `COUNCIL_SCOPE_CODE_RE` now reads
  `^(platform|internal|pkg)/|^cmd/config-key-audit/|^scripts/pattern-check\.py$`. The 097/098
  scripts themselves are still out of scope, so *that* narrower part of the residual stands.
  Both widenings correctly edited BOTH halves — I checked `098:105` `SCOPE_PATHS` and it carries
  `cmd/config-key-audit` and `scripts/pattern-check.py`. The trap this lane wrote up did not fire.
- **"A FOURTH copy … already drifted" (`pattern-check.py:385`)** — STILL LIVE, untouched, and
  now measured rather than asserted.

### The fourth copy, measured [MEASURED 2026-09-02]

The close-out banner asserted the drift from one example file. The honest number is bigger than
one and smaller than alarming, and both halves matter.

In `docs/agent_docs/sql_for_agents` (1,132 `.sql` files):

| set | count |
|---|---|
| matches the runner's name rule `^[0-9]{3}_[A-Za-z0-9_]+\.sql$` (`run-migrations.sh:283`) | 1,127 |
| …of those, sidecars per `SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$'` (`:65`) | 384 |
| **= APPLIABLE by the runner** | **743** |
| what `pattern-check.py:385` can see (`^\d{3}_[a-z0-9_]+\.sql$`) | 738 |
| **BLIND SPOT** | **5** |

```sh
export LC_ALL=C; D=docs/agent_docs/sql_for_agents
ls -1 $D | grep -E '^[0-9]{3}_[A-Za-z0-9_]+\.sql$' \
         | grep -vE '_[A-Z][A-Z0-9_]*\.sql$' \
         | grep -vE '^[0-9]{3}_[a-z0-9_]+\.sql$'
```

The five: `482_ROLLBACK_claim_timeout_exclusion.sql`, `582_dispatch_sibling_A_task_name_on_trigger_row.sql`,
`583_dispatch_sibling_B_parameterise_notify_stamps.sql`, `584_dispatch_sibling_C_insert_trigger_2.sql`,
`637_dispatch_lever_B_interval_not_sibling.sql`.

**MISSTEP, logged in `WRONG_CALLS.md`.** My first census used
`comm -23 <(…) <(…)` and printed `comm: file 1 is not in sorted order` — then a count of **660**,
and a listing containing plainly-lowercase files that the pattern does match. I nearly took the
660. It is a collation failure, not a finding. `grep -vE` under `LC_ALL=C` gives 5.

### The gap is LATENT, not live damage — say so plainly

All five already carry `DO $` guards, so `check_unguarded_migration_insert` would have exempted
them even if it could see them. **There is no unguarded migration hiding in the blind spot today.**
What is exposed is the next one: an appliable migration with a capital in its name carrying a bare
INSERT would pass the commit-time lint in silence.

`584_dispatch_sibling_C_insert_trigger_2.sql` is the useful illustration — it INSERTs into
`scheduled_tasks`, precisely the durable non-idempotent sink 007 Class C is about. It is guarded,
so nothing is wrong; but it is the shape that would have been missed.

### Why this is the same defect one level down, not merely a stale regex

`pattern-check.py:373-379` states its own contract: *"The runner's own dry run already warns
(lint_idempotency in run-migrations.sh — **SAME semantics, keep them in step**)"*. So the lint's
question **is** the runner's question — "will the runner replay this file?" — and its predicate
must therefore be the runner's appliable predicate. It is not.

And the comment beside :385 says `# sidecars (_ROLLBACK etc.) excluded`, which is TRUE but for the
wrong reason: sidecars are excluded only because they happen to be uppercase. That is a proxy
standing in for the rule — **exactly the defect the council's `editquality` seat caught inside this
lane's own fix**, where `_HOLD.sql` was excluded by reusing the runner's `SIDECAR_RE` (which answers
"will `--apply` run this?") as a proxy for "is this the change?". Same mistake, one level further
down, and it was sitting in the file the whole time.

### Vocabulary census — four copies confirmed, no fifth

```sh
grep -rnE '\[0-9\]\{3\}_|\\d\{3\}_' --include='*.sh' --include='*.py' --include='*.go' --include='*.yaml' .
```
1. `run-migrations.sh:283/293/313` + `SIDECAR_RE:65` — source of truth
2. `council-scope.sh:124` — verbatim copy, watched by `council_scope_drift_warn()`
3. `098:105` `SCOPE_PATHS` — pathspec pre-filter + `in_council_scope` post-filter; in step
4. `pattern-check.py:385` — **drifted, unwatched**

Copies 2 and 3 are watched. Copy 4 is the only one with no drift guard on it, which is why it
drifted and why nobody noticed.

### Consumers told, not merely measured (owner ruling 2026-07-29 §3)

Four of the five blind-spot files belong to `dispatch_throughput`, an ACTIVE lane
(14 commits in 5 days, `HANDOFF_2026-08-30_continue_here.md`) whose naming habit — `_sibling_A_`,
`_sibling_B_`, `_lever_B_` — is a *standing generator* of names this lint cannot see. The fifth is
`bugs_open/317`'s (`29670a50f`). Messaged the `throughput` session directly with the mechanism, the
reproduction, the reassurance that none of their files is broken, and two questions only they can
answer: whether a `_HOLD` is ever renamed to become runner-appliable (this decides whether the lint
should cover `_HOLD`), and whether the `_A_/_B_/_C_` convention is deliberate (it is their call;
the detector should match the runner, not the other way round).

### The `_HOLD` question, answered from git — and it inverts the obvious answer

I posed this to the `throughput` lane as a question only they could answer. It turned out the
repository answers it, decisively, and I should have asked it first.

**Premise I started with (WRONG):** the runner never applies a `_HOLD`, so it never replays one, so
the idempotency lint should keep skipping them.

[MEASURED 2026-09-02]
```sh
git log --all --diff-filter=R --format='%h|%ad' --date=short --name-status -M \
  -- docs/agent_docs/sql_for_agents/
```
`_HOLD.sql` files are **routinely renamed to drop the suffix** once ordering allows: **37 rename
events across 26 distinct files**, running continuously from 2026-08-01 to **2026-08-31**
(`6e8fa6a3c` 645_…, `9873bac59` 588_…, `6531e694b` 541_…, `eb347dc12` 599_…). **40 `_HOLD` files
exist right now.** This is a standing workflow.

So the whole life of a `_HOLD` is 007 Class C, start to finish:

1. written as `_HOLD` and **applied by hand** — i.e. out of band, which is the hazard's precondition;
2. **not ledger-recorded**, because the runner skipped it;
3. **renamed into the appliable set**, where the runner finds it pending and unrecorded and
   **replays it** — and a bare INSERT dies on a raw 23505 that reads as broken SQL. 151 blocked the
   runner 3 days exactly this way.

**`_HOLD` is therefore the HIGHEST-risk category for this lint, not the lowest** — the only one
guaranteed to be applied out of band before the runner ever sees it. A fix that mirrored the runner's
appliable predicate would have written the worst case out of the lint while looking rigorous.

**Consequence for the fix.** The lint's predicate is NOT "the runner's appliable set". Its question is
*"could this file's SQL ever be executed by the runner on replay?"* — appliable files **plus** `_HOLD`.
The exclusion is the not-the-change enumeration (`_ROLLBACK` undo, `_VERIFY` assertions, `_SUPERSEDED`
retired), not the runner's catch-all uppercase-suffix rule.

**Convergence with `council-scope.sh`, and why it is not a reason to share code.** That file's
`COUNCIL_SCOPE_NOT_THE_CHANGE_RE` reaches the same three-suffix list from a *different* question — "is
this the change?". Same set today, two different questions, and a future suffix could be the change
without ever being replayed (or the reverse). Sourcing one into the other to "single-source" it would
be this bug's own defect committed a third time; the council caught precisely that inside 314's own
fix. Derive independently, note the convergence in a comment, and point the drift guard at the
**runner** (the source of truth for the name shape), not at council-scope.sh.

### `_HOLD` coverage — SETTLED between the two lanes, with the counter-argument recorded

The `dispatch_throughput` lane argued the other way and then withdrew it, and both halves are
worth keeping because the distinction they landed on is the one a reviewer is most likely to blur.

**Their argument (withdrawn):** *"a live `_HOLD` is by definition not yet runner-replayable, so the
rename is the moment it enters your jurisdiction."*

**Why it fails, in their own words:** it conflates **two different jurisdictions**. The RUNNER's
exclusion of `_HOLD` is correct and must stay — it must never auto-apply one. The LINT's job is to
advise the author *while the text is still editable*. Those are different questions about the same
file, which is this bug's whole subject.

**Their supporting point, which is stronger than the one I had, and goes in the submission:** a
`_HOLD` is **guaranteed** to traverse the risky path. It will be hand-applied at least once (that
is what the suffix means), and after the rename it sits in the probed, replay-eligible population.
So an unguarded `_HOLD` is not a maybe-risk — it is **a scheduled appointment** with the replay
hazard. Commit time is the cheapest point on that whole trajectory, and the only one where the SQL
is still being written.

**The boundary to state EXPLICITLY in the submission** (their suggestion, to pre-empt an
objection): the coverage widens to `_HOLD` — a trailing-uppercase suffix that **IS the change** —
and *not* to the true sidecars `_ROLLBACK` / `_VERIFY` / `_SUPERSEDED`, which are hand-run against
an already-decided state. The two-part predicate already draws that line; the submission must say
so in words as well, because "you widened past the sidecar rule" is the obvious objection and the
answer is that the sidecar rule was never the right rule for this question.

### CORRECTED 2026-09-02 — it was never drift. The predicate was WRONG AT BIRTH.

> **CORRECTED 2026-09-02:** everything above this line that calls the fourth copy
> "drifted" — and the bug 314 close-out banner that says "**already drifted**", and my
> commit message `60739e2c4` — is **wrong**. Caught by the planning pass (fable), verified
> by me before accepting it:
>
> ```
> git log -S'[0-9]{3}_[A-Za-z0-9_]+' -- scripts/migration/run-migrations.sh   -> a51333fd7  2026-07-20
> git log -S'_[a-z0-9_]+\.sql$'      -- scripts/pattern-check.py             -> 9d95e1c31  2026-07-25
> ```
>
> The runner gained `[A-Za-z]` on **2026-07-20**. The lint was written on **2026-07-25**,
> five days LATER, already lowercase-only. **The two literals have never matched.** There
> was no moment of divergence; the copy was made wrong.

**Why this is not pedantry — it changed the fix.** I had built a drift guard modelled on
`council_scope_drift_warn()`: read the runner, assert its lines are still present, warn if
they moved. Against a birth defect **that guard sits green forever** — the runner never
moved; the copy was simply never right. I would have shipped a guard that could not have
caught the very bug it was written for, and it would have looked like diligence.

**So the guard changed shape**, to the one the evidence actually demands: a parity test that
compares **the two literals** and pins **the decisions** in a must-lint/must-not-lint table.
`cmd/config-key-audit/migration_lint_predicate_parity_test.go`, reached from
`scripts/check-migration-lint-parity.sh` when either source file is staged.

Three further reasons that shape is right, and the third is the one I would not have found:

1. **Three precedents exist in that exact package** — `optional_budget_cron_parity_test.go`,
   `cron_parity_test.go`, `retention_parity_test.go` — all extracting a literal from a
   non-Go file and asserting it against its source. `cmd/config-key-audit/` is in council scope.
2. **`scripts/lib/precommit-gotest.sh` already exists** for the scoped-advisory-Go-test
   shape, and its header records *why*: the council's reuse seat objected (corr `be252395`
   r2) to two hand-rolled implementations of it in this same package. A Python arm of my own
   would have drawn that objection a third time, correctly.
3. **The realistic future regression is not the runner moving** — it is a session reading
   `SIDECAR_RE`, seeing the lint "disagrees", and reconciling it, **silently dropping `_HOLD`**.
   That is 314's own defect one level down. A literal-only check passes happily while the
   RULE is wrong; only the fixture table catches it. Proven, not asserted — see the mutation
   run below.

### Controls, all run against the final code (not reasoned about)

Pre-fix baseline pinned at **`b6c4311`** (a `git show HEAD:` baseline expires the moment you
commit — MEMORY has the lesson and I took it).

| control | result |
|---|---|
| predicate matrix: 14 cases incl. stacked suffixes, 1-digit prefix, non-`.sql` | PASS |
| end-to-end through the SHIPPED check, pre-fix module vs final | `701_…_D_…` and `702_…_HOLD` **missed pre-fix, caught now**; 0 regressions |
| blast radius, live corpus **[MEASURED 2026-09-02]** | 1,142 files; newly visible **45** (39 `_HOLD` + 6 appliable); **0 new findings; 0 coverage lost** |
| Go parity suite, 4 tests | PASS |
| **MUTATION** M1 lowercase pattern restored | **3 tests fail** |
| **MUTATION** M2 `_HOLD` added to the exclusion (the predicted regression) | **2 tests fail** |
| **MUTATION** M3 function body broken, both constants left correct | **exactly 1 fails** — test 4, proving it reads the CODE not the literals |
| hook path A: parity broken **and staged** | prints the guidance, **exit 0** |
| hook path B: parity broken, **not** staged | 0 bytes — relevance gate holds |
| hook path C: parity broken, unrelated file staged | 0 bytes |

Mutations were run in a **throwaway `git worktree`**, never in the shared tree, and each was
asserted to be a real edit before its result was believed. The worktree was removed and
`git worktree list` confirmed clean.

**The parity test earned its keep on its first run**, before any mutation: it FAILED, because
my extraction regex matched a THIRD `grep -E '^[0-9]{3}…'` in the runner — `:312`'s
near-miss warner, `^[0-9]{3}.*\.sql$`, which is deliberately WIDER and is not the appliable
rule. Re-anchored structurally on the pairing that actually identifies the rule (a name grep
immediately followed by a `SIDECAR_RE` filter), and it now asserts it finds **exactly 2** such
sites so a removed one cannot pass unnoticed.

### Also corrected by the same pass, and verified here

- **`_HOLD` promotions were already partly linted**, at the RENAME commit: `changed_files()`
  uses `--diff-filter=ACMR`, and a rename lists the destination path, which is lowercase. So
  the coverage this fix adds for `_HOLD` is the **write-time** hit (and any modify while
  held) — the rename hit already existed. My earlier "`_HOLD` files are invisible" was too
  strong; the write-time argument is the one that carries the change.
- **`668_terms_publish_RELOCK.sql` exists** — an unclassified suffix. Under the new predicate
  it is linted (the safe default). I confirmed `council_scope_drift_warn()` **is** currently
  warning about it on every 097/098 run, exactly as 314 designed it to. That mechanism works.
- **The `--record-only` refusal is already a LANDMINE entry** (2026-08-18). Cited, not re-filed.
