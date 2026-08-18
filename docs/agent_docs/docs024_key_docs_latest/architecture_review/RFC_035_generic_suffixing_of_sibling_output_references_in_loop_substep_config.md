# RFC 035 — loop expansion rewrites ANY reference-shaped sibling-output config value (WFA-017): ratify, narrow, or revert

## STATUS: **RULED 2026-08-18 — OPTION (1), RATIFIED BY THE OWNER. Shape-based rewriting is now the estate's convention for loop-substep config references; the allow-list is retired.** The code is LIVE (chassis v1.0.1307, 2026-08-17 17:05Z) and PROVEN (§8). Council round 1 was REJECTED on SCOPE (guardian veto, corr `cba35b35`) and the owner has overridden that veto on the merits — the same disposition as `bugs_closed/124` / BLD-019: **the code stands, and "live ≠ approved" still holds, so no `Council-Reviewed:` trailer may ever be written against that correlation.**

### What the ruling settles, in one sentence each

- **A loop-substep config value that is reference-SHAPED and whose first dotted segment names a sibling `output_field` IS a reference**, and loop expansion will iteration-suffix it — for every action, present and future, without anyone having to add a key to a list.
- **The 8-key `dataRefKeys` allow-list is retired as a contract.** It survives in the code only as an ordering detail (it runs first); nothing may be added to it as a way of opting a new key in, because opting in is no longer a thing an author does.
- **`condition` expressions stay out of scope** (the shape gate excludes anything with spaces or operators). Conditional resolution remains its own seam; a partial rewrite of an `OR` carrying two references would be worse than today.
- **What an author now owes:** if you write a literal into a loop-substep config, it must not collide with a sibling `output_field` name. That is the LANDMINE line on WFA-017, and the census query in `bugfix_287_spawn_record/RUNBOOK` is how you check a config you did not write.
- **The guardian's objection is not dismissed, it is answered by monitoring rather than by argument:** a snapshot census cannot bound future configs, so the standing check is the resolver instrument (`RESOLVER_%` rows) plus that census query — a mis-suffixed literal surfaces as a resolution failure in whichever agent's loop runs.

Filed by the bugfix-287 lane at the council's direction; ruled by the owner one day later.

## 1. What happened, plainly

Fixing `bugs_open/287` (spawn_record slug — dispatch loops storing the spawn record as a work
item's `result`), the fixing lane shipped two halves:

- **Config half** — RFC_029 `!` strict markers + an error route on the three dispatch agents'
  `mark_complete` (migrations `448` applied / `452_HOLD` gated). The council's guardian seat
  explicitly endorsed this half ("a contained, agent-scoped fix … should proceed").
- **Go half (THIS RFC's subject)** — `prefixConfigStepReferences` (coordinator.go), the loop
  expansion's config rewriter, stopped enumerating: instead of suffixing sibling-output
  references only for an 8-key allow-list + `input_mapping`, a generic pass now rewrites ANY
  top-level string config value that is *reference-shaped*
  (`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z0-9_]+)*$`) and whose first dotted segment names a
  sibling `output_field`. Registered as **WFA-017**; mutation-proven tests
  (`loop_config_reference_suffixing_test.go`).

The guardian vetoed the Go half on SCOPE: a fleet-wide change to shared loop-expansion
behaviour, justified as closing a class of future bugs, riding a 3-agent bug fix — "exactly
'architecture change dressed as a point fix'". The architecture seat: `needs_rfc`, split from
the migrations. Per the 2026-07-28 owner ruling (a scope veto is not answered by resubmitting;
record it where the change lives, route the seam here, let a human break it) — this file.

## 2. The question for the owner

**Should loop expansion's data-reference rewriting be enumerated (allow-list) or shape-based
(generic)?** Three options, costed in §5. This is a real convention choice: the generic pass
makes "a loop-substep config string that names a sibling output" MEAN a reference, fleet-wide,
for every future workflow author.

## 3. Evidence base (all measured, 2026-08-17; queries in `bugfix_287_spawn_record/RUNBOOK`)

- The allow-list's own comment says *"IMPORTANT: Any config key that references step outputs
  must be listed here"* — and three actions had already broken the promise unnoticed:
  `complete_work_item` (`result`, `work_item_id`, `commit_sha`), `mark_maintenance_complete`
  (`result_field`), plus the `condition` strings. That gap is one of 287's two doors.
- Census of every live loop substep: **22 config strings across 7 agents** have a sibling-
  output first segment through a non-allow-listed key; **all 22 are read-references**; zero
  literals, zero write-destinations; the only expression-shaped values are 3 `condition`
  strings, which the shape gate excludes by construction.
- Nested sibling-output strings exist ONLY inside `input_fields` ARRAYS (Strategy-1 field
  names) — the generic pass is top-level-string-only and a test pins the arrays untouched.
- Behaviour delta today ≈ **path, not value**: `propagateIterationOutputs` runs before every
  substep (coordinator.go:1355), so the base key a dotted reference resolves today holds the
  same value as the suffixed key the generic pass makes it resolve. (This same fact is why
  the strict markers alone close 287 — the veto is right about that, and 287 §11 says so.)
  **⚠ §7.3: this bullet was asserted before the arithmetic that makes it load-bearing; it now
  rests on 201 measured `MAPPING_BYPASSED` events vs 155 completions, not on the code path
  alone. Read it with §7.**

## 4. The seats' substantive concerns, kept whole

1. **Guardian:** a snapshot census "only proves today's config is safe — any future or
   currently-unindexed agent config matching the shape gets silently reprefixed too". True:
   the census is a point-in-time survey; the generic pass is a standing convention.
2. **Tooling-provenance:** `referenceShapedConfigValue` is a SECOND definition of "this
   string is a reference", beside the resolver's own context-dependent classification
   (`IsDottedPathReference`, Strategy-4 dotless references, spec-Default literals — the
   CTS-059 landmine: a dotless string means VALUE on a spec-defaulted field and REFERENCE on
   every other one). Two classifiers for one concept is the drift class this estate reviews for.
3. **Editquality:** the shape gate bounds what gets REWRITTEN (first-segment collision), but
   the pass EXAMINES every top-level string; an enum-like value (`mode: fast`) that happens
   to collide with a sibling output name would be rewritten. Censused zero today; not zero by
   construction.
4. **Reuse:** the estate now has this pass AND `input_mapping`'s own allow-list gate
   unreconciled — two gates for one problem, mirroring a known pattern.

## 5. Options, costed

**(1) Ratify WFA-017 as-is (generic, on by default).**
Cost: the concerns in §4 stand; the mitigation is the register entry's LANDMINE line ("a
loop-substep literal that exactly names a sibling output_field is now a reference") plus the
census query for audits. Benefit: the 287 door class is closed for every future action; the
allow-list promise (already broken three times) stops existing; conventions-over-enumeration
is how `input_mapping` already behaves.

**(2) Narrow: keep the pass but make it opt-in per key (`*_from`-style) or per action** —
i.e. revert to enumeration with `result`/`work_item_id`/`commit_sha`/`result_field` added.
Cost: restores the exact promise that was broken unnoticed three times, and the next action
author must know the list exists (they demonstrably do not); the 2026-08-02 §2 ruling's
reasoning ("a comment is not a control on a tree this many sessions share") cuts against it.
Benefit: zero risk to unexamined configs; smallest possible surface.

**(3) Revert the Go half (forward commit) and rely on strict markers + propagation.**
Cost: `!` on loop agents then resolves the BASE key via the per-substep propagation
side-channel — correct today (287 §11's measured correction), but coupled to
`setLoopVariable`'s silent early-returns, and every future loop action with an un-listed
reference key silently re-enters the whole-tree search until someone marks it strict.
Benefit: shared mechanism untouched; 287 still closed (the migrations do that alone).

The filing lane's recommendation is **(1)**, with (2) as the acceptable fallback; the
evidence for "the allow-list is a promise nobody keeps" is the strongest single fact in the
file. But the guardian is right that this is a convention decision, not a bug fix, and it is
the owner's to make.

## 6. Interactions

- **Migration `452` did NOT depend on this ruling and is now APPLIED** (16:28:57Z, gate converted
  on measured evidence — guardian's own note plus §7): strict-on-base resolves on the current
  binary. So this RFC no longer blocks 287's closure; it decides the convention only.
- If (3) is chosen, revert forward (`git revert` of the coordinator.go hunk + test file),
  keep WFA-017's register entry with status REVERTED (the register records what existed).
- RFC_029 Phase 2's triage reads these same resolver rows; its lane was notified via
  RFC_029 §10.6b either way.

## Sources

`bugs_open/287` §9a/§11 · council report corr `cba35b35` (verdict REJECTED, guardian veto;
`diagnosis_artifacts` kind=`council_report`) · `0ed96c7eb` (the committed change + tests) ·
WFA-017 (`register/workflow-authoring.md`) · CTS-059/CTS-060 · the 2026-07-28 scope-veto
ruling and `bugs_closed/124` / BLD-019 precedents (code-stays-pending-human, "Live ≠ approved").

## 7. ADDENDUM 2026-08-17 ~16:30Z — one option got cheaper and one premise got measured

Two facts arrived after this RFC was filed; both bear on the owner's choice.

1. **The config half is now LIVE WITHOUT the Go half** (migration `452` applied 16:28:57Z after
   its ordering gate was converted — `bugs_open/287` §11b). It had to be: the "fresh chassis
   build" of 14:42Z shipped **no new code** (same `IMAGE_TAG v1.0.1305`, cached image; old stamp
   `6a782274b` probed present on both pods, Half 1's `0ed96c7eb` absent), so waiting for the roll
   meant leaving ~25 wrong records/hour running indefinitely. **This makes option (3) —
   forward-revert the Go half — genuinely viable rather than merely arguable:** 287 can be
   closed by config alone, if the live verification holds.
2. **What the Go half still buys, stated precisely.** Strict-on-base resolves today only because
   `propagateIterationOutputs` refreshes the un-suffixed key before every substep
   (`coordinator.go:1355`), measured at the resolution moment (**201 `MAPPING_BYPASSED` rows for
   `field=result` vs 155 completions in 6 h**). That is a side-channel: `setLoopVariable` has
   silent early-returns (missing `loop_metadata` / item key), and under `!` those become failed
   items rather than wrong records. The generic suffixing removes the dependency — strict then
   names the reply's own `handler_result_N` directly. So the choice is **"correct via a
   side-channel we can measure" (revert) vs "correct by construction, at the cost of a new
   fleet-wide convention" (ratify)**.
3. **Correction to this file's §3, in the interest of not over-selling (`WRONG_CALLS.md`
   2026-08-17):** the filing lane (me) asserted the base-key presence flatly in `bugs_open/287`
   §11 before doing the rows-per-demand arithmetic, and then briefly mis-refuted it from final
   `collected_data` (a lossy instrument on this agent). The presence claim now rests on the
   event rows above, not on the code path alone. Nothing in §5's option costing changes; the
   evidence under it is firmer and its provenance is now visible.

Also worth the owner's attention when ruling: the same `retry_payload` sibling sits on BOTH the
base and the suffixed keys (140/140 and 548/548 respectively), so the item's stored `result`
will be `{retry_payload, response}` under either option — fatter than the reply alone, and a
separate (small) cleanliness question that neither this RFC nor 287 proposes to solve.

## 8. THE CODE IS LIVE AND PROVEN — 2026-08-17 17:05Z, chassis v1.0.1307 (this changes what the options COST, not what they are)

The owner rebuilt with a bumped tag (the 14:42Z attempt had shipped nothing — same tag, cached
image). Evidence, all measured within 75 minutes of the roll:

- **The generic pass is observably firing:** every `build-dispatch-loop` plan expanded since the
  roll carries `"result!": "handler_result_0"`. The old binary could not produce that — `result`
  was never on the allow-list. Two controls ride in the same object: `work_item_id!` stays
  `current_item.id` (loop variable, not a sibling output) and `error_step` is step-prefixed, not
  data-suffixed. So the shape gate and the first-segment rule are both behaving as specified.
- **No regression anywhere the census predicted one:** `field=result` resolver rows **0** (from
  ~455/day), 10 orchestrations of demand, 11/11 loop completions carrying the handler's reply,
  0 spawn records. The other 22 live sites in the census belong to agents whose loops ran in
  this window (page-content-writer, page-build-handler, tool-generator all appear in the
  resolver rows, i.e. they executed) with **no new error class** attributable to suffixing.

**What this does to the three options:**

- **(1) Ratify** — now costs nothing to enact (it is the running behaviour) and has a live
  disconfirmation window running: any mis-suffixed literal would surface as a resolution failure
  in an agent whose loop runs. None so far, 75 minutes.
- **(2) Narrow to enumeration** — now a behaviour CHANGE to live code, not a decision to hold
  something back. Same argument as before (the allow-list is a promise three actions already
  broke), but it must now be shipped and re-verified.
- **(3) Forward-revert** — likewise a live change; and note it would return `mark_complete`'s
  strict reference to the base key, which resolves only via `propagateIterationOutputs`' refresh.
  That works (it is what migrations 448/452 relied on for the four hours before this roll, with
  the same 4/4 and 1/1 proofs) — but it puts a measured side-channel back under a strict marker
  whose whole purpose is to remove guessing.

**The filing lane's recommendation is unchanged: (1).** The guardian's objection — that a
snapshot census cannot bound future configs — also stands unchanged and is not answered by this
section; what it now has is a live behavioural window on top of the static census, and the
register entry's LANDMINE line telling the next author what the convention is.
