# HANDOFF — clearing the `/bugs_open/` backlog by hand (2026-07-21)

**Cold-start for a new thread continuing this work.** Read this top to bottom;
it is self-contained. Origin: the session began as "experience loop 2" and
pivoted on the owner's directive — *"lets start fixing the bugs so we don't keep
being tempted to write workarounds."* The experience-loop decision that was
open when it pivoted is still open; see § "The origin thread" at the bottom.

---

## 1. What is now LIVE in production (v1.0.1144, pod-verified 2026-07-21)

Image `v1.0.1144` rolled and was verified against the **running pod's binary**,
not git and not the tag (`agent-chassis-59c675c4f-pxr9f`). All four grep to a
non-zero count in `/app/agent-chassis`:

| bug | what shipped | marker string | state |
|---|---|---|---|
| `008` item 5 | `RefusalError` — `stop_reason=refusal` decoded, names the cause instead of "no text content (had 1 blocks)"; provider-parity CI guard | `model declined to answer` | **CLOSED → `/bugs_closed/`** |
| `013` | `formatGeneratedGo` runs `go/format` at commit-prep so trivia stops burning implementer runs; unparseable body still fails loud | `is not valid Go (cannot format` | **CLOSED → `/bugs_closed/`** |
| `032` | completion verifier returns an error (not `Resolved:true`) when the component row is gone — deletion no longer reads as a fix | `cannot verify: component` | **CLOSED → `/bugs_closed/`** (on its **safe floor** — the stronger "absence = deletion when the page still expects it" verdict is left to the `empty_sections_loop_integrity` thread) |
| `015` | `MissingNewsPageCheck` says *re-type an existing stranded page*, not *create a duplicate* | `RE-TYPE that page to page_type` | **STAYS OPEN** — partial (candidate 3 only); the planner still emits `section-index` (candidate 2 = the real fix, untouched) |

Commits (all ancestors of HEAD): `45e90acbb` (008), `a467baa11` (032),
`fc38c6058` (013), `48838c090` (015), plus doc commits `867049c04`, `53ee1f95f`,
`ed1e20602`, and 015's live-status note.

**Verify command pattern** (use this, never git/tag, for any "is it live" claim):
```
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "<marker>"'
```

## 2. The method that actually worked — this is the transferable part

Every one of these earned its place by catching a real error **this session**:

1. **Verify against the committed ref, never the shared tree.** A delegated
   triage read the working tree, found *my own uncommitted* 008 fix, and reported
   008 as "already shipped in `f32b208e5`" — a commit containing neither the code
   nor the test. In a many-session tree, "check current code" silently means
   "check everyone's WIP". Claims about the codebase → `git show HEAD:<path>`.
   Claims about prod → grep the pod. A verdict with no named ref is not a
   verification. Filed as a pattern in **016b §9** ("Verify against the current
   code silently means against every session's uncommitted work").

2. **Ground blast-radius against the live DB before asserting it.** The triage
   called `015` "fleet-wide, non-English sites worst". Live query: **one** site
   qualifies. And my *first* detection predicate (copied from the case file's own
   diagnostic query) false-positived on **six deployed, working** tool/blog pages
   on ai-agent-orchestration.com — it would have invited the planner to re-type a
   live tool page into a news index. `build_status <> 'deployed'` fixed it; a
   test regex pins the clause. **An empty `sections` array only means "stranded"
   when the page never built.** The case file's own query carries the same false
   positive — noted in-file.

3. **Structural detection, not name heuristics.** `015` detects the stranded
   page by nav-visible + sectionless + not-deployed, never by matching
   "news"/"noticias". A vocabulary list is the `bugs_open/044` shape and fails
   worst on the non-English sites the bug hurts most.

4. **Change the field the consumer actually reads.** `015`'s handler is an LLM
   (`content-gap-planner`); the intervention is its natural-language
   `description`/`suggestion`. A new structured key *alone* is the dead-config
   shape of `bugs_open/025` and `/042` — written by one side, read by nobody.

5. **A fix can close on a safe floor with a residual left open** — say so in the
   file and the index, or the floor reads as the finished shape (`032`).

## 3. The collision reality — READ before picking a target

Many "open" cases are already fixed or actively in-flight. This session reached
`022`, `042`, `036`, and the whole `003`/`029`/`030` dispatch family only to find
them owned or done. **"Has a fix commit" ≠ closeable** — several are partial
(`003 F1` of F1–F4, `021 §2`, `023 P2.2/P2.3`). On 2026-07-20 roughly **13 bugs
were filed while ~7 closed**; discovery outpaces fixing, so working the list to
zero is not the goal. Backlog now: **46 open, 18 closed**.

**Before touching any target, run BOTH checks** (this is not optional — it cost
real runs when skipped):
```
git status --short <the file(s) the fix touches>     # empty = no session has it open
git log --oneline -3 -- 'bugs_open/NNN*'             # someone may have a fix commit already
```
And check the diagnosis queue for in-flight coverage:
```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT summary,status FROM site_work_items WHERE item_type='needs_diagnosis' AND status='awaiting_diagnosis';"
```

## 4. Live issues the next thread should know

- **`discovery_checks` test package is RED at HEAD** — `check_contact_form_undeliverable.go`
  (another session's) produces item_type `contact_form_undeliverable` with **no
  registered verifier**. The `021 §2` coverage guard (`verifier_coverage_test.go`)
  is correctly failing the build for it — this is the guard's first real catch,
  working as designed. The image still builds (`go build`, not `go test`), which
  is why v1.0.1144 shipped. **Someone must either register a verifier for it or
  add it to `itemTypesWithoutVerifiers` with a reason.** Not this session's file.
- **`035`/`036` council type-slip** (a reviewer emitting `"3"` instead of `3`
  voids a whole round): the fix lives in `diagnose_council_decide_action.go`,
  which another session has been extending for the `019` truncation path — it adds
  a `Degraded` flag. **Extend that flag to the schema-failure case; do not build a
  second degraded-round path beside it.** A coordination note is appended to the
  036 case file. `036` already has a round-2 fix commit (`ab158c32a`) — check its
  state before starting.
- **`015` stays open** — candidate 2 (teach the planner to emit `news-index` for
  a news listing) is the real fix and is untouched.

## 5. Good next targets (unowned as of this writing — RE-CHECK per §3)

These looked tractable and unclaimed, but the tree moves fast:

- **`025`** (`content_direction` dead column) — TRIVIAL, self-evidencing, docs+SQL.
- **`028`-avoid-lists** (`banana/provider.go:105` drops `NegativePrompt`) and
  **`027`-content_hero §4b** (palette truncated by the 200-char cap) — the triage
  flagged these as ONE defect stated twice: *a structured brand instruction
  silently discarded between style guide and model.* Fix in one pass or `028`'s
  candidate truncates under `027`'s cap. Fleet-wide imagery.
- **`033`** (human-review queue has no working surface — 292 items, 0 actioned;
  `HandleConfirmWorkItem` written but never registered) — **blocked on an owner
  decision**, do not start unilaterally.
- Avoid the LARGE ones as a batch item: `021` (verifier coverage across ~50 item
  types), `023` (CTA label/URL pairing), `003`/`029`/`030` (spawn-dispatch, owned).

## 6. Standing constraints (from CLAUDE.md — non-negotiable)

- **Commit per task, explicit pathspec on `git commit <paths>`.** Never
  `git add -A/./*`, never `git commit -a`. New files: `git add` first, then name
  again on commit. Read the yellow commit-scope block — anything listed that
  isn't yours belongs to another session.
- **Forward-only** — no resets, amends, rebases.
- **Go changes are inert until an image rolls; config/SQL is live immediately.**
- **Verify deploys against the running pod's binary**, never git, never the tag.
- **Image first, then seeds. Bump `IMAGE_TAG` every build.** No orchestration
  dispatch within ~300s of a chassis (re)start.
- **Grep BOTH `/bugs_open/` and `/bugs_closed/`** (and 016b §10) before filing —
  numbering is one sequence across both dirs, and `016`/`017`/`018`/`027`/`028`/
  `029`/`040` are each two different cases; **resolve by slug, never bare number.**
- Every `.go` fix: `gofmt -l` clean + a regression test (`sqlmock` and
  `httptest` are already deps; `go/parser` for structural guards).

## 7. The origin thread — experience loop (decision still pending)

The session pivoted away from this; it was NOT resolved. Docs in
`docs/agent_docs/docs024_key_docs_latest/experience_loop/` (RUNNING_NOTES,
RUNBOOK §8a, README_where_we_are, SUMMARY). State at pivot:

- CP2 closed (run 8, first unanimously-approved EXPERIENCE_PLAN) but that plan
  carries contract debt the 4-seat council never saw.
- The `review_contracts` seat (migration 174) finds real defects every round but
  **blocks every greenfield step** with an over-strict rule: *"a pair you cannot
  verify from context is itself an objection"* — correct for an EXISTING consumer,
  wrong for one the plan will CREATE. Runs 9/10/11 did not converge; ~75 min +
  credits spent.
- **Pending owner decision:** split the rule (existing-contradicts-plan = hard
  objection; plan-creates-it = check the plan specifies the access path + has an
  acceptance criterion), run once more, and if it still can't converge take the
  objections as a hand-worked to-do list rather than buying more rounds.
- Note `008`'s refusal fix (now LIVE) closes the exact mechanism that killed an
  experience-planner council round on 2026-07-18 — combined with migration 171 a
  refusing seat now degrades to abstention instead of voiding the round.

---

*This is a HANDOFF, not a standing-five workstream — the bug backlog is shared
ongoing work, not a bounded project, so one self-contained doc is the right
artifact. The durable learnings live in 016b §9/§10 and the individual case
files; this doc points at them.*
