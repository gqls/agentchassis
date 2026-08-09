# PLAN 2026-08-06 — executing RFC_012's second-sitting rulings

Owner rulings (recorded in the RFC, commit `3851e90b5`), three pieces, this lane owns all:

1. **(d) → a STANDING check, online if the framework allows.** Bound by addendum 1
   (13-key routing graph, different-action discriminator; both naive detectors return 0
   on the known bug — a candidate must prove itself against that case). Vehicle to be
   chosen against the RFC_006 precedent: live config is guarded by a scheduled check,
   never a commit-time hook.
2. **The §3(a) reader census — COMMISSIONED.** Deliverable: an artefact in
   `architecture_review/` naming every reader of an awaited step's key (config side from
   live `agent_definitions`; Go side mechanically-findable readers, honestly bounded)
   with a per-reader verdict under merge-not-replace. This is what (a)/(a′) stays gated
   behind; the census does not decide (a).
3. **Option B implementation — ASSIGNED here, now.** A leaf-package shared
   `agent_error_log` writer importable from `actions` (retiring the hand-copied INSERT
   column lists) + a named findings-plus-await helper generalising 098 debt-5b's proven
   direct-DB-row pattern. Through the council gate; concept-register entry in the same
   commit as the seam.

## Decisions and their reasons

- Research delegated to three parallel agents (helper ground truth; online-check
  machinery precedents; the census itself) — this session's context is deep, and the
  census in particular is exactly the shape an agent can run end-to-end with the method
  written into the artefact for re-running.
- Order: census lands independently (its own commit); helper + check are code and go
  through one council round as a coherent task (they share the RFC and the register
  category) unless research shows they are better split.

## Verification (to be completed as research lands)

- Helper: unit tests + the debt-5b shape re-expressed through it; `go build ./...` on
  `git archive HEAD`; the import-cycle claim proven by the build itself.
- Standing check: MUST detect the known-bug case addendum 1 names (the disconfirmable
  half); a no-finding run on today's fleet is only meaningful beside that.
- Census: queries verbatim in the artefact so it can be re-run; totals stated;
  [UNVERIFIED] marked where a grep cannot see.

---

## 2026-08-09 — §1a, the last piece, and two decisions that resized it

§1a (hoist the `RunAgentType` ladder) was not in the three owner rulings above. It was carved out
of the RSH-008 hardening deliberately, because this lane's round 2 was REJECTED for bundling, and
it is now shipped as `1bc08d1ce` / `RSH-009` / `RFC_019`.

**DECISION 1 — the `os.Getenv("AGENT_TYPE")` rung stays coordinator-side; the shared method stops
at the context.** The 08-08 handoff called this "the design decision of the job" and left it open.
Reasons, in the order they weigh: `AGENT_TYPE` is a property of the process and this type is
deserialised from Kafka headers, so a method reading it would answer differently per pod,
invisibly at the call site; and the two consumers' tails *legitimately differ* — the actions door's
floor (`params.AgentType` = `state.OwnerAgentType`) is strictly more specific than the pod, and it
must never reach for the `"generic"` filler because that is the value RSH-008 chose `unattributed`
to avoid colliding with. `ResolvedAgentType(fallback string)` was considered and rejected: it
looks like one ladder while two callers pass different fallbacks. `t.Setenv` pins the exclusion so
the decision is enforced rather than documented.

**CORRECTION to the brief this work was commissioned on, recorded here rather than silently
absorbed.** The 08-08 handoff sized §1a at "**559** rows across 25 distinct `step_name`s" plus 25
`REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows. Measured before building: **499 of 555 predate
2026-07-26**, the day `RunAgentType` shipped; the 25 rows are all from **one day**, 2026-07-23.
Live residue: **~36 rows in 13 days**. The change is justified structurally — one question, two
ladders, in packages that cannot import each other — and **not** by volume. `RFC_019` §3 therefore
lists "do nothing" as a serious option. The class of error (a retention-bounded table prices a
fixed defect identically to a live one) is filed in `LANDMINES.md` and `WRONG_CALLS.md`.

**DECISION 2 — the change went to BOTH forums, and the gate REJECTED it on scope.** RSH-008's
round-1 `architecture` seat had ruled that change `point_fix` explicitly because it *"stays inside
`platform/orchestration/actions`"*; this one does not, so the precedent was not claimed. Round 1
(corr `6186ab10-a006-4c34-b9ea-ecedfde8ea2d`): **REJECTED**, hard veto from `guardian`, with
`architecture` returning `ARCHITECTURE_SIGNAL: needs_rfc` — so my own §8 argument that this was
gate scope was wrong, and is corrected in `RFC_019` §10.

**It is deliberately NOT resubmitted and NOT reverted.** The guardian's contained alternative
(duplicate the two-line read locally, leave `types` and `coordinator.go` alone) is the second
ladder this change exists to retire, and the `architecture`, `reuse_agent` and `constitution`
seats say so in the same round. Per CLAUDE.md 2026-07-28 a scope veto is not answered by
resubmitting with better measurements, especially when seats disagree with each other; the
guardian itself routes the call to `RFC_019`. **The open item is an owner decision, not a task.**

**Still open, and it is a measurement not an argument:** whether §1a is a partial no-op on resumed
steps (`ensureFullExecutionContext` never backfills `RunAgentType`). Undecidable before the roll —
`orchestration_states` keeps ~24h and every affected row is weeks old. The post-roll query and its
36-row baseline are in `RSH-009`'s `verify-later`.

---

## 2026-08-09 (later) — OWNER RULING on Decision 2, and the residuals commissioned

**The owner ruled: the shared ladder ships** (*"I think shared code wins this one"*) — recorded in
full at `RFC_019` §11, with the handoff's Open Item 1 marked decided. The same message directed
*"please go ahead and fix all those other problems"*, which this lane reads as the residual
problems its handoffs recorded. Five pieces, dispatched 2026-08-09 (planning in this session,
implementation delegated per problem):

1. **The §7 resumed-step backfill** — no longer waiting on the post-roll measurement; one `if` in
   `ensureFullExecutionContext` backfilling `RunAgentType` from `state.OwnerAgentType`, its own
   council round as promised in RFC_019 §3's last row. The post-roll measurement stays as the
   acceptance evidence for the pair.
2. **`validDocSubjectTypes` lockstep** — HEAD's failing test (migration 340 added `decision` to
   the doc_notes CHECK without moving the Go slice; RFC_015 lane, `e1628f7df`). Add `decision`.
3. **`ExtractNestedField` array-index support** — makes vet-practice-verifier's
   `fallback_url_field: search_results.results.0.url` resolvable at last (census §6.1), plus the
   fallback's silent-failure logging. Additive, opt-in, council-gate scope per OWNER RULING
   2026-07-29 §1. NOTE the census wording is off by one segment: the walk dies at `0` (array), not
   at `results` (which resolves via the `.response` unwrap).
4. **Dead config keys retired** — `commit_from` (6 agents) and the HITL `output_format` map
   (4 templates in ONE agent, `simple-content-writer-with-approval` — the handoff's "4 agents" was
   wrong); plus `ActionInputSpec` opt-ins for `update_page_status` and `process_approval_decision`
   so the class is detected next time (the bugfix-136 mechanism, 67 actions already in). Migration
   352.
5. **The hero/logo canary** — measurement only, one dispatch, exactly as the 08-06 handoff §2
   prescribed. Any fix routes back through a decision because it may land in (a)/(a') territory.

Also executed under the same ruling: `PROCESS_architecture_review.md`'s trigger now reads "adds,
changes or removes" (§10's withheld one-liner, owner-sanctioned).

### Outcome, same day — all five delivered, and four corrections came back with them

1. **§7 backfill** — `58aefe282`, council `b0deddf7` **APPROVED** round 1. 2. **`decision` lockstep**
— `5019bf2b7`, `c88c0a84` **APPROVED**. 3. **Array-index + fallback logging** — `f7111f4d8`
(+`6cb41ae06`), `c961b79e` **APPROVED**; registered as **WFA-012**. 4. **Dead config keys** —
`96f8075fb`, `501ef561` pending; **migration 356, NOT APPLIED**. 5. **Hero/logo** — no canary
needed, see below.

**A sixth, unplanned:** `TestEveryActionInputSpecHasARegistryEntry` had been red at HEAD since the
bugfix-136 probes landed — and red in a *particularly* misleading way, passing under `-run` on its
own name while failing whenever the package ran, because the spec registry is process-global with
no removal path. Fixed at the assumption (`a6c0498f2`): a registered spec is not necessarily an
action. **All four packages are now green at a clean `git archive HEAD` tree.**

**The corrections are the valuable part, and every one of them contradicted this lane's own brief:**

- **The dead-key census was WRONG in three ways, all found by looking rather than by trusting.**
  Three of the six `commit_from` steps are nested inside a loop `sub_workflow`, so the obvious
  top-level census sees **3, not 6** — this file's own earlier figure came from the depth-aware
  walk and survived; a naive re-check would have "corrected" it downwards. There were **two more
  dead keys nobody had named** (`content-reviewer`'s `notes_field`, `validation_issues_field`),
  left standing as declared true positives per the `create_work_item`/`spec` precedent. And a
  **seventh seed file** (`034_page_rerender_agent.sql`) seeds one of the six live agents, while
  `033` — the one the brief named — seeds a *different* agent whose live row has already drifted
  clean.
- **The live HITL step spells its action `process_data`, not `process_approval_decision`.**
  Registering only the canonical name would have opted in **zero** live steps. The deprecated
  alias registration is load-bearing, not tidiness — the kind of thing that would have shipped
  looking exactly like a working fix.
- **Migration numbering raced four times in one afternoon.** 352 was taken before the work
  started; 353, 354 and 355 were claimed by other sessions *during* it. Landed at **356**.
- **The census correction the lane already knew about bit again in a new place:** the array-index
  blast-radius query reads live config *text*, and the `guardian` seat pointed out it cannot see a
  path Go builds at runtime. Checked: seven such sites, all appending a field name rather than an
  index. Filed as a landmine, because the clean answer is exactly what hides the blind spot.

**Hero/logo (§5) was overtaken by evidence and is now `bugs_open/236`.** The keys appear in live
rows at last, a *successful* logo response merged without `image_url`, and the three readers are
silent by construction. **The obvious mechanism is REFUTED by the merge code** (it preserves rather
than replaces), so the root cause is deliberately not asserted; a `090` re-run came back
UNVERIFIABLE on **harness** grounds and surfaced a standing blind spot worth more than the bug —
`orchestration_states` is absent from the diagnosis bundle's schema section, so the loop cannot
address the platform's central state table. My own premature "confirmed end to end" is logged in
`WRONG_CALLS.md`.

**Still owed, and neither is this lane's to take:** the post-roll measurement (both ladder halves
inert until a chassis roll), and migration 356's application. 236's real fix may land in the
`(a)`/`(a′)` merge design, which remains the owner's decision.
