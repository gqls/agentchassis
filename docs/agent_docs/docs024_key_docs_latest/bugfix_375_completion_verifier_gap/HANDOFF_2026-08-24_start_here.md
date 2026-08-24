# HANDOFF — start here: `bugs_open/375`, the unguarded completion writer

**Written 2026-08-24 ~19:00Z** by a session that was handed the *closed* `bugfix_367_router_remit`
lane, checked what was actually left, and did the one piece of groundwork `375` said it owed.
**This lane has no prior session. You are the first — claim it.**

> **Your task, in one line:** `update_work_item_status` can stamp `complete` without ever consulting
> the verifier framework. The blast radius is now MEASURED (it was not when the bug was filed), and
> the measurement changes which fix to choose. Read §3 before §4.

---

## 1. Why you are taking 375 and NOT 333

Both were named as candidates. The answer is 375, and the reason is ownership, not interest.

**`bugs_open/333` is OWNED and moving** — `bugfix_333_owned_page_door`, six commits in five days,
the most recent this afternoon. Its fix is committed, council-**approved** (round 2, `9813dec8`,
`complete_approved` 2026-08-24 17:01:54Z), and **live in the running chassis since 18:32:19Z
today**. `scripts/who-owns.py 333` says OWNED. CLAUDE.md's rule is explicit: *"If it says OWNED:
contribute into the bug file, do not compete."* I contributed the state change into the bug file
(commit `839242d82`) and stopped there. **So should you.**

**`bugs_open/375` is genuinely UNOWNED.** `who-owns.py 375` prints "OWNED or recently active", and
that is a **false positive you should not be fooled by**: the only workstream it names is
`bugfix_367_router_remit`, which *filed* 375 and has since closed. Nobody is working it.

### What 333 did NOT do for you, since 367's handoff pointed at it

367's residual list said its population "can only ever park" until 333 is fixed. **333 is now live
and that is still true.** 333's own fix section says it plainly: *"It does not repair owned pages.
It converts 'silently refused and forgotten' into 'visibly parked, with its reason and the route
that works'."* The repair half is `bugs_open/277`'s question and the owner ruled 2026-08-19 that the
two must not be merged. **Do not let anyone tell you 333 unblocked repair. It unblocked legibility.**

---

## 2. Verified state as of 2026-08-24 19:00Z — with how each was checked

| thing | state | how it was established |
|---|---|---|
| `bugs_open/375` | **OPEN, UNOWNED** | in `bugs_open/`, one commit ever (`fe0df0cd0`, the filing) |
| its core claim | **HOLDS** | `grep -rn 'GetVerifier('` over `platform/ internal/` → exactly **2** hits: the definition, and **one** caller, `complete_work_item_verification.go:122`. None in `UpdateWorkItemStatusAction`. |
| registered verifiers | **13** as of 2026-08-24 | `grep -rhn 'RegisterVerifier\(WithPolicy\)\?('`, `_test` excluded |
| the blast radius | **4 agents, 6 `complete` arms, 5 item types, 134 completions** | §3 below; live DB, 2026-08-24 18:49Z |
| verifiers actually bypassed | **ZERO** | §3; controlled — see the positive control there |
| `bugs_open/333` | **OPEN, owned, fix APPROVED and LIVE** | council row `complete_approved` 17:01:54Z; chassis `v1.0.1335` from 18:32:19Z; door literals in `/proc/1/exe` with a must-be-absent control |
| 333's door observed firing | **NO — and it is not a failure** | 0 parked rows, but the **demand control is also 0**: only 2 work items exist fleet-wide since the roll |
| chassis | `v1.0.1335`, both pods, started 18:32:19Z | `kubectl get pods -l app=agent-chassis` |

⚠ **Every row above is a snapshot on a tree many sessions share. Re-run before you act on it.**
The one that will go stale first is the last: another roll changes the chassis out from under you.

---

## 3. The census — the single most useful thing I can hand you

`375` §3 carried a ⚠ saying *"a count is owed here and this file does not have it … Do not size this
from `bugs_open/367` alone."* **I ran it. The result is in the bug file as new §3a (commit
`b1a385b9b`), and it changes the answer.**

**Four agents reach `complete` through the unguarded writer**, out of **200** live agent
definitions: `image-build-handler`, `image-source-unsatisfiable-handler`, `image-url-404-handler`,
and `required-fields-missing-handler` (367's router, 3 arms). Two other agents name the action but
use it only for `failed` / `needs_human_review` — statuses a verifier has no opinion about.

**Those four handle exactly five item types, and NONE of the five has a registered verifier.**
So **no verifier is being bypassed today**. The defect is **latent, not active**. 134 completions
all-history took the unguarded path.

**Why that is not a reason to close it.** `verifier_coverage_test.go` maintains the list of
unverified types with a reason each, and says of one category *in its own words*: **"these SHOULD
get verifiers — this is the actionable backlog, not an excuse list."** Two of your five sit in that
category. **The framework is inviting somebody to write the verifier this action will silently
ignore.** They will register it, the coverage test will go green, and nothing will be protected.
`CQ-023` warns that the `required_fields_missing` one *would fail-close the `converted` arm* — so
the first person to work that backlog meets a trap from both sides at once.

**That is the bug. Not "a writer lacks a guard" — "a trap is set for the next person, by name."**
Frame your submission that way; it is both truer and more persuasive than the filing framing.

### What it does to the fix candidates in §4 of the bug file

- **Candidate 1 (opt-in verifier consult, unsafe default OFF) is now the recommended start**, and
  the census is what unblocks it. **`RFC_022`'s narrowing (owner ruling 2026-08-11) applies**: an
  opt-in field whose unsafe default is OFF and which **no live consumer names** is NOT
  architecture-scope. All three conditions hold, and RFC_022 demands the consumers be **enumerated,
  not asserted** — *"asserting it without the query is itself the objection"*. §3a is that
  enumeration. **Cite it.** On this footing candidate 1 goes through the normal council gate.
- **Candidate 4, which the bug file did not have** — teach the coverage guard that *"has a
  verifier"* ≠ *"is verified"* unless every completer of that type consults one. **This is the only
  candidate that protects the person the trap is set for**, and it keeps §3a's zero true rather
  than merely true-today. I would pair it with candidate 1.
- **Candidate 2 (unify the two writers) stays architecture-scope** and is the real structural fix.
  `bugs_closed/284` is the precedent (owner ruling 2026-08-17) — duplicate writers unified with a
  structural single-definition test to stop a fourth copy appearing. Bigger, slower, correct.
- **Candidate 3 (record it honestly) is now provable rather than merely honest** — it has a number.

**My recommendation, stated as a recommendation:** candidate 1 + candidate 4 together, through the
normal council gate, citing §3a for the RFC_022 conditions. Do **not** arm the field for
`required_fields_missing` in the same change — read `CQ-023` first, and make arming a separate,
per-type decision with its close paths read. **[JUDGEMENT, not measurement]** — you own this call,
and a reviewer may reasonably prefer candidate 2.

---

## 4. Commands, each with the gotcha attached

```bash
# The census, re-runnable. Re-run it before quoting §3a — types get added.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db
```
```sql
-- who reaches complete through the UNGUARDED writer
WITH live AS (SELECT type, default_config FROM agent_definitions
              WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)
SELECT l.type, s.key AS step_name, s.value->'config'->>'status' AS sets_status
FROM live l, jsonb_each(l.default_config->'workflow'->'steps') s
WHERE s.value->>'action'='update_work_item_status' ORDER BY 1,2;
```

- ⚠ **Enumerate the verifiers with BOTH registration functions.** `RegisterVerifier(` does **not**
  match `RegisterVerifierWithPolicy(`. Use
  `grep -rhn 'RegisterVerifier\(WithPolicy\)\?(' platform/ internal/ --include=*.go | grep -v _test`.
  **I got 11 with the naive grep and only caught it because the bug file named two types my own
  grep had not found.** The bug file's §3 carries the same defective grep; §3a corrects it.
- ⚠ **Control any zero you report.** A zero from a mis-spelled `IN` list is indistinguishable from
  a real one. The control that worked: run the same type list *without* the handler filter and
  require real rows back (12 of 13 came back).
- ⚠ **Ask the pod what it is running; never `strings`.** The `build provenance` startup line
  scrolls — it was already out of `--tail=300` for the chassis today, and **empty there means "not
  in range", not "unstamped"**. Fall back to
  `kubectl -n ai-persona-system exec <pod> -- grep -aq "<literal>" /proc/1/exe`, **always with a
  must-be-absent control in the same breath**.
- ⚠ **A count of things carries the date it was counted** (owner ruling 2026-08-22). Every figure in
  §3a is written `as of 2026-08-24`. Keep that up — a census does not go wrong, it goes stale by
  addition, and it reads as current for ever.

---

## 5. Traps specific to this lane

1. **`who-owns.py` will tell you 375 is owned. It is not** — see §1. The tool reads mentions, and
   the lane that mentions it is closed.
2. **A verifier that fires is not automatically an improvement.** `CQ-023` is explicit that a
   `required_fields_missing` verifier fail-closes the `converted` arm of 367's router. Registering
   one to "test the fix" would break a live route. Use a type with no live close path, or a
   fixture.
3. **A mock's own bookkeeping cannot assert this negative.** The bug file says it and it is right:
   mutate the guard and require the test to fail, or you have proven nothing
   (`LANDMINES.md`, *"a mock's own bookkeeping cannot assert a NEGATIVE"*). Note the sibling trap —
   *"a mutation that PASSES may have hit a guard in SERIES"*: the terminal-decision guard already on
   this arm can mask a missing verifier consult.
4. **The bug file's own `v3_site_actions.go` line numbers have DRIFTED and will drift again.**
   Filed 08-23 as `:6010` / `:6290-6300`; the real ones on 08-24 are `:5978` / `:6260-6268` — about
   30 lines in one day on a tree taking ~1,500 commits a week. **Re-locate by symbol, never by
   line**: `grep -n 'func UpdateWorkItemStatusAction'`. I re-verified the substantive claim at the
   current lines — no `GetVerifier` anywhere in `5978..6296` — so the defect is real; only the
   coordinates were stale.
5. **Do not read `410` or any seed as the live router config.** 367's lane was burned by this:
   `410`'s comments say "v3", the live row is `version = 1`. Read `agent_definitions`.
6. **The window between appending to a shared doc and committing it belongs to everybody.** 367's
   lane had three `WRONG_CALLS` entries swept into another lane's commit in that gap. Commit
   narrowly, by pathspec, the moment your edit is coherent.
7. **Your commit is a deploy.** `make build-*` builds from committed HEAD and any session's roll
   ships it. There is no "hold it pending review" on this tree — which is exactly why review here
   is after the fact by design (owner ruling 2026-07-29 §2).

---

## 6. What is NOT established — say so, do not quietly inherit it

- **Whether any of the five types SHOULD have a verifier.** §3a establishes only that they do not,
  and that the coverage test's own categories put two of them on the actionable backlog. Whether
  writing one is right is a separate judgement with `CQ-023` attached.
- **Whether the 134 unguarded completions contain any FALSE completions.** Nobody has re-run those
  items' predicates. `bugs_open/367` found one by accident; that is one, not a rate. **Do not
  extrapolate.** If you want the number, that is its own measurement and probably its own `090`.
- ~~**Anything about `image-url-404-handler`.**~~ **CHECKED before handing over, and it turned up
  something that is NOT yours** `[MEASURED 2026-08-24 19:00Z]`: that handler has handled **0 rows,
  ever**, which is why it is absent from the five. But its item type exists and is stuck — 42
  `image_url_404` rows (**38 `detected`**, 3 cancelled, 1 complete) all carry an **empty
  `handler_agent`**, so nothing ever routes them to the handler built for them. That is an
  undispatched population, a different defect from this one, and **[OBSERVED, NOT DIAGNOSED]** —
  I did not look for the cause. Note it, do not absorb it: 375 is about a guard, not about
  routing. If you file it, grep `bugs_open/` and `bugs_closed/` first and check
  `scripts/who-owns.py`.
- **Whether candidate 2 is feasible.** I did not read `CompleteWorkItemAction`'s signature or its
  call sites. `bugs_closed/284` is cited as precedent by shape, not by verified similarity.

---

## 7. What to do first, in order

1. **Claim the bug** — edit `375`'s `**Status: OPEN, UNOWNED.**` line to name this directory, and
   commit that alone. It is how the next session's `who-owns.py` stops lying.
2. **Start the standing five in this directory** (`PLAN_`, `RUNBOOK_`, `NOTES_`,
   `README_where_we_are.md`). Not at handoff time — at the start. This handoff is not one of them.
3. **Re-run the §3a census yourself** before you build on it. It is 18 hours stale the moment you
   read this, and re-running it is how you inherit it honestly rather than on trust.
4. **Read `CQ-023` and `verifier_coverage_test.go`'s header** before choosing between candidates.
5. **Decide, then submit to the council gate** citing §3a for RFC_022's three conditions. Budget
   ~30 minutes, not ~2 — the council takes 2–5 but the dispatch queues behind the fleet.

---

## 8. Where everything is

| what | where |
|---|---|
| the bug | `bugs_open/375_HANDOFF_2026-08-23_update_work_item_status_completes_without_ever_consulting_the_verifier_framework.md` |
| the census | that file, **§3a** (added 2026-08-24, commit `b1a385b9b`) |
| the unguarded action | `platform/orchestration/actions/v3_site_actions.go:5978` (`UpdateWorkItemStatusAction`), `complete` arm's UPDATE at **`:6260-6268`**; registered `registry.go:939` ✅ |
| the guarded one | `platform/orchestration/actions/complete_work_item_verification.go:122` — the only `GetVerifier` caller |
| the framework | `platform/orchestration/actions/discovery_checks/verifiers.go` |
| the coverage guard | `platform/orchestration/actions/discovery_checks/verifier_coverage_test.go` (categories at `:60-95`) |
| the sibling bug | `bugs_open/333` (owned — contribute, do not compete); my state CONTRIB is at its foot |
| where 375 came from | `docs/agent_docs/docs024_key_docs_latest/bugfix_367_router_remit/` — closed, read `HANDOFF_2026-08-24_continue_here.md` §5 for traps |
| related | `bugs_open/017`, `bugs_open/213`, `bugs_open/021` §INSTANCE 2, `bugs_closed/284`, register `CQ-023`, `WII-003` |
