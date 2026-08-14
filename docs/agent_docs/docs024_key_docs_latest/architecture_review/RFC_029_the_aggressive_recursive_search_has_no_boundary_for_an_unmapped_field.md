# RFC 029 — the aggressive recursive search (`findFieldRecursive`) has no stated boundary for a field the caller never mapped, and `bugs_open/248` is its second production incident in one bug file

## STATUS: **OPEN — routed here from a council REVISE, at the reviewer's own direction, not raised speculatively.**

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
