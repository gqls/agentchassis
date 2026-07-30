# 149 — the discovery checker layer: defect queue

> **STATUS 2026-07-30 — 3 of 12 items done and LIVE; the file STAYS OPEN.**
> **C1 ⚠ (structural half only) · B4 ✅ · C3 ✅ (corrected — it was not a defect)** —
> **C1 is NOT closed.** The floor now sits on all six persisting agents, but it would
> **not** have caught the fabrication another session witnessed live the same day —
> checked, 0 findings, see the ⚠ banner on C1. What is fixed is that no gate existed
> on four of six; what remains is what the engine can *recognise*. All in commit
> `f61dce806`. **LIVE on chassis `v1.0.1211`**, carried into another session's roll
> at 17:33 UTC and **pod-verified on BOTH replicas**: `CONTENT_CLAIMS_FLOOR_DETAIL`
> 1, `claims floor blocked` 1, `silently omitted it` 1, `checks_unregistered` 1,
> positive control `CONTENT_LINK_REPAIR_DETAIL` 1, negative control 0. Concept
> register **CLM-018**.
>
> **Council `2d0dbc2e-e125-41f6-876d-0f8d6cf96688`: round 1 REVISE (evidence), round
> 2 REVISE (HARNESS). Read `unreadable`, not `abstained`, before reading round 2 as a
> judgement.** Round 2: `decided_by = "unreadable reviewer(s): review_guardian.result,
> review_improvement_guardian.result"`, `unreadable` **0 → 2**. **No high-severity
> objection was raised in round 2 at all** (3 medium, 6 low), 12 of 15 seats approved,
> and **`prior_art_librarian` — the seat whose HIGH objection gated round 1 — flipped
> to APPROVE.** So the evidence answer worked and the second REVISE is `bugs_open/119`'s
> shape (a seat's malformed result costing a round), not a verdict on the change.
> Not resubmitted a third time: the change is already live, the council is advisory,
> and a round that was decided by two unreadable results is not evidence of anything
> to fix.
>
> **The medium objections that survive round 2, recorded as follow-ups rather than
> silently closed:**
> - `bug_historian` (edit 3): an **erroring** check is reported in the step output but
>   nothing durable records it — pod logs roll. This is the same "a pod log line is not
>   a record" rule (`bugs_open/071` gap 3) that the claims floor **does** honour with
>   `CONTENT_CLAIMS_FLOOR_DETAIL`, and B4 does not. **A real inconsistency in my own
>   change; small, and deliberately not patched in after the verdict** — adding code
>   post-review is the exact pattern the guardian seat objects to.
> - `bug_historian` (edit 1) and `architecture`: the three unguarded persistence paths.
>   Round 2 bounded two of them (`create_report_page` 2 components, `rebuild_blog_listing`
>   7, together <1% of the 949-component surface) and stated plainly that
>   `ApplySectionEditAction` **cannot** be bounded from `page_components` — it edits in
>   place and the table has no provenance column. That residual is `bugs_open/136`'s.
> - `editquality` (edit 2): the rationale should have named the consequence that a green
>   orchestration status no longer implies the sections were written. It is in
>   `LANDMINES.md`; it should have been in the submission too.
>
> Round 1 detail follows, kept because it is the more useful of the two:
> 15 seats, **12 approve / 3 object**, 2 abstained, 0 unreadable, not truncated.
> The gating objection (`prior_art_librarian`, high) and four of its five points
> were **one objection**: load-bearing absence claims asserted as live measurements
> with **no query attached** for the council to check. `debug_historian` raised the
> same point independently on `checks_run`. **Nothing was refuted** — the counts all
> stand — so round 2 answers it with the SQL and its verbatim output, and **the code
> is unchanged**. Two of the council's own verification requests had themselves
> failed with `column "agent_type" does not exist`, so the seats could not check for
> themselves even having asked the right question.
>
> **One objection is deliberately NOT answered by resubmission.** The `guardian`
> seat (medium) asks that the owners of the four newly-gated pipelines
> (`pageflow-builder`, `page-rebuild`, `page-rerender`, `site-work-orchestrator`)
> sign off on a save that can now refuse. That is a judgement about how a capability
> reached production, not an evidence gap, and the estate's rule is that a scope
> objection needs a human. **Owner: this is the decision that wants you.**
>
> **Still open: A1–A6, B1, B2, B3, C2.** Several need an owner ruling rather than
> code (B1 is seat-or-delete; A4/A5 are shared-schema and shared-mechanism changes
> wanting their own council round), which is why the queue is not closed.
>
> **Read the correction banners on C1 and C3 before using either.** Both items were
> *wrong in a way that pointed at the wrong file*, and both corrections were made
> by re-measuring rather than by reasoning:
> - **C1's seam is PERSISTENCE, not the handler** — and `page-content-writer`, the
>   agent it named, persists nothing.
> - **C3's "zero rows ever" was false 2m28s before this file was committed** — the
>   detector had already fired automatically and found two real things.
>
> The pattern across both, and across this file's earlier A1 correction: **a figure
> measured at the start of a session and written up at the end of it is not a
> measurement, it is a memory.**

**Filed 2026-07-29 at the owner's request ("there are a lot of fixes to be done with
the checkers — list them and we'll work through them"). This is a QUEUE, not a
diagnosis: every item below is measured, with the query that measured it, so the
next thread can re-run it rather than trust it.**

Everything here was found by pulling one thread — `bugs_open/146`, an unreachable
tool page — and asking why it was never linked. Nothing here is speculative; where
a cause is not yet established it is marked `[UNMEASURED]` and says so.

**Standing caution for this whole queue:** several of these checks are *correct* and
the defect is in the seam between them and their handler. Read the handler's action
before "fixing" a check. That mistake is what produced `146`'s first, wrong write-up.

> **THE RULE THIS FILE KEEPS BREAKING — owner's correction, 2026-07-29:**
> **a lack of evidence that a check or handler works is NOT evidence that it doesn't.
> It may simply not have run.** A1 originally read "`check_orphan_pages` has never
> repaired a page by any of its three branches", built from `0 complete` over three
> months. The handlers had **never been offered a single one of those items** —
> `claimed_by` NULL on all 37 — because the rows sit in statuses the dispatcher does
> not claim. The corrected finding is sharper *and points somewhere else entirely*.
>
> So every item below is now labelled by **what kind of evidence it rests on**:
> - **MECHANISM** — the code path cannot do the thing, and the artefact confirms it.
>   This survives "it hasn't run much". (A2, A4, A5, A6, B4, and C1's structural half.)
> - **NEVER RAN** — a count of zero that means *not exercised*, and says nothing
>   about correctness. Useful for prioritising cadence, useless for judging code.
>   (B1, B2, B3, C3, and now A1.)
>
> **Before writing "X does not work", ask what would have had to happen for X to
> leave a trace, and check that first.** For a work item the check is one column:
> `claimed_by IS NOT NULL`.

---

## Group C — the claims gate. **Owner's explicit requirement; top of the list.**

> *"The rewrite for `run_discovery_checks` must write copy that follows the claims
> checking like everything else."*

### C1. Copy written by discovery-triggered handlers is never claims-checked

> **⚠ READ THIS BEFORE TREATING C1 AS CLOSED — the floor would NOT have caught the
> witnessed case (checked 2026-07-30, by me, on my own change).**
>
> Another session contributed a WITNESSED instance of C1 the same day (`4494162af`):
> `page-content-writer` wrote four false claims onto gamesdesign.co.uk's homepage via
> `page-build-handler`, and they deployed. That path DOES go through
> `save_page_sections`, so the floor now sits on it — which makes it tempting to read
> the two entries together as "witnessed, then fixed". **They do not compose that way.**
>
> I ran the witnessed copy through the floor's own engine (`cmd/claimscan`, same
> functions, gamesdesign's own register, `page_type=content` so prose numbers ARE
> scanned): **0 findings.** Neither half fires.
> - *"10,000 Monte Carlo trials per query"* — the number scan needs
>   `businessClaimContextRe` to match the window (clients, customers, records, uptime,
>   years of experience…). "Monte Carlo trials per query" is not that vocabulary, so the
>   number is never scanned. That is **CLM-003's documented blind spot**, not a new one.
> - *"built **by** a shipped live-service designer"* — a fabricated human credential
>   matching no banned pattern. Nothing in the set is shaped like it.
>
> **So what C1's fix actually closes is the STRUCTURAL half: there was no gate at all on
> four of the six persisting agents, and now there is one on all six.** What it does not
> close is DETECTOR COVERAGE — and the witnessed case is a coverage failure, not a
> placement failure. A floor can only enforce what the engine can recognise.
>
> **This is the more valuable finding of the two and it is not yet an item.** The
> witnessed fabrication's shape is the dangerous one: the model took a true sentence,
> made one grammatical substitution, then **invented corroborating specifics to justify
> it** — so the falsehood arrives looking better-researched than the honest copy it
> replaced. Nothing lexical catches that. Whoever picks this up: it belongs as a new
> Group C item about the SET, not about the seam, and it should not be filed as "the
> floor didn't work".

> **✅ FIXED AND LIVE 2026-07-30 — commit `f61dce806`, chassis `v1.0.1211`,
> pod-verified on both replicas** (see the status block at the top of this file for
> the marker counts and the controls). Concept register **CLM-018**. Council
> `2d0dbc2e-e125-41f6-876d-0f8d6cf96688` — round 1 REVISE on evidence-attachment,
> round 2 REVISE on two unreadable seat results with **no high-severity objection
> raised**; nothing refuted in either round. See the status block at the top.
> **The fix is NOT where this item said it was — read the correction below before
> using anything in it.**
>
> **CORRECTED 2026-07-30: the seam is PERSISTENCE, not the handler.** Re-measured
> live that day (recursive `$.**.action` over `agent_definitions`, active
> non-snapshot): **six** agents persist page body sections through
> `save_page_sections` — `page-build-handler`, `pageflow-builder`, `page-rebuild`,
> `page-rerender`, `site-work-orchestrator`, `tool-recreation-handler` — and
> **two** of those run `validate_page_content`. That is the honest denominator.
>
> **And `page-content-writer`, which the rest of this item is built around,
> PERSISTS NOTHING.** Its workflow ends `compile_page → complete_workflow
> {output_field: page_content}` and it is **called by four of those six parents**
> (`page-build-handler`, `pageflow-builder`, `page-rebuild`,
> `site-work-orchestrator`). Adding `validate_page_content` to it would have
> gated a value on its way to a caller, not a write. `check_empty_sections.go:7`
> records the same route being corrected once already:
> *`HandlerAgent: "page-build-handler" (was "page-content-writer")`*.
>
> **The general lesson, which is the transferable part:** *"which agent writes the
> copy"* and *"which agent persists it"* are different questions, and an
> enforcement point has to answer the second. This item asked the first and got a
> plausible, wrong target — the same shape as this file's own A1 correction, one
> level along the pipeline.
>
> **What shipped instead:** a claims FLOOR inside `SavePageSectionsAction`
> (`save_sections_claims_guard.go`), reusing the same engine the gate calls. A
> banned claim (blocker — a known falsehood) **refuses the save**; an unregistered
> number (error) is **recorded durably and allowed**. Severity-driven, not
> check-by-check, so it stays correct as checks are added. Same argument
> `save_sections_link_repair.go` already settled for `bugs_open/079`: a gate can be
> forgotten by whoever writes the seventh agent; a floor cannot.
>
> **Blast radius, measured before submission** (`cmd/claimscan`, the gate's own
> engine, over the **complete** live corpus — 949 components / 14 sites, each
> against its own register): **3 banned-claim findings on 2 sites (0.32%)** —
> webdesign.co.uk `tool-blueprint-compiler` ("never invents"), robot-hands.com
> `how-it-works` and `gripper-catalog` (both "independently verified" = the
> already-filed `bugs_open/147`). Those three **cannot be re-rendered until the
> copy changes**; all three assert something untrue today. Plus **59 unregistered
> numbers** on the 4 armed sites, which record only. **Re-measure, do not quote:
> the surface was 908 components on 07-28, 919 on 07-29, 949 on 07-30.**
>
> **Still open after this fix:** `ApplySectionEditAction`, `create_report_page` and
> `rebuild_blog_listing` also persist LLM prose and are untouched (the first is
> `bugs_open/136`'s territory, the second is C2's question below), and
> `bugs_open/123`'s content-creator path has no site and no page row, so no
> persistence-seam fix can reach it.

**Measured 2026-07-29.** Of the 22 handler agents that discovery checks route work
to, **exactly 2 gate on `validate_page_content`** — `page-build-handler` and
`tool-recreation-handler`. The rest have no validation step and the string
`validate_page_content` appears nowhere in their config.

The important one is **`page-content-writer`**, the handler for `placeholder_contact`
and for `validate_component_standards`' `needs_content_page`. Its entire workflow:

```
spawn_research_agent → spawn_link_resolver → load_site_specs → prepare_link_context
  → build_render_context → resolve_links → select_sections
  → process_sections_loop   [loop, sub_workflow:
        call_researcher (call_agent research-agent)
        → generate_content (execute_llm_prompt, gemini-pro-latest, max_tokens 8000)
        → render_section (render_component)]
  → compile_page → complete
```

There is no validation anywhere in it. An LLM writes prose, it renders, it compiles,
it completes. `internal-linker` is the same shape — `plan_links` is an
`execute_llm_prompt` with no gate.

**Two aggravating factors, both already known and both landing exactly here:**
- The LLM write happens **inside a `loop` sub_workflow** — and per `bugs_open/144`,
  `sub_workflow` steps are validated by *nothing* (not cycle, not depends_on, not
  must-have-action, not config-key). So even adding a validation step inside that
  sub-workflow needs 144 resolved before anyone can trust it is wired.
- `validate_page_content`'s claims settings **default to `true`**
  (`validate_page_content.go:223,231,247` — `check_claims`, `check_stat_claims`,
  `check_claims_fleet_wide`). So the fix is mostly *adding the step*, not configuring
  it. Do not repeat the `in_header` mistake (A4) of relying on a default silently:
  set it explicitly so the intent is on the record.

**The fix.** Every handler that emits prose either runs `validate_page_content` with
claims checking on, or routes its output through one that does. `run_discovery_checks`
itself writes no copy — the rewrite the owner is asking for is at the handler seam,
and the honest scope is "every content-emitting handler", not one action.

**Verification query** (re-run after any fix; expect `validate_steps ≥ 1` for every
handler with a content-generating step):
```sql
SELECT a.type,
       (SELECT count(*) FROM jsonb_each(a.default_config->'workflow'->'steps') s
        WHERE s.value->>'action'='validate_page_content') AS validate_steps
FROM agent_definitions a
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;
```

### C2. `report-builder` explicitly disables claims checking

It is the **only** agent that sets `check_claims: false` on `validate_page_content`.
Either there is a reason that belongs on the record, or it is a leftover. Find out
which before switching it on — a report that legitimately restates figures from a
cited upstream may need different handling, not the same gate.

### C3. The claims DETECTOR lives in the least-run discovery agent

> **⚠ CORRECTED 2026-07-30 — THE HEADLINE BELOW WAS FALSE WHEN IT WAS COMMITTED,
> BY TWO AND A HALF MINUTES.** Re-measured live 2026-07-30:
> `claims_unverified` has **2 rows**, and `created_by` on both is
> **`quality-discovery-agent` itself** — not a session firing by hand:
>
> | created_at (UTC) | created_by | summary |
> |---|---|---|
> | 2026-07-29 17:06:02 | `quality-discovery-agent` | Unverified claims on about-index: 1 unregistered stat field(s) |
> | 2026-07-29 17:06:02 | `quality-discovery-agent` | Unverified claims on bayesian-ranking: 2 unregistered stat field(s) |
>
> **This file was committed at 17:08:30 UTC** (`b15b1456f`, `2026-07-29 18:08:30
> +0100`). The rows predate it by **2m28s**. The count was taken earlier in that
> session and written up without re-running — the exact failure the standing rule
> *"ground every figure against the live system before repeating it"* exists to
> catch. Logged in `WRONG_CALLS.md`.
>
> **So the detector is not unexercised and it is not broken: it RAN, automatically,
> and it FOUND things.** Both findings are `needs_human_review` with `claimed_by`
> NULL, which is HITL-terminal **by design** (`check_unverified_claims.go:145`) —
> correct behaviour, not a stuck queue.
>
> **What survives the correction is narrower and is a CADENCE fact, not a code
> one:** `quality-discovery-agent` carries **5 checks and has raised 9 items in its
> whole history**; its siblings carry **30 and 22** and have raised **172 and 152**
> (all three ran on 2026-07-29). Whether `unverified_claims` should ALSO be seated
> in `completeness-discovery-agent` is a real decision and is deliberately **not**
> taken here: double-seating interacts with `insertWorkItem` dedup on `item_key`,
> which is one of B2's three candidate causes, so seating it before B2 is
> established could produce a change whose effect nobody can attribute.
>
> **The C1+C3 pairing this file argued for is now half-closed from the other
> side:** C1's floor records every unregistered number it sees at write time under
> `CONTENT_CLAIMS_FLOOR_DETAIL`, so the platform gets a fleet-wide claims
> measurement from the write path without seating anything new.

`check_unverified_claims.go:145` is explicit that this is HITL-terminal by design —
`HandlerAgent: ""`, `Status: needs_human_review`, *"no automated handler, ever"* —
and that is right. But it means detection is the only backstop, and the backstop is
not running.

**NEVER RAN, not broken.** The zero rows are what you would expect from a check
sitting in an agent that has raised 7 items in its life; they are **not** evidence
the check is faulty, and it should not be rewritten on this basis. The cheap first
move is to seat `unverified_claims` in an agent that actually runs and see what it
says — the fleet claims dry run (`claimscan`) already finds real instances on live
sites (`bugs_open/147`), so a working detector on a working cadence should too. If it
then stays silent, *that* is a finding.

**So today there is neither a write-time gate (C1) nor a working after-the-fact
detector (C3).** That pairing is the actual exposure, and it is why C1 should not be
deferred behind the routing work in Group A.

---

## Group A — routing: detection that cannot repair

### A1. Two of the three branches never reach a handler at all — **the handlers are not implicated**

> **CORRECTED 2026-07-29, before anyone worked this item.** This entry first read
> *"`check_orphan_pages` has never repaired a page by any of its three branches"* and
> presented `needs_internal_links` at 33 items / 0 complete as a handler that does not
> work. **The owner's challenge was right: a lack of evidence that these handlers
> work is not evidence that they don't — they may simply not have run.** They have
> not. Measured below. `internal-linker` is not implicated by anything here.

| branch | item type | items | **ever claimed** | reachable by the dispatcher? |
|---|---|---|---|---|
| nav-flagged | `nav_drift` | 17 | **15 of 17** | yes — and 15 completed |
| unflagged | `needs_internal_links` | 33 | **0 of 33** | no |
| blog | `orphan_blog_posts` | 4 | **0 of 4** | 1 of 4 (the one raised today) |

`claimed_by` and `claimed_at` are **NULL on all 37** `needs_internal_links` and
`orphan_blog_posts` rows. `internal-linker` and the blog rebuild have **never been
offered one of these items**, so nothing here says anything about whether they work.

**Why they are unreachable, which is the actual defect:**

- `claim_work_item_action.go:102` claims only `status IN ('triaged','approved')`.
- **`unresolved` is a TERMINAL status** (`work_items_common.go:29-35`), i.e. closed
  and non-dispatchable. I read 27 terminal rows as a three-month backlog. They are
  not a backlog; they are parked.
- `detected` is not dispatchable either — `triage_detect_items_action.go` exists
  precisely to promote `detected` → `triaged` ("*This action bridges the gap*"). The
  9 `detected` items are awaiting triage, not awaiting a handler.

**And the sharper finding that replaces the wrong one: 20 of the 24 `unresolved`
`needs_internal_links` rows were BORN `unresolved`** (`updated_at` within 5s of
`created_at` — created non-dispatchable and never touched again), across **16
distinct `item_key`s for 24 rows**, i.e. repeat detections of the same pages. That is
the documented recurrence-branding failure, already pinned in
`work_item_recurrence_test.go:20,103`: *"later re-renders were born 'unresolved' and
never dispatched … which is how the fix loop silently died."*

**So the real item is:** why does a re-detected orphan get branded terminal at birth,
and is the same true for other item types? That is a dispatch/recurrence defect, and
it is testable. `[UNMEASURED]`: whether the branding is `insertWorkItem`'s dedup
behaviour on a repeat `item_key` or something later. Establish that before touching
either handler.

**What still stands, on mechanism rather than absence:** the `nav_drift` branch. Those
items **were** claimed (15 of 17) and **did** complete, and the handler provably
cannot repair a `/tools/` page — see A2, where the evidence is the code path plus the
artefact, not a missing row.

### A2. `nav_drift` for a `/tools/` URL is structurally unfixable by its own handler

`nav_drift` → `nav-updater` → `populate_nav_tables`, which **skips every URL under
`/tools/`, `/blog/`, `/guides/`, `/articles/`, `/case-studies/`, `/news/`,
`/resources/`, `/insights/`** (`populate_nav_tables_action.go:294,339`) on the stated
ground that the parent listing represents them. Proof: a `nav_drift` item raised
2026-07-24 for `tool-ai-agent-roi-estimator` is `complete`; the page has **0 nav
items and 0 chrome links** today. Fleet-wide, **2 nav items point at a tool page, out
of 95 deployed tool pages.** Full mechanism in `bugs_open/146` §4.

### A3. Missing route: `orphan_tool_pages` → rebuild the tools listing

The platform's own contract (that same comment) is that the parent listing
represents child pages — but nothing keeps a tools listing in sync when a tool is
added. The exact analogue exists for blogs (`orphan_blog_posts` → `rerender-pages`).
A listing alone is not sufficient: **gamesdesign.co.uk has `/tools/index.html`
(`page_type='section-index'`) and still has 4 orphans**, because the listing
enumerates only the tools using one of its two URL conventions
(`/tools/<name>/index.html` vs `/tools/tool-<name>.html`).

### A4. `pages.in_header` / `in_footer` **DEFAULT TO TRUE**, and that default does the routing

`create_tool_component_action.go:280` omits both columns from its INSERT;
`deploy_tool_action.go:117` defaults `inHeader := true` in Go. So the nav flag
records no decision — and `check_orphan_pages` branches on it, sending these pages to
`nav_drift` (A2, unfixable) instead of `needs_internal_links` (the branch that would
be right). Changing the column defaults is a **shared-schema change** — architecture
scope, own council round, and the fleet needs sweeping first for rows that are only
correct by inheriting `true`.

### A5. Two nav builders with opposite predicates

`populate_nav_tables` **excludes** tool pages from nav; `buildServicesHTML`
(`render_site_components_action.go:950`) **includes** any `in_header OR in_footer`
page in the chrome footer. Identical rows produce different live sites depending on
which ran last. One of the two is wrong; decide which.

### A6. The two tool creators each do half the nav write

`deploy_tool_to_site` sets the flags and writes **no `site_nav_items` row at all**
(only `populate_nav_tables_action.go` and `create_tool_component_action.go` write
that table). `create_tool_component` writes the nav item but never sets the flags.
Neither re-renders chrome. Fix at creation time — it makes the bad state
unrepresentable, which is why `146` now ranks it above detection.

---

## Group B — coverage: checks that never run

### B1. Six registered checks are configured in NO agent — **NEVER RAN**

`backend_unreachable` · `cross_site_contamination` · `orphan_element_refs` ·
`tool_recreation_needed` · `unrendered_templates` · `validate_component_standards`

Confirmed two ways: absent from every live `run_discovery_checks` `checks` array, and
**0 rows across all 11 item types they raise**. `validate_component_standards` alone
raises seven (`broken_template_slots`, `missing_logo_in_header`, `needs_content_page`,
`slot_name_mismatch`, `stacked_nav`, `unlinked_site_component`, `unwanted_nav_element`).
**The primary evidence is structural** — they appear in no live `checks` array, so
they *cannot* have run; the zero item count is corroboration, not the finding.
**Nothing here suggests any of these six is broken** — they are unexercised, which is
a different problem with a different fix. Each needs a decision: seat it in an agent
and see what it finds, or delete it. Dead-but-plausible checks are worse than absent
ones — they read as coverage. Expect a seated check to raise a burst on first run;
that is the check working, not a regression.

> **Not a defect, checked and cleared:** six configured names
> (`missing_*_tracker_page/section`, `missing_model_directory_*`) look unregistered to
> a grep for literal `Name()` returns, but `check_directory.go:111-116` registers them
> dynamically per profile. They are fine. *A dynamic registration is invisible to a
> grep for string literals — enumerate via the registry, not the source.*

### B2. No `nav_drift` item has ever been raised by a discovery agent

All 16 came from named sessions or `created_by='generic'` — i.e. threads firing
checks by hand. Meanwhile `completeness-discovery-agent` (which carries
`orphan_pages`) *is* running — 144 items, most recent 2026-07-25 — but its
orphan-branch output stopped: `needs_internal_links` last 2026-07-17,
`orphan_blog_posts` last 2026-07-15, `nav_drift` never.

**Cause undetermined `[UNMEASURED]`.** Three candidates, none excluded: (a) the agent
is only dispatched at some sites and never at the ones with orphans; (b) the check
errors and is swallowed — `discovery_checks.go:152` logs a failed check at WARN and
`continue`s; (c) `insertWorkItem` dedup suppresses re-raising against an existing
non-terminal item with the same `item_key` (`nav_drift:<site_id>` is one item per
site). **Establish which before changing anything** — the three have different fixes.

### B3. `quality-discovery-agent` is effectively dead

5 checks, 7 work items in its whole history, nothing since 2026-07-17 — while its
two siblings ran to 07-25 (144 and 108 items). It carries `unverified_claims` and
`voice_tells`, so this is the same defect as C3 viewed from the cadence side.

### B4. An unregistered check name is a WARN and a `continue`

> **✅ FIXED AND LIVE 2026-07-30 — commit `f61dce806`, chassis `v1.0.1211`,
> pod-verified on both replicas.** An unregistered name now **fails the step**,
> naming the bad name and the registered set; an **erroring** check is reported
> rather than fatal, because failing all thirty over one transient error would
> discard twenty-nine checks' findings. `checks_run` reports what **actually ran**,
> with new `checks_requested` / `checks_unregistered` / `checks_failed` keys
> alongside. Lever `allow_unregistered_checks` (default false) covers the
> seed-ahead-of-image window and is declared in `ConfigKeys`, or the audit would
> report the correct config as a stray key.
>
> **The safety proof, because a hard failure is a production risk:** all **57**
> check names configured across the three live agents that call this action were
> verified to resolve, and a fixture test (`discovery_checks_registration_test.go`)
> now pins them so a rename fails in CI rather than in the fleet.
>
> **One trap worth inheriting:** a `jsonb_path_query($.**.checks)` over
> `agent_definitions` also returns **`maintenance-triage`'s** array
> (`stale_pages`, `missing_content`, `orphan_nav`). Those are **not** discovery
> checks — that agent has no `run_discovery_checks` step and the array belongs to
> `scan_sites_for_maintenance`. Treating them as unregistered would have been a
> fabricated defect, and it would have made this fix look unshippable.
>
> **Consequence for the rest of Group B: its numbers are still the pre-fix ones.**
> Everything measured before this rolls was measured under the silent-skip
> behaviour. Re-run B1/B2/B3 after the roll before acting on them.

`discovery_checks.go:141-146`: `checks.Get(name)` returning nil logs
*"Unknown discovery check — not registered"* and moves on. A typo in a `checks` array
is invisible from outside the pod — the run reports success with a silently smaller
check set. Same for a check that *errors* (`:152`). **`checks_run` in the output
should record what actually ran, not what was requested** — and a name that resolves
to nothing should fail the step, not shrink it.

---

## Corrections to `bugs_open/146`, made while measuring this queue

- **The filename says "the check that finds them has not run since 07_17". That is
  wrong as stated.** `completeness-discovery-agent`, which carries `orphan_pages`,
  ran to **2026-07-25**. The 07-17 date belongs to `quality-discovery-agent`
  (B3) and to that agent's `needs_internal_links` output. The file is not renamed —
  forward-only — so read this correction with it. The accurate statement is B2:
  the agent runs, and its orphan branch has produced nothing automatically.
- `146`'s "cadence, not code" framing was corrected in that file earlier today; the
  creation-time causes are now its candidate 0. See also `WRONG_CALLS.md` 2026-07-29.

---

## Suggested order of work

1. **C1 + C3** — the claims gate. No write-time gate *and* no working detector is the
   only item here with a content-correctness consequence rather than a
   discoverability one.
2. **B4** — make silent check-skipping loud. It is small, it is self-contained, and
   every measurement in Group B is untrustworthy until it lands.
3. **B2** — establish why the orphan branch never fires automatically. Cheap, and it
   decides whether Group A's fixes would even be reached in production.
4. **A6 → A2 → A3** — creation-time first (unrepresentable beats detectable), then
   stop routing `/tools/` to a handler that cannot act, then add the real route.
5. **A4, A5** — schema default and the two nav builders. Both are shared-mechanism
   changes wanting their own council round; neither blocks the above.
6. **B1** — seat or delete the six dead checks.
7. **A1** — **not** a handler defect (corrected). The real item is the
   **recurrence branding**: 20 of 24 repeat detections born `unresolved`, i.e.
   terminal and non-dispatchable, across 16 distinct keys. Worth pairing with B2,
   since dedup-on-repeat-`item_key` is one of B2's three candidate causes and this
   may be the same mechanism seen from the other side.

## Relations

`bugs_open/146` (the case this came from) · `bugs_open/144` (`sub_workflow` validated
by nothing — blocks trusting any fix inside `page-content-writer`'s loop) ·
`bugs_open/128` (`check_image_url_404` masks working images and misses broken ones —
same layer, already owned) · `bugs_open/140` (a component asserting what the site
never said) · `bugs_open/117`/`118` (chrome is a stored artefact) ·
`bugs_open/098` (deployed ≠ fetchable; A1 is fetchable ≠ reachable) ·
016b §9 *"a detected defect whose handler cannot act on it"*.

---

## C1 — WITNESSED, with the false copy live and deployed, 2026-07-29/30

C1 was measured statically (2 of 22 handlers gate on `validate_page_content`).
**It has now happened.** `page-content-writer` — the agent C1 names as "the
important one" — wrote four false claims into gamesdesign.co.uk's homepage, which
rendered, compiled, completed and **deployed**, with nothing anywhere in the path
objecting.

Provenance: a `content_rewrite` item raised by `site-review-agent` during the
hand-fired `improvement-sweep` run of 07-29 (`bugs_open/083` contribution;
`bugs_open/150`). Item `b587fe1e`, `page-build-handler` → `page-content-writer`
17:17:34→17:26Z, page `deployed_at 17:26:22`.

**What it wrote, and what was true:**

| claim | reality |
|---|---|
| "built **by** a shipped live-service designer" | the previous copy said "built **for** live-service and tabletop designers". **The audience was rewritten into a credential** — a fabricated human qualification on a generated site |
| "drop-rate simulators run **10,000 Monte Carlo trials per query**" | the tool computes the binomial distribution **analytically** (Lanczos log-gamma for `C(n,k)`); `Math.random` appears **nowhere**. No sampling, no trials. 10,000 is `Math.min(val, 10000)`, a browser-freeze ceiling on attempts modelled, and the input defaults to `value="50"` |
| "11 interactive tools" / "10 technical guides" | **both exact** — and worth recording, because the same component's own `stat1_value`/`stat4_value` already held `11` and `10` |

**The mechanism-shaped lesson, which is worse than a wrong number.** The false
claims were not random: the LLM took a true sentence, made one grammatical
substitution ("for" → "by"), and then **invented supporting specifics to justify
the new sentence** — a trial count, a technique name, a "strict baseline". A
fabrication that arrives with corroborating detail reads as *more* researched than
the honest copy it replaced. No claims gate ran, and a human reading it for tone
would not have caught it either.

**The claim was repeated in four components, not one** — `hero.subheadline`,
`game-list.section_intro`, `guide-list.cta_subtext`, and `system-stats` as a
**headline stat card** (`stat2_label: "Monte Carlo Trials"`, `stat2_value:
"10,000"`). So a write-time gate must consider the page, not the field: fixing one
mention would have left a corrected sentence directly above a stat card asserting
the same falsehood.

**Partially self-corrected, by luck rather than by a gate.** A second rewrite at
18:10:15 removed the fabricated credential and kept everything else. Nothing
recorded *why*, and no gate was involved — so the one genuinely dangerous claim
(a human qualification) survived on the live site for ~44 minutes and was then
removed by an unrelated iteration. That is not a control.

**Remediated by hand 2026-07-30** (owner instruction) across all four components,
writing **both** `content_data` and `rendered_html` per component, each guarded on
its own `updated_at`, so neither a regeneration-from-source nor an assemble-only
rerender can restore it. The figure 10,000 was KEPT where it is true (the attempts
ceiling) and the mechanism claim dropped; `stat2` became "Max Attempts Modelled".
The 10 other site components mentioning Monte Carlo were left alone — they are
guides *about* RNG design, i.e. legitimate exposition, and C1's gate must not flag
those.

**What this adds to C1's fix.** The verification query in C1 is necessary and not
sufficient: it proves a `validate_page_content` step exists. It cannot prove the
step would have caught THIS, because the two wrong counts were right and the
detectable falsehoods were a mechanism name and a credential — the
`banned_claims`/`stat_claims` families from `bugs_closed/104`. **Before declaring
C1 fixed, run the gate against this exact stored copy** (recoverable from
`page_component_history` for page `6e988cc4-4898-4021-aa5e-2ab0271f9b75`) and
confirm it fires. A gate that passes this text is not a gate.
