# SUMMARY — the component Go gate is proven, 2026-07-31 (evening)

## What we're trying to do

Add `component` as a fifth `subject_type` travelling docs (`doc_plans`/`doc_notes`) can carry,
so a component — not just a tool or a pipeline — can have a PLAN and a NOTES trail the same
enforcement machinery already gives everything else. The wider goal this sits inside is
`features_open/027`: a staged, gated way to build tool and component fences without eight
stages of ceremony (that ladder was cut to three — D8 — earlier this lane).

The type has two independent enforcement points that have drifted apart twice before
(`bugs_open/064`, migrations 163 and 184): the Postgres CHECK constraints, and a Go-side
allowlist compiled into the running chassis binary. Both have to accept `component` or the
type is only half-shipped, and the two previous drifts each shipped one half without the
other and called it done.

## Where we've come from

The DB half went live by hand on 2026-07-31 (migration 273, applied ahead of the image per
the standing image-then-seed discipline this platform requires). The Go half — the
`validDocSubjectTypes` slice in `platform/orchestration/actions/doc_subjects_common.go` and
its consumers — was believed shipped because the running pod's build date postdated the
commit that added `component` to the slice. That is exactly the inference this lane's own
landmine list already warned against: a build date is a claim about *when*, not about
*what*, and the vocabulary entries are short string literals Go compiles to immediate
comparisons rather than anything a `strings`/grep pass on the binary can see either way.

Two consecutive handoff documents (`HANDOFF_2026-07-31_continue_here.md`, then
`HANDOFF_2026-07-31b_continue_here.md`) named this the single open item and specified how to
close it: seed a scratch agent with a `load_doc_context` step configured for
`subject_type: component`, dispatch it, and read back either the PLAN body (pass) or the
string `unsupported subject_type "component"` (fail).

## What we've done

Read the code before writing the probe, rather than trusting the handoff's recipe verbatim —
and that reading changed the plan twice:

1. **The predicted failure string was wrong.** `unsupported subject_type "…"` is
   `docSubjectGateReason`'s message, and its only caller is `persist_diagnosis_note`.
   `load_doc_context` goes through a different function, `docResolveSubject`, whose message
   is `subject_type must be one of …, got "…"`. Following the handoff literally would have
   left a real pass indistinguishable from a run that failed for an unrelated reason, because
   the string it told you to look for appears on neither branch of the route it recommended.
   Corrected in both handoff files and logged in `WRONG_CALLS.md`, because the error was
   invisible on a successful test and would only have surfaced as confusion, not as a wrong
   answer.
2. **No scratch `agent_definitions` row was needed at all.** `selectWorkflow`'s first priority
   (`platform/messaging/processor.go:922-928`) takes a workflow travelling inline in the
   dispatched message over any database lookup. So the whole probe — including its own
   control arm — rides in one Kafka message, with nothing to seed and nothing to snapshot.

Built `scripts/PROBE_doc_subject_go_gate.sh`: one dispatch, two steps. The first loads
`subject_type=component` for real. The second, in the same run, loads a deliberately invalid
type as a control — because a green first step and a step that silently never ran look
identical unless something is watching for the failure that *should* happen. The control's
error message is assembled at runtime from the live `validDocSubjectTypes` slice
(`docSubjectTypesQuoted()`), which is the only way to read the vocabulary as the pod actually
compiled it, rather than as the source tree says it should be.

Dispatched it against `v1.0.1215`. Both replicas answered the same way: `has_plan=true`, a
827-byte PLAN body byte-identical to the one written for the probe, a criteria fence
extracted from it, and the control arm errored with the pod's own compiled list —
`'tool', 'pipeline', 'experience', 'action', 'experience-pattern', 'component', 'landmine'`.
Correlation `8f564028-6fc6-488c-96d2-c2e362b243b2`.

Cleaned up the throwaway PLAN row afterwards (`doc_plans` is back to 0 `component` rows,
reconfirmed again this evening after further unrelated chassis rolls) and wrote up all five
standing docs plus the concept register (`DOC-068` corrected, `DOC-070` added for the new
instrument) and `LANDMINES.md`. Committed as `2efae4d29`, after another session's broad
commit (`96799d39d`, `c4307578f`) had already swept some of the same edits into HEAD ahead of
me — nothing was lost; the remaining commit is genuinely just what was still outstanding.

## Where we are now

The Go half of `subject_type='component'` is proven live, at the artefact, not inferred from
a build date. Both enforcement points now agree. Everything is committed and confirmed
present at HEAD.

What is deliberately **not** claimed: `doc_plans` holds zero `component` rows again — the
capability is proven, a *use* of it is not. `persist_diagnosis_note`'s own gate
(`docSubjectGateReason`) was never dispatched in this probe; it shares the same
single-sourced slice, so membership carries to it, but that is an argument from the code, not
an independent observation the way `load_doc_context`'s result is.

Since this proof ran, the chassis image has moved twice more for reasons unconnected to this
lane: `v1.0.1216` shipped `bugs_closed/157`'s fix to `has_visible_area` in
`browser-runner-adapter` (that check had been deliberately left out of every fence this lane
built, and is now safe to use), and the chassis pods are currently on `v1.0.1219` following
further unrelated fixes from other threads (`bugs_open/165`, `138`, `137` per `git log`).
Neither changes anything claimed above — the `doc_subjects_common.go` vocabulary was not
touched by either — but it is the reason any figure in this file should be re-grounded rather
than assumed current if it is read much later.

## Where we're going

The handoff's ordered open-items list moves to its second entry: **S6 for components** —
dispatching a component's fence to `browser-runner-adapter` the way `tool-acceptance-agent`
already does for tools. This lane's own tooling (`try_fence.go`,
`prove_fence_can_fail.go`, TL-036) makes it possible to author and prove a component fence
before any dispatch path exists, which is why the handoff calls this step "wiring, not
construction." A fresh, detailed handoff for that step follows this summary, because it is
new scope rather than a continuation of the work this file records.

Behind it, in order: the ten-tool authoring backlog (a cleanup, explicitly not a rule to wire
into tool birth); `features_open/028` (rename-orphan detection, filed and unowned); and
`has_visible_area` checks being owed to every existing fence now that `157` no longer blocks
them — owned by another thread, not to be duplicated here.
