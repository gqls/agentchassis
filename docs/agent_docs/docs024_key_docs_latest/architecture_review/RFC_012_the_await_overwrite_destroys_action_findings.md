# RFC 012 — the await machinery destroys whatever an action computed, and every action pays the tax separately

**Filed** 2026-08-04 by the `bugfix_098_unpublish_primitive` thread, at the direction of
the council's **`architecture` seat** in an APPROVED round (correlation
`5a965452-a9a0-40a6-a990-410f14ac32b0`): *"the landmine registry already treats this as a
named recurring class, which is evidence enough that the coordinator's overwrite semantics
deserve their own RFC even though this specific fix should proceed unblocked."*

**Status: DECIDED 2026-08-06 — OWNER RULING: option B, as a DB-BACKED helper** (the
addendum-2 amendment is part of the ruling — a reserved collected_data namespace does
not survive the park and is NOT what was decided). See the ruling block at the end of
this file. No code change was proposed alongside this RFC; the point fix that
occasioned it (098 debt 5/5b) is shipped, approved, live and NOT dependent on it.

---

## 1. The mechanism, in one paragraph

An action that returns a result with `await_response: true` has that result stored under
its step name and its `output_field` (`storeActionResult`, coordinator.go). When the
awaited adapter reply lands, `applyResponseToState`'s default branch **replaces both keys
wholesale** with the reply. So any action that both *computes findings* and *dispatches a
request* loses the findings from the durable record the moment the reply arrives — they
survive only in pod logs. The status is `complete`, the reply looks like the step's
output, and nothing indicates anything was lost. This is the detected-then-discarded
class (`bugs_open/071`, `083`, `091`) built into the platform's own response plumbing.

## 2. Why a point fix was right THIS time, and what the running cost is

098's retraction action now writes its audit to a **sibling collected_data key**
(`retraction_audit`) — outside the overwrite's write set — and its refusals to
`agent_error_log`. The council approved that as correctly scoped: no shared mechanism
touched, established idioms reused.

But the `architecture` seat's point stands and is the reason for this RFC: this is at
least the **third** documented instance of an action independently discovering the
overwrite and hand-rolling the same escape hatch (the sibling-key pattern:
`image_result`, `final_html`, `__spawn_input_data__`, now `retraction_audit`). Every
future findings-plus-await action pays the same tax three times over:

1. **rediscover** the overwrite (usually by losing data in production first — 098 found
   it only because the session read the durable record after a green run);
2. **invent** its own sibling key, with its own collision risk against `output_field`
   names, checked (if at all) by a hand query;
3. **duplicate** the `agent_error_log` INSERT column list, because the one shared writer
   (`orchestration.LogAgentError`) lives in the package that imports `actions` and cannot
   be called from an action without an import cycle — there are now ~15 hand-copied
   INSERTs against that table in the actions package alone, and a future schema change
   must find every one.

## 3. The questions this RFC asks

> **(a) Should `applyResponseToState` MERGE the adapter reply into the step's existing
> record instead of REPLACING it** (e.g. reply under a `response` sub-key, as the
> call_agent branch already does; or pre-dispatch result preserved under a `dispatch`
> sub-key)? This is a change to what every workflow's response handling GUARANTEES —
> architecture scope by the 2026-07-29 ruling §1 — and any consumer that reads the step
> key expecting the bare reply shape would break. A census of readers is the first step;
> nobody has run it.

> **(b) Failing (a), should the sibling-key escape hatch be PROMOTED from folklore to a
> named helper** — e.g. `datahelpers.PreserveStepFindings(collected, stepName, findings)`
> writing to a reserved, documented namespace (say `<step>__findings`) that
> `applyResponseToState` is tested never to touch — so the pattern is one function call
> with one collision rule instead of N inventions?

> **(c) Should the actions package get a shared `agent_error_log` writer** (the import
> cycle runs coordinator → actions, so a writer in `datahelpers` or a new leaf package is
> importable by both sides), retiring the ~15 duplicated column lists? `bugs_open/185`'s
> fix candidate 2 already asks the sibling question for the eligibility predicates.

## 4. Options, costed

- **Option A (merge-not-replace in the coordinator):** closes the class for every future
  action; largest blast radius — every awaited step's readers see a changed shape unless
  the merge is additive (reply keys preserved at top level, prior keys kept where they
  don't collide). Needs a reader census before it is even costed honestly.
- **Option B (named helper + reserved namespace + guard test):** closes taxes 2 and 3 of
  §2, leaves tax 1 (you must still know to call it — though a landmine entry now exists,
  and the helper's existence in datahelpers is discoverable). Small, additive, testable.
- **Option C (do nothing beyond the landmine entry):** the class stays open; the landmine
  (filed 2026-08-04, footprint `applyResponseToState`/`await_response`) is now the only
  guard. Zero cost until the next action loses data it never knew it had.

**Recommendation of the filing thread:** B now (it is four small pieces: helper, reserved
namespace, a coordinator test pinning the namespace as untouched, and a shared error-log
writer), and A only if a reader census says the additive merge breaks nobody. C is what
we had before 098 debt 5, and it cost a production data loss to notice.

## 5. Evidence base

- `coordinator.go` `applyResponseToState` default branch — the replacement, read not
  inferred: `state.CollectedData[stepName] = normalisedData`, then the same for
  `output_field`.
- The measured loss: orchestration `fc00db29…` (the one real page retraction) — record
  held only `{paths, success, repo_url, …}`; candidates, refusals, the whole graph audit
  gone. `bugs_open/098` STATUS block and NOTES entry 2026-08-03.
- The landmine entry: `LANDMINES.md` "An action that RETURNS findings and AWAITS a
  response loses the findings" (2026-08-04) — the prospective guard until this RFC is
  decided.
- The council round asking for this RFC: `diagnosis_artifacts` correlation
  `5a965452-a9a0-40a6-a990-410f14ac32b0`, `council_report`, `architecture` seat notes.
- The import cycle forcing INSERT duplication: `coordinator.go:23` imports
  `platform/orchestration/actions`; `orchestration.LogAgentError`'s own comment says it
  exists "so there is ONE INSERT against agent_error_log" — a guarantee the actions
  package structurally cannot use.

---

# ADDENDUM 2026-08-04 — the SAME write is unguarded on the other side too, and it has now cost a fleet-wide outage. Plus the census §3(a) asked for, partially run.

**Added by the `bugfix_192_select_sections_wrapper` lane, at the direction of the council's
`bug_historian` seat** (submission `7afbf531-5ddd-484e-88c8-091994a0f51f`, verdict REVISE,
gating objection **high**): *"The true enabling mechanism is coordinator.go's
storeActionResult writing `result` wholesale under `output_field` with no check for
shape/collision against an existing key of the same name … leaves the generic mechanism
itself unguarded … Worth a human architecture look."* It also named this RFC's own case as
evidence that the class recurs.

Contributed **here rather than as RFC 013** because it is the same write, and a second
account of one mechanism is how a class becomes folklore.

## The other face of the same write

This RFC is about the record being replaced **when the awaited reply lands**
(`applyResponseToState`). `bugs_open/192` is the record being replaced **when the action
returns at all** (`storeActionResult`). Same two lines, same absence of a merge or a
collision check:

```go
state.CollectedData[state.CurrentStep] = result
if step.OutputField != "" { state.CollectedData[step.OutputField] = result }   // no check
```

A step whose `output_field` names a key an earlier step wrote does not annotate that key —
it **replaces** it. So an action that returns "the value, plus a note about what I did"
demotes the real value one level, on every run, and **the wrong result looks exactly like
the right one** because all the data is still present at `<key>.<key>.<field>`.

**Cost, measured:** `bugs_open/192` — every page build in the fleet failed from 08:20 to
09:01 on 2026-08-04. The producing step reported success; the error everyone read was
raised two steps downstream, in `loop_actions.go`, naming a missing key. Three lanes hit
it independently inside forty minutes.

## The census §3(a) asks for — run for the `output_field` side, and it found a SECOND live instance

Nobody had run it. The general precondition is *two steps in one workflow sharing an
`output_field`*:

```sql
WITH steps AS (
  SELECT ad.type, s.key AS step_name, s.value->>'action' AS action, s.value->>'output_field' AS of
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.value->>'output_field' IS NOT NULL)
SELECT type, of, count(*), string_agg(step_name||'('||action||')', ', ' ORDER BY step_name)
FROM steps GROUP BY 1,2 HAVING count(*) > 1 ORDER BY 3 DESC;
```

**24 (agent, output_field) pairs are shared by 2–5 steps.** The raw count is not the
finding, and my first reading of it was **wrong in a way worth recording**, because the
standing-check proposal in (d) below depends on getting the discriminator right.

> **CORRECTED 2026-08-04, same day, before the round-2 verdict.** I first reported to the
> council that "the large majority are mutually exclusive branches; only 2 are sequential".
> **The conclusion (2) was right and the reason was wrong.** I had read branch-vs-sequential
> off the step *names*. Checked structurally, several of those "branches" ARE sequentially
> reachable — the `propose → reframe → repropose → repair_plan` retry loops all write
> `proposal`. They are harmless for a completely different reason: **same action, so same
> shape** — a retry legitimately replaces the whole value with another of its own kind.

**Two naive detectors both give a clean, wrong answer.** Anyone implementing (d) needs this:

1. **Direct-edge check** (`a.next_step = b.step_name`) → **0 rows, including for
   `bugs_open/192` itself.** The real path is `plan_sections → … → check_has_ready_sections
   → load_current_section_content`; reachability is transitive, never adjacent.
2. **Transitive check over `next_step`/`error_step` only** → **still 0 rows.**
   `check_has_ready_sections` is a `conditional`: it routes through **`config.then_step`**.
   Routing lives in the config for **13 distinct keys** fleet-wide — `then_step` and
   `else_step` at **117 occurrences each**, plus `repair_step`, `ok_step`, `emit_step`,
   `alive_step`, `unreachable_step`, `lost_step`, `failed_step`, `gather_step`,
   `complete_step`, `probe_step`. A graph walk that reads only the two top-level keys is
   **blind to the majority of the fleet's control flow.**

**The discriminator that actually works: same `output_field`, transitively reachable, and
DIFFERENT `action`.** Run over the complete routing graph, the fleet splits exactly:

| agent | key | producer → refiner | same action? |
|---|---|---|---|
| `page-build-handler` | `section_plan` | `plan_sections` → `load_current_section_content` | **no** ← `bugs_open/192` |
| `site-adoption-agent` | `design_fingerprint` | `extract_design_fingerprint` → `enrich_fingerprint_with_css` | **no** ← found here |
| `experience-planner` | `proposal` | `execute_llm_prompt` → itself | yes |
| `feature-designer` | `proposal` | `execute_llm_prompt` → itself | yes |
| `fix-proposer` | `proposal` | `execute_llm_prompt` → itself | yes |
| `feature-implementer` | `gate` | `diagnose_build_gate` → itself | yes |
| `feature-implementer` | `stage` | `feature_stage_route` → itself | yes |

**2 hazards, 5 benign, 0 false negatives against the one known bug** — and the benign five
are benign *structurally* (a step cannot change the shape of a value it re-derives with the
same code), not by anybody's judgement. That is what makes it checkable rather than a
matter of taste, and it drops the "would fire on all 24 and become noise" objection I
raised against a naive warn: keyed this way it fires on **two**, both of which were real.

**The second one, read not inferred** (`enrich_fingerprint_with_css_action.go`, pre-fix):
its two success paths correctly returned `fp` — the fingerprint itself — but its two
early-outs returned `{"status": "no_fingerprint"}` and `{"status": "invalid_fingerprint"}`.
With `output_field: design_fingerprint`, the second **overwrites a real fingerprint with a
status stub**, destroying it and handing every downstream consumer a status object where a
fingerprint belongs. Fixed in the same commit as 192's revision (returns the caller's value
unchanged on both paths; reason goes to the log) — but note *how it was found*: not by a
failure, because nobody has reported one. **By the census.** That is the argument for the
census being a standing check rather than a one-off.

## What this adds to the questions in §3

The three questions stand. This addendum widens (a) and adds one:

> **(a′) The same merge-not-replace question applies to `storeActionResult`, not only to
> `applyResponseToState`** — and here it is *cheaper*, because the collision is knowable
> statically: at the moment of the write the coordinator can see that `output_field`
> already exists in `CollectedData` and that the incoming result is a map containing that
> same key as its only structural difference. That is the exact signature of the wrap.
> A **warn** costs nothing and is not a guarantee change; an **error** is.

> **(d) Should the shared-`output_field` census become a standing offline check?**
> **If so, the detector is specified above and the two naive versions must not be shipped:**
> key on *same `output_field` + transitively reachable over the FULL routing graph (13
> config keys, not just `next_step`/`error_step`) + different `action`*. Both simpler
> versions return 0 rows on a fleet containing a bug that had just taken every page build
> down, which is the worst possible failure for a check nobody re-reads.
> `config-key-audit` already walks every live step (WFA-003's `WalkSteps`) and already
> hosts `--unregistered-actions` (WFA-004), `--relay-gaps` (WFA-007) and the `SingleOwner`
> check (WFA-006, which runs **daily** as a CronJob under RFC 006). A
> `--shared-output-fields` mode reporting *sequential* refiners — branches excluded, which
> is the whole difficulty and is what makes 24 into 2 — would have found the fingerprint
> instance before it was written. This is the same shape RFC 006 already ruled on for
> `SingleOwner`, so there is precedent for the venue and the mechanism.

## What this lane did NOT do, and why it is the human's call

`bugs_open/192`'s fix is **shape-preservation at both instances plus an opt-in fail-loud
primitive** (`extract_fields.required`, WFA-009). It deliberately does **not** touch
`storeActionResult`. Two reasons, and the second is the honest one:

1. By the owner ruling of 2026-07-29 §1, changing what the coordinator's write
   **guarantees** is architecture scope, not council-gate scope — which is precisely why
   this is an addendum to an RFC rather than a fifth edit in a bug patch.
2. **A warn-on-collision would fire on all 24 pairs today**, including the 22 legitimate
   branch cases, so shipping it naively converts a real signal into log noise that
   everyone learns to ignore. Distinguishing "branch" from "sequential refiner" needs the
   reachability analysis question (d) describes. That is a design decision with a cost,
   and it belongs to whoever answers (a).

**Nothing in `bugs_open/192` is blocked on this**, exactly as §"Status" says of 098's fix.

---

# ADDENDUM 2 — 2026-08-04 (evening): a THIRD face, and it defeats this RFC's own worked example. The sibling-key escape hatch does NOT work for awaited actions.

**Added by the filing lane itself (098), from the first live batch retraction after the
fix shipped.** Run `e23b7257-e579-4766-9674-106eca5b66ba` — 10 pages retracted from
leopardessconsulting.co.uk, adapter success, curls all 404 — completed on a binary that
provably carries the fix (`strings /app/agent-chassis | grep -c retraction_audit` = 1 on
the executing pod, with a pre-existing control = 1). **The persisted record still has no
`retraction_audit` key.**

## The mechanism, read not inferred

`persistAwaitingStateWithRetry` (coordinator.go:2052) is what parks a step to await. It
**loads FRESH state from the DB** (:2058), copies onto that fresh copy ONLY the awaited
request entries (:2078-2080) and `Status`/`LastActivity` (:2083-2084), and saves the
fresh copy. Every in-memory `CollectedData` mutation made during the step's execution —
the action's sibling keys AND `storeActionResult`'s own step-name/output_field writes —
is **discarded at park time**. The reply later lands on another fresh load
(`handleCompleteResponse`), which is why an awaited step's record always ends up holding
exactly the reply and nothing the action computed.

So the class has three faces now, one per write path:
1. `applyResponseToState` — replaces the step keys when the reply lands (§1 above);
2. `storeActionResult` — same-name collisions between sequential steps (addendum 1, 192);
3. `persistAwaitingStateWithRetry` — **discards ALL CollectedData mutations at park
   time**, which makes face 1 mostly moot for awaited steps: the data it would have
   replaced was never persisted at all.

## What this corrects in the RFC above

- §2's premise that the sibling-key pattern is a working escape hatch is **true only for
  non-awaited actions** (where the ordinary step-completion persist saves the live map —
  `image_result`, `final_html` all live on that path). For findings-plus-await actions —
  the exact class this RFC is about — the sibling key is a NO-OP: my own tests proved the
  in-memory contract and could not see the park-persist discard, and the first live run
  did.
- The LANDMINES entry this RFC cites gave the sibling key as "the check"; it is corrected
  as of tonight (the durable half of the guidance — a direct DB row — stands and is the
  only half that works).
- **Option B must therefore be a DB-backed helper, not a reserved collected_data
  namespace**: a namespace `applyResponseToState` never touches still dies at :2058's
  fresh load. The cheap immediate fix for 098 (next session): persist the audit as an
  always-on `agent_error_log` row alongside the refusal rows, whose INSERT path is proven
  durable (that half of debt 5 works — the run's 0 refusal rows for 0 refusals is the
  mechanism behaving, not failing).

**Evidence:** run `e23b7257…` collected_data keys (no `retraction_audit`; `retraction`
holds only the wrapped reply); pod `agent-chassis-5455ddcdcc-gpr92` strings census;
coordinator.go:2052-2096 read in full. First-hand chain declared per the 2026-07-31
ruling in place of a 090 run: three artefacts (binary, persisted row, source), each
checked independently, agreeing.

---

# OWNER RULING 2026-08-06 — OPTION B, DB-BACKED

Recorded by the 098 lane on the owner's word ("RFC012 can be B: DB-backed helper"),
the same sitting that closed `bugs_closed/098`.

**The decision:** the sibling-key escape hatch is promoted from folklore to a named,
shared, **DB-backed** mechanism — addendum 2's amendment is binding, because it was
proven live that an in-memory namespace, however well reserved and tested, dies at
`persistAwaitingStateWithRetry`'s fresh load before the DB ever sees it. Findings that
must survive an await are persisted as **direct DB rows through a shared writer**, the
pattern 098 debt 5b proved end to end (`RETRACTION_AUDIT`/`RETRACTION_REFUSED` rows in
`agent_error_log`, written before dispatch, live-probed 2026-08-05).

**What B comprises** (per §4, amended): a shared `agent_error_log` writer in a leaf
package importable from `actions` (retiring the ~15 hand-copied INSERT column lists —
§3(c) is answered YES by inclusion), plus a named helper so a findings-plus-await
action makes one call instead of inventing its own escape hatch. The reserved
collected_data namespace and its coordinator guard test — §4-B's original in-memory
half — are **dropped from B**, not deferred: they are the refuted mechanism.

**Not decided here, left explicitly open:**
- **(a)/(a′)** merge-not-replace in `applyResponseToState`/`storeActionResult` — not
  taken now; available later only behind the reader census §3(a) names, which nobody
  has run. B does not preclude A.
- **(d)** the shared-`output_field` standing check — undecided. Anyone picking it up
  is bound by the addendum-1 specification (full 13-key routing graph, different-action
  discriminator); both naive detectors return 0 on the known bug.

**Implementation: unassigned.** No thread owns building the helper yet. The next
findings-plus-await action (or a thread adopting `bugs_open/158`/185-adjacent work)
should build it as its own coherent task through the council gate, and register it in
the concept register in the same commit. Until it exists, the LANDMINES entry (as
corrected 2026-08-04: durable = direct DB row) is the guard.

---

# OWNER RULING 2026-08-06 (second sitting) — (d) DECIDED, the census COMMISSIONED, B ASSIGNED

Recorded by the bugfix_080 session on the owner's word, three rulings in one message:

1. **(d) is DECIDED YES: the shared-`output_field` census becomes a STANDING check** —
   and the owner asks it be made ONLINE within the framework if possible (i.e. run by
   the platform against live `agent_definitions`, not only by a hand-run script). The
   addendum-1 specification remains binding: full 13-key routing graph,
   different-action discriminator; both naive detectors return 0 on the known bug, so
   a candidate implementation must prove itself against that case before it counts.
   Precedent for the online form: RFC_006's ruling that live config is guarded by a
   scheduled check, never a commit-time hook (at commit time the config is unapplied).

2. **The §3(a) reader census is COMMISSIONED.** Scope per §3(a): every consumer that
   reads an awaited step's key expecting the bare reply shape — config-side
   (`input_mapping`s and template references over the fleet's `agent_definitions`) and
   Go-side readers of `collected_data[<step>]`. Deliverable: a census artefact in this
   RFC's directory naming each reader and whether a merge-not-replace under a
   `response` sub-key would break it — the artefact (a)/(a′) is gated behind.

3. **Implementation of B is ASSIGNED — "let's do it now"** (the bugfix_080 session
   takes it): the leaf-package shared `agent_error_log` writer + the named
   findings-plus-await helper, through the council gate, concept-register entry in the
   same commit, per the first sitting's prescription.
