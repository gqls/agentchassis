# NOTES — bugs_open/208, owned page committed before the guard refuses it

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-08-06 — session pickup and first-hand re-verification

Picked up `bugs_open/208`, filed hours earlier by the `bugfix_201_page_content_writer_dispatch`
lane in pre-flight for an owner-authorised rebuild of `ai-agent-orchestration.com`. That lane
did **not** take the fix on — its own transcript says "208 is handed off. Continuing with 201
symptom 2", so the file's "OPEN, unowned" is accurate.

**Ownership check, and why the advisory VERDICT was overridden.** `scripts/who-owns.py 208`
returns **OWNED or recently active** — but only because the *filing* commit (`aaf8779e2`,
today) touches the file. That is the tool's documented blind spot: it reads commits, so a
filing looks identical to a claim. Per the memory note that every ownership check is lagging,
I also grepped the live session transcripts under
`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`, which is the only source that
sees an *uncommitted* session:

```
for f in $(ls -t *.jsonl | head -30); do
  hits=$(tail -c 400000 "$f" | grep -oE "208_HANDOFF|rebuild_policy|queryPagesForBuild|get_pages_to_build" | sort | uniq -c | tr '\n' ' ')
  [ -n "$hits" ] && echo "$f [$(date -r $f +%H:%M)] $hits"
done
```

One session (`fef871d4…`, active) was deep in the mechanism — the filer. Reading its last
outputs showed it had handed 208 off and moved to 201 symptom 2. Two other sessions had 2
incidental `rebuild_policy` hits each, no 208 involvement. **Taken on.**

### The bug is still valid, and the blast radius is bigger than filed

Every link re-read first-hand rather than taken from the handoff.

**1. Selection ignores ownership — confirmed.** `queryPagesForBuild`
(`platform/orchestration/actions/get_pages_to_build_actions.go:88-165`) filters on
`site_id + status='active' + COALESCE(build_status,'planned') IN (...)`. No `rebuild_policy`
clause. Note a second branch the handoff did not mention: **`include_all: true` drops the
status filter entirely**, so that branch would sweep every `deployed` owned page too.

**2. Live consumers — TWO, both `include_all:false`** (nested walk over `agent_definitions`,
because a top-level `jsonb_each` under-reports steps nested in a loop sub_workflow — the
landmine noted at `save_page_sections_action.go:180`):

| agent | step | build_statuses |
|---|---|---|
| `page-rebuild` | `get_pages_to_rebuild` | `["needs_rebuild"]` |
| `pageflow-builder` | `get_pages_to_build` | `["planned","needs_rebuild"]` |

So `pageflow-builder` is exposed too, and it additionally selects `planned`.

**3. The commit-before-guard order — confirmed live, and it is THREE agents, not one.**
`assemble_page → deploy_page (action `git_commit`) → save_sections (action
`save_page_sections`) → update_page_status` is the order in `page-rebuild`,
`pageflow-builder` **and** `site-work-orchestrator` (the third selects from work items via
`load_work_items`, not from `get_pages_to_build`).

**4. Live exposure census.** 14 active pages are `rebuild_policy='owned'` AND
`build_status IN ('needs_rebuild','planned')`, across **6 domains** — the handoff named 2 on
1 site:

```
agent-complexity-estimator, password-entropy      ai-agent-orchestration.com  needs_rebuild
ai-agent-roi-estimator                            finetuning.uk               needs_rebuild
tool-drop-rate-simulator, tool-ehp-calculator,
tool-jump-physics, tool-lanchester-sim,
tool-progression-architect, tool-ttk-calculator   gamesdesign.co.uk           needs_rebuild
password-entropy                                  leopardessconsulting.co.uk  needs_rebuild
provocation (blog-post)                           vonc.com                    planned
tool-archetype-taster-quiz, tool-arena-interface,
tool-gauntlet                                     vonc.com                    needs_rebuild
```

`tool-arena-interface` and `tool-gauntlet` are on **vonc.com** — the same site as the "vonc
arena clobber" that motivated the ownership marker in the first place. A further **189** owned
pages sit at `deployed`: not selected today, but selected by the `include_all:true` branch.

### The finding that decided the design

`assemble_page` has **exactly three live consumers — the same three agents above — and in all
three its `next_step` is `deploy_page`.** The *sanctioned* owned-page deploy paths do **not**
use it: `page-rerender` goes through `rerender_single_page`, `section-editor` through
`apply_section_edit`. Both of those still `git_commit`.

That is what rules out the tempting fix. A guard inside `git_commit` would be the widest net,
and it would **break the only paths by which owned pages legitimately deploy** — which
migration 164 says in terms: *"page_rerender / assemble (re-assembly of EXISTING
page_components) is deliberately NOT gated — it is how owned pages deploy."* `assemble_page`,
by contrast, means precisely "generic composition of freshly generated content, about to be
committed". It is the seam.

**Verified the skip protocol reaches the commit.** All three `assemble_page` steps declare
`output_field: "assembled_page"`, and `checkUpstreamSkipped` (`git_deployer_actions.go:576-588`)
reads `collectedData["assembled_page"].skipped` as its first branch. `AssemblePageAction`
already returns `{"html":"", "skipped":true, "skip_reason":…}` in two existing cases
(`multipage_actions.go:38-62`). So a refusal expressed as a skip needs **no change to
`git_commit` and no config change on any agent** — existing machinery, not new.

### Prior art that shapes the fix rather than duplicating it

- **Migration 164** (`docs/agent_docs/sql_for_agents/164_pages_rebuild_policy.sql`) — created
  the column as Experience Loop guard rail 1, mechanising TL-001. It names exactly two Go
  refusals (reconcile emits `owned_page_review`; `save_page_sections` hard-refuses) because it
  assumed the only route into a generic build was `reconcile → needs_page`. **A path that
  selects straight off `pages.build_status` was never in its model — that is the gap 208 is.**
- **`reconcile_site_plan_action.go:232-270`** — the framework's existing answer when an owned
  page reaches a generic builder: exclude it and emit a `site_work_items` row of
  `item_type='owned_page_review'`, `status='needs_human_review'`, deduped by
  `item_key='owned_page_review:'+name` with `ON CONFLICT DO NOTHING`.
- **`features_open/012`** (LIVE v1.0.1149, council-approved) — establishes the precedent that
  regenerating an existing page's composition requires **explicit named intent**
  (`recompose_pages`). An opt-in override on selection is the same philosophy one layer down,
  not a new invention.
- **`features_open/021`** (LIVE and proven **today**) — the operator bulk-rebuild entry point
  whose first real dispatch surfaced this. **Owned by another workstream**, so it is a
  consumer to be told, not just measured.

### The damage is still UNREALISED — measured, not assumed

The filing marked the damage `[INFERRED]` (correctly: it declined to induce it). I checked the
current served state of all 14 exposed pages directly —
`BASELINE_2026-08-06_owned_pages_served.txt`, http status + byte count + `<script>` count +
interactive-element count + a body sha256 per URL.

**[MEASURED 2026-08-06] 13 of 14 serve HTTP 200 with 5-7 `<script>` blocks and 2-15
interactive elements — i.e. the tools are intact. The trap is armed and has NOT fired.**

What would have disconfirmed this: a 200 of ~10-15KB with 0-1 scripts and no interactive
elements, which is what generic regenerated prose looks like on this estate. None of the 13 is
that shape.

The 14th, `vonc.com/blog/provocation.html`, is a **404 with 0 `page_components`** — the
`planned` owned page migration 164 deliberately parked ("the parked per-provocation page the
experience spec will define (T4.3)"). So it has never been built at all.

Two things follow, and both matter to the fix:

1. We are fixing **before** the fire, not after it, and the baseline above is a clean control
   set: after the fix these 14 bodies must be byte-identical (same sha256).
2. **The no-op case is checked, not just the damage case.** The only owned page a
   selection-level exclusion would remove from `pageflow-builder`'s `planned` set is a page
   that 404s and has no components today — so excluding owned pages costs the live fleet
   nothing observable, while protecting 13 working tools. A fix whose downside I had not
   measured would be a guess dressed as a trade-off.

### A second symptom the filing did not report: one owned page aborts the WHOLE batch

`continue_on_error` is what makes a loop tolerate a failed iteration
(`platform/orchestration/loop_error_handler.go:66-80`: `shouldContinueLoopOnError` requires
the loop step's `continue_on_error` to be true). **[MEASURED] It is unset (NULL) on all four
build loops** — `page-rebuild.build_pages_loop`, `pageflow-builder.build_pages_loop`,
`site-work-orchestrator.build_items_loop` and `.fix_items_loop`.

So today, when `save_page_sections` hard-refuses the owned page, the refusal does not merely
lose that page — **it fails the entire workflow**. Pages are selected `ORDER BY nav_order ASC,
name ASC`, so every page positioned after the owned one is never rebuilt either. The operator
gets a partially-rebuilt site, one destroyed tool, and one error.

That settles a design question I had been holding open: **Layer 2's refusal must be a SKIP, not
a hard error.** `AssemblePageAction`'s existing skip comment says exactly why — *"Return
success with skipped flag - allows loop to continue to next page"* (`multipage_actions.go:39`).

`SavePageSectionsAction` has **no** upstream-skip check (read `save_page_sections_action.go:41-120`:
it guards `DB == nil`, missing page name and bad site_id, then goes straight on). It resolves
the page name from `current_page.name`, which survives an assembly skip — so after Layer 2
skips the assembly and `deploy_page` skips the commit, `save_sections` would still reach its
owned refusal and abort the run. Damage prevented, batch still dead. Hence a possible third,
smaller edit (ranked as optional, not required for safety) — see PLAN.

### The sanctioned path already does it in the right order

`page-rerender`'s live step graph is `rerender_sections → save_sections → render_page
(rerender_single_page) → deploy_page (git_commit) → update_status`. It **saves before it
commits**, and it models a skip as a first-class outcome (`check_skipped` conditional).

So `commit → save` is not the platform's convention that 208 happens to sit inside; it is an
anomaly confined to the three generic composition loops, and the path that handles owned pages
correctly is the one that already inverted it. Useful for the council submission: the fix moves
the generic loops *towards* the existing norm rather than inventing one.

### Open question logged before it is answered

`[UNMEASURED]` Whether excluding an owned page at selection leaves it stuck at
`needs_rebuild` for ever and re-selected on every subsequent run — i.e. what
`update_page_status` does for a page the loop skipped. Being answered before the design is
fixed, not after. (Note: harmless in the damage sense — re-selecting and re-excluding costs
nothing — but it would make an operator's explicit request silently do nothing for ever, which
is the visibility problem, not a safety one.)

> **ANSWERED 2026-08-06, and my premise was WRONG in the more dangerous direction.**
> `UpdatePageStatusAction` has no skip check, and for an owned page **both** existing deploy
> guards pass — `pageHasComponents` is true (the tool's own component) and
> `pageSectionShortfall` compares planned `pages.sections`, typically empty for an owned page.
> So the page is **stamped `deployed`**, not left at `needs_rebuild`. Worse than a wrong status:
> the same statement writes `built_from_plan_version = the current plan`, which makes
> `ReconcileSitePlanAction`'s `decideEmit` return `skip_built` and **permanently suppresses the
> `owned_page_review` item that is this design's own visibility channel.** So the honest-looking
> option ("let it move off needs_rebuild") actively blinds guard rail 1. Found by the second
> model reviewing the design, not by me — I had assumed the page would simply sit there.

---

## 2026-08-06 — what the second model (fable) corrected, and what I corrected in its plan

Recorded both ways round, because the point of the exercise is the disagreements.

**It corrected me on three things.** (1) The `deployed`-stamp finding above — the most valuable
of the session, because my assumption was safe-sounding and wrong. (2) It found the second Go
caller of `queryPagesForBuild` (`WriteBuildItemsAction`) independently; my compiler found it at
the same time, which is a weaker form of the same catch. (3) It dismantled my reason for holding
the visibility arm back: I had cited the 114-junk-items landmine, and that incident
(`bugs_open/204`, `needs_new_component` per unresolvable slot) had **no stable key and an
unbounded population**, whereas `owned_page_review` has one deterministic key per page arbitrated
by the `(site_id, item_key)` partial unique index. Different shape, so the objection did not
transfer. The emitter went in.

**I corrected it on one thing that matters.** Its plan assumed `site-work-orchestrator`'s loop is
fed by `write_build_items`' `needs_page` items via `current_item.spec`. Live measurement says
otherwise: that agent's `load_work_items` filters `handler_agent='page-content-writer'`, while
**every one of the 158 `needs_page` rows fleet-wide (all history) has
`handler_agent='page-build-handler'`** — a different agent, whose order is
`save_sections → update_status → deploy_page` and which is therefore **safe**. The third loop's
real exposure is discovery-check fix items (`literal_markdown`, `placeholder_contact`), of which
**11 of 14 targeted `owned` pages on webdesign.co.uk on 2026-08-04, all failed.** So the door is
real and observed, but it is a different door from the one the plan defended, and that changed
which collected-data shape the guard has to resolve (`current_item.spec` for a *fix* item, not a
page record).

`[UNDETERMINED — EVIDENCE REAPED]` Whether those 11 runs reached `deploy_page` before failing.
All 11 pages serve working tools today and terminal `orchestration_states` are reaped at ~24h, so
the deciding `collected_data` is gone. **Not asserted.** The plausible innocent explanation is
that they failed earlier (an owned page has no planned sections for `plan_sections` to work
with); the plausible guilty one is a commit followed by a repair on 08-05, when all 11 rows were
touched. Both fit. Neither is evidence.

## 2026-08-06 — mutation testing, and the one that mattered

Four mutations, each expected to break a named test:

| mutation | guard killed | outcome |
|---|---|---|
| M1 | `ownedPageExclusionSQL` emptied | 3 × `TestQueryPagesForBuild_*` failed ✓ |
| M2 | `pageIsOwnedForGuard` → `false` | `TestAssemblePage_OwnedPageIsSkippedNotAssembled`, `_WorkItemShapeIsGuarded` failed ✓ |
| M3 | `upstreamAssemblySkipped` → `false` | **`TestSavePageSections_HonoursUpstreamSkip` PASSED — the test was worthless** |
| M4 | stamp condition dropped | `TestUpdatePageStatus_RefusesDeployStampAfterOwnershipSkip` failed ✓ |

**M3 is the entry that earns this section.** My test asserted `res["skipped"] == true` with no
sqlmock expectations. With the guard deleted the action runs on, fails the page lookup against a
mock that expects nothing, and takes its **pre-existing** "page not found, skipping" branch —
which returns the identical shape. Two guards in series, one indistinguishable outcome; the test
would have vouched for the absence of the thing it was written to pin. Fixed by asserting the
**discriminator** (`reason`, which carries `OWNED_PAGE_GUARD` only on the intended path), after
which M3 fails as it should. Logged in `WRONG_CALLS.md`, and it is the reason the estate's
"mutate to prove a guard" rule is worth its cost: three mutations behaved and told me nothing I
did not already believe.

## 2026-08-06 — build discipline on a tree three sessions were editing

`go build ./platform/...` failed on `plan_sections_action.go:1007: undefined:
composeScopedWriterBlock` — a symbol that exists **nowhere**, in a file I never touched, from
another session's uncommitted WIP (they have written a call ahead of the function). A green or
red tree build proves nothing about my change either way.

So everything was verified in an isolated overlay — `git archive HEAD | tar -x` into a scratch
dir, then **only my six files** copied over it. That is exactly what `make build-*` will archive
once committed, and it excludes every other session's WIP by construction. `go build
./platform/...` and `go test ./platform/orchestration/...` both green there. Command in the
RUNBOOK as R6.

## 2026-08-06 — council APPROVED, and the three objections that improved the fix

Corr `5d1dcb10-7929-431e-b9e5-496992ce3229`. 13 reviewers, 4 abstained,
`decided_by: "approved with 5 advisory objection(s) — none high-severity"`. Read in full
(`scratchpad/verdict_body.txt`), not just the decision field — the whole point of the gate is the
objections, and three of these were right.

**Acted on (`f5710d6b0`):**

1. **`reuse_agent` MEDIUM — I had reimplemented the ownership predicate** rather than extracting
   the one already inline in `save_page_sections`. Exactly the drift class this guard exists to
   prevent, and I walked into it while writing the guard. The inline query is now a call to
   `pageIsOwnedForGuard`, declared in its doc comment as the only place ownership policy may be
   read. Behaviour-preservation is evidenced rather than asserted: the pre-existing
   `save_sections_stored_slot_identity_test.go` mocks that exact query and still passes unchanged.
2. **`bug_historian` MEDIUM — fail-open was silent.** An owned page whose policy read times out
   gets composed and committed, with only a log line. Answered by making the window *loud*, not by
   failing closed (which would stop generic page building fleet-wide on one flaky query, and the
   seat's own framing — "two fail-open gates instead of one" — is the argument for reporting, not
   for reversing the posture). The predicate now returns `(owned, checked)`; `!checked` writes
   `OWNED_PAGE_GUARD_UNCHECKED` to `agent_error_log`, so the window is a number:
   `SELECT count(*), max(created_at) FROM agent_error_log WHERE error_code='OWNED_PAGE_GUARD_UNCHECKED';`
3. **`architecture` LOW** — the single-source doc comment, folded into (1).

**Answered with evidence rather than a change:**

4. **`guidelines` MEDIUM — "WORK-ITEM DEDUP mandates DELETE+INSERT, not ON CONFLICT"**, because
   `idx_swi_dedup` is partial. The mechanism does not apply to the form used here: a bare
   `ON CONFLICT DO NOTHING` names **no conflict target**, so there is no partial-index inference
   and no 42P10 — that hazard belongs to the *targeted* `ON CONFLICT ... WHERE` shape
   `insertWorkItem` uses, which is why `work_items_common.go` carries its lockstep warning.
   **Induced rather than argued**, in a transaction rolled back afterwards:

   ```
   INSERT 0 1   -- first keyed insert
   INSERT 0 0   -- identical key, open row exists   => deduped
   UPDATE 1     -- drive it to 'cancelled' (terminal)
   INSERT 0 1   -- same key again => 2 total, exactly 1 open  => correct re-arm
   ROLLBACK     -- 0 rows remain
   ```

   Also: `insertWorkItem` requires a `*sql.Tx` and adds two-strike/within-cycle suppression, and
   **reconcile — the existing producer of this very `item_type` — uses the same bare form.** Using
   a different one here would make two producers of one key namespace behave differently, which is
   worse than the objection.
5. **`editquality` LOW** — the `write_build_items` rationale misattributed the causal route.
   Correct, and already self-caught before the verdict (see the fable/measurement entry above and
   `WRONG_CALLS.md`).

**The objection that was wrong, and it is our own fault it was made.** `prior_art_librarian`
(MEDIUM) inferred from the landmine corpus that `ownedPageExclusionSQL` / `include_owned` /
`checkUpstreamSkipped` already existed "from a prior attempt at this exact bug", so
`owned_page_guard.go` was *"a rebuild, not an add"*. There was no prior attempt. Those
`subject_keys` exist because **this lane wrote the landmine an hour earlier in this same session,
footprinted on its own new symbols, and `landmines-sync.py --apply` published them to `doc_notes`
before the council ran.** Writing a landmine for your own unshipped change manufactures prior-art
evidence about that change. Nothing to fix in the sync tool — but note the shape: a reviewer
searching a corpus that already contains your own footprints will find you, and report it as
precedent.

### Mutation, round two — and it caught two guards with no test at all

Re-running the mutations after the refactor (a refactor is exactly when a guard quietly stops
being load-bearing) found **two gaps my green suite had hidden**:

- **M7** — force `checked=true`, silencing the new fail-open report: **nothing failed.** The
  report I had just added to answer a council objection had no test.
- **M2b** — make the shared predicate never return `owned`: only the *assemble* tests failed.
  **Nothing anywhere asserted `save_page_sections` refuses an owned page** — the pre-existing suite
  mocks the policy as `generic`, so migration 164's own guard was untested, and my unification
  inherited that blind spot.

Both now pinned (`TestAssemblePage_FailsOpenWhenPolicyUnreadable` gained an `agent_error_log`
expectation; `TestSavePageSections_RefusesOwnedPage` is new) and both re-mutated afterwards to
confirm they fail. **Three mutation rounds, three findings — each time the informative mutation
was the one I expected to be boring.**
