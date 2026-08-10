# RFC_019 — One ladder for "which agent is running", and where its bottom rungs live

**Status: DECIDED 2026-08-09 — the shared ladder ships (owner ruling, §11)** (filed 2026-08-08 by the `rfc012_await_findings` lane, §1a)
**Code: already shipped** — `1bc08d1ce`, alongside concept-register entry `RSH-009`.
**Council gate: `Council-Submitted: 6186ab10-a006-4c34-b9ea-ecedfde8ea2d`** — round 1 **REJECTED**, hard veto from `guardian`; the `architecture` seat returned **`ARCHITECTURE_SIGNAL: needs_rfc`**. Read §10 before anything else: it answers §8's question, and it answers it against me.

> **This is a RETROSPECTIVE RFC, like `RFC_002`, and that is the design of the track, not a
> corner cut.** OWNER RULING 2026-07-29 §2 retired the ordering-exemption condition that
> assumed a thread could hold a change out of the fleet: on this tree HEAD is shared,
> `make build-*` builds from committed HEAD, and any other session's roll ships your commit.
> So review here is after the fact. What is required — and was done — is registration in the
> same commit (condition 2) and submission to the gate before or alongside it.

> **WHY IT IS HERE AT ALL, given I argue below that it is close to the line.** The handoff
> that commissioned this work flagged the OWNER RULING 2026-07-29 §1 trigger and said: *"Read
> that ruling and consider an RFC rather than the council gate."* I have, and my measurement
> moves the answer — but the measurement is exactly what a reader six weeks from now will not
> have. `PROCESS_architecture_review.md`: *"When in doubt, the cost of an RFC is one document
> — write it."* The decision this paper wants is not "may I ship it" (it is shipped) but
> **"is this the right forum for the next seam of this shape"**, because `RSH-008`'s round-1
> `architecture` seat created a precedent last week and this change sits just outside it.

---

## 1. Problem + evidence

One question — *"which agent's workflow is this context executing?"* — was answered by two
ladders, in two packages, and they disagreed in production.

`types/context.go:62` documents the field that answers it:

> `RunAgentType` is the RESOLVED real agent type whose workflow is executing (from
> `config.agent_type` / the loaded agent definition), **as opposed to the dispatch-path sender
> which is often 'generic'**. Set by the processor before executing a workflow; consumed by the
> coordinator's `determineOwnerAgentType` so `owner_agent_type` records the real agent, not
> 'generic' (`bugs_open/060`).

It is set at `messaging/processor.go:1828`. Until `1bc08d1ce`, `coordinator.determineOwnerAgentType`
(`coordinator.go:3466`) was its **only** reader:

```
RunAgentType → Sender.AgentType → os.Getenv("AGENT_TYPE") → log Error → "generic"
```

`actions.runningStepProvenance` (`actions/log_action_error.go:118`) answers the same question
for `agent_error_log.agent_type`/`step_name` — the provenance half of the `RSH-008` door — and
implemented **only the second rung**:

```go
agentType, stepName = params.AgentType, params.CurrentStep
if params.ExecutionContext.Sender.AgentType != "" {
    agentType = params.ExecutionContext.Sender.AgentType   // ← the dispatch sender wins
}
```

So a row filed under "the running step" recorded the dispatch-path sender — literally `generic`
on the generic dispatch path — while the orchestration row for the *same run* recorded the real
agent. The column it degrades is the one `idx_error_log_agent (agent_type, occurred_at DESC)`
makes the entry point of every investigation.

### The evidence, re-measured — and the sizing I was handed was wrong

The commissioning handoff cited `generic` = **559 rows over 25 distinct `step_name`s**, and 25
`REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows all carrying `generic`, as the live damage. Measured at
the rows on 2026-08-08, that framing does not survive:

| measurement | result |
|---|---|
| `generic` rows before / after 2026-07-27 | **499 / 56** |
| the boundary | `baf887a8e` (2026-07-26) — the commit that added `RunAgentType` **itself** |
| dominant producer `call_agent`/`call_dispatch` | 394 rows, `max(occurred_at)` = **2026-07-25** |
| all 25 `REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows | **2026-07-23**, one day; none since |
| residue reachable by THIS change (13 days) | **~36** — `diagnose_council_decide` 31, `retract_page_deployment` 4, `emit_tool_cross_link_items` 1 |
| the other 20 residual rows | `orchestrate` / `process_message` / coordinator paths, untouched by this |

**The coordinator's own ladder had already removed ~89% of the damage.** Anyone citing 559 is
citing a table dominated by pre-fix history. That correction is the most useful thing in this
paper, and it is why §3's do-less option is a serious one rather than a formality.

## 2. Design

Hoist the two **context** rungs onto the shared type; leave the two **process** rungs where they
are.

```go
func (ec *ExecutionContext) ResolvedAgentType() string {
    if ec == nil { return "" }
    if ec.RunAgentType != "" { return ec.RunAgentType }
    return ec.Sender.AgentType
}
```

- `coordinator.determineOwnerAgentType` delegates, then keeps `os.Getenv("AGENT_TYPE")` and the
  `"generic"` filler. **Its answer is identical for every input** — this consumer's behaviour
  does not change. The drift closes by making the *other* consumer agree with this one.
- `actions.runningStepProvenance` calls it, and keeps `params.AgentType` as its floor.

Reachability was checked with the compiler, not asserted: `orchestration` imports `actions`
(the cycle is real, so delegating coordinator-ward is impossible), and `types` is imported by
both (`go list -f '{{range .Imports}}…'` → 1 and 1).

### The load-bearing decision: the bottom rungs do NOT move

This is the part worth an architecture reader's attention, because it is where the change could
have re-created the drift while appearing to fix it.

1. **`os.Getenv("AGENT_TYPE")` is a property of the PROCESS, not of the message.**
   `ExecutionContext` is deserialised from Kafka headers and passed between packages as data. A
   method whose answer changes with the pod it happens to be called on is invisible at the call
   site and untestable without mutating the environment.
2. **The two consumers legitimately DISAGREE about what follows the context, and the
   disagreement is correct rather than drift.** The coordinator falls back to the pod and then
   to a filler because `owner_agent_type` is NOT NULL and a row must be written. The actions
   door falls back to `params.AgentType` — which *is* `state.OwnerAgentType` (`coordinator.go:1691`),
   this ladder's own durable output, so strictly more specific than the pod — and must **not**
   reach for a filler at all, because an unresolvable provenance there lands as `unattributed`
   by `RSH-008`'s design (OWNER RULING 2026-08-02, RFC_010 §2). Folding both tails into one
   method hands one consumer the other's answer.
3. **`ResolvedAgentType(fallback string)` was the obvious alternative and is worse.** It looks
   like one ladder while two callers pass different fallbacks — the same drift wearing a
   shared-helper costume, and harder to see because it now has a shared name.

`TestResolvedAgentType_IgnoresTheProcessEnvironment` pins the exclusion with `t.Setenv`, so a
future author cannot "complete" the ladder without a red test. That test is the enforcement; the
paragraphs above are only its explanation.

## 3. Alternatives considered

| option | why it was ruled out, with evidence |
|---|---|
| **Do nothing** — the residue is ~36 rows in 13 days | Genuinely arguable, and I nearly took it. Ruled out because the cost of the defect is not the row count: the two answers to one question sit in packages that cannot import each other, so nothing *structural* stops the gap widening, and the next reader of `RunAgentType`'s doc comment is told it has one consumer. `bugs_closed/098` debt-5b is the estate's precedent for converging duplicated answers rather than counting their current damage. |
| **Fix `runningStepProvenance` in place** — add `RunAgentType` to its own ladder | Two ladders, both correct today, drifting again on the next rung. This is the shape the whole exercise exists to retire; it also leaves `determineOwnerAgentType` with no test, which is how the second consumer became invisible in the first place. |
| **Move the WHOLE ladder, env rung included** | Rejected on §2's two grounds. It would also silently give the actions door the `"generic"` filler, which is the exact value `RSH-008` chose `unattributed` *to avoid colliding with* (`generic` has real traffic, so a detector keyed on it could never separate a mistake from noise). |
| **`ResolvedAgentType(fallback string)`** | §2.3. |
| **Backfill `RunAgentType` in `ensureFullExecutionContext` as well** | Would probably make the fix land on resumed steps too (§7), and is *not* taken here: it adds a rung sourced from durable state, which is a different claim needing its own evidence. Deliberately left as the named next round rather than bundled — this lane's round 2 was REJECTED for exactly that kind of bundling. |

## 4. Blast radius, named

- **Behaviour changes in:** `agent-chassis` (and every service built from it) — one column's
  value on rows written through the six `RSH-008` doors, on the declared-inheritance path only.
  Write-path only: the 499 historical `generic` rows are untouched.
- **Merely relinks:** everything else importing `platform/orchestration/types` — the method is
  additive and unread elsewhere.
- **Consumers of the changed column, named rather than merely measured** (OWNER RULING
  2026-07-29 §3). **No automated consumer branches on the VALUE of `agent_error_log.agent_type`:**
  - `diagnose_load_runtime_action.go:265` — renders it into the diagnosis bundle (display);
  - `reconcile_superseded_reviews_action.go:223` — filters `site_id` + `error_code` + `context->>'page_name'`;
  - `page_build_failure_guard.go:131` — filters `error_code` + `context->>'page_id'`;
  - live agent config: **0** steps filter on it. The only three whose SQL names the table
    (`council-gate`, `feature-designer`, `fix-proposer`) are `load_schema_hint` steps reading
    `information_schema.columns`;
  - repo SQL: `sql_for_agents/214_build_dispatch_watchdog.sql:109` selects it for display.

  **So the consumers are HUMANS and the diagnosis bundle**, through `idx_error_log_agent` — which
  is who the change is for, and who a volume claim would have misled. `orchestration_states.owner_agent_type`
  is unchanged: the coordinator's answer is identical for every input.
- **No schema, no migration, no wire shape, no config vocabulary, no new reserved key.**

## 5. Staged rollout

One stage; there is nothing to sequence. The Go change is inert until the next chassis roll,
which any session's build will carry. **Induced-fault testing, not happy-path greps** — six
mutations, and the failure *identities* are the finding:

| mutation | types | actions | orchestration |
|---|---|---|---|
| baseline | 0 | 0 | 0 |
| drop rung 1 (`RunAgentType` ignored) | 3 | 3 | 2 |
| drop rung 2 (`Sender` ignored) | 3 | **4 pre-existing** | 2 |
| drop the nil guard | 1 | 0 | 0 |
| **revert the actions door to its old one-rung read** | 0 | **exactly the 3 new** | 0 |
| **revert the coordinator's delegation** | 0 | 0 | **exactly the new coordinator test** |

Row 3 is the no-op half — the four *pre-existing* tests are what prove the old behaviour is
preserved rather than assumed. Row 5 is the defect reproduced, and simultaneously the proof the
old suite could not see it: every pre-existing test in the actions package passes
`sqlmock.AnyArg()` for `agent_type`, so **a broken ladder was green on the whole package**. Row 6
matters because `determineOwnerAgentType` had **no test at all** before this change — a claim of
the form "one ladder, two consumers" is worth only as much as its least-pinned consumer.

Row 4 is stated honestly: the nil guard is exercised by the `types` test only, because both
consumers nil-check before calling. It is defence for future callers, not a live path.

## 6. Rollback

Image-only. No schema, no config, no migration; the previous binary tolerates every row this one
writes, and reverting restores the previous column value for subsequent rows. Nothing to undo in
the database.

## 7. Acceptance evidence — and a declared way this could be a partial no-op

**Do not mark this IMPLEMENTED on the binary check alone.**

1. **In the binary**, after the next roll, one Running pod per **distinct image tag** (the fleet
   is routinely mid-roll, and a label-picked pod answers for either): POS
   `provenance from the dispatch sender rather than the resolved run agent` = 1 and
   `whose workflow is this context executing` = 1; NEG a phrase in no version = 0.
   ⚠ **never anchor the needle** — `grep -c "^ResolvedAgentType$"` reads 0 on a binary that
   carries it, because the Go linker packs constants into contiguous blobs (landmine filed).
2. **In production behaviour**, which is the claim that matters:
   ```sql
   SELECT count(*) FROM agent_error_log
   WHERE agent_type = 'generic'
     AND action IN ('diagnose_council_decide','retract_page_deployment','emit_tool_cross_link_items')
     AND occurred_at > '<roll time>';           -- baseline: 36 in the 13 days before
   SELECT count(*), count(DISTINCT agent_type) FROM agent_error_log
   WHERE occurred_at > '<roll time>';           -- the positive control
   ```
   **A zero with no traffic is a dead path, not a fix.**

### The declared limitation

This may be a **partial no-op on RESUMED steps**, and it is stated here rather than left to be
discovered. `RunAgentType` reaches the actions door on the first-step / same-message path
(processor sets it → `ToHeaders` → `FromHeaders` → `buildActionParams` hands the same `execCtx`
to the action; the round trip is pinned by a test). On a step resumed after an await, `execCtx`
is rebuilt from the **response** message's headers, and `ensureFullExecutionContext`
(`coordinator.go:1589`) backfills `Sender` from `state.OwnerAgentType` **only when `Sender` is
empty** — it does not backfill `RunAgentType` at all.

I could not settle this by measurement: `orchestration_states` retains ~24h and every
actions-door `generic` row is older than that, so the join returns nothing to evaluate. The
decisive check is therefore post-roll, above. If the residue does not fall, the contained remedy
is one `if` in `ensureFullExecutionContext`, and it belongs in its own round.

## 8. The question this paper actually wants ruled on

Not "may this ship" — it has shipped, and under the 2026-07-29 §2 ruling it could not have been
held. The question is **which forum the next seam of this shape belongs in**, because the
precedent is one week old and this change sits just outside it.

`RSH-008`'s round-1 `architecture` seat returned `ARCHITECTURE_SIGNAL: point_fix`, reasoning:
*"stays inside `platform/orchestration/actions`, adds no schema column, no new reserved key on a
shared action config, no wire-shape change."* Three of those four hold here. **The first does
not** — this adds an exported method to a type that `orchestration`, `actions` and `messaging`
all import.

Applying OWNER RULING 2026-07-29 §1 (*"needs an RFC only when it changes what the shared
mechanism GUARANTEES"*) as I read it, my own answer is **no, this was council-gate scope**:

- it adds an opt-in capability reachable by nothing until a caller names it — the
  additive-and-inert case the ruling explicitly narrowed *out* of architecture scope;
- it grants no new authority; nothing can now refute, retract or re-type anything it could not
  before (the RFC_002 trigger);
- the guarantee readers actually rely on is *"`agent_type` names the agent this row belongs to"*,
  and the change makes the value **more** faithful to it, not different in kind;
- and the consumer census (§4) is the measurement the 2026-08-02 narrowing turns on: **no
  automated consumer**, so there is no set of owners whose agreement is being presumed. The
  humans who *are* the consumers are told by this paper and by `RSH-009`.

**The counter-argument, kept visible because a seat may well prefer it:** "no automated consumer
today" is a fact with a timestamp, and an exported method on a wire type is a standing invitation
to a third consumer with a third tail. If the estate's answer is that *any* new method on
`ExecutionContext` is architecture-scope regardless of what it does, that is a clean and
defensible line — and it is a line the owner should draw, not me, because I am the interested
party.

**What would change my own answer:** a reader who can name an automated consumer of
`agent_error_log.agent_type`'s *value* that my census missed. That census is a grep and a live
query, both in `RSH-009`; falsify it and the paper's §4 collapses.

## 10. VERDICT, 2026-08-09 — the gate answered §8, and my own answer was wrong

Round 1, corr `6186ab10-a006-4c34-b9ea-ecedfde8ea2d`: **REJECTED**, `decided_by: hard veto from
guardian`. Ten seats approved, one objected, one vetoed.

### §8 is CORRECTED, visibly rather than quietly

§8 argued, at length, that this was council-gate scope. **The `architecture` seat disagrees and
returned `ARCHITECTURE_SIGNAL: needs_rfc`, severity MEDIUM**, and I was wrong. Its ground is not
the consumer census I offered to be falsified on — it is the trigger test's *other* clause:

> "Trigger test: fires on 'exported symbol multiple packages depend on.' The author names this
> themselves and routes it to `RFC_019` in the same commit — that is exactly the discipline this
> seat wants to see, and **it does not relocate the scope, it just means the right conversation is
> already scheduled.**"

So §8's framing — "what would change my answer is an automated consumer I missed" — **named the
wrong disconfirmer**. The trigger fires on the shape of the symbol, independent of blast radius,
and I had already conceded the shape in the same paragraph while arguing past it. Recorded here
rather than edited away, because a paper that argued itself into the wrong forum and was corrected
by the forum is exactly the kind of thing this track exists to keep.

> **One ambiguity in `PROCESS_architecture_review.md` this exposed, offered as a question and not
> a defence.** The trigger clause reads *"it **changes or removes** an exported symbol other
> packages depend on (signature changes count)"*. This change **adds** one; it changes and removes
> nothing. Two seats read the clause as firing on an addition anyway, and given they are the seats
> that hold the rule, **their reading is the operative one and this paper accepts it.** But the
> written clause and the applied clause differ, and the next author will read the written one. If
> the owner agrees with the seats, the words should say *"adds, changes or removes"*. That is a
> one-line fix to `PROCESS` that I have deliberately NOT made — amending the trigger test I was
> just caught by, on my own authority, is not mine to do.

### The seats disagree with each other about the fix, and that is the decision

`guardian`'s veto (two HIGH objections on edits 1 and 2, one MEDIUM on edit 3) names a contained
alternative, faithfully quoted:

> "duplicate the 2-line `RunAgentType`/`Sender.AgentType` read locally inside
> `actions/log_action_error.go` (both fields are already reachable via `params.ExecutionContext`,
> which `actions` already imports), and leave `coordinator.go` and `types.ExecutionContext`
> untouched entirely."

**That alternative is precisely the thing this RFC exists to retire** — it closes the row-level
symptom by writing the ladder a second time, in the package that cannot see the first. And the
`architecture` seat says so, in the same round, unprompted:

> "A contained non-hoist fix (just correcting the actions-door rung locally) **would have
> re-created the drift risk the author is trying to close**, so the shared-method direction is
> defensible even though the volume argument is weak. … agent-identity resolution was already
> duplicated and drifting across two packages, and **a THIRD site would have been next**. …
> **I'd rather see this land than not.**"

`reuse_agent` and `constitution` reach the same place independently — *"not a fresh SQL pair
reinvented, but an existing duplication found and collapsed to one implementation … one way is
being restored"*; *"REUSE BEFORE RECREATE done right, collapsing duplication rather than adding
it"*.

So the round produced a genuine, load-bearing disagreement between seats: **the guardian's safest
contained fix is the reuse seat's founding violation.** Per CLAUDE.md's 2026-07-28 ruling — *"A
veto on SCOPE is not answered by resubmitting with better measurements … especially when seats
disagree with each other"* — **this is not being resubmitted, and it is not being reverted.** The
guardian itself routes the decision here:

> "If council disagrees and wants the shared-ladder design, **that decision belongs to `RFC_019`,
> not to this gate**."

and flags as `missing` whether this paper had resolved first, noting *"if it hasn't, this gate
should not pre-empt it"*. It had not. **That is the ruling this RFC now asks for**, and it is a
narrower question than §8's: not "was the forum right" — the seats have settled that — but
**"which of the two designs ships"**, given the safest-contained rule and the
reuse-before-recreate rule point in opposite directions on the same edit.

⚠ **`Council-Reviewed:` is NOT written and may not be**, on this or any future commit citing this
correlation. `1bc08d1ce` carries `Council-Submitted:`, which asserts nothing; a rejected
correlation simply never gets credited by `098`, which is the correct outcome.

### The objections that were NOT about scope, each with what was done

- **`prior_art_librarian` (MEDIUM, edit 3), the only substantive technical challenge.** It could
  not read `doc_notes` from the seat, and asked whether the `LogActionEntry` merge landmine means
  *"the same merge could silently overwrite the newly-resolved `agentType` downstream of
  `runningStepProvenance` without any test catching it"*. **Answered from the code, and the answer
  is no.** `resolveProvenance` is the only consumer of `runningStepProvenance`'s output, and it
  assigns `entry.AgentType = runningAgentType` **only** when `inheritProvenance` is true *and*
  `entry.AgentType == ""`; the strict branch returns without touching it, and the unattributed
  branch overwrites with the sentinel by design. There is no path that reads the resolved value
  and then discards it. More importantly the concern is structurally answered by *how* the tests
  assert: they pin argument **5 of `agenterrors.Write`'s INSERT via `sqlmock`**, i.e. the value
  that reaches the database — not the helper's return — so an overwrite anywhere downstream fails
  the test. That is why the pins were written at the SQL boundary rather than as unit assertions
  on the helper, and the seat was right to ask.
- **`debug_historian` (LOW, edit 5):** the plan never named a pod-grep of the shipped binary. Fair
  — it was in `RSH-009`'s `verify-later` but not in the submission, and the submission is what a
  seat reads. §7 above carries it, needle and anti-anchor warning included.
- **`editquality` (LOW, edit 4):** the comment-only correction "is not an edit" and should not
  count toward the fix. Accepted; it was labelled as documentation hygiene precisely so it could be
  discounted, and discounting it is correct.
- **`bug_historian` (no objection, flagged for the record):** the declared resumed-step gap is
  structurally the shape of **`bugs_open/093`** (*one guarded call site, sibling path unchecked*)
  and closed case 077 (*detector wider than handler*). Recorded in `RSH-009`; the seat's own note
  is that disclosing and deferring it *"is the correct response to this pattern, not a violation
  of it"*, and that it is non-regressive — resumed steps behave exactly as before.

### What a reader should take from this round

The measurement work in §1 and §4 was not what decided it and would not have changed it. **Five
seats independently praised the mutation discipline and the declared limitation, and the paper was
rejected anyway, on the shape of one exported symbol.** If there is a transferable lesson it is
that on this estate *evidence answers "is the fix right"; it does not answer "may this seam
exist"* — and only the second question was ever open here.

## 12. POST-ROLL, 2026-08-10 — the code is LIVE; the acceptance test was BROKEN TWICE and proves nothing

Chassis `v1.0.1277` (both replicas, started 2026-08-09 21:35Z). **Both ladder halves are present in
the shipped binary.** But §7's acceptance evidence — this paper's own recipe — failed in two
independent ways, and neither would have announced itself.

### (1) The prescribed pod-grep needles are Go COMMENTS, so the check returns 0 on a correct binary

§7 says to grep the binary for `provenance from the dispatch sender rather than the resolved run
agent` and `whose workflow is this context executing`. **Neither string exists in the source at
all.** They are (approximate, and misquoted) renderings of the *doc comment* at
`types/context.go:734` — `// ResolvedAgentType answers "which agent's workflow is this context
executing?"`. Comments do not ship in a binary. **Run as written, the check reports 0/0 on a binary
that carries the fix perfectly**, and the obvious reading of that is "the roll did not include it".
This is worse than the anchoring trap §7 already warns about: an anchored needle is a real string
matched wrongly; this one was never a string.

**Neither change contributes a usable needle of its own** — `ResolvedAgentType` is pure control flow
and the §7 backfill is one `if`. So the binary must be dated by a *neighbouring* literal from a
commit that is a descendant. Done, and it is a sound argument rather than a lucky grep:

```
POS "fallback_url_field is configured but resolved to no URL"   → 1  (both replicas)
POS "check 'url_field', 'fallback_url_field'"                   → 1  (both replicas)
NEG "zzz_this_phrase_is_in_no_version_of_the_binary"            → 0  (both replicas)
```
Both literals arrived in `f7111f4d8` (08-09 15:29). `git merge-base --is-ancestor` confirms
`1bc08d1ce` and `58aefe282` (15:24) are both its ancestors, so a build containing them contains the
ladder and the backfill. **One image tag fleet-wide, checked on every Running replica.**

### (2) The behavioural test could not have come out otherwise — the producers went dormant BEFORE the roll

| query | result |
|---|---|
| **the claim** — `generic` rows, the 3 residual actions, post-roll | **0** |
| **§7's control** — all rows post-roll / distinct `agent_type` | 288 rows / 20 types ✅ *passes* |
| **the control §7 did NOT specify** — rows from those 3 actions post-roll, **any** `agent_type` | **0** |
| baseline, same 3 actions, `generic`, 13 days pre-roll | 33 |

**The third row is the finding.** Those three actions produced *no rows of any kind* after the roll,
so there was nothing for the fix to relabel, and the headline 0 is guaranteed by absent demand
rather than by working code. §7's specified control — fleet-wide traffic — **passes cleanly while
being blind to this**, because fleet traffic is not demand on the path under test.

Bucketed by day, the baseline is worse than merely thin: the three producers' `generic` rows stop on
**2026-08-05**, four days before the roll (`08-01: 1, 08-02: 1, 08-04: 2, 08-05: 2`, then nothing).
`diagnose_council_decide`, which contributes 42 of the 47 `generic` rows these actions have ever
written, last filed on **2026-08-02** — a week before the fix shipped. The table retains from
07-11, so this is dormancy, not retention. **The acceptance test was already incapable of returning
a non-zero the moment those producers went quiet, which was before the code it tests existed.**

**This is the SAME defect this paper's own §1 correction identified, one level up.** §1 caught a
`count(*)` that priced a fixed defect like a raging one. §7 then specified a follow-up measurement
with exactly the same property — dated, marked, controlled, and unfalsifiable. Being burned by it
once did not prevent designing it in again.

**A genuine positive control does exist, and it is worth keeping:** three `generic` rows *were*
written post-roll, from `process_message` (2) and `orchestrate` (1). Those are precisely the
coordinator-path producers §1's table scoped **out** of this change, so their survival is expected —
and it proves the `generic` value can still be written at all, which is what stops the headline 0
being read as a dead write path.

### What would actually settle it

Waiting is not a plan: the producers are dormant and may stay so indefinitely. The decisive test is
to **induce** one — deliberately fail a step that has been resumed after an await, through one of the
six RSH-008 doors, and read the `agent_type` on the row it writes. That is the only check whose
result depends on the code rather than on whether anything happened to break this week. Until it is
run, the honest status is: **present in the binary, behaviourally unproven.**

## 11. OWNER RULING, 2026-08-09 — "shared code wins this one"

The owner read §10's question — which of the two designs ships, given the safest-contained
rule and the reuse-before-recreate rule point in opposite directions on the same edit — and
ruled for the shared ladder. Verbatim: *"I think shared code wins this one."*

What follows from it, all executed the same day:

1. **`1bc08d1ce` stands as shipped.** No revert, no resubmission (unchanged from §10 — the
   ruling confirms the design rather than reopening the round). The guardian's contained
   alternative — duplicate the two-line read locally — is DECLINED: it is the second ladder
   this paper exists to retire, as the `architecture`, `reuse_agent` and `constitution` seats
   said in the round itself.
2. **The declared resumed-step gap (§7) is commissioned as its own round, immediately** —
   the owner's same message directed the residual problems fixed ("please go ahead and fix
   all those other problems"). The one-`if` backfill in `ensureFullExecutionContext` goes
   through the council gate as the separate round §3's last row promised, with its own
   submission and `Council-Submitted:` trailer.
3. **The `PROCESS_architecture_review.md` trigger wording is amended** to *"adds, changes
   or removes"* — the one-line fix §10 raised and deliberately withheld as not this paper's
   author's to make. Read as sanctioned by the owner's blanket direction above; if the owner
   meant the narrower reading, it is one line to revert.
4. **The post-roll measurement (§7) remains the acceptance evidence** and still cannot be
   taken before the next chassis roll. Baseline unchanged: ~36 reachable rows in the 13 days
   before. With the §7 backfill also shipping, the residue is now expected to fall on BOTH
   the first-step and resumed paths; if it does not, the remaining suspects are the
   `orchestrate`/`process_message` coordinator-path rows §1's table already scoped out.

## 9. Sources

- Code: `1bc08d1ce`. `platform/orchestration/types/context.go` (+`resolved_agent_type_test.go`);
  `platform/orchestration/coordinator.go` (+`owner_agent_type_ladder_test.go`);
  `platform/orchestration/actions/log_action_error.go` (+`log_action_error_test.go`).
- Register: `docs026_concept_register/register/resilience-self-heal.md` → **RSH-009**
  (and **RSH-008**, the door this feeds).
- Lane: `docs024_key_docs_latest/rfc012_await_findings/` — the standing five; the mutation
  harness and the pod-grep recipe are in its `RUNBOOK`.
- Submission: `rfc012_await_findings/SUBMISSION_2026-08-08_resolved_agent_type_ladder.json`,
  corr `6186ab10-a006-4c34-b9ea-ecedfde8ea2d`.
- **Not this**: `bugs_open/060` is a different bug (no durable agent-run record; `usage_count`
  dead). It is where `RunAgentType` came from and must not be folded in.
