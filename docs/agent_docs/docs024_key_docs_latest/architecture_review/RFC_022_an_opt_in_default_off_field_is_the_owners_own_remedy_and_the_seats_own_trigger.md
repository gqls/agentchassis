# RFC 022 — an opt-in, default-OFF field on a shared action is BOTH the owner's prescribed remedy AND the architecture seat's RFC trigger

## STATUS: **CLOSED — OWNER RULING 2026-08-14: the budget is N = 10.** (Option (3) reached: counter built 08-13, N ruled 08-14.)

The owner's ruling, with its framing correction recorded verbatim in spirit: *"when we
started we wanted every agent to be 'somewhat independent' so it could be used in other
workflows when required. We want to continue this idea. The 'shared action nobody
understands' sounds like it isn't right."* So: **sharing is estate design, not the
defect.** A budget finding means an action's ACCUMULATED optional surface has grown past
N=10 and owes one architecture review **as a whole** — it never means the action's reuse
is a problem, and de-sharing is not the default remedy. At N=10 the standing stock is
exactly three: `analyse_repo_local` (12), `append_doc_note` (11), and
`diagnose_prepare_fix_commit` (11) — each owes one review round, after which its
acknowledged level is the baseline. `scripts/audit-optional-key-budget.sh` now defaults
to N=10 (a bare run enforces the ruling; `--census` keeps the no-budget census). Roster
clauses updated to the ruled form by migrations `404`/`405` (the "N is not yet ruled"
sentence 402/403 shipped was falsified by this ruling).

> **UPDATE 2026-08-17 — the cron half SHIPPED, so nothing here is open any more.**
> `optional-key-budget-check` is live in the cluster (CronJob, `50 6 * * *` UTC,
> created 2026-08-14) and has written its `doc_notes` row every day since — 08-14,
> 15, 16, 17 — including on clean results, which is the property that makes a
> MISSING row mean "the job did not run" rather than "nothing is wrong". The
> paragraph below (and CLAUDE.md's RFC_022 clause, corrected the same day) still
> read "whether to build it at all is the remaining choice", which sent a session
> to build what already existed.
>
> **The one live defect found while verifying it, now fixed:** `check.py`'s
> `OPTIONAL_KEY_COUNTS` literal had drifted from the registry — `retract_asset_files`
> (4 keys) and `publish_site` (3) had entered the registry counted as **ZERO**, so
> their optional surface was invisible to the accumulation check. The parity test
> `TestBudgetCronCountsLiteralMatchesTheRegistry` existed and was FAILING at HEAD;
> nobody had run it. Regenerated (118 → 120 entries, +2 lines, nothing else
> touched), committed, and the kustomize overlay re-applied so the CLUSTER carries
> it (verified at the live ConfigMap, not at the file). **Blast radius, stated
> honestly: no finding was missed** — both actions are far under N (4 and 3 keys,
> 0 and 1 carriers). The blindness was latent, not live; it would have mattered to
> an action that grew while unlisted.

**Still open, operational not architectural:** whether the counter runs on a daily
CronJob or stays on-demand. The owner ruled the language half (2026-08-14, "we can keep
the python"): the RFC_006 check's Python mirror STAYS — no Go-native rework — and a
counter CronJob, if built, may follow that same shape. Whether to build it at all is
the remaining choice; until then the counter is run-on-demand with N=10 as its default.

*The section below is the 2026-08-11 ruling as it stood before N was set, kept as the record.*

## Previous STATUS (2026-08-11): **DECIDED — OWNER RULING 2026-08-11: option (3), a BUDGET, with option (1) as the interim.**

The owner took the recommendation in §"The three options, costed" as written. Both halves,
and the ordering between them, matter:

| half | what it means | state |
|---|---|---|
| **interim — option (1)** | An opt-in field whose unsafe default is **OFF** and which **no live consumer names** is **NOT architecture-scope**. The architecture seat should not raise `needs_rfc` on that shape alone. This makes the 2026-08-02 ruling self-consistent: the prescribed remedy stops being penalised. | **effective immediately** |
| **destination — option (3)** | The trigger moves to the **accumulated count**, not the single addition: an RFC when a shared action's optional-key set grows past a threshold. Needs the mechanical counter over `RegisterActionInputSpec` declarations per action, ~~which does not exist yet~~ **BUILT 2026-08-13** (below). | **counter built; budget N + cron half await the owner** |

### The counter — BUILT 2026-08-13 (register WFA-013)

`cmd/config-key-audit --optional-key-budget [N]` / `scripts/audit-optional-key-budget.sh [--json] [N]`
(bugfix_223_index_answerability lane, per §0 item 2 of its 2026-08-10 handoff). It counts
`len(ActionInputSpec.Optional)` per registered action — exactly the "sweep over
`RegisterActionInputSpec` declarations" this RFC named — joined against DISTINCT live
carrier agents from the same export the RFC_006 check reads, walked with
`validation.WalkSteps` (a `substeps` loop body cannot hide a carrier). `over_budget`
fires only on a SHARED action (≥2 carriers) past N; without N it is a report-only
census. Mirrors WFA-006's shape deliberately: same binary, same decode, refuse-on-empty
both sides, wrapper discriminates refusal by empty stdout (`go run` folds exit codes).

**The live distribution, measured 2026-08-13 — the sizing input for the owner's N:**
118 actions declare optional keys; **21 are shared**. Widest shared surfaces:
`analyse_repo_local` **12** (2 carriers), `append_doc_note` **11** (**8 carriers** —
this RFC's own motivating action), `diagnose_prepare_fix_commit` **11** (2 carriers),
then `diagnose_build_gate` 8, `write_doc_plan` 8 (3), `create_rerender_items` 7 (3),
`diagnose_persist_fix_plan` 7 (3), `plan_sections` 7 (3). A budget of **10** flags
exactly the top three today; **12** flags none.

**What still awaits the owner, and this RFC stays open until the first is ruled:**
1. **The budget N.** A governance choice, not a technical finding (§options, (3)).
2. **The cron half** — RFC_006's shape came with a daily CronJob because a pre-commit
   hook cannot gate live config; this starts **report-only** because that choice was
   explicitly reserved ("decide with the owner whether the counter needs the cron half").

The roster clause shipped by `381`/`383` said "that counter is not built yet" — updated
by the follow-up pair `402` (fix-proposer) / `403` (council-gate), same surgical anchored
pattern, so the seats now cite the counter for the exact figure instead of "several".

**The interim is not a weaker version of the destination — it is the half that is safe to
ship without the counter.** Option (1) alone deliberately gives up the *accumulation* signal,
which is the real harm the seat named; option (3) is what gets it back, aimed at the tenth
field instead of the first. **So (1) must not be allowed to become the settled answer by
default.** Whoever builds the counter closes this RFC; until then the estate is running with
a known, stated blind spot rather than an unknown one.

**Answering the question in §"The question for the owner, in one line":** yes — an opt-in
field with the unsafe default OFF, named by no live consumer, satisfies the 2026-08-02
ruling and falls **outside** the trigger. Compliant work should **not** expect `needs_rfc`
as routine.

### What shipped, same day

| file | target | state |
|---|---|---|
| `381_rfc022_narrow_the_architecture_seat_opt_in_default_off.sql` | `fix-proposer.review_architecture` | **APPLIED + recorded**, prompt 10,442 → 11,829 chars |
| `383_rfc022_narrow_the_gate_architecture_seat_without_reverting_377.sql` | `council-gate.review_architecture` | **APPLIED + recorded**, 10,479 → 11,866 chars |

Both insert a byte-identical clause at a verbatim anchor, so the two rosters cannot say
different things about the same trigger. The clause states the exception, its three
conditions, and — deliberately, inside the prompt — **what the seat is no longer watching**,
so the accumulation blind spot is visible to the reviewer rather than silent. It also keeps
a reduced form of the signal: if the plan shows an action already carrying several optional
keys, report the observed count as `insufficient`. That is the most of (3) that can be had
before the counter exists. Verified live in both rows: clause present, original trigger text
intact, routing-signal contract intact.

> **⚠ AND ONE THING THIS RULING FLUSHED OUT, which matters beyond RFC_022.**
> `099_SYNC_gate_roster.py --apply` — the mirror CLAUDE.md tells you to run after any seat
> change — **would have reverted migration 377** and destroyed the council-gate prompt
> caching (**68% measured saving**, on ~85% of fleet LLM spend). Its transform predates 377
> and rebuilds every gate prompt in the pre-hoist order with no `<!--CACHE_BREAKPOINT-->`.
> Its dry run says `drift (steps that would change): [all 17]`, which reads as "mirror me"
> and means "the gate is **ahead** of the mirror". That is why `383` exists as a surgical,
> guarded migration instead of a mirror run — it asserts the breakpoint does not move and
> that the shared prefix does not fragment. **`099 --apply` is suspended** (CLAUDE.md +
> `LANDMINES.md`) until it learns about 377. **Fix `099` and both the suspension and `383`'s
> hand-patch exception end.**

---

**Filed 2026-08-10 by the `bugfix_223_index_answerability` lane. For a human to break, not
for a thread to argue.** Raised because the architecture seat signalled `needs_rfc` at
MEDIUM on a change that was built to follow the owner's own ruling about how such changes
should be built — and it did not misread anything. The two rules genuinely point in
opposite directions for one specific shape, and that shape is now common.

**Nothing is blocked and nothing is being appealed.** Council `495df717-4010-491f-aec0-92c13aaf3809`
returned **APPROVED** (6 advisory objections, none high) and the change is live. Five of the
six objections were acted on the same afternoon. This RFC is about the sixth, which cannot
be closed by better measurements because it is not a measurement question.

## The exact shape

`bugs_open/223` phase 1 added, to `append_doc_note` — a shared action with **8 live
consumers** (`component-template-fixer`, `council-gate`, `domain-research-classifier`,
`experience-planner`, `landmine-verifier`, `tool-acceptance-agent`, `tool-improver`,
`tool-recreation-handler`) — one optional config key, `note_body_suffix_field`:

```go
func applyBodySuffix(body, suffix, fieldName string) string {
	if fieldName == "" { return body }          // absent ⇒ byte-identical to before
	…
}
```

It is unreachable until a workflow names it; measured 0 of 8 consumers name it; the default
is the unsafe-side-OFF; and the same commit registered it (DIAG-042).

## The two rules, quoted

**The owner's ruling of 2026-08-02 §2** — the prescription this change followed:

> **New authority on a shared seam ships as an OPT-IN FIELD, not a documented contract.**
> … when a seam's widest branch is licensed by "callers must all be X", make X a field with
> the unsafe default OFF. It costs about four lines, it moves the decision to where a
> reviewer of the CALLER can see it, and it is the only version that survives a session that
> did not read the helper.

**The architecture seat's own trigger**, applied to this change (verbatim from the round):

> `note_body_suffix_field` is a new reserved config key on `append_doc_note`, a widely-reused
> shared action; even default-off, this is the precedent shape (`bugs_closed/124`, `129`) the
> RFC trigger exists to catch. … The design is genuinely careful — opt-in, default-OFF,
> byte-identical when unset, 0 measured prior consumers of the new keys, no ordering
> constraint claimed — **which is the sanctioned pattern from the 2026-08-02 owner ruling for
> extending a shared seam without a formal RFC gate. That mitigates severity but does not
> relocate the trigger** … That is architecture-scope by the trigger test regardless of the
> author's declaration.

Read those together. **The remedy the owner mandated for shipping new authority safely is
itself the thing the seat is required to flag as architecture-scope.** A thread that follows
the ruling exactly still draws the signal, so the signal stops discriminating between careful
work and careless work — which is the property that makes a signal worth having.

## Why this is not resolved by the 2026-07-29 narrowing

That ruling already narrowed the trigger once:

> An addition to a shared vocabulary needs an RFC only when it changes what the shared
> mechanism GUARANTEES … A type that only adds an opt-in capability, reachable by nothing
> until a document names it, goes through the normal council gate.

By that text this change is plainly *not* RFC-scope: it is reachable by nothing until a
document names it, and it changes no guarantee for the other 7 consumers. **But the seat's
brief keys on the SHAPE** — a new reserved key on a shared action — and the two 2026-08-02
clauses do not obviously yield to the 2026-07-29 clause, because the later ruling is about
*how to ship new authority* and the earlier one is about *when a vocabulary needs an RFC*.
An opt-in field is simultaneously both. Nobody has said which reading wins, so both seats and
threads are guessing, and they guess differently.

## The three options, costed

1. **Narrow the seat's trigger explicitly: an opt-in field whose unsafe default is OFF and
   which no live consumer names is NOT architecture-scope.** Cheapest, and it makes the
   2026-08-02 ruling self-consistent — the prescribed remedy stops being penalised. Cost: the
   seat loses sight of a real class of drift, *accumulation*. Ten such fields, each
   individually inert, are a shared action nobody understands; the trigger is currently the
   only thing that would notice the tenth.
2. **Keep the trigger and accept the signal as routine** for this shape, with the seat's own
   MEDIUM-not-HIGH reasoning doing the work (it explicitly said the cost of *not* changing —
   81% of landmine footprints unresolvable, a measured 1-in-4 false STALE — argues for
   proceeding now). Cost: `needs_rfc` fires on compliant work, so it decays into noise, and
   the next thread reads it as a formality. **This is the status quo and it is what this RFC
   exists to name.**
3. **A budget rather than a per-change gate.** Let an opt-in field ship without an RFC, and
   have the seat trigger on the *count* — e.g. an RFC when a shared action's optional-key set
   grows past N, or when two are added inside one quarter. Most faithful to what the trigger
   is actually protecting (accumulated surface, not any single addition), and it needs a
   mechanical counter nobody has built: `SELECT` over `RegisterActionInputSpec` declarations
   per action, which is a real but small piece of work.

**This lane's recommendation is (3), with (1) as the interim** — because the harm the seat
names is real and is about the tenth field, while the tax it levies falls on the first. But
this is a judgement about how the estate wants to be governed, not a technical finding, and
it is the owner's to make.

## What was already done, so this is not asking for the same work twice

Everything measurable in the round was measured and acted on:

- **8 consumers of `append_doc_note` enumerated** (query above), 0 naming the new key —
  the guardian's "asserted, not enumerated per-caller" objection, closed.
- **`answerCodeCheck`'s callers swept repo-wide** (`grep -rn`, 2 hits: the action and
  `diagnose_load_runtime_action.go:484`) — the low-severity signature objection, closed.
- **The "no compose action exists" claim swept** rather than resting on one candidate: the
  only formatter-shaped actions are `format_research_content` and `format_crawl_for_analysis`,
  both web-domain content formatters, neither a generic collected-data composer —
  `prior_art_librarian`, closed.
- **This council's own precedent on the seam checked** — 4 prior `council_report` rows
  mentioning `diagnose_code_lookup`/`landmine-verifier`, most recent 2026-08-06, all
  approved, none addressing answerability. No verdict is being repeated or contradicted —
  `prior_art_librarian`, closed.
- **`bug_historian`'s strongest objection closed in code, not prose:** the gate "could ship,
  look wired, and never actually gate a single verdict". `codeEvidenceGateField` is now a Go
  constant and a test asserts seed 365's condition string equals
  `"lookup." + codeEvidenceGateField + " == true"`, so a rename fails a test instead of
  silently unwiring production. Proven by mutation. The half that cannot be bought at build
  time — that the TRUE branch is reachable on a live run — is a named acceptance step.
- **`debug_historian`'s needle gate added and PROVEN** by inducing it against the
  already-applied state: it refused with `needle gate: run_checks.next_step is
  gate_evidence, expected the pre-365 value 'verify'`. The seed is now recorded in
  `schema_migrations` via `--record-only`.
- **`editquality`'s MEDIUM was a real gap and is fixed in code:** the third false-positive
  mode (a `content` check aimed at a non-Go file answered by a same-named Go symbol) now
  carries a caveat on non-empty answers, mutation-proven at the call site.

**The one process objection accepted without remedy:** the guardian was right that a
workflow-JSON edit should be filed as `operation: "config_change"` naming the owning
pipeline, not `"add"` on a new `.sql` file. Recorded here rather than fixed, because a
submitted plan cannot be amended and forward-only forbids rewriting the round. Next
submission in this lane will use `config_change`.

## The question for the owner, in one line

**Does an opt-in field with the unsafe default OFF, named by no live consumer, satisfy the
2026-08-02 ruling and therefore fall OUTSIDE the architecture seat's RFC trigger — or does
the trigger stand and compliant work should expect `needs_rfc` as routine?**

Related: `bugs_open/223`; `bugs_closed/124` and `129` (the precedent shape the seat cites);
`RFC_002` (the last time two seats reached opposite defensible conclusions in one round, and
the ruling that followed); register DIAG-042.

---

## ADDENDUM 2026-08-25 — a blind spot in the counter, surfaced by `bugs_open/345`, and the decisions it puts to the owner

> **RESOLVED 2026-08-25 — OWNER RULING: option (c). Leave the blind spot; RFC_022 stays CLOSED.**
> The counter continues to count `spec.Optional`; directly-read optional literals stay uncountable,
> and `censusUncountedActions` printing spec-less actions as "unknowable" is the accepted coverage.
> No change to the counter, no spec added to `update_work_item_status`, and `bugs_open/345`'s
> `stop_on_repeat_failure_item_types` stays uncounted (harmless at 3 vs N=10). **The analysis below
> is kept as the record of the known gap — it is NOT an open work item; do not action decisions
> 1–3.** If a directly-read carrier ever grows toward N=10 in practice, reopen then, not now.

The RFC is CLOSED and its ruling (N=10) stands. This addendum does not reopen it — it records a
**measurement gap in the counter that enforces it**, found when `bugs_open/345` tried to make one
new opt-in key (`stop_on_repeat_failure_item_types`, on `fail_work_item`) visible to the budget and
discovered the counter structurally cannot see its whole class. It is written as decisions because
they are the owner's, not a session's.

### What was found (measured 2026-08-25)

- The budget counts **`len(ActionInputSpec.Optional)`** per registered action
  (`cmd/config-key-audit/optionalbudget.go`).
- But an action reads its optional config two ways: **through `ExtractActionInputs`** (which walks the
  spec, so those keys are in `Optional`), and **directly from `StepConfig.Config`** (a literal read,
  which the spec deliberately does NOT list). `FailWorkItemInputSpec` says so in a comment:
  `error_message` is read directly and kept out of the spec on purpose.
- So **every directly-read optional key is invisible to the budget.** `fail_work_item` really carries
  at least three (`error_message`, `status_override`, and now `stop_on_repeat_failure_item_types`)
  and declares **0** Optional. `update_work_item_status` reads ~6 optional keys directly and registers
  **no spec at all** (it never calls `ExtractActionInputs`).
- **~35 actions** read optional config literals directly (`grep -l 'params.StepConfig.Config\["'`),
  and specs include such keys **inconsistently** — `feed_actions.go` puts `error_message` IN its
  Optional list while `fail_work_item` keeps it OUT. So there is no single convention to lean on.
- The `censusUncountedActions` half already prints spec-less actions as "unknowable" (not a finding,
  by its own comment). So `update_work_item_status`'s gap is *shown* but not *counted*.

**Consequence:** an action can accumulate optional authority past N=10 entirely in directly-read
keys and never trip the budget — the exact accumulation the RFC exists to catch, in the exact blind
spot the counter cannot see. It is **latent, not live**: today's worst directly-read carrier
(`fail_work_item`, 3) is far under N. But it grows by addition and reads as a clean bill.

### Why `bugs_open/345` did NOT just register its key

The obvious "fix" (add the key to `FailWorkItemInputSpec.Optional`) is wrong three ways: it fights
that spec's stated convention; the spec is **also the runtime extraction spec** (passed to
`ExtractActionInputs`), so adding a key changes what the extractor walks on the failure hot path, not
just what the counter sees; and it would still leave `error_message`/`status_override` and 34 other
actions invisible — fixing the count for one key while lying about the same action's other two. So
345 routed it here rather than patch one action into a false green.

### The decisions, for the owner

1. **Should the budget count what actions actually READ, or what their specs DECLARE?** The honest
   count is "optional config keys the action consumes", which today's counter only approximates via
   `spec.Optional`. Options: **(a)** make the counter scan actual config-key usage (a static read of
   each action for `StepConfig.Config["…"]` plus its extracted keys) — accurate, but a real change to
   the audit and its parity test; **(b)** rule that all optional config keys, directly-read included,
   MUST be declared in `Optional` (making the list mean "the budgeted surface", not "the extracted
   surface") and add a check that a directly-read key with no `Optional` entry is a finding — cheap
   per-action but requires touching ~35 specs and severing "Optional == extracted"; **(c)** accept
   the blind spot as documented and rely on `censusUncountedActions` plus review — zero work, the
   count stays an undercount by design. This is the load-bearing choice.

2. **If (b): does putting a directly-read literal in `Optional` have to stay inert at runtime?** It
   does today for `fail_work_item` only because `ExtractActionInputs`'s Strategy 0 skips non-string
   values — a fragile guarantee to rest a convention on. (b) would want a test pinning that a
   budget-only `Optional` entry is never extracted, or a separate `BudgetedConfigKeys` field so the
   two surfaces stop sharing one list.

3. **Does `update_work_item_status` get a spec?** It has none and uses no extraction, so a spec would
   be counter-only — safe, but under (c) it adds nothing `censusUncountedActions` doesn't already
   show, and under (a) it is unnecessary. Only (b) needs it.

**Recommendation (a session's, not a ruling):** (a) — count what is read — because it is the only
option that makes the number honest without taxing 35 authors or resting a convention on an
extractor's type check. But it is the most work, and the RFC is closed, so it is genuinely the
owner's call whether the blind spot is worth closing now or left documented. No code shipped for
this; `bugs_open/345`'s key stays uncounted (and harmless at 3) until the owner rules.

---

## Addendum 2026-09-03 — a recorded escalation trigger for `on_approve.fan_out_from` (`site_delivery_and_editor`)

Filed here rather than left in a council submission's prose, because a decision that lives only in a
submission has to be re-derived by whoever needs it next — which is the exact failure this RFC is
about, one level up. Raised by the council's `tooling_provenance` seat on correlation `d04c1bc1`.

**What shipped.** `bugs_open/466` added two optional keys to the `on_approve` block that
`checkpoint_for_review` embeds in a human-review work item: `fan_out_from` (file one follow-on per
element of a named array in the approved data) and `defaults`. Both default **absent**, so an
`on_approve` naming neither behaves byte-identically to before — the 2026-08-02 §2 shape.

**Why the RFC_022 exemption was NOT claimed.** All three of its conditions must hold, and the third
does not: migration `750` wires a live consumer (copy-editor) in the same commit. The change went
through the ordinary council gate instead, and the architecture seat read it as a contained
`point_fix` — correctly, for **one** wired consumer.

**The trigger, stated so nobody re-derives it.** `[MEASURED 2026-09-03]` — by a **recursive** walk of
every live, non-snapshot, undeleted `agent_definitions.default_config`, exactly **one** step anywhere
configures `on_approve`: `copy-editor / workflow.steps.request_review`. Zero others, sub-workflows
included.

> **A SECOND live consumer setting `fan_out_from` comes back as `needs_rfc`, not as another
> same-commit wiring.** At that point `on_approve` is a small dispatch contract for the estate's only
> human-in-the-loop copy-approval route, the "zero live consumers" safety property is gone for real,
> and the accumulation this RFC exists to catch has happened.

⚠ **Re-run the census recursively; the obvious query is blind.** `jsonb_each(default_config->'workflow'->'steps')`
is a TOP-LEVEL walk and cannot see a step nested in a `sub_workflow` — the council caught exactly
that in round 2 of this change, where the top-level version happened to give the right answer for the
wrong reason. Use:

```sql
WITH RECURSIVE walk(agent, path, node) AS (
  SELECT d.type, 'workflow.steps', d.default_config->'workflow'->'steps'
  FROM agent_definitions d
  WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
    AND d.default_config->'workflow' ? 'steps'
  UNION ALL
  SELECT w.agent, w.path || '.' || kv.key, kv.value
  FROM walk w, jsonb_each(w.node) kv WHERE jsonb_typeof(w.node)='object'
)
SELECT agent, path FROM walk
WHERE jsonb_typeof(node)='object' AND node->'config' ? 'on_approve';
```

**Sources:** `bugs_open/466` · council `d04c1bc1` (rounds 1–3) · concept register `ADM-012` ·
`internal/core-manager/admin/site_admin_handlers.go`, where the same query and the same trigger
condition are recorded in the handler itself.
