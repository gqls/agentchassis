# HANDOFF — RFC_012 lane · **START HERE** · written 2026-08-09

**Supersedes `HANDOFF_2026-08-08_continue_here.md`.** Go there only for the pod-grep recipe for the
provenance hardening and the two threads that are not this lane's (they are unchanged and repeated
at the bottom). Its §1a brief is **superseded and partly WRONG** — see "the correction" below,
which is the single most useful thing on this page.

Then: `RUNBOOK_rfc012_await_findings.md` (every command with its gotcha; the
`agent_error_log` date-bucketing section and the ladder mutation matrix are new) and
`NOTES_rfc012_await_findings.md` (the missteps — there are sixteen, and 14/15/16 are today's).

---

## THE LANE HAS NO OUTSTANDING WORK. What is left is one OWNER DECISION and one MEASUREMENT.

| piece | id | state |
|---|---|---|
| Three owner rulings (RFC_012 second sitting) | `3851e90b5` | **ALL DELIVERED** |
| B core + 18 conversions (`agenterrors`) | `5f49b4cfd`, `f930de86b` | **LIVE** |
| (d) detector, offline + the CronJob | `abf5e8266`, `867037f5a`, `22ed9aa04` | **LIVE AND PROVEN** |
| (a)/(a′) reader census | `40992cbce` | **DELIVERED** |
| Provenance hardening (merge split + step_name symmetry) | `f993554f6`, `0dc2d71a2` | **LIVE on v1.0.1268** |
| **§1a — the `ResolvedAgentType` ladder** | `1bc08d1ce` | **BUILT, NOT ROLLED. Council REJECTED on scope. OWNER RULED 2026-08-09: it ships (RFC_019 §11).** |
| RFC_019 + numbering-ledger restoration | `44f950522` | **DECIDED 2026-08-09 — see Open Item 1 below** |

**Nothing is in flight. Nothing is broken. Do not start coding — read the two open items first.**

---

## OPEN ITEM 1 — an OWNER DECISION ~~, and it is not answerable with more evidence~~

> **DECIDED 2026-08-09 — the owner ruled for the SHARED LADDER** (*"I think shared code
> wins this one"*), and directed the residual problems fixed in the same message. The full
> ruling and its four consequences: `RFC_019` **§11**. In short: `1bc08d1ce` stands; the §7
> resumed-step backfill is commissioned as its own council round immediately rather than
> waiting on the post-roll measurement; the `PROCESS` trigger wording is amended to "adds,
> changes or removes"; the post-roll measurement (Open Item 2) is unchanged and still due
> after the next roll. The text below is kept as the decision record it was.

§1a shipped `(*types.ExecutionContext).ResolvedAgentType()` — one ladder for "which agent is
running", called by both `coordinator.determineOwnerAgentType` and `actions.runningStepProvenance`,
which had been answering the same question two different ways.

**Council round 1 (`6186ab10-a006-4c34-b9ea-ecedfde8ea2d`): REJECTED, hard veto from `guardian`.**
Ten of twelve seats approved. `architecture` returned `ARCHITECTURE_SIGNAL: needs_rfc`.

**The seats contradict each other, and that is the whole decision:**

- `guardian` (2×HIGH, 1×MEDIUM) wants the contained fix — *"duplicate the 2-line
  `RunAgentType`/`Sender.AgentType` read locally inside `actions/log_action_error.go` … leave
  `coordinator.go` and `types.ExecutionContext` untouched entirely."*
- `architecture`, same round, unprompted — *"A contained non-hoist fix … **would have re-created
  the drift risk the author is trying to close** … a THIRD site would have been next … I'd rather
  see this land than not."* `reuse_agent` and `constitution` land there independently.

So the guardian's safest fix is the reuse seat's founding violation. This is the `bugs_closed/124`
shape, which CLAUDE.md already names as a human's call.

**WHAT YOU MUST NOT DO:**

- ⚠ **Do not resubmit.** CLAUDE.md 2026-07-28: *"A veto on SCOPE is not answered by resubmitting
  with better measurements … especially when seats disagree with each other."* Nothing about the
  measurements was disputed — five seats praised them.
- ⚠ **Do not revert.** Forward-only, and `guardian` itself says *"that decision belongs to
  `RFC_019`, not to this gate"*, flagging that the gate *"should not pre-empt it"*.
- ⚠ **Never write `Council-Reviewed: 6186ab10…`.** It is a rejected correlation. `1bc08d1ce`
  carries `Council-Submitted:`, which asserts nothing; `098` simply never credits it, which is the
  honest outcome.

**Full record with both sides quoted: `architecture_review/RFC_019_one_ladder_for_which_agent_is_running.md` §10.**
That section also corrects my own §8, which argued this was gate scope and named the wrong
disconfirmer. And it raises one wording question for the owner and **deliberately does not fix
it**: `PROCESS_architecture_review.md`'s trigger says *"changes or removes an exported symbol"*
while two seats applied it to an **addition**. Their reading governs; amending the trigger test I
was just caught by is not mine to do.

## OPEN ITEM 2 — one MEASUREMENT, and it can only be taken AFTER the next roll

§1a may be a **partial no-op on resumed steps**, and this is declared, not discovered.
`RunAgentType` reaches the actions door on the first-step / same-message path (pinned by a
round-trip test). On a step resumed after an await, `execCtx` is rebuilt from the *response's*
headers, and `ensureFullExecutionContext` (`coordinator.go:1589`) backfills `Sender` from
`state.OwnerAgentType` **only when `Sender` is empty** — and never backfills `RunAgentType`.

It cannot be settled by archaeology: `orchestration_states` keeps ~24h and every affected row is
weeks old (see the landmine). So, after the next chassis roll:

```sql
-- the claim. Baseline: 36 rows in the 13 days before the roll.
SELECT count(*) FROM agent_error_log
WHERE agent_type = 'generic'
  AND action IN ('diagnose_council_decide','retract_page_deployment','emit_tool_cross_link_items')
  AND occurred_at > '<roll time>';
-- the positive control, WITHOUT WHICH THE ZERO MEANS NOTHING
SELECT count(*), count(DISTINCT agent_type) FROM agent_error_log WHERE occurred_at > '<roll time>';
```

Plus the binary check, one Running pod **per distinct image tag** (the fleet is routinely mid-roll):
POS `provenance from the dispatch sender rather than the resolved run agent` = 1,
`whose workflow is this context executing` = 1, NEG a phrase in no version = 0.
⚠ **never anchor the needle** — `grep -c "^ResolvedAgentType$"` reads 0 on a binary that carries
it (the linker packs constants).

**If the residue does NOT fall**, the cause is the resumed-step gap and the remedy is one `if` in
`ensureFullExecutionContext` backfilling `RunAgentType` from `state.OwnerAgentType`. It is
deliberately not taken now: it adds a rung sourced from durable state, and this lane's round 2 was
REJECTED for bundling. `bug_historian` notes the gap has a precedent title — **`bugs_open/093`**,
one guarded call site with the sibling path unchecked.

---

## THE CORRECTION — read this before quoting any figure from the last handoff

Its §1a brief sized the defect at *"`generic` = **559 rows across 25 distinct `step_name`s**"* and
*"all **25** live `REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows carry `generic`"*. Both were correctly
measured and both are **misleading as statements about live damage**:

- **499 of the 555 predate 2026-07-26** — the day `RunAgentType` itself shipped (`baf887a8e`). The
  coordinator's own ladder had already removed ~89% of them.
- The dominant producer `call_agent`/`call_dispatch` (394 rows) **stops dead on 2026-07-25**.
- All 25 `REVIEW_SUPERSEDED` rows were written on **one day, 2026-07-23**; that action has not
  filed one since.
- Live residue reachable by §1a: **~36 rows in 13 days.**

**The class, because it will bite you on any table with retention:** a month-wide `count(*)` prices
a FIXED defect exactly like a raging one, and there is no tell. `[MEASURED]` is satisfied in full
by a figure that could never have come out otherwise. Bucket by date — `min`/`max(occurred_at)` per
group — and split at the date of the commit that could have changed it. Two landmines filed
(the count, and the `orchestration_states` join that also cannot see it); `WRONG_CALLS.md`
2026-08-09; `RUNBOOK` has the queries.

---

## Things that were true yesterday and are still true

- **`LANDMINES.md` is the highest-collision file in the repo.** Append and commit it **alone**,
  immediately; then `git show <sha> --numstat -- <path>` **and** grep the sha for text you added.
  A non-zero numstat is not evidence.
- **My tool channel rewrites plain ASCII.** Finish every file with
  `grep -o -P '[^\x00-\x7F]' <file> | sort | uniq -c` and account for each character, plus
  `grep -n -P '[\x{2018}\x{2019}\x{201C}\x{201D}]' <file>`.
- **A fresh landmine's verifier verdict carries no information** (stale code index). Settle it
  yourself and write the disposition in.
- **NEW: the council can be much faster than the ~29 minutes CLAUDE.md budgets.** This round was
  dispatched 22:45Z and decided **22:52Z**. That 29 is a measurement under load from 2026-07-20,
  not a constant — check before you go away.
- **NEW: two `097` submission-schema refusals**, both clean and client-side, one round trip each:
  `.plan.edits[].operation` must be `modify|add|remove|config_change` (a new file is **`add`**, not
  `create`), and `.plan.risks` must be a **STRING** while `.plan.grounded_in` beside it is an
  **array**.
- **NOT MINE, do not fix it here:** `go test ./platform/orchestration/actions/` has one failing
  test at committed HEAD — `TestValidDocSubjectTypes_LockstepWithMigrationCheck`, another lane
  adding a `decision` doc `subject_type` without moving `validDocSubjectTypes`. Confirmed
  pre-existing against a clean `git archive HEAD` tree. ⚠ and use the test's **real name** with
  `-run`: I first typed `TestDocSubject`, matched nothing, got `ok`, and briefly believed it was
  mine.

## Other threads, both explicitly NOT this lane's (unchanged from 08-08)

- **The hero/logo silent breaks.** 090 came back UNVERIFIABLE (`dce40cf4`) — the evidence would not
  grow, *not* that the premise is false. `hero_deployed` appears in 0 of 1,667 retained
  `orchestration_states` rows, so it cannot be observed after the fact. **The cheapest decisive
  test is a CANARY**: run a page build on `pageflow-builder` or `site-work-orchestrator` and read
  `collected_data->'hero_deployed'` while it is in flight. Detail in the 08-06 handoff §2.
- **`search_results.results.0.url` can never resolve** — `vet-practice-verifier`'s
  `fallback_url_field` uses an array index and `ExtractNestedField` does map access only. That
  fallback has never fired. Belongs to the vet lane.
- **Dead config keys survive indefinitely** — `commit_from` configured in 6 agents, read by
  nothing; 4 HITL `output_format` templates never rendered. No drift check notices.

## Milestone read-outs
`SUMMARY_2026-08-06_two_of_three_built.md` · `SUMMARY_2026-08-08_shipped_proven_and_the_survey_done.md`
· `SUMMARY_2026-08-08b_all_three_delivered.md` · `SUMMARY_2026-08-08c_the_seam_is_strict_now.md` ·
**`SUMMARY_2026-08-09_the_last_job_and_the_reason_that_was_wrong.md`** — the current one. Each is a
NEW file; never edit an earlier one.
