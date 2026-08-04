# RFC 012 — the await machinery destroys whatever an action computed, and every action pays the tax separately

**Filed** 2026-08-04 by the `bugfix_098_unpublish_primitive` thread, at the direction of
the council's **`architecture` seat** in an APPROVED round (correlation
`5a965452-a9a0-40a6-a990-410f14ac32b0`): *"the landmine registry already treats this as a
named recurring class, which is evidence enough that the coordinator's overwrite semantics
deserve their own RFC even though this specific fix should proceed unblocked."*

**Status: OPEN — needs a human.** No code change is proposed alongside this RFC. The
point fix that occasioned it (098 debt 5) is shipped, approved, and NOT dependent on any
answer here.

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
