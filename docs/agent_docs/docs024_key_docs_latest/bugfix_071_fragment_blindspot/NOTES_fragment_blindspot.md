# NOTES — fragment blind spot (append-only, newest at the bottom)

## 2026-08-06 — selection, verification, research

**Selection.** Ranked the 38 open bugs by reference-heat over live transcripts
(41 files, last 4h): coldest 085 (30), 093 (41), 113 (43), 146 (44), 203 (45).
Walked them in order: 085 nearly closed (both paths verified live, brochure
lane's); 093 blocked on 083 with no code work left; 113 fixed in code 07-27,
needs verify+close by its lane; 146 site-lane repairs + one architecture-scope
template change; 203 active (fix committed 08-05). 114 looked free at heat 49
but the symbol grep found session ad5665d0 building `asset_reference_404`
(commit e526a5196 — filed under 084) and 7b4e88a8 draining image_url_404 queues
— **the cold-heat pick would have collided twice; the symbol grep is what
caught it.** Settled on 071's fragment residue: named unowned by 071's own
triage note, symbols cold, 9 days quiet.

**Validity re-measurement (the bug MOVED).** 071's 07-25 figure (24/25 anchored
links dead) is stale. Today: 5 path#fragment links (idea.uk, all resolve), 61
bare-# links (57 = `#content` skip-links, ids present in stored rows and on
served pages; the rest probed resolving). Live damage ≈0; the check gap
unchanged at HEAD (links.go:113/199, validate_page_content.go:910,
accumulateLinkIssues). Evidence queries in RUNBOOK.

**Register status catch (LANDMINE class: stale register STATUS).** LNK-009 says
check_phantom_internal_links is "deliberately not yet enabled". Live
`agent_definitions` shows `phantom_internal_links` IN
completeness-discovery-agent's checks array, with `phantom_internal_link` items
complete as recently as 08-04. Will correct the entry visibly in this lane's
commit.

**Misstep (logged in WRONG_CALLS).** Queried
`item_type='phantom_internal_links'` (the CHECK name), got 0 rows, and said
"zero items ever" in a visible message. The item type is the singular
`phantom_internal_link` (ItemType: f.IssueType literal in the check). 119 items
exist, 55 complete. The check: take the item_type spelling from the `ItemType:`
literal in the check's source, never from the check name — they differ in this
package by design (one check, three item types).

**Reuse found (why D2 is a refactor, not new code).**
`datahelpers.OrphanElementRefs` already answers "does this document contain or
create this id", with paid-for conservatisms (dynamic ids, interpolated-id
loosening — the css-filter-playground false positive). The fragment arm's
presence test must be THAT test extracted, or the two will disagree about what
an id is.

**Assembler check.** Served loancash.co.uk/index.html ids ⊆ stored
(page_components ∪ site_components) ids — deploy-time assembly adds no ids on
that page, so stored-row resolution matches the served document. Single-page
sample; the pre-roll harness covers the rest.

**Concurrent-lane map.** d361e826 (active 10:00) builds a page-pairs discovery
check — same package, different files; shared surface = the two coverage tests;
keep edits additive, re-read before edit. 203's lane owns the
`primary_cta_url`/`secondary_cta_url` defaults map still at
component_library.go:1136-1147 (their fix removed only the `cta_url` scalar
defaults) — recorded for them in 071's update, not taken.

## 2026-08-06 (later) — built, measured, submitted, committed

**Built.** `SplitFragment` + `DocumentIDs` (extracted from `OrphanElementRefs`,
which now runs on top of it) + the `dead_fragment_link` arm + its verifier + the
writer constraint. Commit `af2667453`. Council `bbbb4132-4abe-4db1-a1ba-755377dab009`
(submitted before the commit; `Council-Submitted:` trailer, so 098 credits it
automatically when the verdict lands).

**The coverage guard did its job on me.** Registering the verifier broke the
build instantly — `TestRegisteredVerifiersMatchClaimTimeoutExclusion` named the
obligation I did not know existed: an item type with a verifier must ALSO be
excluded from the claimed-item-timeout sweep, or the 15-minute auto-complete
branch walks past the verifier. Hence migration `322`, plus `220`'s declared
list. I read the LIVE column before writing the replace (it matched 220 exactly,
6 entries), which is what `305`'s header says to do — it had to carry another
lane's unapplied entry because nobody checked.

**Measurements, both disconfirmable.** Fleet harness over the shipping
functions: 67 fragment-bearing hrefs, 0 findings; same corpus + 2 planted dead
fragments → exactly 2, one per arm (bare and cross-page). Mutation: 3 mutations,
3 distinct test failures, tree restored green.

**Second misstep this session** (both now in WRONG_CALLS): I nearly reported the
clean 0 as evidence before inducing a non-zero on the same corpus. Ninety
seconds of planting turned a vacuous number into a real one.

**Roll state at the time of writing.** A fresh build landed mid-session:
`v1.0.1257` on both replicas, pod-grepped `dead_fragment_link` = 0, positive
control `phantom_internal_link` = 9, negative control 0 — i.e. correctly NOT
carrying this work, which was uncommitted at that moment. The arm needs the NEXT
roll.

**Still owed:** the verdict; migration 322 applied; the post-roll pod-grep and an
induced live finding; then this file's damage/no-op pair re-run.

## 2026-08-06 (evening) — APPROVED round 1, and the guardian's objection found a caller I had not checked

**Verdict: APPROVED, 3 advisory objections, none high-severity**, correlation
`bbbb4132-4abe-4db1-a1ba-755377dab009` (11 seats; `architecture` returned
`point_fix`; `guardian`, `bug_historian` and `debug_historian` objected at
medium and approved-with-objections overall). Every objection that named a
checkable fact has been checked rather than argued with; the answers:

**1. `guardian` [medium] — "an exported-signature change to a shared helper needs
a check for other callers before 'strictly a refactor' is safe to accept."
THE SEAT WAS RIGHT AND I HAD NOT LOOKED.** There IS a third caller outside the
two packages I had in mind: `deploy_tool_action.go:182` — and it is a **hard
pre-deploy refusal gate** for tool birth, i.e. the worst place to change
behaviour silently. My submission asserted "existing datahelpers tests pass
unchanged", which is true and is not the same claim.

Settled by differential test, not by argument: the pre-refactor implementation
was restored verbatim from `af2667453^` into a temporary harness and run against
the new one over **every** component template in the estate (2.3 MB) plus every
page component, every site component and every whole-page document in the fleet
dump. **4,036 documents, 0 mismatches.**

> **The first run of that differential was VACUOUS and my own guard caught it.**
> 2,018 real documents, 0 mismatches — and **0 documents where the old
> implementation returned anything non-empty**, so the two agreed only about nil.
> I had written `if nonEmpty == 0 { t.Error("VACUOUS: …") }` into the harness
> before running it, which is the only reason I did not report the clean pass.
> Re-run with each real document ALSO compared in an id-stripped variant — which
> turns every script-referenced id into an orphan and exercises the
> present/dynamic/interpolated branches on real markup — gives **403
> discriminating cases and still 0 mismatches.** Third time this session the
> "could this measurement have come out otherwise?" question changed the answer.

**2. `guardian` [medium] — does anything else consume this check's finding counts
or severities?** No. `grep -rn phantom_internal_link --include=*.go platform/
internal/ pkg/` returns only the check itself plus **comments** in nine other
files that reference the class boundary (dead_controls, misdirected_cta,
backend_entry_orphaned, link_repair, …). Nothing reads the counts, so a new
low-severity arm cannot skew a consumer.

**3. `guardian` / `prior_art_librarian` / `debug_historian` — migration ledger and
live-column state.** Applied by hand and verified after the fact, not just at
plan-authoring time: live `pre_query` now carries `dead_fragment_link` as the
7th exclusion; `schema_migrations` has
`322_dead_fragment_link_claim_timeout_exclusion.sql | record-only | 10:20:14Z`,
recorded via `--record-only` with a note (never a hand-written INSERT). The
runner's own probe now REFUSES a replay — its dry run reports *"expected exactly
1 scheduled_task carrying the known 6-entry exclusion list, found 0"*, which is
the pre-assertion doing its job, not a fault.

**4. `reuse_agent` [low] — was any existing item_type already claiming
fragment/anchor territory?** Queried, which I had not done from the DB side:
`SELECT DISTINCT item_type … ~* 'frag|anchor|link|nav'` →
`nav_drift, nav_rebuild_refused_incomplete, needs_internal_links,
phantom_internal_link, unbuilt_internal_link, unlinked_site_component`. **None
resolves a fragment**; no collision.

**5. `bug_historian` [medium] — a shared predicate reused on a new INPUT SHAPE.**
Accepted as a real limit and stated rather than closed: `presentIDRe` harvests
ids from the whole page text **including inside script string literals**, which
`OrphanElementRefs` does deliberately to avoid false positives. For fragment
resolution that inherited looseness produces **false NEGATIVES** — a `#pricing`
whose id exists only inside a script string is called resolved. That is the same
direction the other consumer chose (under-report, never accuse a working page),
and it is the direction this arm should fail in. Recorded here and in the file
header rather than "fixed", because tightening it is what would produce findings
against working pages.

**6. `bug_historian` [medium] + `architecture` — three unaligned consumers now
reason about link-target resolution (gate, this arm, `link_repair.go`).** Agreed,
and explicitly NOT fixed here: the architecture seat read the same fact and still
returned `point_fix`, noting `DocumentIDs` is positioned so the deferred
section-id-emission work has a validator ready. Logged as this file's open item 1
and named in `bugs_open/071`. That split is a candidate for the architecture
track, not for a bug patch — which is the 2026-07-28 ruling's whole point.

**7. `debug_historian` [medium] — no separate backup/rollback artefact for the
live `scheduled_tasks` mutation.** Fair. The file carries a counted
pre-assertion, a two-directional post-assertion (new list present AND old list
consumed) and a stated inverse-replace rollback in its header, but no dumped
"before" row. The blast radius is one column of one row and the before-state is
recorded verbatim in the migration's own header and in this lane's RUNBOOK, so I
have not re-run it; noting the house standard is a separate `_ROLLBACK.sql`.

**No re-submission.** The verdict is APPROVED and the objections are advisory;
`af2667453` already carries `Council-Submitted:`, so `098` credits it
automatically at report time — and forward-only forbids an amend to add
`Council-Reviewed:`.

## 2026-08-06 (post-roll) — LIVE on v1.0.1259 and INDUCTION-PROVEN, all four cases

**1. Shipped.** `v1.0.1259`, **both replicas**, one exec each:
`dead_fragment_link` **10** (0 pre-roll), `VerifyDeadFragmentLinkResolved` **2**,
`SplitFragment` **2**, positive control `phantom_internal_link` **10**, negative
control `zzz_no_such_string_control` **0**. (The positive control moved 9 → 10
because this change's own file names the sibling type — worth knowing before
someone reads that as drift.)

**2. Induced on a REAL run of the live binary, not a test.** Fixture on the pool
site `pool-ai-agents.internal` (`status='pool'`, 0 pages before, nothing serves
it), two scratch pages carrying **four cases in one run**, then a real
`completeness-discovery-agent` dispatch via a one-shot `scheduled_tasks` row.
Fired in **under a minute** (11:35:51Z).

| case | href | expected | got |
|---|---|---|---|
| bare fragment, no such id | `#zzz-induced-dead` | FIRE | **filed** |
| bare fragment, id on the same page | `#zzz-induced-live` | silent | **silent** |
| cross-page, target lacks the id | `/zzz-induction-b.html#zzz-induced-crosspage-dead` | FIRE | **filed** |
| cross-page, target HAS the id | `/zzz-induction-b.html#zzz-induced-crosspage-live` | silent | **silent** |

Exactly **2** items, both `severity=low`, `handler_agent=page-build-handler`,
`pipeline=content`, `priority=25`, filed against the page **containing** the
link, `item_key` = `dead_fragment_link:page_component:<page>:<slot>:<href>`.

**The boundary held, and it produced its own positive control.** The `<a href="#">`
noop in the same component was NOT claimed by this arm — it was filed by
`dead_control` at `high`, which is exactly the division of remit the header
asserts. One fixture proved both "mine fires on mine" and "mine leaves the
neighbour's alone".

**3. Retraction proven by a before/after with one variable.** Deleted the two
items, added `<div id="zzz-induced-dead">` to page A — **repairing case 1 only**,
leaving its href in place — and refired. Result: **1** item, and it is the
cross-page one. Same binary, same run, same page, same data except one `<div>`.
So the arm is genuinely resolving fragments against document ids rather than
pattern-matching hrefs, which no unit test can establish about the deployed code.

**4. Verifier: SQL validated in both directions, Go function NOT yet executed —
stated as a gap rather than glossed.** Its three query shapes were run against
the live fixture: href-presence returns `t` for the rendered href and `f` for an
absent one; `concatPageHTMLByPath`'s normalisation resolves
`/zzz-induction-b.html` to page B's 106-byte document, which contains the live id
and not the dead one — the exact discrimination the verifier branches on. What
this does NOT prove is `VerifyDeadFragmentLinkResolved` itself running: it is
called only by `CompleteWorkItemAction`, and the only live callers are the
dispatch loops. `build-dispatch-loop` takes `item_domain='build'` and these items
are `content`, so reaching it would have meant spawning `page-build-handler`
against a pool-site scratch page — more side effect than the evidence is worth.
**Owed: the first real completion of a `dead_fragment_link` item exercises it.**

**5. Fleet re-measured post-roll:** 67 fragment-bearing hrefs, **0 findings** —
unchanged from pre-roll, as predicted.

**6. Fixture removed.** Pool site back to 0 pages / 0 work items, one-shot task
deleted, no `dead_fragment_link` rows anywhere. Verified in the same statement.
The three junk items the full 32-check run filed against the pool site
(`needs_rerender`, `nav_drift`, `dead_control`) were deleted with it — they are
artefacts of my fixture, not findings about anything real.

## 2026-08-07 — CORRECTION to my own framing: I avoided 093's ENABLEMENT trap and inherited its CADENCE one

Re-verified on `v1.0.1262` (a second roll since): arm intact on both replicas,
`dead_fragment_link` 10 / verifier 2 / `SplitFragment` 2, positive control 10,
negative 0. No regression.

Then I asked whether any REAL finding had appeared, and the answer was more
interesting than the zero.

> **CORRECTED — "it rides an already-enabled check, so it cannot land inert"
> (PLAN, the register entry, the commit message, and the 016b corollary) is TRUE
> ABOUT ENABLEMENT AND MISLEADING ABOUT CADENCE.** The check is genuinely enabled
> in `completeness-discovery-agent`'s `checks` array — that part stands, and it is
> why no config change was needed. What I did not measure before asserting it was
> **how often that agent runs.** Measured now:
>
> ```
> items created by completeness-discovery-agent, by day, last 21 days:
> 08-05: 9 (2 sites) · 08-04: 256 (6) · 08-03: 266 (5) · 08-02: 5 (1)
> 07-31: 6 (1) · 07-29: 9 (1) · 07-22: 7 (1) · 07-20: 8 (1) · 07-17: 46 (3)
> ```
>
> **Nine days out of twenty-one, 1–6 sites each time, most recently 2026-08-05 —
> i.e. manual/one-shot dispatch, not a schedule.** `improvement-sweep`, the thing
> that would drive it fleet-wide, is `enabled=f` and last fired 2026-05-02
> (`bugs_open/083`/`116`). So **zero `dead_fragment_link` items exist and that is
> NOT evidence the fleet is clean** — no dispatch has touched a real site since
> the arm shipped. The only run it has ever had is the one I induced.
>
> This is `bugs_open/093`'s shape at one level out, and the distinction is the
> transferable part: **an arm on an enabled check escapes "nobody switched it on"
> but inherits whatever drives the agent.** "Enabled" and "driven" are two
> questions and I answered one of them. The cheap check I skipped is the query
> above — item counts by day for the producing agent — which takes one query and
> would have stopped me writing "cannot land inert" in four places.

**What is actually true, stated to survive:** the arm is live, correct and
proven-by-induction; it will run the next time ANY lane dispatches
`completeness-discovery-agent` at a site, which happens every few days; fleet-wide
coverage needs either the improvement sweep re-enabled (`083`/`116`, owner ruled
staged supervised re-enablement on 08-06) or per-site dispatches. **Fleet evidence
for "no dead fragments today" comes from the offline harness (67 hrefs, 0
findings), not from the queue**, and that distinction is now the one to quote.

**Still owed, unchanged in substance:** `VerifyDeadFragmentLinkResolved` has not
executed (needs a real completion). Now with a second, smaller owed item: the
first real dispatch that includes a site with fragment links is the arm's first
production exercise — worth watching, and worth *not* reading its silence as a
clean bill of health.

---

## 2026-08-07 (later) — state re-verified unchanged; and 213 lands on this lane's registry

Picked the lane up cold from the handoff. **Nothing in it needed correcting** —
both owed items are still owed and both are still passive waits.

**Re-verified live, `v1.0.1262`, both replicas** (same exec, four strings):
`dead_fragment_link` 10 · `VerifyDeadFragmentLinkResolved` 2 · positive control
`phantom_internal_link` 10 · negative control 0. Image confirmed at the pod spec
(`docker.io/aqls/agent-chassis:v1.0.1262`), not inferred from the makefile.

**Cadence, re-measured — unchanged, so the handoff's framing holds:**

```
completeness-discovery-agent, items by day: 08-05: 9 · 08-04: 256 · 08-03: 266 …
```

Still nothing after 08-05, i.e. **no dispatch since the arm rolled**, and
`SELECT … item_type='dead_fragment_link'` returns **0 rows**. That remains "it has
not looked", not "the fleet is clean". [MEASURED 2026-08-07]

### The new thing: `bugs_open/213` is about the registry this lane just joined

Filed today by the `bugfix_122` lane: two producers file under one `item_type`,
the registered verifier implements only one producer's predicate, and the other
producer's items close `complete` untouched. **071 owns the newest verifier in
that registry**, so the first question is whether our own arm is exposed.

**It is not, today.** Three checks, all measured:

1. **One producer in code.** `grep -rn dead_fragment_link --include=*.go` over
   `platform/ internal/ pkg/` returns only the check pair that owns it
   (`check_phantom_internal_links_fragments.go`, `check_phantom_internal_links.go`)
   plus doc comments. No second filer.
2. **No design-audit route.** `designItemTypes`
   (`write_audit_findings_action.go:111-119`) maps seven categories onto five
   item_types; `dead_fragment_link` is not among them. Only
   `hardcoded_section_colors` — 213's case — collides with a registered verifier.
3. **No config producer.** Zero live agent definitions file a verified item_type
   via `create_work_item`. **This zero is disconfirmable**: the same query without
   the item_type filter returns **11** config-driven producers (improvement-loop,
   claims-auditor, domain-strategist, …), so the query shape does match when there
   is something to match.

### But the exposure is structural, and this is the part worth carrying to 213

`CreateWorkItemAction` sets `createdBy: source` where
`source, _ := config["source"].(string)` and falls back to `params.AgentType`
(`create_work_item_action.go:129-131, 283`). Two consequences, both measured:

- **`created_by` cannot enumerate producers.** `params.AgentType` itself falls
  back to the literal `"generic"` (`agentbase/agent.go:158`,
  `coordinator.go:3482`, `processor.go:909`). `generic` carries **20+ item_types**
  including `phantom_internal_link` (45 rows), and two live definitions
  (`claims-auditor`, `grounded-explainer`) file with an **empty** `source`. So
  `count(DISTINCT created_by) = 1` is **not** evidence of a single producer —
  distinct real producers collapse into one label. 213 used
  `spec->>'audit_source'`, which is right for producer B specifically, but there
  is no general producer field to generalise it with.
- **The producer set is not enumerable from code at all.** Any agent definition
  can file any `item_type` with any `source` — DB config, no code change, no
  registration. So `dead_fragment_link` could acquire a second producer tomorrow
  and `VerifyDeadFragmentLinkResolved` would grade it against the fragment
  predicate regardless of what it described. **Our arm is safe by luck of nobody
  having done it, not by construction.**

That bears on 213's fix ordering: its candidate 3 ("make a verifier declare the
producers it speaks for") **cannot be satisfied from a code-side list**, because
the producers live in config a Go allow-list cannot see. Enforcement would have to
sit at *creation*, which is closer to its candidate 1.

**Not written into `bugs_open/213`:** its author (`bugfix_122`) was writing that
lane's handoff **26 seconds** before I checked, and 213 is still **untracked**.
Appending to a live session's uncommitted file is the same-file-passenger case in
LANDMINES — my section would be lost on their next Write, or theirs on mine. Left
here, flagged for handover instead. **[UNVERIFIED whether 213's author agrees with
the qualification above — it has not been put to them.]**

213's own `090` (`84c3da66-06c0-41a5-94dc-21fbf71260f0`) had **no orchestration row
yet** when I looked, which per CLAUDE.md is latency, not a dropped dispatch. Not
retried.

---

## 2026-08-08 — the roll, and the handoff's stated blocker is WRONG

**Re-verified on `v1.0.1264`** (fresh pods, started 13:08Z), both replicas, five
strings: `dead_fragment_link` 10 · `VerifyDeadFragmentLinkResolved` 2 ·
`SplitFragment` 2 · positive control `phantom_internal_link` 10 · negative control
0. **Three rolls, no regression.** Still 0 `dead_fragment_link` rows; still no
`completeness-discovery-agent` dispatch since 08-05. [MEASURED 2026-08-08]

Then I went after owed item 1 properly instead of waiting for it.

> **CORRECTED — the handoff's reason the verifier cannot run is FALSE.** It said:
> *"`build-dispatch-loop` takes `item_domain='build'` while these items are
> `content`."* There is **no such filter.** Measured: that loop's only
> `load_work_items` step carries **no `item_pipeline`, no `item_domain`, no
> `handler_agent`** — and the Go filter is optional
> (`load_work_item_actions.go:635-673`, applied only `if pipelineFilter != ""`).
> The loop therefore loads items of **any** pipeline for the site.
>
> **The real gate is `status`**: `load_work_items` selects
> `wi.status IN ('triaged','approved')` (`:653`), and every discovery item —
> including ours — is filed `Status: "detected"`
> (`check_phantom_internal_links.go:175`). A `dead_fragment_link` item would sit
> at `detected` until something triages it. `emit_imagery_items_action.go:18`
> names the split: *"'triaged' (build path auto-dispatch) vs 'detected' (loop
> triages)"*.
>
> The cheap check that would have caught it: dump the loop's own config
> (one query) instead of reasoning from the item's pipeline. Logged in
> `WRONG_CALLS.md`.

**Also corrected: `dead_fragment_link` is not always `content`.** I misread
`check_phantom_internal_links.go:145` as the fragment arm's routing; that line
belongs to `unbuilt_internal_link`. The `dead_fragment_link` block (`:133-142`)
only adjusts priority, `spec["fix"]` and `summary` — routing falls through to
`routeBySurface` (`:195-201`): **chrome/`site_component` → `nav-link-fixer`,
pipeline `build`**; page/`page_component` → `page-build-handler`, pipeline
`content`. So a chrome-surface dead fragment lands in the pipeline where the
verifier path demonstrably works.

### The verifier registry DOES work — measured, including the refusal direction

`result ? '_verification'` (note the underscore — `verification` returns 0 and
reads like "never ran"): **11 items fleet-wide**, and they are informative:

| pipeline | detected | triaged | complete | carry `_verification` | total |
|---|---|---|---|---|---|
| build | 118 | 2 | 2,447 | **11** | 3,420 |
| content | 25 | **0** | 29 | **0** | 321 |
| design | 31 | 0 | 2 | 0 | 121 |

One of the 11 is `literal_markdown`, **`failed`**, 2026-08-07 — a completion the
verifier **refused** in production (the `bugfix_201` lane's `45e0020af`). So the
registry's dangerous direction is proven live by another lane, which is the thing
this lane's owed item was worried about.

**But zero `content`-pipeline items have ever been closed through the verifier
path — 0 of 321** — and content currently holds 0 `triaged`/0 `approved`. The 29
complete content items were closed by `revalidate_review_queue` (a `revalidation`
block in `result`, not `_verification`) — the "count the CLOSERS" landmine's own
worked case, appearing again.

**So owed item 1 restated, accurately:** the verifier is reachable — via
`build-dispatch-loop` → `process_item` (a `loop` action) → sub_workflow step
`mark_complete` → `complete_work_item`. It needs a `dead_fragment_link` item that
(a) exists at all and (b) reaches `triaged`. A chrome-surface one is the likelier
first exercise, because it lands in `build`.

**Method note:** the nested `mark_complete` step is invisible to a top-level
`jsonb_each(default_config->'workflow'->'steps')` scan — my first "who calls
`complete_work_item`?" query returned only `report-`/`diagnose-dispatch-loop` and
**missed `build-dispatch-loop`, the one that matters**. Steps inside a `loop`
action live under `config.sub_workflow.steps`.

### `improvement-sweep`: this lane has been quoting ONE of two disagreeing columns

```
name=improvement-sweep  enabled=f  target_agent_type=improvement-loop
last_triggered_at = 2026-05-02 10:11:07+00
last_completed_at = 2026-08-05 12:24:20+00   <-- three days before this session
```

The lane (and 016b, and my memory line) repeat *"last fired 2026-05-02"* from
`bugs_open/083`/`116`. That is `last_triggered_at`. **`last_completed_at` is three
months later**, and 08-05 is also the day `completeness-discovery-agent` last
filed (9 items). **[UNRESOLVED — I did not establish what wrote that stamp, and
`enabled=f` means the scheduler did not trigger it.]** Do not repeat either column
alone; quote both, or check what actually ran.

> **MISSTEP, caught before it was written anywhere durable.** To corroborate a run
> I queried `orchestration_states WHERE workflow_plan::text ~ 'improvement-loop'`
> and got **7 COMPLETED rows dated today**, which I briefly read as "the loop is
> running after all". They are **council-gate** runs — `current_step` is
> `complete_approved` / `complete_revise` — that merely *mention* the string in
> their payload. A text match over a jsonb blob proves the spelling occurs, not
> that the agent ran. (Also: `orchestration_states` has **no `agent_type` column**,
> and `client_id` is only `demo_client`/`system`, so my first attempt answered a
> question I had not asked. Retention reaches back to 2026-07-13, so an absence
> there *would* be meaningful — if asked correctly.)

---

## 2026-08-08 (later) — the owed verifier ran, in production, and REFUSED. All three branches now proven.

Chassis `v1.0.1264`, both pods started 13:08Z (≫300s, so no spawn-drop window).
Owner's instruction was "go ahead with nav-link-fixer", and on being shown the
side effect below, "do both": exercise the verifier cleanly first, then the full
handler recipe. In the event the second half **fired by itself** — see the
timeline.

### The fixture — chrome surface this time, which had never been induced

Previous induction (08-06) was `page_component` only. `site_component` is the
other half of the arm and it is judged by a *different rule* (rule 2, "resolves if
ANY page does"), routed by a different branch (`routeBySurface` → `nav-link-fixer`
/ `build` / 40−10) and verified down a different SQL path (`concatAllPageHTML`,
not `concatPageHTML`). None of that had ever executed.

Pool site `pool-ai-agents.internal` (`29e0ffc4-…`), measured clean first:
0 pages / 0 page_components / 0 site_components / 0 work items.

- one page `zzz-frag-chrome-a` with a component carrying `id="zzz-page-anchor-live"`
- one `site_components` row, slot `footer`, with **two** anchors:
  `#zzz-page-anchor-live` (resolves on page A → must be SILENT) and
  `#zzz-induced-chrome-dead` (resolves nowhere → must FIRE)

Both controls in one run, so the zero half is disconfirmable by construction.

### Result: exactly one item, routed exactly as the corrected handoff predicted

```
item_type   dead_fragment_link      surface site_component   slot footer
href        #zzz-induced-chrome-dead
pipeline    build     status detected    severity low    priority 30
handler_agent nav-link-fixer
item_key    dead_fragment_link:site_component:footer:#zzz-induced-chrome-dead
```

Priority **30** = `routeBySurface`'s 40 minus the fragment arm's `priority -= 10`.
The resolving control filed nothing. Two junk items from unrelated checks
(`needs_internal_links`, `needs_rerender`) — fixture artefacts, deleted.

**This confirms by observation what 08-08's earlier correction had only read off
the source:** a chrome-surface dead fragment routes to `nav-link-fixer`/`build`,
not to `page-build-handler`/`content`.

### Timeline of the completion runs (all times UTC, 2026-08-08)

| time | what | verdict |
|---|---|---|
| 15:16:54 | one-shot `completeness-discovery-agent` at the pool site | 1 `dead_fragment_link`, `detected` |
| 15:33:54 | probe fire #1 | **FAILED to reach the verifier** — see misstep below |
| 15:35:24 | probe fire #2 → `complete_work_item` | **REFUSED**: `defect_persists` |
| 15:36:25 | `build-pipeline-trigger` → `build-dispatch-loop` → **`nav-link-fixer`** (unprompted) | handler rewrote the chrome |
| 15:36:31 | that loop's `mark_complete` → verifier again | **verified**, branch 1 (href gone) |
| 15:40:24 | probe fire #3, href replanted with the id now present | **verified**, branch 2 (fragment resolves) |
| 15:41:54 | discovery re-run over the repaired state | **0** `dead_fragment_link` |

### 1. The refusal — the assertion the handoff said was the point

```
status        detected -> triaged      attempt_count 0 -> 1      claimed_by cleared
error         completion blocked: post-fix verification found the defect still present:
              href "#zzz-induced-chrome-dead" is still rendered and
              #zzz-induced-chrome-dead still resolves to no element
result._verification  {"status":"defect_persists","item_type":"dead_fragment_link","detail": …}
```

Corroborated by the verifier's **own** log line — the `logger.Info` at
`check_phantom_internal_links_fragments.go:290`, which sits *past* both queries, so
it cannot fire unless the function body ran:

```
"msg":"dead_fragment_link verifier: fragment still unresolved"
pod agent-chassis-dc56548fb-zn2bl  orchestration 11e14fc2-d4bd-4575-ac04-40f0146bfb68
step_name complete_item  action complete_work_item  item_id 53983137-…  href #zzz-induced-chrome-dead
```

**`VerifyDeadFragmentLinkResolved` had never executed before this line.** The owed
item is discharged.

### 2. The full recipe ran ON ITS OWN, because a refusal promotes the item

This was not staged and it is the most useful thing in the session.
`failUnverifiedCompletion` sets `status = 'triaged'` on a retryable refusal
(`complete_work_item_verification.go:224-231`). Our item was `detected`; the
refusal made it **`triaged`**, i.e. dispatchable. 61 seconds later
`build-pipeline-trigger` (120s, enabled) selected the pool site, `build-dispatch-loop`
claimed the item and spawned its real `handler_agent`, `nav-link-fixer`.

What the handler actually did (from the item's own `result.response`):

```json
"template_fix_result": {"reason":"no header/footer component templates assigned to site","updated":0}
"rerender_result":     {"success":true,"rendered":{"header":true,"footer":true}}
```

`render_site_components` force-rerendered header and footer from **generic**
templates — wiping the planted anchor — and `mark_complete` then ran the verifier a
second time, which took branch 1 and agreed:

```
_verification {"status":"verified","detail":"href \"#zzz-induced-chrome-dead\" is no longer rendered on this site_component"}
status complete   handled_by build-dispatch-loop
```

So the handler leg is proven end to end for this item type, by the real dispatcher,
without my having fired it.

### 3. Branch 2 — the one that proves it resolves IDS, not hrefs

Branch 1's agreement is weak evidence about the predicate: "the href vanished" is
a string test. So I replanted the identical href into the footer, with
`<div id="zzz-induced-chrome-dead">` now present on page A, and re-fired:

```
_verification {"status":"verified","detail":"fragment #zzz-induced-chrome-dead now resolves on the target page"}
```

**Same href string as the refusal; opposite verdict.** The only difference is a
`<div id="…">` on a different row of a different table, reached through
`concatAllPageHTML` (rule 2 — any page on the site). A verifier that pattern-matched
the href could not have produced both answers. All three of the function's exits are
now demonstrated live: refuse · agree-because-gone · agree-because-resolves.

### 4. Convergence — and the contrast with `bugs_open/220`

Re-ran discovery over the repaired state (href rendered, id present): **0
`dead_fragment_link`**, while the same check on the same site had filed exactly 1 an
hour earlier. So the zero is disconfirmable, and the **check** and the **verifier**
agree on the same live data despite building "the document" through different SQL
(the discovery loader vs `concatAllPageHTML` + `concatSiteComponentHTML`).

`bugs_open/220` describes the same dispatcher, the same `mark_complete`, and a loop
that reads green for ever because nothing re-checks. The difference here is only that
this `item_type` has a registered verifier. That is direct evidence for the verifier
registry as 220's fix shape, from a different lane's item type.

### Missteps, in the order I made them

**(a) I authored the probe with a literal UUID in the step's `config`, and it could
not be read.** `complete_work_item` returned
`input extraction failed: missing required fields: [work_item_id]` — which reads like
"you forgot it" when it is sitting right there in the config. `ExtractActionInputs`
only resolves a config value when it is a **multi-segment dot-path**
(`action_inputs.go:472-488`, "Strategy 0"); everything else comes from an aggressive
recursive search of `collectedData` (Strategy 2). A literal is therefore never a
value. Fix: put the value in the scheduled task's `input_data` and point the config
at `input_data.<key>`. Now a LANDMINE.

**(b) I told the owner the pool-site handler run "can error before mark_complete"
because the site has no templates. It did not — it succeeded.** I inferred that from
`fix_nav_link_templates` having nothing to fix; but `render_site_components` falls
back to generic templates and rendered both slots. Logged in `WRONG_CALLS.md`: a
stated *reason a thing will fail* is a prediction, and I offered it as a
decision-input to the owner without checking it.

**(c) `href_still_rendered` came back 0 when I expected 1**, four minutes after the
refusal — I briefly read it as a bad assertion. It was true: `nav-link-fixer` had
rewritten the footer 6 seconds after being dispatched. The tell was
`site_components.updated_at = 15:36:25`, which I only looked at second. **On a shared
cluster an assertion that "fails" may be reporting a change someone else's machinery
made while you were typing** — read the row's `updated_at` before doubting the query.

### Two things about the completion record that will mislead somebody

Both found by reading the row after each run rather than trusting the status.

1. **A successful completion does NOT clear `error`.** The success UPDATE
   (`load_work_item_actions.go:941-949`) writes `status`, `result`, `completed_at`,
   `handled_by` — and leaves `error` holding the earlier refusal text. Our item
   finished as `status=complete`, `_verification.status=verified`, and
   `error = "completion blocked: … the defect still present …"` **at the same time.**
   Any audit keyed on `error IS NOT NULL` will call this item broken.
2. **`result._verification` is OVERWRITTEN by each attempt, so a refusal leaves no
   durable trace in `result`.** The `defect_persists` verdict from 15:35 was gone by
   15:36. "How often has a verifier refused?" is therefore **not** answerable from
   `result` — only from `error` (stale-prone, per 1) and pod logs (they rotate). The
   11-item `result ? '_verification'` census this lane has been quoting counts
   **surviving** verdicts, not verifications performed.

### The side effect the owner accepted, and its cleanup — NOT FINISHED

`nav-link-fixer`'s last two steps render a JS snippets bundle and `git_commit` it,
and `render_js_snippets_for_site_action.go:86-94` returns a `files` map **even for a
site with zero snippets**, so the commit is unconditional. Dispatching it at the pool
site therefore pushed a real file for a domain that does not exist:

```
repo gqls/sites   path pool-ai-agents.internal/assets/js/snippets.js
commit "Update JS snippets bundle" 15:36:28Z   GH Actions run 31264883288 succeeded 15:36:30Z (so it also synced to B2)
```

**Still present.** My `gh api -X DELETE` was refused by the tool sandbox, and I did
not work around it. Owed: delete that path from `gqls/sites` (and confirm at the B2
origin with a cache-buster — the `b2 sync --skip-newer` landmine bites reverts).

### Teardown

Ran in one transaction with an all-zero post-assertion: pool pages 0 · pool
page_components 0 · pool site_components 0 (the handler had created a `header` row
too) · pool work items 0 · **`dead_fragment_link` fleet-wide 0** · lane one-shot
`scheduled_tasks` 0 · probe `agent_definitions` row 0.
