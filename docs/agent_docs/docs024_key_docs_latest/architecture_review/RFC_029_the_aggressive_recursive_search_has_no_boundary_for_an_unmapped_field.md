# RFC 029 — the aggressive recursive search (`findFieldRecursive`) has no stated boundary for a field the caller never mapped, and `bugs_open/248` is its second production incident in one bug file

## STATUS: **RULED 2026-08-15 (§9). Phase 1 IMPLEMENTED 2026-08-15 (§10) — committed, inert until an image roll; Phase 2 OPEN pending the observation window. One D3 premise corrected on evidence (§10.3).**

> The owner, in chat on 2026-08-15, delegated the determination in writing: *"Please think
> about this and determine the best course of action, we'd rather there be no wrong fields
> but if there are what best to do with them. The extensive search at least works some
> (most?) of the time, or should we do a deterministic order - think hard and find the best
> solution."* The determination below (§9) was made by the session after reading the actual
> mechanism (`extractSingleField`, `findFieldRecursive`) rather than this file's prose, and
> records why the owner's own suggested alternative (a deterministic order alone) was
> examined and not chosen as the whole answer.

Filed 2026-08-14 by the `staged_component_build` lane. `bugs_open/248`'s council round 4
(`bug_historian`, corr `7f0c1535-25cb-4645-adba-f7429e357a79`, run
`8a36998f-7188-4cda-9a44-f8a99f71e6a0`, verdict `complete_revise`) raised an objection that is
**not** a defect in the two edits under review:

> should the shared fallback strategy ever run for a field with no explicit `input_fields`
> entry, at all? […] Not a veto — should be flagged to a human independent of this round's
> disposition.

By CLAUDE.md's own council-gate section, a scope-shaped objection "is not answered by
resubmitting with better measurements … route the seam to architecture review on its own
merits, and let a human break it." The lane stopped resubmitting on R4 and is filing this
instead. Full round-by-round record: `bugs_open/248_HANDOFF_2026-08-10_undeployed_asset_repair_deploys_every_asset_as_a_hero_under_a_placeholder_name.md`.

## 1. What the mechanism actually is, plainly

Every action a workflow step calls declares which fields it needs (`Required`/`Optional` in
its `ActionInputSpec`). The step's own config can name, explicitly, where each field comes
from — a literal value, or a dot-path into the run's accumulated data (`input_fields` /
`input_mapping`). `ExtractActionInputs` (`platform/orchestration/datahelpers/action_inputs.go`)
is the function that turns "what the action needs" plus "what the config said" into the actual
values the action runs with.

When the config's explicit mapping does not name a field — either because the step's
`input_mapping` never mentions it, or because it maps a whole nested object (`"spec":
"current_item.spec"`) instead of the field itself — `ExtractActionInputs` does not fail or
leave the field unset. It calls `ExtractFields`
(`platform/orchestration/datahelpers/unified_extractor.go`), which for each still-missing field
runs its **own** five-strategy chain, the last resort of which is `findFieldRecursive`: an
unbounded (depth-capped at 20, not field-capped) walk of the **entire** `collected_data` tree
for a map key matching the field's name, unwrapping and recursing into every nested value
that is not on an infrastructure-key skip-list.

**The rule this makes true, and which nobody had stated as a rule before `bugs_open/248`:**
a field with no explicit mapping is not "absent" to this platform — it is "resolved by
whatever else in the run's entire accumulated state happens to carry the same key name,"
searched in Go map-iteration order (documented as non-deterministic in the code's own
comments; see §4). That is a much stronger, and much less obviously safe, guarantee than
"the config decides where values come from," and it was never decided on purpose — it fell
out of a function named for what it does (`findFieldRecursive`) rather than for what it
promises.

## 2. This is not RFC_028, and here is exactly how it differs

`RFC_028` (open, unresolved) is about the **same overall resolver** and is the right sibling
reading, but it audits `ExtractActionInputs`'s own eight numbered arms (Strategy 0 through 6,
all living in `action_inputs.go`) and asks whether that chain needs an owner, a deduplicated
discriminator, and an arm-count budget.

**RFC_028's own arm count is a floor, and this bug file is the evidence for why.** Strategy 1
and Strategy 2 of that chain — the two arms RFC_028 counts as one step each — both delegate to
`ExtractFields`, which is a **second, independent five-strategy chain living in a different
file** (`unified_extractor.go`), never enumerated in RFC_028 because its own census query
matched only `action_inputs.go` / `ExtractFields` by name in `fix_plan` text (see RFC_028 §2's
own caveat: "sees the architecture seat's structured signal and misses any round where another
seat made the same point in prose"). `bugs_open/248` is exactly such a round, twice: council
R1 (`editquality`, `image-build-handler`'s `call_asset_deployer`) and R2 (`editquality` again,
`build-dispatch-loop`'s repair path) both gated on this inner chain's last arm, and neither
round's `fix_plan` names `action_inputs.go` as the site of the fix — the fixes are one-line DB
config additions (migrations 401, 402), not code changes to that file at all. **RFC_028's 8-of-27
census undercounts this specific mechanism's architecture-signal rate by at least 2, and by a
file its own query structurally cannot see.**

So: RFC_028 asks "does the outer chain have too many arms and no owner." This RFC asks a
narrower, more urgent question about **one specific arm of the inner chain it never counted**:
should that arm — the whole-tree name-matching search — run at all for a field with zero
explicit mapping, or should "unmapped" be a hard resolution failure instead of a guess?

## 3. Two independent production incidents, same arm, same bug file

- **R1** (migration 401): `image-build-handler`'s `call_asset_deployer` step mapped
  `{domain, s3_uri, purpose, asset_key?}` to the spawned `asset-deployer` agent — no `asset_id`
  key at all. `asset-deployer` declares `asset_id` as a field it wants. With no explicit
  mapping, resolution fell through to `findFieldRecursive` across the whole run's
  `collected_data`. **Measured, not assumed** (migration 401's own text): 5 real pre-fix
  completions all show `asset_id` **empty** in the resolved input — the safe-but-incomplete
  outcome, not a wrong-asset pull, in the sample examined. The council objection this addressed
  (R1, `editquality`, HIGH) was precisely that the *other* outcome — the search finding a
  **different** asset's stale `asset_id` from elsewhere in the tree — was untested and
  plausible, not that it had been observed.
- **R2** (migration 402): the repair path (`check_undeployed_assets` →
  `build-dispatch-loop` → `asset-deployer`) carries `asset_id` **nested** under `spec`
  (`"spec": "current_item.spec"` is the only top-level mapping), so the same arm engages for
  the same reason: the field exists in the run's data, but not at a path the caller's explicit
  config names. Same class of gap, different caller, found because R1's own resubmission had
  asserted this path "safe" by observing the value's presence, not by tracing how it resolves —
  which is the exact shortcut this bug file's own diagnosis section warns against.

**Both were fixed the same way: name the field explicitly** (`"asset_id?": "<real path>"`),
which makes Strategy 0 (an explicit dot-path) resolve it deterministically and the aggressive
search is never reached for that field on that caller again. **Neither fix touched the arm
itself** — it still runs, unbounded, for every other field on every other caller that has not
been individually walked and patched. R4's objection is that this is now the third
patch-one-caller-at-a-time round on the same underlying arm (a fourth if WFA-009's sibling
problem, §5 below, is counted), and nothing yet decides how many more callers get found the
same way, one production symptom at a time.

## 4. What R4's own proposed checks found, run and clean

R4 proposed two concrete checks rather than only naming the concern; both came back clean and
are recorded here so the owner does not need to re-run them:

- **No third live caller of `asset-deployer` exists** beyond the two already patched
  (`image-build-handler`, `build-dispatch-loop`) — checked directly; a plausible-looking match
  in `render-audit-agent` was inspected and is descriptive prose, not a real dispatch edge.
- **`asset-deployer`'s own `input_contract` does not list `asset_id` as required at all** — so
  even the framing "a required field fell through to a guess" is soft; it is an optional field
  the action happens to use when present, which is arguably a stronger reason the resolver
  should not go hunting the whole tree for it uninvited.

Neither check changes the answer to §2's question; both narrow what shipping migrations 401/402
actually closed (two named callers) versus what remains open (the arm itself, for every other
field/caller pair that has not yet produced a symptom).

## 5. Precedent already exists for the shape of an answer

The concept register's **WFA-009** (`ExtractFieldsAction`, a related but distinct mechanism —
see the naming collision noted in §6) solved the sibling problem — a fallback path list that
misses every path and silently continues — with an **opt-in `required` list, default OFF**,
per the owner's 2026-08-02 §2 ruling: *"new authority on a shared seam ships as an opt-in field
with the unsafe default OFF."* That is a live, shipped, owner-approved precedent for exactly
this class of question, on an adjacent mechanism in the same codebase. This RFC's §7 asks
whether the same shape — an opt-in "this field must resolve via an explicit path or the step
fails loudly" — should extend to `ExtractFields`'s aggressive search, rather than asking the
owner to invent a new answer.

## 6. A naming collision worth registering regardless of the ruling

Both files number their fallback strategies from a low integer, independently, and the
migration text that fixed R1/R2 already conflates them: migration 402's own comment says
resolution "falls through to Strategy 4's aggressive `findFieldRecursive` search," but
`action_inputs.go`'s **own** Strategy 4 is a different thing entirely ("resolve remaining
config value references," `action_inputs.go:740`) — the aggressive search is
`unified_extractor.go#extractSingleField`'s Strategy 4, one file and one function call away.
Two independently-numbered "Strategy 4"s implementing two different rules, in two files that
call into each other, is exactly the kind of scattered-guarantee problem RFC_028 §5.1 already
flagged for the outer chain alone ("today its guarantees live in five places … nothing keeps
them so"). This RFC adds a sixth: the migration record that fixed the two live incidents this
mechanism caused already has the arm number wrong. Worth a fix regardless of how §7 is ruled —
renaming either chain's arms so the numbers do not collide, or naming the inner arm
descriptively instead of numerically, removes a standing way to cite the wrong code when
tracing the next incident.

## 7. What this RFC asks the owner to decide

1. **Should a field with no explicit mapping (no `input_fields`/`input_mapping` entry, no
   matching literal or dot-path) ever be resolved by `findFieldRecursive`'s whole-tree search —
   or should "unmapped" be a hard resolution failure** (nil, or an error naming the field and
   the paths that were tried, per `WFA-009`'s already-shipped idiom), gated behind an opt-in
   flag with the unsafe (search) behaviour remaining the default, consistent with the
   2026-08-02 ruling?
2. **If yes to an opt-in hard-fail: who takes it first?** The two callers migrations 401/402
   already patched no longer need it (Strategy 0 now resolves them deterministically) — the
   candidates are the *other* fields on those same two callers, and any future caller of an
   action with optional fields nobody has walked yet. A default-OFF flag with no adopter is a
   mechanism that "rots unexercised," the exact cost the 2026-07-29 ruling weighed when it
   declined a blanket default-OFF requirement — so this wants a named first adopter, not a bare
   capability.
3. **Should RFC_028's arm-count question (its §5.3) be re-scoped to include the inner chain**,
   given §2 above shows the outer count alone already undercounts by at least the two rounds
   this bug file produced? This RFC does not re-ask RFC_028's other two questions (contract
   ownership, discriminator dedup) — they stand as filed — but flags that whichever RFC ends up
   ruled on the arm-budget question should rule on both files' arms together, or a budget on
   one chain while the other is uncounted measures nothing.

## 8. Deliberately not proposed here

- **Reverting migrations 401 or 402.** Both are narrow, already-reviewed fixes for named,
  measured callers; §7 is about the arm they routed around, not about undoing them.
- **A redesign of `ExtractFields` or its five internal strategies as a whole.** Out of scope —
  this RFC is about whether one arm (the last-resort whole-tree search) should have a boundary
  for the unmapped case, not about the other four.
- **Auditing every other caller of every action with optional fields for the same latent gap.**
  That is the kind of sweep an owner decision on §7.2 would actually schedule, not something to
  pre-empt here by guessing at scope.

## 9. RULING 2026-08-15 — "unique-or-nothing": the search stays, but it may never guess

First, what the mechanism does today, in plain terms: when a field has no explicit mapping,
the platform walks the run's entire accumulated data looking for any key with the same name,
and takes the FIRST one it meets. Go randomises the order maps are walked in, so when two
different values share that key name, which one wins is a per-run coin flip.

The owner's stated preference ranks the outcomes: **no field at all is better than a wrong
field.** Measured against that ranking, the three candidate answers come out as follows.

- **Remove the search (hard-fail all unmapped fields):** rejected. The dependency surface is
  unknown and currently unmeasurable — the only evidence the arm fired is an INFO log line,
  and this session measured the chassis scrolling 50,000 log lines in minutes with zero
  extraction events in the window, so no one can say today how many field/caller pairs
  resolve this way. Removing the arm risks breaking every silent dependent at once, on a
  tree where every Go change ships fleet-wide on the next roll.
- **Deterministic order alone (the owner's suggested alternative):** examined and not chosen
  as the whole answer, because determinism makes a wrong pick REPRODUCIBLE, not RIGHT. A
  sorted traversal replaces "random winner" with "lexicographically-first winner" — arbitrary
  in a different way. Where candidates agree, determinism costs nothing and we take it; where
  they disagree, ANY picking rule is a guess, and a guess is exactly what the owner's ranking
  forbids.
- **Unique-or-nothing (chosen):** the search collects ALL matches (same depth cap as today).
  If every match carries the same value, it resolves — deterministically, shallowest path
  first — and behaviour is unchanged from today's happy path. If the candidates CONFLICT, it
  resolves NOTHING and logs a WARN naming the field and every candidate path. A pipeline that
  today depends on the coin flip landing right is already broken nondeterministically; this
  converts its failure from a silent wrong value into a visible, stable absence with its own
  log line — the only option on the table that can never produce a wrong field.

### The four decisions

**D1 — `findFieldRecursive` becomes collect-all / unique-or-nothing** as above. Same depth
cap, same infrastructure-key skip list; traversal order made stable (sorted keys) so the
candidate list and the shallowest-first winner are reproducible.

**D2 — Instrument first, refuse second (two chassis builds).** Phase 1 ships the collector
and the determinism but conflicts still resolve (to the stable shallowest winner) while a
WARN (`aggressive search: conflicting candidates`) records field, candidate paths, and the
winner. After an observation window (48h minimum, a week preferred), Phase 2 flips conflicts
to refusal. Precondition for the flip: zero conflict WARNs observed, or every observed
field/caller pair given an explicit mapping first. This is the closest thing to a staged
rollout this estate's deployment model permits, and it directly answers the "works some
(most?) of the time" question with a measurement instead of a guess.

**D3 — the opt-in strict marker ships (Q1's second half), and Q2's first adopter is
`asset_id` itself.** Per the owner's 2026-08-02 §2 ruling (and WFA-009's shipped precedent):
a per-field `!` suffix in a caller's `input_mapping` — the mirror of the existing `?`
optional suffix — meaning "explicit resolution only; if unmapped or unresolved, fail the
step loudly; never the search." Default OFF. First adopters: the `asset_id` mappings on both
already-patched callers (migrations 401/402, `image-build-handler`'s `call_asset_deployer`
and `build-dispatch-loop`'s repair path). Both are already explicitly mapped, so adopting
`!` there changes nothing on the happy path while converting the two shipped fixes from
current-state accidents into enforced invariants — a regression that re-exposes `asset_id`
to the search becomes a loud failure instead of a silent guess. A default-OFF mechanism with
a named, shipped adopter is what the 2026-07-29 ruling's "rots unexercised" cost demanded.

**D4 — Q3 answered: the arm budget covers BOTH chains.** RFC_028's D3 (ruled 2026-08-15,
commit `260cb2393`) built `resolver_arm_budget_test.go`: an AST-walk count of the outer
chain's write sites, floor pinned at the exact count (10), ceiling 15, mutation-proven both
ways. Extend the same pattern to the inner chain: count `extractSingleField`'s resolution
return sites (5 today), floor 5, ceiling **8** — generous, per the owner's instruction on the
sibling question. And the §6 naming collision is resolved as part of the same change: the
inner chain's log/comment names drop "Strategy N" for descriptive names (direct-path,
input-data-prefix, input-data-map, whole-tree-search, alias), so no future migration text can
cite the wrong file's arm number again; migration 402's already-wrong comment gets a dated
correction note when the rename lands.

### Implementation notes (for whoever takes it)

- All of it is chassis Go — inert until an image roll; platform code, so one council-gate
  submission for the coherent task, before or alongside the commit, per CLAUDE.md.
- **Repair `TestDefaultBeatsTheRecursiveSearch` first** (`action_inputs_strategy6_test.go`):
  it fails on pristine HEAD with its own vacuity control reported stale (found and filed by
  the RFC_028 implementation session, commit `260cb2393`; also referenced in
  `bugs_open/274`'s context). It is the exact invariant D1 extends — building on a failing
  control is how vacuous passes are born.
- **What would disconfirm this ruling:** Phase 1 showing a substantial population of
  conflict WARNs whose lucky winner is load-bearing (each mapping to a pipeline that would
  break on absence). In that case Phase 2 does not proceed on schedule; the pairs get
  explicit mappings first, and this section gets a dated correction re-examining the premise
  that conflicts are rare and usually already broken.

## Sources

- Council rounds: `bugs_open/248`, corr `7f0c1535-25cb-4645-adba-f7429e357a79` (R1 through R4,
  four REVISE verdicts, `editquality` gating R1/R2, `guardian` gating R3, `bug_historian`
  raising the architecture question in R4).
- Code: `platform/orchestration/datahelpers/action_inputs.go` (Strategy 0/1/2, the delegation
  to `ExtractFields`), `platform/orchestration/datahelpers/unified_extractor.go`
  (`extractSingleField`, `findFieldRecursive`, lines ~399–520).
- Migrations: `docs/agent_docs/sql_for_agents/401_image_build_handler_explicit_asset_id_mapping.sql`,
  `docs/agent_docs/sql_for_agents/402_build_dispatch_loop_maps_asset_id_top_level.sql` (both
  quote the mechanism in their own header comments; migration 402's is the source of the
  Strategy-4-naming collision noted in §6).
- Sibling precedent: register **WFA-009** (`ExtractFieldsAction`'s opt-in `required` list,
  `platform/orchestration/actions/v3_site_actions.go`), shipped under the owner's 2026-08-02 §2
  ruling on opt-in fields with an unsafe default.
- Companion RFC: `RFC_028_the_input_resolver_precedence_chain_has_no_owner_and_draws_the_architecture_signal_in_30_percent_of_its_rounds.md`
  (same overall resolver, the outer chain's ownership/budget question — read together, per §7.3).
- Prior rulings this RFC rests on: 2026-07-28 (platform seams and the ordering exemption),
  2026-07-29 §1/§2 (the guarantee test; no blanket default-OFF requirement), 2026-08-02 §2
  (opt-in fields on shared seams ship with the unsafe default OFF), 2026-08-11 (RFC_022's
  accumulated-count narrowing).

---

## CONTRIBUTION 2026-08-15 from the `bugfix_213_verifier_producer_join` lane — a THIRD production incident, with a DIFFERENT entry condition that your ruled remedy does not obviously cover

Contributed into this file rather than filed separately: it is your mechanism, your file, and
your ruling, and this is a case for it — not a competing account. **No action taken on the
code.**

### 1. The incident

`bugs_closed/213` §D: `dark_section_audit` work items reach `complete` carrying a `result`
payload their own handler never produced. **10 of 14 completed rows**, and the split reproduced
**live at 3:1** when gate 1b's abstain arm instrumented it on 2026-08-14 — so it is systematic,
not historical. The two foreign shapes are `[color_scheme design_notes spacing typography]` and
`[add_to_page approach new_page not_actionable reasoning retype_existing update_spec]`, which a
`090` run cited as matching **`webdesign-agent`'s `design_spec`** and
**`content-gap-planner`'s `gap_plan`** `complete_workflow` `output_fields` exactly.

### 2. The chain, read at source today — every step verified, none inferred

1. `build-dispatch-loop` completes items at
   `workflow.steps.process_item.config.` **`sub_workflow`** `.mark_complete_step`:
   `{"action": "complete_work_item", "config": {"result": "handler_result",
   "work_item_id": "current_item.id"}}`.
   ⚠ *Nested inside another step's config — a `$.*` search of `workflow->steps` does not see it.
   Ask `$.**` of the whole `default_config`. That path cost our lane a week.*
2. `CompleteWorkItemInputSpec` declares `Optional: []string{"result", "commit_sha"}`, so
   **`result` is in `allFields`** (`load_work_item_actions.go:52-62`).
3. `ExtractActionInputs` **Strategy 0** resolves config values as explicit paths — but only when
   `IsDottedPathReference(pathStr)`, which is literally `strings.Contains(s, ".")`
   (`action_inputs.go:597`). **`"handler_result"` has no dot, so Strategy 0 SKIPS it.**
4. **Strategy 2** then calls `ExtractFields(collectedData, allFields, …)` →
   `extractSingleField(data, "result", …)` → its own Strategy 4,
   **`findFieldRecursive(data, "result", 0, …)`** — any key named `result`, anywhere, to depth
   20 (`unified_extractor.go:440-445`).
5. `result.Values["result"]` is now set.
6. **`ExtractActionInputs` Strategy 4 — the arm that exists precisely to resolve single-segment
   references, and whose own comment gives `"spec_data": "site_plan"` as its example — then
   SKIPS the field, because `if _, hasValue := result.Values[field]; hasValue { continue }`.**

### 3. Why this is a different entry condition from the two incidents in §3, and why that matters to the ruling

Your §1 frames the trigger as *"when the config's explicit mapping does not name a field"*.
**Here the config DID name it, and named it correctly.** The mapping was silently overridden
because it lacked a dot — the aggressive search ran on a field the caller had explicitly bound.

Strategy 0's own comment says it was added *"because ExtractFields uses aggressive recursive
search that can find stale values from previous loop iterations"*. **It fixed exactly this
problem — for dot-paths only** — and single-segment mappings were left on the far side of the
guard that was built for them. So the fleet has a mapped-field class as well as an unmapped-field
class, and only the unmapped one is written up.

### 4. ⚠ THE PART WORTH YOUR ATTENTION BEFORE IMPLEMENTATION: "unique-or-nothing" may not fix this case

Your §9 ruling lets the search *"resolve a unique candidate, never guess between conflicting
ones"*. **If `collectedData` holds exactly ONE key named `result` — the foreign one — it is
unique, and unique-or-nothing resolves it wrongly with full confidence.** That is precisely our
shape: the handler's own envelope is under `handler_result`, so the only bare `result` in scope
may well be a sibling step's. Uniqueness is a defence against *ambiguity*; this incident is not
ambiguous, it is *confidently wrong*.

The cheap addition, offered rather than argued: **when the config explicitly maps a field, the
mapping should win — or the field should fail — regardless of whether the mapping contains a
dot.** That is Strategy 0's existing intent with the `strings.Contains(s, ".")` test relaxed,
and it needs no new arm (RFC_028's budget is untouched: it moves work between two existing arms
rather than adding a ninth).

### 5. What this contribution does NOT establish

**[NOT VERIFIED]** that a foreign bare `result` key was actually present in `collectedData` at
the moment those specific items completed. The chain above proves the mechanism is *available*;
it does not prove it *fired* for those 10 rows. The `090` on this symptom returned
**`UNVERIFIABLE`** (correlation `adecf408-1e60-4293-8b22-351ddbb52a08`) — it did not confirm a
mechanism, and its `ExtractFields` citation is what sent this lane to read the code. Confirming
the firing needs a live `collectedData` capture from a `build-dispatch-loop` run, which nothing
currently retains.

**Free instrumentation you already have:** gate 1b (`WII-017`) records every unreadable payload
to `agent_error_log` as `NO_CHANGE_GATE_UNREADABLE_RESULT`, **with the payload's top-level keys
in the message**. If this RFC's fix lands, that stream should stop naming foreign shapes — a
ready-made before/after that costs nothing to collect. ⚠ It is currently quiet only because
`improvement-sweep` is `enabled=false`.

— `bugfix_213_verifier_producer_join` lane (bug closed 2026-08-15; this is a post-closure
contribution, not a reopening)

> **ANSWERED by §10 below (same day):** Phase 1 now measures this exact class — a WARN
> (`aggressive search: explicit single-segment mapping bypassed`) fires whenever the search's
> answer for a field with a dotless config reference differs from what that reference resolves
> to — and the `!` strict marker is the shipped opt-in remedy for any caller that wants the
> mapping to win or fail today (a strict dotless reference resolves as a single-segment
> reference, never the search). Your §4 point is conceded in full: unique-or-nothing does not
> cover this shape, so whether the DEFAULT behaviour flips (your "mapping wins regardless of
> dot" suggestion) is decided by the observation window's data, not pre-empted here.

---

## 10. IMPLEMENTATION 2026-08-15 — Phase 1, committed (inert until an image roll)

Implemented by the `staged_component_build` lane, one coherent council-gated task. All Go;
nothing below is live until a chassis image carrying it rolls.

### 10.1 What shipped, decision by decision

- **D1/D2 Phase 1** — `findFieldRecursive` collects ALL matches (sorted-key DFS, same depth
  cap, same infrastructure skip-list), resolves the shallowest-first winner deterministically,
  and on conflicts logs WARN **`aggressive search: conflicting candidates`** (field, every
  candidate path, winner, `phase: 1-resolve-and-warn`) while still resolving. Tests pin:
  unique-resolves-silently, conflict-warns-with-stable-winner, shallowest-beats-DFS-order,
  200-run determinism. The pre-existing flaky controls this exposed were repaired first, as
  §9's implementation notes required: `TestDefaultBeatsTheRecursiveSearch`'s control asserted
  one winner of a four-way map-iteration race (single-candidate fixture now), and
  `TestBridgeIsUnchangedForANonDefaultedField`'s fixture let the search outvote the bridge on
  a coin flip that determinism froze on the wrong side (sources moved under skip-listed keys).
- **D3** — the `!` strict marker ships on BOTH mapping surfaces. In `ExtractActionInputs`
  step config, `"field!": "<reference>"` = explicit resolution only (dotted path or
  single-segment reference), never the whole-tree search / nested-object fallback /
  deprecated bridge; unresolved (a still-standing Default counts as unresolved) → the
  extraction errors naming the field and marker. In `ResolveInputMapping`/`WithItem`, `!` is
  the loud spelling: the `?`→`!` transition converts silent-skip into hard-fail; the suffix
  is stripped before delivery (test pins that the child receives the bare name).
  `UnknownConfigKeys` recognises the strict spelling of a declared field; `relaygaps.go`
  strips `!` like `?`. First adopter: migration **`417_image_build_handler_asset_id_goes_strict_HOLD.sql`**
  — held until the roll (see §10.3 and the LANDMINES entry for why the order is load-bearing).
- **D4** — the inner chain's arm budget exists: an AST count of `extractSingleField`'s
  non-nil resolution returns in `resolver_arm_budget_test.go` (floor **5** pinned exact,
  ceiling **8**), mutation-proven both ways (a real added return moved 5→6; a lowered ceiling
  failed with the RFC message). The inner arms are renamed descriptively — direct-path,
  input-data-prefix, input-data-map, whole-tree-search, alias — and migration 402 carries a
  dated correction for its "Strategy 4" miscitation (§6 closed).
- **213 contribution** — instrumented, not yet flipped: WARN
  **`aggressive search: explicit single-segment mapping bypassed`** fires when the search's
  answer for a field with a dotless config reference differs from what the reference itself
  resolves to (Default-blocked fields excluded — those are CTS-059's case, not this one).

### 10.2 The observation window (what Phase 2 waits for)

After the roll, grep both WARN messages fleet-wide for 48h minimum (a week preferred):
`kubectl -n ai-persona-system logs -l app=agent-chassis -f | grep "aggressive search:"` —
**stream, do not `--tail` an old pod** (CTS-059's measured 92-second retention window).
Phase 2 (conflicts resolve NOTHING) proceeds only on §9 D2's precondition: zero conflict
WARNs, or every observed field/caller pair explicitly mapped first. The bypass WARN's
population decides whether the 213 class needs the default flipped or stays opt-in via `!`.

### 10.3 DATED CORRECTION to §9 D3 — one named adopter was wrong on the ruling's own evidence

§9 D3 says both 401/402 callers adopt `!` and "adopting `!` there changes nothing on the
happy path". Measured against the two migrations' own texts before implementing:

- **`image-build-handler` (401): ADOPTED**, via `417_HOLD`. Live measurement 2026-08-15:
  13/13 asset-deployer children of image-build-handler parents in the retained window carry
  `asset_id`; zero refusal-branch spawns (the locked/no-asset-URL branches 401 kept non-fatal
  with `?`). After the flip a refusal-branch spawn hard-fails the step loudly — a real change
  on a measured-zero branch, and the intended one: a loud absence beats a silent guess.
- **`build-dispatch-loop` (402): NOT ADOPTED, and must not be.** That mapping is shared by
  EVERY dispatched item type — 402's own measurement: exactly one (item_type, handler_agent)
  pair fleet-wide carries `spec.asset_id`, against 636+ item types flowing through the same
  `call_handler` step. Its `?` is per-item-type optionality doing real work; `!` there
  hard-fails every non-asset dispatch in the fleet. The repair path keeps `?` and is covered
  by D1/D2 instead (post-Phase-2, a conflicting foreign `asset_id` resolves to nothing rather
  than a guess). If the owner wants the repair path strict, that needs a per-item-type
  mechanism this ruling did not design — flagged, not built.

The `!`-before-roll trap (an old binary reads `field!` as an ordinary field and silently
re-arms the search) is registered in LANDMINES.md and in the concept register as **CTS-060**.

### 10.4 DATED REVISION 2026-08-16 — the council said REVISE, the objection was right, the WARNs are now rows

**Verdict on the Phase 1 submission (corr `75091072-9d65-433e-8a30-84719dc3f30f`, run
`ae2a88a7`, 2026-08-15 14:10Z): REVISE**, decided by a gating HIGH objection from the
`reuse_agent` seat. Approve: architecture, constitution, mission, guidelines. Object:
editquality, reuse_agent, tooling_provenance, guardian, debug_historian, prior_art_librarian.

**The gating objection is a real defect in §10.1/§10.2 as written.** Both Phase 1 WARNs were
plain zap log lines, and chassis pod log retention is ~90 seconds (CTS-059) — so the 48h+
observation window that §9 D2 makes the precondition for Phase 2 **could not be read after the
fact**. §10.2's "stream, do not `--tail`" was an honest instruction and still not an
instrument: a window observable only by a human tailing a pod live is not observable. The
platform's own remedy already existed (`agent_error_log` via `platform/orchestration/agenterrors`,
RFC_012's leaf writer), which is exactly the reuse-before-create defect the seat names.

**What the revision adds (all Go; inert until a chassis image ≥ its commit rolls):**

- **Every occurrence of both WARNs is also persisted to `agent_error_log`** — no dedup, no
  sampling, because frequency is the population §9's disconfirmation clause needs. Rows:
  `error_code` **`RESOLVER_CONFLICTING_CANDIDATES`** (context: field, candidate_paths,
  winner_path, phase) and **`RESOLVER_MAPPING_BYPASSED`** (context: field, reference,
  resolved_type); `severity='warning'`; `action='input-resolver'` (deliberately not a
  registered action name — the resolver runs inside every action); `error_message` is the WARN
  text verbatim so a row and a line are joinable by eye. **The log lines stay** (live tailing).
- **Mechanism: a registered sink, not a threaded DB handle.** `findFieldRecursive` and
  `ExtractActionInputs` carry no `*sql.DB`/`ctx`, and threading one through ~115 call sites was
  not on. `datahelpers.SetResolverFindingRecorder` (new, `resolver_findings.go`) is a
  package-level recorder, **nil by default = log-only = the previous build's behaviour** (the
  default-OFF shape of the 2026-08-02 §2 ruling). The chassis registers one at startup in
  `agentbase.initializeComponents`, right after the pool is created — the one place the DB and
  the pod identity both exist — as a thin wrapper over the ONE writer (`orchestration.LogAgentError`
  → `agenterrors.Write`), synchronous under a detached 5s timeout exactly like the other
  agent_error_log recorders in that file. A recorder panic is recovered inside datahelpers; the
  resolver's answer can never change because of the instrument.
- **Known limit, stated on every row:** per-run identity (`orchestration_id`, `step_name`) is
  not reachable from the resolver, so rows carry pod-level attribution only (`pod_name` +
  `agent_type`); each row's context says so (`identity_scope`) so the blank column reads as a
  stated limit, not a bug. Two standing traps for whoever queries: `agent_error_log.domain` —
  "no domain" is `COALESCE(domain,'')=''`, never `IS NULL`; and these rows have empty
  `orchestration_id` by design.
- **Tests** (`resolver_findings_test.go`, `resolver_findings_bridge_test.go`): with a fake
  recorder installed, ONE conflict → exactly one finding (code, field, all 3 candidate paths,
  winner) and a second conflict → a second finding; ONE bypass → exactly one finding; the
  agreeing / never-resolvable controls → zero findings (row population == WARN population);
  no recorder → same value, same WARN (the default-OFF control); a panicking recorder cannot
  change the answer. **Mutation-proven both sites**: removing either `recordResolverFinding`
  call fails its test. Arm budgets unmoved (outer 10/15, inner 5/8) — the recorder writes no
  `result.Values` and adds no return site. Whole tree builds from `git archive HEAD` + these
  files.

**§10.2 is superseded by this:** the observation window is read from rows, not from a log
grep, and **its clock starts at the roll of the REVISED build**, not Phase 1's —
```sql
SELECT error_code, context->>'field' AS field, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log
WHERE error_code IN ('RESOLVER_CONFLICTING_CANDIDATES','RESOLVER_MAPPING_BYPASSED')
GROUP BY 1,2 ORDER BY 3 DESC;
```
Phase 1's log-only build DID roll before this revision — measured 2026-08-16: the running
chassis binary is stamped `5e075a6f9` (v1.0.1303; must-be-absent control HEAD `bc4cd65e7`
absent), a descendant of `1806371ef` — so the log-only WARNs are live now and, by this
section's own argument, unreadable. A ~4,000-line, 2-pod `--tail` sample that morning showed
the resolver's INFO lines (search reached, one "Found via aggressive search") and **zero WARNs
of either kind** — consistent with §9's premise, and far too small a sample to be evidence of
it, which is the whole point of the rows. **The `!` parser is therefore live too, so
migration 417's binary precondition is met** — its header now carries the two debug_historian
checks as measurements (1 active row; the two-arg `snapshot_agent` overload → `agent_definitions_backup`).

**The other five seats' objections, answered without code:** editquality's two "missing D4"
items were in fact shipped in `1806371ef` (the `extractSingleField` renames and migration 402's
dated correction) — the submitted plan failed to LIST them, and the resubmission names them as
edits; guardian's "the winner changes now" is §9 D2's explicit owner-delegated choice (stable
shallowest-first over a coin flip, instrument-then-refuse); tooling_provenance's ledger check
is DONE (`schema_migrations`: `417_brief_fidelity…` applied, `417_image_build_handler…HOLD`
unclaimed, ledger keys on the full filename so no collision) and its doc_notes ask is met by
two rows written the same day (`decision`/`RFC_029` — the ruling's key lines and what shipped,
so a seat can read them in-DB; `decision`/`council-submission-75091072` — this round's evidence
with the queries that produced it), which is also prior_art_librarian's answer (owner rulings
are invisible to seats: known gate landmine).

**ROUND 2 VERDICT — READ 2026-08-16 10:2xZ: APPROVED** (run `b5678c3a`, completed 10:16:40Z,
"approved with 3 advisory objection(s) — none high-severity"; approve: editquality,
guidelines, tooling_provenance, diagnosis_guardian, debug_historian, constitution, mission,
architecture; object (advisory): reuse_agent, guardian, prior_art_librarian; 6 abstained).
Commit `53edef286` carries `Council-Submitted:`, which the coverage report now credits — no
amend (forward-only). A duplicate run (`d1a20669`, my mis-fired second publish) was still in
seats at read time; it judges the identical plan and is a consistency check only.
The advisory points, answered on evidence rather than waved past:
- **reuse_agent — "did the plan consider `LogActionEntryFindings` / `…InheritingProvenance`?"**
  Measured after the verdict: every variant takes an `ActionParams` and lives in
  `platform/orchestration/actions` (`log_action_error.go:278–326`), a package that imports
  `datahelpers` in 260 files — so the resolver cannot call it without a cycle, and the
  provenance those helpers inherit (`orchestration_id`, site, step from `params`) is exactly
  what the resolver does not have. The registered sink is therefore required by the
  dependency direction, not a preference; the WRITE half is the same one writer either way.
- **guardian — "does one chassis process ever host more than one Agent?"** No: `agentbase.New`
  has exactly one caller, `cmd/agent-chassis/main.go:209`; the process-wide recorder's
  identity is the process's identity. The `a.db == nil` guard is belt-and-braces — the
  registration line runs after `a.db = db` and the closure reads `a.db` at call time.
- **prior_art_librarian — the 1-active-row claim "asserted, not checked".** It IS checked, in
  the DB where the seat can read it: `doc_notes` `decision`/`council-submission-75091072`
  item [4] carries the query and its output (`1, {1}`). The seat's own procedure did not
  look there; the row exists for the next round that does.
- **architecture — the sink is not a precedent to be waved through for a second finding
  type.** Accepted; CTS-060's wording is adjusted to say exactly that (a second finding type
  or consumer → a fresh architecture look, not "already established").

### 10.5 FIRST READ OF THE OBSERVATION WINDOW — 2026-08-16 ~15:10Z (window opened 10:41Z): §9's disconfirmation clause is FIRING; Phase 2 does NOT proceed on schedule

**The revision is LIVE**: chassis `v1.0.1304`, binary stamped `5de6cddbe` (probed `/proc/1/exe`;
`53edef286` is its ancestor); pods started 10:41Z; first row 10:42Z. The instrument works.

**Population after ~4.5 h `[MEASURED]`** (`agent_error_log`, both codes): **672 rows**.
- `RESOLVER_CONFLICTING_CANDIDATES`: `current_page` **245**, `work_item_id` **207**, `result` 77,
  `sections` 12, `page_type` 12, `reason` 1.
- `RESOLVER_MAPPING_BYPASSED`: `result` **118**, every one with `reference='handler_result'`
  (dotless) and the search's answer a `map[string]interface{}` from elsewhere — the
  bugs_closed/213 §D shape, at ~26/h.
- By agent: **build-dispatch-loop 608**, page-content-writer 33, page-build-handler 27,
  page-rerender 3, rerender-pages 1.
- Candidate counts per conflict are LARGE: `work_item_id` conflicts carry **21–93 candidate
  paths**, winner `claim_result.work_item_id` (shallowest); `current_page` conflicts carry
  4–62, winner mostly `handler_result.retry_payload.message.body.~unwrap.current_page`
  (a retry payload — whether that is the RIGHT page for the step is exactly the question).
  `result` conflicts: 7 candidates, winner `handler_spawned.result`.

**Reading, marked:** `[INFERRED — not yet diagnosed]` build-dispatch-loop's `collected_data`
accumulates across items in one long-lived orchestration, so each iteration's `work_item_id`,
`current_page`, `result` meets every previous iteration's copies; the stable shallowest winner
may be the current item's (`claim_result.*`) or a stale one (`retry_payload…`) — items are
COMPLETING, which says the winner is at least not fatal, not that it is right. **This is a
mechanism claim about a shared loop; it goes through `090` before anyone acts on it.**

**What this decides now:** §9 D2's precondition ("zero conflict WARNs, or every observed pair
explicitly mapped first") is nowhere near met, and §9's own disconfirmation clause ("a
substantial population of conflict WARNs whose lucky winner is load-bearing") is at least half
satisfied on day one — substantial, yes; load-bearing, to be established per pair. **Phase 2 is
NOT to be flipped on the calendar.** The next work on this lane is the per-pair triage: for each
(agent, field, winner_path) — is the winner the value the step needs? If yes, write the explicit
mapping (or `!`) so it stops being a search; if no, the pipeline has been living on the old coin
flip and needs the mapping even more. Then, and only then, D2's flip. Owner note: this is
precisely the measurement the ruling asked for instead of a guess, and it says the guess would
have been wrong.
