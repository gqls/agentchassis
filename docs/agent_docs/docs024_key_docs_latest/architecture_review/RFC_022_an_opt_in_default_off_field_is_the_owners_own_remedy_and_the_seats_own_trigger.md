# RFC 022 — an opt-in, default-OFF field on a shared action is BOTH the owner's prescribed remedy AND the architecture seat's RFC trigger

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
