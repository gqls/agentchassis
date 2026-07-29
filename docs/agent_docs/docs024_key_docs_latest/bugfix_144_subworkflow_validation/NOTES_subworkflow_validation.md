# NOTES — bugs_open/144 sub-workflow validation

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-29 — picking it up

Checked ownership before starting: `scripts/who-owns.py 144` names only the filing
commit (`cbb08bfb8`) and a docs commit from the filing thread (`396b92a5f`), no owning
workstream. Bug file says "OPEN, unowned". `git log --since=5.days -- platform/validation/`
shows one commit, from the 100/101 lane, not this. Working tree has no edits under
`platform/validation/`. Taken on.

## 2026-07-29 — the measurement, before the design

Exported all 178 live workflow definitions and walked them in python. Numbers in the
PLAN. Two results decided the design before a line was written:

* **All 20 distinct nested actions are `IsLocal: true`, and 0 of the 85 nested steps
  carries a topic.** So the "remote action needs a topic" rule costs nothing today —
  and would have been an instant fleet-wide outage if even one nested action were
  remote. This was the single most load-bearing check of the day and it was one grep.
* **0 nested `(action, key)` pairs trip the strict-config rule.** So the hard-error
  path for a strict action is free today too.

## 2026-07-29 — two things the bug file's fix candidate would have got wrong

Not criticism of the filing thread — both live in the executor, which the bug file
correctly points at but does not quote.

1. **`next_step` out of a sub-workflow is LEGITIMATE.** `coordinator.go:4009-4014`
   passes an external reference through untouched, and
   `loop_expansion_handler.go:192` *injects* `<loop>_complete`, which exists in no
   definition. "Validate each sub-workflow as a workflow in its own right" would make
   both a hard error. Live count of nested steps whose `next_step` leaves the
   sub-workflow: 0 today — so this would have passed the fleet measurement and still
   been wrong, breaking the first workflow anybody wrote that way. **A measurement
   that comes back zero does not tell you the rule is right; it tells you the rule is
   not currently firing.**
2. **`parseSubsteps` drops fields.** `dependencies` is read by the top-level decoder
   and NOT by the nested one. Validating a nested `depends_on` for existence would
   have been enforcing a field the executor discards. Reported as dropped instead.

## 2026-07-29 — misstep: I nearly measured the wrong flag

For "would any nested step hard-fail the strict-config rule?" I measured against
`opted_in` from `config-key-audit --specs`, which is `CheckConfig || len(ConfigKeys)>0`
— NOT `StrictConfig`, which is what `IsStrictConfigAction` actually gates on. I only
noticed when a test failed for registering `CheckConfig: true` and expecting a hard
error.

The measurement survives because the direction is conservative: `opted_in` is a
superset of strict, and the answer was 0 under the superset. **But I had written "0 of
the 66 nested pairs trips the strict rule" before checking which flag I had measured**
— the right answer for the wrong reason, which is indistinguishable from the wrong
answer until someone checks. Corrected in the PLAN to state the population precisely.

## 2026-07-29 — misstep: two tests that could not fail, caught by the compiler and by a control

* `dropped_fields` came back empty because zaptest's `ContextMap()` renders
  `zap.Strings` as `[]interface{}`, and my `.([]string)` assertion silently yielded
  nil — a type assertion with the comma-ok form swallows the failure. The test then
  asserted `len(got) != len(want)` and failed for the right reason by luck. Fixed to
  `[]interface{}`.
* Both new warning tests were written with a positive-control half (a step using only
  honoured fields must report NOTHING). Without it, a warning that fired on every step
  would have passed the first half exactly as well.

## 2026-07-29 — the instrument found something that is not mine

The dry-run reported 3 live definitions rejected. **All three pre-existing** —
`html-developer-chunked`, `multipage-wrapper`, `html-assembler`, each with a top-level
step naming an action that is in **no registry** (`assemble_html_parts`,
`wrap_multipage`, `assemble_full_page`) and carrying no topic, so `ValidateWorkflow`
rejects them on every message.

`bugs_closed/044` already records this family as "retired code still flagged
`is_active=true`" and explicitly scopes the `is_active` hygiene half out as an owner
decision. What 044 does NOT say, and what I measured:

* three LIVE builders still point at two of them — `website-builder`,
  `landing-page-builder`, `content-site-builder`, via `spawn_agent`/`call_agent` with
  `config.agent_type` = `multipage-wrapper` / `html-assembler`;
* zero orchestration rows for those three agent types in the retention window
  (oldest row 2026-07-13, 1,952 rows total). So either those dispatch paths are never
  taken, or they are taken and die — **[UNMEASURED] which**, and the distinction
  matters.

Recorded here and in the bug I filed rather than fixed in this change: it is a
different defect (an action name that exists in no registry is undetectable until a
message arrives) and fixing it inside a validation patch is exactly the scope mistake
the guardian seat vetoes.

## 2026-07-29 — HEAD does not compile, and it is not mine

`go build ./cmd/...` fails at HEAD: `cmd/reasoningset/main.go:504: declared and not
used: planJoined, planMissing, provJoined`. Confirmed against a clean
`git archive HEAD` extract, and `git status` shows no session holding a fix in the
tree. **It does not block image builds** — `build/docker/backend/agent-chassis.dockerfile`
compiles `./cmd/agent-chassis` only — but it does break `go build ./...` and `go vet
./...` for every session, which is the check we all use to answer "did I break HEAD?".
Left alone (another session's in-flight tool, forward-only, not mine to rewrite) and
reported to the owner. My own HEAD check was scoped to `./platform/... ./cmd/config-key-audit`
accordingly, and that is stated rather than implied.

## 2026-07-29 — submitted, and committing rather than waiting

Council submission corr `9194bc97-8475-4022-b658-2ac64f06dd63`. Committed the code
immediately without a trailer, per the standing practice: waiting for a verdict does
not protect anything on a shared tree — it just exposes finished work to the next
`git add -A` while forfeiting attribution. Trailer is earned by APPROVED only.
