# PLAN 2026-08-04 — `bugs_open/173`: per-substep `continue_on_error`

**Bug:** `bugs_open/173_HANDOFF_2026-08-01_loop_error_routing_has_no_substep_granularity.md`
**Status of the bug when picked up:** OPEN, **UNOWNED** (its own header says so). The lane
that filed it — `bugs_open/165` sites B+C — has since **closed** (`bugs_closed/165`,
`SUMMARY_2026-08-02_all_four_guarded_and_the_lane_closes.md`), so the file is genuinely
available rather than parked mid-fix.

**Claim check before starting**, because 30-odd sessions share this tree:
- `scripts/who-owns.py 173` returns OWNED-or-active, naming `bugfix_165_reconciliation_deletes`
  — but that is the **filing** lane and it is closed. The tool reads commits, so a lane that
  filed a follow-on for someone else looks identical to one working it.
- Sharper check: **no live session has opened the file.** Scanning every `.jsonl` transcript
  modified in the last 300 minutes for `"file_path":".../bugs_open/173"` returns zero. The
  files that ARE being read right now are 187, 182, 098, 181, 155, 153, 184, 179, 170, 116,
  191, 190, 188, 186, 178, 177, 175, 169, 163, 159, 139, 132, 117, 100, 080, 075.
- Bare *mentions* of a bug number are worthless as an ownership signal — every session lists
  `bugs_open/`, so all 54 numbers appear in most transcripts. Reads/edits are the signal.

---

## The defect, re-verified in today's source (not taken from the file)

`continue_on_error` is a **loop-level** flag with no per-substep granularity:

- `actions/loop_actions.go:66` reads it from the loop step's config;
- `loop_expansion_handler.go:39` carries it through the expansion;
- `loop_expansion_handler.go:157` stamps **the same value onto every injected iteration
  step**, unconditionally;
- `loop_error_handler.go:83` (`shouldContinueLoopOnError`) reads it back off the injected step.

So a loop has exactly two states — every substep tolerant, or every substep strict. There is
no way to say "a failure in *this* substep should skip the item, but a failure in *that* one
should still fail the build".

### The sharper finding, which the bug file does not state

Line 104 deep-clones the substep's config into the injected step, so a substep's **own**
`continue_on_error` is present in `clonedConfig` — and then line 157 **overwrites it**.

**The key is therefore declared-and-inert today**: an author can write
`config.continue_on_error` on a substep, the config-key audit will not object (see below),
nothing will warn, and the value will be silently discarded. That is a landmine in its own
right and is registered as one.

## Four things established by READING, so the fix does not have to guess

1. **The decoder preserves config verbatim.** `models.DecodeSubWorkflowStep`
   (`pkg/models/substep_decode.go`) does `step.Config = config` — the whole map. Only
   *top-level* substep fields outside the seven are dropped. So a substep's
   `config.continue_on_error` reaches the expander intact, and **no change to
   `SubWorkflowStepFields` or `models.Step` is needed.**
   This is decisive for the design: WFA-003's registered landmine says *"the honoured field
   set is a statement about the EXECUTOR, not about `models.Step`… 'Fixing' that test by
   widening the honoured set makes the validator vouch for a field nothing reads."* Putting
   the knob in `config` steps around that trap entirely.
2. **The key is already legal vocabulary.** `datahelpers/action_inputs.go:144` lists
   `continue_on_error` in `frameworkStepConfigKeys` ("keys the orchestrator itself reads or
   injects, on any step regardless of action"). So a substep declaring it is **not** reported
   as an unknown key by `cmd/config-key-audit`. No audit change is needed, and no audit is
   silenced.
3. **There is exactly one read-side decision point.** All three coordinator sites —
   `coordinator.go:907` (sync step failure), `:3239` (async unrecoverable error), `:3345`
   (request timeout) — call `shouldContinueLoopOnError`, which reads the flag off the
   *injected step's* config. Resolving the value once at expansion time therefore covers all
   three. **This is the "one guarded call site" family the platform keeps finding, and it does
   not apply here** — checked rather than assumed, because that family is exactly what
   `bugs_open/093` is about.
4. **`loop_metadata["continue_on_error"]` (line 85) is loop-level and stays loop-level.**
   `skipToNextLoopIteration` deliberately does not rely on that shared key (it is overwritten
   when a second loop expands in the same orchestration) — it derives everything from the
   workflow plan. So the per-substep value must be written onto the **step**, which is what
   line 157 does and what the read side consults.

## Blast radius — MEASURED before submitting, not asked of the reviewer

CLAUDE.md: *"'No collision is possible' is a query, not an argument."*

```sql
WITH loops AS (
  SELECT a.type, s.key AS loop_step,
         s.value->'config'->>'continue_on_error' AS loop_coe,
         COALESCE(s.value->'config'->'substeps', s.value->'config'->'sub_workflow'->'steps') AS body
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.value->>'action' = 'loop'
)
SELECT type, loop_step, b.key AS substep
FROM loops, LATERAL jsonb_each(body) b
WHERE b.value->'config' ? 'continue_on_error';
```

**Result: 0 rows. [MEASURED 2026-08-04]**

**With its positive control**, because a query that returns 0 for the wrong reason is this
platform's most-recorded self-inflicted wound — the same query grouped instead of filtered
returns **18 loop steps and 79 substeps** (9 loops `true`, 1 `false`, 8 unset). So the
predicate demonstrably finds loops and substeps; a non-zero answer was reachable.

**Therefore the change is inert across the entire live fleet** until some config names the
key. Nothing changes behaviour on any of the 79 substeps on the day it rolls.

> **CORRECTION to the figures carried in `173`'s own contribution section.** It recorded
> (2026-08-01) **20** loops — 10 unset / 9 true / 1 false. Today the census is **18** — 8
> unset / 9 true / 1 false. Two `unset` loops have gone in the three days since, consistent
> with `7a15c3a47` *"retire(agents): three unused builders are out"*. The bug's conclusion is
> unaffected (no loop is mis-shaped in the swallow direction; all nine `true` loops are still
> fan-out/dispatch loops). Recorded because CLAUDE.md requires figures to be grounded against
> the live system before being repeated, and this one had drifted.

## Which candidate, and why not the other two

`173` lists three, unranked. Taking **candidate 1**.

- **Candidate 2** — a named severity vocabulary (`on_error: fail | skip_item | continue`).
  More expressive, but it is a **new shared vocabulary**, which by the owner ruling of
  2026-07-29 wants an RFC rather than a bug patch. The bug file says so itself. Not now; if
  the estate later wants three outcomes rather than two, candidate 1 does not block it — a
  bool is forward-compatible with a vocabulary that keeps `true`/`false` as aliases.
- **Candidate 3** — document per-substep `error_step` as the answer. It works, but it
  requires naming a real recovery step, which is a different thing from "tolerate this one
  and carry on". The bug file's own objection stands: *"it is the option that leaves the next
  thread inventing a bool"* — and the `architecture` seat's recorded objection was precisely
  that the next nested substep *"will invent its own ad hoc bool/sentinel rather than the
  engine offering per-substep `continue_on_error`"*.
- **Candidate 1** makes the bad state **unrepresentable-by-lying**: "one tolerant substep in a
  strict loop" becomes expressible in config, so no action ever again has to launder its own
  failure into a success-shaped return to get proportionate handling. That is the shape the
  four council seats demanded when they rejected the `165` workaround.

Ordering by "what closes the door" (`order-fix-candidates-by-what-closes-the-door`): 1 > 2 > 3.

## The change

**One file: `platform/orchestration/loop_expansion_handler.go`.**

Replace the unconditional stamp at line 157 with a resolution that prefers the substep's own
declaration:

```go
injectedStep.Config["continue_on_error"] = resolveSubstepContinueOnError(
    substep.Config, continueOnError, substepName, logger)
```

and add the helper, whose contract is:

| substep declares | resolves to |
|---|---|
| `true` | `true` — tolerant substep inside a strict loop |
| `false` | `false` — **strict substep inside a tolerant loop** |
| nothing | the loop's value — today's behaviour, byte-identical |
| a non-bool (e.g. `"true"`) | the loop's value, **and a WARN naming the substep** |

**The malformed case must be LOUD.** Silently ignoring a present-but-wrong-typed declaration
is *this bug's own failure mode* (declared-and-inert), and reproducing it inside the fix
would be indefensible. The loop-level read at `loop_actions.go:66` does exactly that silent
`.(bool)` ignore — deliberately left alone here, because widening it is a separate change to
a separate contract, but noted in NOTES as a residual.

### Rulings this change has to satisfy, checked one by one

- **Owner ruling 2026-07-29 §1 — does it change what the shared mechanism GUARANTEES?**
  No. Today's guarantee is "every substep in a loop shares one tolerance"; after, "a substep
  may declare its own, and absent a declaration behaviour is identical". It adds an **opt-in
  capability reachable by nothing until a document names it** — the ruling's own words for
  the case that goes through the normal council gate rather than an RFC. Contrast the RFC-002
  trigger, where a Tier-2 evaluator gained the power to **refute** where its stated rule had
  been "confirm, never refute": no analogous new power is gained here, and the count of
  documents that could reach the new branch today is **0, measured**.
- **Owner ruling 2026-08-02 §2 — new authority ships as an opt-in FIELD with the unsafe
  default OFF.** Satisfied by construction. The unsafe direction is *tolerate* (swallowing a
  failure that should have been loud); the default is *inherit the loop*, i.e. exactly
  today's behaviour, and a substep gains tolerance only by writing it down **at the site a
  reviewer of that substep can see**. This is the ruling's whole point: *"a comment is not a
  control on a tree this many sessions share."*
- **Ordering exemption condition (2) — registered in the concept register in the SAME commit
  that ships it.** Non-negotiable, and the reason is stated in CLAUDE.md: *"later is how a
  seam becomes folklore."* WFA-008, written in the same commit as the code.
- **Owner ruling 2026-07-29 §3 — the other consumers must be TOLD, not merely measured.**
  Named in the submission and in NOTES: the four loops that want this knob.

## Tests — both directions, or it is untested in the direction that matters

`173`'s own bar: *"give a loop two substeps, make the tolerant one fail and confirm the
iteration is skipped while the orchestration continues; then make the strict one fail in the
same loop and confirm the orchestration FAILS. Both branches, or the flag is untested."*

New file `platform/orchestration/substep_continue_on_error_test.go`, beside the existing
`error_step_loop_expansion_test.go` (same package, same fixture style):

1. **tolerant substep in a strict loop** — loop unset, substep `true` → injected step resolves
   `true` and `shouldContinueLoopOnError` returns **true**.
2. **strict substep in a tolerant loop** — loop `true`, substep `false` → resolves `false`
   and `shouldContinueLoopOnError` returns **false**. *This is the direction that prevents the
   silent-drop class, and it is the one a naive "read the substep if set" implementation gets
   wrong by treating `false` as absent.*
3. **no declaration, both loop values** — inherits. The inertness proof for the 79 live substeps.
4. **malformed declaration** — `"true"` (string) falls back to the loop value and does not panic.
5. **the loop-level flag still reaches a non-declaring sibling** in the same loop as a
   declaring one — i.e. the override is per-substep, not per-loop-once.

**Mutation-proven, per `mutate-the-code-to-prove-the-guard`:** reverting line 157 to the
unconditional `= continueOnError` must make tests 1, 2 and 5 fail. A test that passes against
both the fixed and the broken code proves nothing, and this platform has shipped exactly that
mistake (`WRONG_CALLS.md`, 2026-08-03). The implementer must run the mutation and record
which tests failed.

**What these tests do NOT prove**, stated so nobody inherits it as a finding: they exercise
expansion and the tolerance *decision*, not a live orchestration end-to-end. The live
induction — a real loop, a real failure, a real skipped iteration — needs a chassis roll and
is listed under "what is owed" until it happens. `[UNVERIFIED]` end-to-end at time of writing.

## Sequence

1. Workstream docs opened (this file, NOTES, README) — **before** the code, per the directive.
2. Opus implements: helper + line 157 + the five tests + the mutation run.
3. `go build ./...` and `go test ./platform/orchestration/... ./pkg/models/...`, plus
   `gofmt` and `scripts/pattern-check.py` (a prior session had a commit bounced for gofmt).
4. Concept register **WFA-008** + index row + count.
5. `LANDMINES.md` entry (footprint `loop_expansion_handler.go`) + `landmines-sync.py --apply`.
6. Council gate submission (platform code — `platform/` is in scope).
7. Commit by **explicit pathspec**, with `Council-Submitted:` if no verdict yet.
8. Bug file updated with what shipped and what is owed.

## Closure — the honest bar

CLAUDE.md: a fix committed but inert until the next roll **stays OPEN**, because the defect is
still reproducible until it ships. `make release` is whole-fleet and owner-run
(`releases-are-whole-fleet-make-release`), so this session cannot unilaterally make it live.

**So the plan is: commit + prove + submit, then check the pod.** If another session's roll
carries it live within the session, verify at the pod with a discriminating marker **and a
negative control** (`bugs_open/153`: a roll is not evidence your fix shipped) and close. If
not, the file stays OPEN with the roll and the live induction named as owed. **It will not be
closed on a commit.**
