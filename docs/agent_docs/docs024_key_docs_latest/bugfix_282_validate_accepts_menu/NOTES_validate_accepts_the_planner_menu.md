# NOTES — bugfix 282 (validate accepts the planner's menu)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-16 — picking the bug up

`who-owns.py 282` returned OWNED-or-recently-active, which on this estate needs
reading rather than obeying: the "owner" is the **loancalculator lane, which
filed the bug and is BLOCKED on it** ("Sequence to release D2: fix 282 → image
rolls → re-fire phase2_recompose_26.sh"). Its own last action was to confirm no
roll carried a fix. The 285 lane touched the file only to *remove* itself
("Nothing here blocks or is blocked by 282"). So the bug was unowned in the sense
that matters — nobody was fixing it — and taking it on is what unblocks the lane
that filed it. `who-owns.py` reads COMMITS, so I also grepped the live session
transcripts: two other sessions mention 282, neither is fixing it.

## Validity re-checked first-hand, not inherited [MEASURED]

Bug files decay. Re-proved the whole chain at the live DB before writing a line
of code (queries in the RUNBOOK): menu 151 rows incl. 11 tool functions → raw
LLM plan names a tool on 12 pages → `collected_data.validate_plan` has none.
`validate_plan` is the action's own output field, so the drop is inside
`ValidateSitePlanAction`. Code at HEAD is as the bug file describes.

## MISSTEP CORRECTED — the ADDENDUM's "second arm" does not exist

The bug file's addendum (2026-08-15 ~19:15Z) records that
`loans-credit-health-check` — "a name matching NO component at all" — survived
validate and spawned `needs_new_component`, while known tool-level names
vanished, and concludes the tool names "likely die in a branch where the
unresolved-name handler finds an EXISTING component row (level 'tool') and
discards the section".

**There is no such branch.** One query settles it:

```sql
SELECT id, "function", component_level, is_active, created_at
  FROM content_components WHERE "function"='loans-credit-health-check';
-- 824e3309-… | loans-credit-health-check | section | t | 2026-08-13 14:19
```

It is a **section-level** component, created 2026-08-13 for
loanandmortgagecalculator.co.uk. Arm 1 of `resolve()` (`validFunctions[raw]`)
accepts it, exactly as designed. The asymmetry was never asymmetric: one name
was section-level and passed, the others were tool-level and were dropped. The
simple account in step 3 of the bug file is complete and correct.

Why it mattered: the addendum instructed the fixing thread to "locate that branch
precisely — it is probably the cleanest seam for the fix". Following that would
have meant hunting a branch that does not exist, in a 7,000-line file. The check
that dissolves it costs 0.2 s and is now in the RUNBOOK. Logged in `WRONG_CALLS.md`.

## The design turn: not mirroring the predicate

The bug file's preferred candidate ("mirror the menu's own predicate in the
resolver … factored into ONE shared helper") is right about the goal and wrong
about the mechanism, because **the menu is not Go**. It is a SQL string in
`agent_definitions.default_config`, and it had already drifted past the text the
bug file assumed: migration **419** (2026-08-15, the `bugs_open/276` family)
added a `requires-backend` clause to the same query, and guards its apply by
asserting 407's exact bytes as its pre-state. Mirroring would have produced a
third hand-maintained copy across the SQL/Go line — the very lockstep class 016b
§9 catalogues, whose stated verdict is *"single-sourcing is a guarantee, a
lockstep test is a backstop."*

Consuming the offer's OUTPUT (`available_components`, already in CollectedData)
is the single-sourced version: no predicate is duplicated, and any future gate on
the menu flows through with nothing to keep in step.

## Deliberate non-widening of the shared resolver

`loadComponentNameResolver` is shared with `apply_gap_plan_action.go` (3 call
sites, content-gap-planner, 131 dispatches/30d). 407's header and PLAN-049's
landmine both record that the gap-planner's menu is **deliberately** not widened.
Changing the resolver's query or signature would have handed that path an
authority an owner-facing decision explicitly withheld — silently. The new arm is
a method the gap-plan path never calls, and the DB query is untouched.

## Tests were mutated, not merely run [MEASURED]

Nine tests pass. That is not evidence on its own, so both arms were broken in
turn:

- `addMenu` → no-op: `…MenuFieldKeepsAToolSectionTheBaseWouldDrop` fails with
  `sections = [hero faq], want [hero tool-loan-repayment faq]`;
- the `menu_field` gate forced always-on:
  `…WithoutMenuFieldTheToolSectionIsStillDropped` fails with
  `got [hero tool-loan-repayment faq]`.

Restored → all nine pass. The second mutation is the one that matters: it proves
the opt-in is real, i.e. that no un-opted-in caller's behaviour changed.

## Tooling notes

- `orchestration_states` has `owner_agent_type`, **not** `agent_type`. Two of my
  queries and one subagent's died on this.
- `go vet` on the actions package reports `load_component_library_actions.go:207:
  unreachable code` — pre-existing, another lane's file, not touched here.
- A Plan subagent I dispatched to stress-test the design died mid-run on the
  session usage limit. Its partial output confirmed the two live configs and the
  `site-planner` shape; the design decisions above are my own first-hand reads,
  not its conclusions. Recording this because a half-finished subagent's findings
  are easy to inherit as if they were complete — this lane's claims are cited to
  queries and file:line, not to that agent.

## 2026-08-16 (later) — shipped, and one shared-tree event worth recording

Order of operations, as executed: build + 9 tests + **mutation of both arms** →
council submission (`bbf49822`) → commit `5534e9f71` (code + migration + lane
docs, pathspec, `Council-Submitted:` trailer) → `git archive HEAD` build and
test in a clean tree (exit 0) → migration `439` applied **by hand, single file**
→ recorded via `--record-only`.

**Why the migration was NOT applied with `--apply`.** The dry run listed six
pending files; four belong to other lanes (429, 432, 433, 434 — two of them
already flagged LIKELY ALREADY APPLIED). `--apply` takes every pending file in
the directory, so firing it would have applied four other threads' migrations on
their behalf. Applied 439 alone by piping it to psql (it carries its own
BEGIN/COMMIT and guard), then recorded it with a note saying exactly that.

Post-apply verification, with its control: `menu_field='available_components'`
AND it equals `load_components.output_field` on build-site-planner; **`site-planner`
returns `<absent>`** — the control that proves the change is scoped to the one
agent rather than to the action.

**My WRONG_CALLS and LANDMINES appends were swept into another session's commit
(`2079aca42`) as same-file passengers.** Both entries are intact at HEAD, and
that session left a note (`4434cd61b`) naming the passengers for their authors.
Nothing lost, forward-only holds — recording it because this is the exact hazard
CLAUDE.md describes, and because it means the git history attributes those two
entries to another lane's commit message. If you are tracing where this lane's
landmine came from, `git log` the FILE, not the message.

**RFC_022 checked rather than asserted** (the ruling says asserting it without
the query IS the objection). Before the migration: `SELECT count(*) FROM
agent_definitions WHERE default_config::text LIKE '%menu_field%' AND is_active
...` → **0**, and a repo grep for `menu_field` outside my own change → **0**. So
all three conditions held (opt-in; the unsafe side — accepting the menu — is the
default; zero live consumers named it). After 439 there is exactly one consumer,
which is the deliberate act, enumerated in the migration header.

**Docs the fix owed, and where they went:** bug file (status + a CORRECTED block
refuting its own ADDENDUM + fix-as-built) · `WRONG_CALLS.md` · `LANDMINES.md`
(+ `landmines-verify-dispatch.sh`, corr `eab5a7cd`) · register **PLAN-050** new,
**PLAN-027** amended, **PLAN-049** given the back-relation it never had — their
mutual silence was the documentary form of this bug · `016b` §10 row (282 had a
§9 pattern but no index row) and a §9 addendum correcting the pattern's own
advice · the loancalculator lane's NOTES, because the lane that is BLOCKED on
this fix would otherwise have no way to learn it had landed.

**The §9 correction is the part worth re-reading.** The pattern entry said the
acceptance predicate "should be the SAME code as the offer predicate". On this
seam that is impossible — the offer is config, the acceptance is Go, and they
can never ship together. The sharper rule: when two surfaces must agree about a
SET, have one read the other's ANSWER rather than recompute its QUESTION.
