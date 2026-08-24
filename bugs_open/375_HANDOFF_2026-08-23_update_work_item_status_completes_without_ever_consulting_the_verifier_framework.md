# 375 — `update_work_item_status` stamps `complete` without ever consulting the verifier framework, so one of the three completion writers has no false-completion guard at all

**Filed 2026-08-23** by the `bugfix_367_router_remit` lane, found while tracing why a router's
wrong close was silent. **Status: OPEN — CLAIMED 2026-08-24 by `docs/agent_docs/docs024_key_docs_latest/bugfix_375_completion_verifier_gap/` (read `HANDOFF_2026-08-24_start_here.md` there first).** Deliberately NOT fixed inside `bugs_open/367`:
this changes what a **shared completion path** guarantees, which is architecture-scope under the
owner ruling of 2026-07-29, and `bugs_closed/124` drew a REJECTED verdict for exactly that shape
of change arriving inside a bug patch.

## 1. The defect in one paragraph

The platform has a **verifier framework** whose entire purpose is stated in
`platform/orchestration/actions/discovery_checks/verifier_coverage_test.go`:

> *"`CompleteWorkItemAction` consults a per-item_type verifier before stamping 'complete'. …
> That is the same class as `bugs_open/017` (a saga reporting success without touching the
> defect), one level up: 017 stops a saga that says it FAILED from being stamped complete; a
> **verifier is what stops one that says it SUCCEEDED but did nothing**."*

There are **three** writers of `complete` on `site_work_items`. `CompleteWorkItemAction` consults
the verifier. `UpdateWorkItemStatusAction` — `platform/orchestration/actions/v3_site_actions.go:6010`,
registered as action `update_work_item_status` (`registry.go:939`) — **does not**. Its `complete`
arm (`v3_site_actions.go:6290-6300`) carries only the *terminal-decision* guard:

```sql
UPDATE site_work_items
   SET status = $2, completed_at = NOW(), updated_at = NOW(),
       attempt_count = attempt_count + 1,
       result = COALESCE(result,'{}'::jsonb) || $3::jsonb,
       error  = COALESCE(NULLIF($4,''), error)
 WHERE id = $1
   AND status NOT IN (workItemCompletionGuardStatuses)
```

`GetVerifier` is never called on this path. The code's own comment says as much about what it *did*
add, and the omission is visible in it:

> *"The `complete` arm carries the terminal-decision guard `CompleteWorkItemAction` has (WII-003,
> load_work_item_actions.go) — this action is a **third writer of `complete`** and had no guard at
> all … Same defect, same remedy, one writer over."*

The remedy that was carried over was the terminal-decision guard (don't overwrite a row that
already failed or was given up). The **verifier** was not.

## 2. Why this matters — it is the reason `bugs_open/367` was silent

`bugs_open/367` is one instance. `required-fields-missing-handler`'s `close_stale` step uses
`update_work_item_status` with `status: complete`. A true finding — schema-required fields
genuinely empty on a component that genuinely exists — was stamped `complete`, `attempt_count 1`,
`error` NULL, and disappeared into the "actioned" bucket. A verifier for
`required_fields_missing` would have re-run the finding's own predicate and refused the completion.

367 was fixed at the router (migration `574`, live 2026-08-23) so that particular wrong close can
no longer be constructed. **That fixes one disposer's reasoning; it does not restore the guard.**
Any agent definition can call `update_work_item_status` with `status: complete` from DB config,
with no code change and no review — which is precisely the property `bugs_open/213` relies on when
it refuses to enumerate producers in code.

## 3. Evidence `[MEASURED 2026-08-23]`

- `grep -rn "RegisterVerifier\|GetVerifier" platform/orchestration/actions/` — **12+ registered
  verifiers** (`content_duplication`, `decision_regression`, `empty_section`,
  `unbuilt_internal_link`, `page_canonical_collision`, `literal_markdown`,
  `hardcoded_section_colors`, `orphan_element_refs`, `truncated_component`, `revenue_shape_cta`,
  `missing_conversion_path`, `dead_fragment_link`, `needs_brand_head_assets`, …) — and **no call
  to `GetVerifier` anywhere in `UpdateWorkItemStatusAction`**.
- `required_fields_missing` is itself an **unverified type by declaration**:
  `verifier_coverage_test.go:237` lists it `{catMechanical, "carries page_id and component_id"}`.
- ⚠ **A count is owed here and this file does not have it.** How many live agent definitions
  reach `complete` through `update_work_item_status` rather than `complete_work_item` — and how
  many of those carry an item_type that *does* have a registered verifier — is the number that
  sizes this bug, and it was not measured. Start here:
  ```sql
  SELECT type, jsonb_path_query_array(default_config, '$.**.action') @> '["update_work_item_status"]' AS uses_uwis
  FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
  Then cross that against the registered verifier list. **Do not size this from `bugs_open/367`
  alone** — that is one router, found by accident.

## 3a. THE OWED CENSUS, RUN `[MEASURED 2026-08-24 18:49Z, live DB + live code]`

Run by a reader from the (closed) `bugfix_367_router_remit` lane, at the invitation of the ⚠
above. **The ⚠ is now ANSWERED — it stays on the page because the warning it carries about
sizing is still right, and because the answer changes the priority order in §4.**

### The blast radius is FOUR agents and SIX `complete` arms — of **200** live agent definitions

`update_work_item_status` is named by **6** live agents; `complete_work_item` by **4**. Only these
four reach `complete` through the unguarded writer (`$.workflow.steps.*` where
`action='update_work_item_status'` and `config.status='complete'`):

| agent | step(s) setting `complete` |
|---|---|
| `image-build-handler` | `mark_work_item_complete` |
| `image-source-unsatisfiable-handler` | `close_complete` |
| `image-url-404-handler` | `close_complete` |
| `required-fields-missing-handler` | `close_converted`, `close_resolved`, `close_stale` ← `bugs_open/367`'s router |

The other two agents naming the action (`css-patch-agent`, `page-build-handler`) use it **only for
`failed` and `needs_human_review`** — statuses a verifier has no opinion about, since the framework
exists to stop a saga that *claims success*. `page-build-handler` matters here: it handles four
VERIFIED types and completes them, but its completions do **not** go through this writer.

### The intersection with registered verifiers is **ZERO**, and that is the headline

Those four agents handle exactly five item types, all history:

| item_type | rows | completed | last filed |
|---|---|---|---|
| `needs_imagery` | 183 | 92 | 2026-08-24 |
| `required_fields_missing` | 64 | 38 | 2026-08-23 |
| `image_source_unsatisfiable` | 15 | **0** | 2026-08-09 |
| `needs_hero_image` | 5 | 3 | 2026-08-24 |
| `needs_logo` | 3 | 1 | 2026-08-22 |

**None of the five has a registered verifier.** So **no verifier is being bypassed today** — the
defect is LATENT, not active. **134** completions all-history have taken the unguarded path
(92+38+3+1); every one of them was unverifiable by any writer, guarded or not.

⚠ **This zero is disconfirmable and was controlled**, because a zero from a mis-spelled `IN` list
looks identical to a real one. Positive control: the same 13-type list run without the handler
filter returns **12 of 13 types with real rows** (`unbuilt_internal_link` 89, `empty_section` 56,
`literal_markdown` 52+10+8, `needs_brand_head_assets` 14, …), every one of them handled by an agent
NOT in the table above. The spellings are right and the separation is real.

**Correction to §3's first bullet, same measurement:** the registered count is **13** as of
2026-08-24, not "12+". And the grep in that bullet **under-reports by construction** —
`RegisterVerifier(` does not match `RegisterVerifierWithPolicy(`, which is how
`hardcoded_section_colors` and `needs_brand_head_assets` are registered even though §3 lists them.
Use `grep -rhn 'RegisterVerifier\(WithPolicy\)\?('`. (I hit this myself: my first census returned
11 and I only caught it because §3 named two types my own grep had not found.)

### What makes it a real bug anyway: it is a trap laid for the NEXT person, by name

`verifier_coverage_test.go` classifies every unverified type and says of one category, in its own
words: *"catMechanical: the defect has a re-runnable predicate and the item carries enough identity
to locate it. **These SHOULD get verifiers — this is the actionable backlog, not an excuse list.**"*

Two of the five are on that backlog — `required_fields_missing` (`:237`, catMechanical) and
`image_source_unsatisfiable` (`:240`, catMechanical) — and the other three are `catCreation`,
*"verifiable in principle by an existence check"*. **So the framework's own maintained list invites
somebody to write exactly the verifier that this action will silently ignore.** They will register
it, the coverage test will go green, and the type will be no more protected than before.

And `CQ-023` already warns that the `required_fields_missing` one *"would fail-close the `converted`
arm's completion"* — so the first person to work that backlog walks into a trap from both sides at
once: the guard that will not fire, and the arm that will over-fire if it ever does.

### What this does to §4's ordering

- **Candidate 1 gets cheaper and less urgent at the same time.** Less urgent because no verifier is
  being skipped today. Cheaper because **`RFC_022`'s narrowing (owner ruling 2026-08-11) applies
  squarely**: an opt-in field whose unsafe default is OFF and which **no live consumer names** is
  NOT architecture-scope — and the enumeration is now done, above, rather than asserted. On this
  measurement candidate 1 can go through the **normal council gate**.
- **Candidate 3 is now provable rather than merely honest.** The overstatement in the coverage
  guard's promise has a number: 4 agents, 6 arms, 5 types, 134 completions.
- **Candidate 2 (unify the writers) is unchanged** — still the structural fix, still squarely
  architecture-scope.
- **A fourth candidate the file did not have:** teach the coverage guard that *"has a verifier"* is
  not *"is verified"* unless every completer of that type consults one. That is the change that
  keeps the zero above TRUE instead of merely true-today, and it is the only one that protects the
  person the trap is set for.

## 4. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Consult the verifier on this path too**, behind an **opt-in step-config key whose unsafe
   default is OFF** — the shape the owner ruled for new authority on a shared seam
   (2026-08-02 §2: *"when a seam's widest branch is licensed by 'callers must all be X', make X a
   field with the unsafe default OFF"*). Arm it per step, so a reviewer of the **caller** sees the
   decision. Note `RFC_022`'s narrowing: an opt-in field with an unsafe default OFF that no live
   consumer names is **not** architecture-scope — but the moment a consumer names it, and
   certainly if the intent is to arm it fleet-wide, it is.
   ⚠ **`CQ-023` already warns of a live consequence:** a verifier registered for
   `required_fields_missing` *"would fail-closed the `converted` arm's completion"*. So arming is
   not free per type, and whoever arms one must read that type's close paths first.
2. **Make the two writers one.** Have `UpdateWorkItemStatusAction`'s `complete` arm delegate to
   `CompleteWorkItemAction`'s guarded path, the way `workItemHandlerRegisteredSQL` was unified in
   `bugs_closed/284` (owner ruling 2026-08-17) with a structural single-definition test to stop a
   fourth copy appearing. Cleanest, largest blast radius, and squarely architecture-scope.
3. **Do nothing, but record it honestly** — the verifier framework's coverage guard
   (`verifier_coverage_test.go`) currently reads as though registering a verifier protects a type.
   For any type completed via `update_work_item_status`, it does not. At minimum that test's header
   should say so, or its promise is overstated for an unmeasured share of the fleet.

## 5. How to verify a fix

Register a verifier for a type that is completed via `update_work_item_status`, make its predicate
return "still failing", drive one item through, and require the completion to be **refused** — then
the negative control in the same run: the same path with the verifier's predicate satisfied must
still complete. Assert at the item's status and at the verifier's own record, not at the saga's
report — the saga reporting success is the thing under test.

⚠ A mock's own bookkeeping cannot assert this negative: mutate the guard to prove it is load-bearing
(`LANDMINES.md`, *"a mock's own bookkeeping cannot assert a NEGATIVE"*).

## 6. Where the record lives

`docs/agent_docs/docs024_key_docs_latest/bugfix_367_router_remit/` (found here, in NOTES and
README_where_we_are). Related: `bugs_open/367` (the instance), `bugs_open/017` (a saga reporting
success without touching the defect), `bugs_open/213` (`GradesFunc` — why a verifier keyed on
item_type alone mis-grades a second producer's items, and why a code-side producer list is
refuted), `bugs_open/021` §INSTANCE 2 (the coverage guard), `bugs_closed/284` (the
single-definition precedent for unifying duplicate writers), register `CQ-023`, `WII-003`.

**No `090` diagnosis run.** Stated plainly per the owner ruling of 2026-07-31, because this file
asserts a structural property. The substitute is first-hand verification and it is direct: the
action was read end to end, its `complete` arm quoted above, and the absence of any `GetVerifier`
call on that path confirmed by grep over the package. **What is NOT established is the blast
radius** — §3 says so and gives the query. A thread taking this on should run that census before
choosing between the candidates, and should file `090` if it intends to assert a cause beyond
"the call is absent".

---

## 7. WHAT THIS LANE DID `[2026-08-24, docs024_key_docs_latest/bugfix_375_completion_verifier_gap/]`

**Candidates 1 and 3 are SHIPPED (committed, inert until the next chassis roll).
Candidate 2 is untouched and still the structural fix. Candidate 4 is NOT built —
and §7c below is the concrete shape for whoever takes it, which the file did not have.**

Commits: `c735bfd9c` (the gate + tests), `c94212ad3` (the documents).
Council: `Council-Submitted: 7a6add95-30e9-4576-85e5-df5bad0f7119`.
Register: **`WII-030`** (`register/work-item-integrity.md`). Landmine appended and synced.

### 7a. The census was re-run before anything was built, and it holds

Identical on the headline: **4** agents of **200**, **6** `complete` arms, **5** item types,
**134** completions, **ZERO** intersection with the **13** registered verifiers, controlled.
Two corrections to §3a, both about the *neighbouring* numbers, not the bug's own:

- §3a's control reports "**12 of 13** types with real rows". That is **12 `(item_type,
  handler_agent)` PAIRS from 10 DISTINCT types** — `literal_markdown` alone has three
  handlers, and `orphan_element_refs` / `page_canonical_collision` / `revenue_shape_cta`
  have **no rows at all**. The control still passes; the figure was a row count read as a
  type count.
- The census query in §4 walks `$.workflow.steps`, which **cannot see a step nested inside
  a loop-step config**. On `update_work_item_status` it happens to be complete (a recursive
  `strict $.**{0 to last}` scan returns the same 22 steps) — but on the *other* writer it
  finds only **2** of the **4** live `complete_work_item` callers. **Use the recursive form**
  (RUNBOOK in this lane's directory).
- ⚠ A third, not previously noted: `status` **defaults to `complete`** when the key is
  absent, so `WHERE config->>'status'='complete'` cannot see a step that omits it. All 22
  name it explicitly as of 2026-08-24; `COALESCE` it or that stops being visible.

### 7b. The finding that decided the design: `CQ-023`'s landmine was FALSE

`CQ-023` warns whoever registers a `required_fields_missing` verifier that it *"would
fail-closed the `converted` arm's completion"*. **All three of that router's `complete` arms
(`close_converted`, `close_resolved`, `close_stale`) are `update_work_item_status` steps** —
so registering it would not fail-close anything; it would do **nothing at all, silently**,
while `verifier_coverage_test.go` went green.

So the trap in §3a is worse than "nobody warned them": **they were warned, and warned wrong.**
And it rules out the tempting fix — making the consult automatic would make that sentence
come true, i.e. break a live route as a side effect of arming a guard nobody asked for.
Hence opt-in **per step**. `CQ-023` is corrected in place; the warning is now true per arm,
once that arm is armed.

### 7c. Candidate 4, made concrete — and the class is ALREADY SOLVED ONCE, for a THIRD writer

The file (and the handoff) framed candidate 4 as "teach the coverage guard that *has a
verifier* ≠ *is verified*". Half of that shipped as documentation. **The enforcing half has a
precedent nobody in this lane's history had noticed, and it is the shape to copy.**

**There is a THIRD writer of `complete`.** The `claimed-item-timeout` sweep auto-completes a
claimed item past its timeout **by writing the row directly, so NEITHER completion gate runs**
(`bugs_closed/317`). Its protection is:

1. a **declaration** in Go — `livespec.ClaimedItemTimeoutExclusions`;
2. a **build-enforced lockstep** — `platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go`,
   asserting **both directions**: `excluded ⇔ (has a verifier) OR (has a noChangeGates entry)`.
   Forward, so a gated type cannot be swept past its gate; reverse, so an exclusion that no
   gate can earn does not create the `bugs_open/006` §C churn;
3. a **live-drift auditor** for the declaration-vs-production half —
   `cmd/config-key-audit --live-declaration-drift` (`bugs_open/363` phase 2).

That guard is what caught this lane's own fixture defect (§misstep 3 in NOTES), which is the
best possible evidence that it works.

**So candidate 4 is: the same three parts, for THIS writer.** A declaration of which
`(agent, step)` pairs complete through an **unarmed** `update_work_item_status`, a lockstep
asserting that no type with a registered verifier is reachable by one, and the drift auditor
comparing the declaration to live `agent_definitions`. What it buys over what shipped: the
runtime record (`result._verification.status='verifier_not_consulted'`) fires only **after** a
live item has already completed unverified, whereas a lockstep fails **the build**, at the
moment somebody writes `RegisterVerifier(…)`. They are complements, not alternatives.

⚠ **What that costs, stated rather than discovered:** a hand-maintained declaration that goes
stale by ADDITION — a new agent completing through this writer is invisible until somebody
refreshes it. That is exactly the criticism the council levelled at the first cut of
`verifier_coverage_test.go` ("a guard against *someone must remember* that itself relies on
someone remembering"), and `--live-declaration-drift` is the estate's answer to it. Do not
build the declaration without the auditor arm.

### 7d. What is still NOT established

- **Whether any of the five types should have a verifier.** Unchanged from §6 of the handoff.
- **Whether the 134 unguarded completions contain false completions.** Nobody has re-run
  those predicates. Still its own measurement, and probably its own `090`.
- **Whether candidate 2 is feasible.** I read `CompleteWorkItemAction` end to end and the two
  paths now share the gate and the row read, which is its first half — but I did **not**
  enumerate its call sites or judge the merge. `bugs_closed/284` is still cited by shape.
- **The `image_url_404` undispatched population** the handoff observed (42 rows, 38
  `detected`, all with an empty `handler_agent`). Untouched, unfiled, still not this bug.
