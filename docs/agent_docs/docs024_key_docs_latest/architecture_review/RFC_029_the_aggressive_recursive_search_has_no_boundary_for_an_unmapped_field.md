# RFC 029 — the aggressive recursive search (`findFieldRecursive`) has no stated boundary for a field the caller never mapped, and `bugs_open/248` is its second production incident in one bug file

## STATUS: **RULED 2026-08-15 (owner-delegated determination) — see §9. Implementation OPEN, phased, not yet started.**

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
