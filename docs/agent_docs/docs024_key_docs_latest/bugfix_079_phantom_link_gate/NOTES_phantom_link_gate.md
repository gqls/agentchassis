# NOTES — bugs_open/079 phantom link gate

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-26 — coverage check first

079 was filed by `tool_suggester_phantom_links` while closing 029, and it overlaps almost
exactly with `bugs_open/071` (same file, same guard, same `valid :=` line). So the first
question was not "how do I fix it" but "is someone already on it".

All three adjacent workstreams disclaimed it **in writing**:

- `tool_suggester_phantom_links`, `README_where_we_are.md:99` — *"I have not quietly folded
  it into this change, because making that check block publication is a fleet-wide decision
  that could stop pages shipping for a reason nobody has measured yet."*
- `brochure_component_library` (files 071), `HANDOFF_2026-07-26_continue_here.md` §6 routes
  it to *"platform thread; candidate 1 (persist warnings) is small and worth doing alone"*.
  071 is not on their next-action list.
- `cta_link_integrity` closed 049 and wrote into 071: *"This is evidence, not a competing
  fix."*

No commit has ever touched `validateInternalLinks`, the severity, or the `valid :=`
predicate. Working tree clean of edits to it.

**Trap worth recording:** `scripts/who-owns.py 071` looks alarmingly busy — CLOSED, council
APPROVED, tombstone verified live. All of that belongs to
`bugs_closed/071_…agent_job_cleanup…`, a **different** bug sharing the number. The doubled
numbers are documented in `bugs_closed/README.md`; resolve by slug. I nearly read that
activity as "someone is fixing my bug".

## 2026-07-26 — the measurement, and what it ruled out

079 said explicitly: do not promote the severity without measuring first. Two things I had
wrong going in:

1. **I expected `collected_data` to be pruned at ~24h** (071 gap 3 says so, and it is true
   of the retention *policy*). In fact `orchestration_states` reaches back to 2026-07-13 —
   13 days. The census 079 asked for was possible after all. Had I trusted the doc I would
   have chosen a fix candidate with no data at all.
2. **I expected a mix of repairable and invented targets**, because 049's post-deploy census
   found 8 extension-less targets that resolve at `.html`. In the pre-deploy gate's own
   sample, **15 of 15 were pure inventions** — nothing resolved in any form. The repairable
   class is real but it is not what the gate is mostly seeing.

The numbers: 16 validated builds in the window, 3 carried phantom links (19%), 17 instances
/ 15 unique targets, and all 3 deployed with `valid: true`. Both affected pages were
**homepages** — oufe.com and webdesign.co.uk. That is what killed the "promote to error"
candidate: on `!valid` the action returns `(nil, error)` and routes to `mark_needs_review`,
so the page never saves and never deploys. Two homepages would not have shipped.

`bugs_open/083` killed the work-item candidate independently: `phantom_internal_link`
detected 22 times, fixed **zero** times ever, 98 rows fleet-wide stuck at `detected`,
because the only promoter lives in the disabled `improvement-sweep`.

So the repair has to happen in-band at the gate or it does not happen.

## 2026-07-26 — the upstream finding (out of scope, filed separately)

While tracing where the invented links come from: `prepare_link_context` IS wired into
`page-content-writer` and is supposed to hand the model the site's real page list. It finds
nothing. `db_sync.pages` and all three fallbacks are absent from that workflow's
`collected_data`, so `page_count: 0`, so `link_constraint_text` is empty, so the prompt's
`{{if .link_context.link_constraint_text}}` guard elides the whole "ONLY link to these
pages" block. **20 of 20 recent writer runs: zero pages.** The writer is unconstrained by
construction.

**Misstep avoided by checking, not assuming:** my first instinct was that
`InjectLinkConstraints` (defined, never called) was the missing piece and wiring it was the
fix. It is not — it is dead code duplicating `prepare_link_context`, which already runs.
Wiring it would have added a second, competing implementation of the same prompt block.

Filed separately rather than folded in: different mechanism, different file, different
agent's path, and it is a content-generation behaviour change nobody has measured. Noted
into `bugs_open/071` under its candidate 4, which owns the writer-side class.

## 2026-07-26 — MISSTEP: I wrote a test that could not fail, and nearly shipped it

After the code was green I ran the induced-fault probe (disable `RepairPageLinks`, expect
every repair assertion to fail). It reported **4 failures out of 8 repair tests**. I read
that as "4 tests are vacuous" and started looking for what was wrong with them.

That reading was wrong, and the truth was worse. The four "passing" tests **never ran**.
`TestRepairPageLinks_RewriteEmitsTheStoredURLNotAConstructedOne` indexed `repairs[0]`
without a length check; with the fault induced, `repairs` was empty, `repairs[0]` panicked,
and the panic took down the whole test binary — so every test declared after it was silently
skipped. `go test` without `-v` prints only failures, so a run that executed 4 of 12 tests
looked identical to a run that executed all 12.

Fixed by guarding the index with `t.Fatalf`. Re-probed: **8 of 8 repair tests now fail**
under the induced fault, while the four "must not change anything" tests (byte-identical
clean input, non-page scopes, runtime-fill exemption, empty index) correctly still pass —
which is the discrimination that makes the probe worth running.

Two transferable lessons, and the second is the one I did not know:

- A green suite proves nothing until you have watched it go red. Already in `WRONG_CALLS`
  in other forms; this is another instance.
- **A panic in one test masks every test after it, and the default output makes that look
  like success.** Read `=== RUN` lines, not the FAIL count. Any `slice[i]` in a test needs a
  `len()` guard with `Fatalf`, or it is a hidden kill switch for the rest of the file.

## 2026-07-26 — a second fail-open found while making the first one safe

`loadValidPagePaths` swallowed a query error and returned an **empty** page set. While
findings were only warnings that was survivable — it produced a spurious phantom warning for
every link on the page, noise and nothing more. Once the findings drive a rewrite it is not
survivable: an empty set means *every link is a phantom* and the repair pass would strip the
lot from a page whose links were all fine.

It also never checked `rows.Err()`, so a mid-iteration failure silently **truncates** the
page list. That is the same hazard in disguise and slightly worse, because it would unlink
only *some* links — much harder to notice than losing all of them.

Now returns `(index, ok)`; both detection and repair are skipped when not ok. A NULL
`pages.url` is skipped rather than treated as a load failure — checked the live data first
(0 NULLs across 408 active pages today), but the column is nullable, and treating one
malformed row as "list untrustworthy" would disable link checking for that whole site.

## 2026-07-26 — committed, submitted, NOT live; and two multi-session collisions

Committed `43f254be5` (code + tests + docs; scope report clean, 7 files, no passengers) and
`31d8ac7dc` (gofmt — the pre-commit pattern check caught a struct-tag misalignment that the
build gate would have rejected in CI).

Council submitted: **`SUBMISSION_CORR = 97904892-5c09-4782-aeda-37dd944abdfc`**. All six
`grounded_in` quotes machine-verified byte-identical against the pre-fix file (`git show
f804b84ed:…`) before submitting — a trimmed quote manufactured a false MEDIUM objection on a
previous run, and reviewers cannot open the file to check. No orchestration row after 15
minutes; that is the documented queue latency, **not** a dropped dispatch. Not resubmitting.

**Live state, measured not assumed.** The chassis pod runs `v1.0.1170`, built by another
session while I worked, and it does **not** carry this fix:

```
strings /app/agent-chassis | grep -c "CONTENT_LINK_REPAIR_DETAIL"     -> 0   (my new string)
strings /app/agent-chassis | grep -c "CONTENT_VALIDATION_BLOCKER_DETAIL" -> 2   (positive control)
strings /app/agent-chassis | grep -c "repair_internal_links"          -> 0   (my new config key)
```

That is a *discriminating* pre-state, which is the point of running it before the roll: the
control proves the grep works, and the two zeros prove the fix is absent. Post-roll the same
three commands must read ≥1, ≥1, ≥1. **079 therefore stays OPEN** — the bar is fixed AND
live, and the defect is still reproducible in production until an image ships.

Build held: the owner reported another thread mid-deploy, and racing on `IMAGE_TAG` is the
multi-session hazard this repo has a whole handoff about.

### Two collisions, both while this task was in flight

1. **My `016b` §9 append was swept into another session's commit** (`d5988a8ed`, a
   `bugs_open/006` closure) before I could commit it. Nothing lost, forward-only holds. This
   is precisely the case CLAUDE.md documents — commit-per-task stops *me* sweeping *others*,
   and cannot stop a session running `git add -A` from sweeping *me*. The only real defence
   is committing sooner, and I had a ~4-minute window open.
2. **Number collision on a fresh bug file.** I checked, found 090 free, wrote the file — and
   another session filed *their* 090 sixty-seven seconds before my commit landed. Renumbered
   mine to `092` (`cf2cafcdd`) rather than leave a sixth doubled number in a scheme where
   `bugs_closed/README.md` already lists doubled numbers as a standing trap. **Checking a
   number is free is not the same as reserving it**; on a busy day the check is stale before
   you finish writing the file. Cheap to undo at 67 seconds old, permanent if left.

## 2026-07-26 21:1x–21:3xZ — LIVE in v1.0.1171; first induction attempt died before reaching the code

Another session's build rolled `v1.0.1171` and it carries this fix. Pod-grep against the
baseline recorded above — this is the discriminating pair, not a bare grep:

| string | pre-roll (v1.0.1170) | post-roll (v1.0.1171) |
|---|---|---|
| `CONTENT_LINK_REPAIR_DETAIL` (new) | 0 | **1** |
| `repair_internal_links` (new) | 0 | **1** |
| `link check and repair SKIPPED` (new) | 0 | **3** |
| `CONTENT_VALIDATION_BLOCKER_DETAIL` (control) | 2 | 2 |

**A marker I nearly used and did not:** "the old policy comment `improvement loop resolves it` is
gone" reads 0 on both binaries — **comments are not compiled into the binary**, so it can never be
anything but 0 and proves nothing whatsoever. It is the same vacuous-marker shape already logged
against `052`. Only compiled strings — error codes, config keys, log messages — discriminate.

### The induction, attempt 1: FAILED before it reached the gate

Dispatched `page-build-handler` for `webdesign.co.uk / learn-design-digital-grain`
(corr `a1dfbf68`). Result:

```
FAILED | spawn_content_writer | Request 295ff5da-… timed out after 3 retries
```

It never got as far as `validate_content`, so **it proves nothing about the fix either way** —
neither that it works nor that it does not. Recording it because a failed run that never reached
your code is the easiest thing in the world to quietly discount, and the temptation is to treat
"no contradiction" as "confirmation".

Checked whether it was systemic before retrying rather than assuming: **1** spawn timeout
fleet-wide in 2 hours (mine), 8 orchestrations in flight. So not saturation. Not within 300s of a
pod restart either (pod up 21:02:56Z, dispatch ~21:14:55Z). Cause unresolved — plausibly the
`bugs_open/003` spawn-loss class. Retried as corr `df7437f2`.

### Two route findings worth keeping

1. **The work-item dispatcher is PER-SITE** — `load_work_item_actions.go:559`,
   `WHERE wi.site_id = $1 AND wi.status IN ('triaged','approved')`. A `triaged` row therefore sits
   untouched indefinitely until something triggers *that site's* build pipeline. Inserting a work
   item is not dispatching it. To exercise one page, publish to `page-build-handler` directly by
   kcat (envelope in the RUNBOOK).
2. **`page-rebuild` and `page-rerender` do NOT call `validate_page_content`.** Only
   `page-build-handler`, `content-reviewer`, `tool-recreation-handler` and `report-builder` do.
   The obvious-looking `TRIGGER_rerender_page.sh` would have run green and tested nothing —
   the "verify the failing branch" trap wearing a convenient disguise.

### Induction attempt 2 (corr `df7437f2`): the gate RAN live — but on a page with no links

```
current_step  complete_error
valid         true      warnings 1 (short_content)
checked_links 0
links_rewritten 0    links_unlinked 0    link_repairs []
```

**What this DOES prove.** `links_rewritten` / `links_unlinked` / `link_repairs` are keys that
exist **only in the new binary** — the pre-fix action's return map has no such fields. Their
presence in `collected_data` is proof that the new code path executed on a real build in
production, was reached, and returned without error. That is more than pod-grep gives (a string
in a binary is not a code path that runs).

**What it does NOT prove, and this is the honest limit.** `checked_links: 0` — the writer
produced 26 characters and **no anchors at all**, so the repair had nothing to repair. Zero
repairs on a page with zero links is a null result, not a passing test. It is exactly the shape
of evidence that is tempting to write up as "verified live" and must not be.

Two incidental findings, both of them the system behaving correctly:

- `save_sections` then refused: *"page learn-design-digital-grain is rebuild_policy=owned
  (tool/widget)"*. A protective guard doing its job — the page was NOT modified, so the failed
  run cost nothing. But it means **an `owned` page can never serve as an induction target**: the
  build cannot complete past validation.
- A writer returning 26 characters for a real content page is itself suspicious and smells like
  `bugs_open/087` (page-rebuild writer has no section plan). Not chased — out of scope here, but
  noted in case someone sees the same shape.

**Attempt 3** targets `dartsonline.com / new-arrivals` (corr `119e1bb7`): `rebuild_policy=generic`
(so `save_sections` can complete), 2 components carrying anchors, and it is the fix-loop's own
example site rather than a client's. Selection query in RUNBOOK.

## 2026-07-27 ~11:15Z — the council submission was DROPPED, not queued; and why the crafted induction keeps failing

### CORRECTION to yesterday's entry: the council run was dropped

Yesterday I wrote that the missing orchestration row was "the documented queue latency, **not** a
dropped dispatch", and declined to resubmit on that basis. **That was the right call on the
evidence available then and it is now falsified.** Thirteen hours later:

- submission `97904892-5c09-4782-aeda-37dd944abdfc` — still **zero** orchestration rows, zero
  `diagnosis_artifacts`;
- meanwhile **676 orchestrations** were created fleet-wide in that window.

A lane that processed 676 runs is not a lane my message is queued behind. The submission was
**dropped**. The standing advice "a missing orchestration row is latency, not a drop — do not
retry" is sound as a *first* reading and becomes wrong once you can show the lane is draining;
the discriminating check is **not** the age of your own row, it is *whether anything else ran*.
That check costs one query and I should have run it at the 1h mark rather than at 13h.

(Same shape as the `bugfix_006` finding — "the council submission was DROPPED, not queued".)

### Attempts 4–6: why a crafted induction keeps getting dropped

The natural induction is unavailable: overnight the gate ran on **4 more real builds** under the
new binary (the new-binary-only keys are present, so the code path executed), and **every one had
`checked_links: 0`**. The writer is currently emitting no anchors at all, so no phantom can occur
— nothing to repair. I cannot make an LLM invent a link on demand.

So I tried to feed the gate controlled HTML. Three routes, and the pattern in the failures is the
useful part:

| # | route | result |
|---|---|---|
| 4 | `content-reviewer`, plain dispatch | **ran** — but returned `"reason": "no content to validate"` |
| 5 | `content-reviewer` + inline `config.workflow` | **no orchestration row, ever** |
| 6 | brand-new `verify079-gate` agent type | **no orchestration row, ever** |

Attempt 4 ran because content-reviewer has a **live pod**; it found nothing because the action
resolves `html_field` against `collected_data` root and the default is
`page_content.response.page_html`, whereas a dispatch puts your payload at
**`collected_data.input_data.*`** — verified, `collected_data ? 'page_content'` is false. The
`input_fields: ["page_content","site_record"]` in that step's config does **not** lift them.

Attempts 5 and 6 produced no row at all. The discriminator is a **live pod for the agent type**:
`page-build-handler` and `content-reviewer` have running pods and always produced rows; an inline
workflow on this path and a freshly-seeded type with no pod both vanished silently. Worth knowing
before anyone spends an afternoon on `body.config.workflow` here — the chassis honouring it
(`bugs_closed/074`) does not mean this dispatch path will.

Throwaway `verify079-gate` row deleted; test work item `560d50cd-…` cancelled. No litter left.

### Where that leaves the proof, stated exactly

The only remaining route was to add `html_field` to content-reviewer's live step config for ~2
minutes (it ran **once** in 24h, and that run was mine, so the collision risk was ~nil) and revert.
**That UPDATE was refused by the permission layer**, correctly — mutating a live agent definition
is not something to do unattended. Not worked around; handed to the owner.

So the honest ledger is unchanged from yesterday and should not be dressed up:

- transform correct — **proven** (13 unit cases, induced-fault probed 8/8);
- code deployed — **proven** (discriminating pod-grep, `v1.0.1171`);
- code path executes in production — **proven** (5 real builds now carry the new-binary-only
  keys);
- **a repair actually mutating a page — NOT proven.** Zero repairs so far, because zero links
  have been offered to it.

## 2026-07-27 12:23Z — PROVEN END TO END; 079 CLOSED

Owner approved the two-minute config patch. Result, first try:

```
checked_links 3 · links_rewritten 1 · links_unlinked 1

/definitely-not-a-page.html  -> unlink   (text "pricing guide" survives)
/contact                     -> rewrite  -> /contact.html   (the STORED pages.url)
/new-arrivals.html           -> untouched (already valid)
https://example.com/x        -> untouched (external)
```

`agent_error_log` carries `CONTENT_LINK_REPAIR_DETAIL` with both hrefs in `context.repairs` —
`071` gap 3 answered for this class.

The four "must not change anything" unit tests were the ones I cared about most here, and the
live run agreed with them: the two links that should not have been touched were not touched. A
repair pass that over-reaches would have been a worse bug than the one being fixed.

**Config restored and byte-checked**, not merely "restored": snapshot taken before the patch,
compared with `json.dumps(sort_keys=True)` after — IDENTICAL. Saying "I put it back" is not
evidence; the comparison is. Window ~2 minutes on an agent that had run once in 24h (that run
also mine), so the collision risk was as close to nil as this cluster allows.

**Why the induction needed four attempts — the transferable bit.** A crafted dispatch only
materialises if the target agent type has a **live pod**. `page-build-handler` and
`content-reviewer` always produced orchestration rows; an inline `config.workflow` and a
freshly-seeded `verify079-gate` type both vanished with no row and no error. Do not read that
silence as queue latency — check whether *anything else* is running before concluding a message
is merely queued. That is the same check that finally settled the council question.

Also worth stating plainly: **the natural induction never arrived and I stopped waiting for it.**
Five real builds ran the new code and all had `checked_links: 0`. Waiting for the fleet to
produce a phantom would have looked like diligence and delivered nothing, because the writer is
currently emitting no links at all (`bugs_open/092`). Crafting the input was the only route that
could ever have closed this.

---

## 2026-07-28 — REOPENED, by another thread (brochure_component_library). The repair's output never persists on a natural build.

The closure above was falsified today. The repair action is exactly as good as this file
says — and `save_page_sections` never reads `validation_result.clean_html` on the
primary build plan: the structured `sections_metadata` path wins whenever metadata
exists, which `require_sections_metadata: true` guarantees. The repaired string is
discarded on every natural build. First natural production run with real phantom links
(fundamentallyai capabilities, 2026-07-28 10:45Z — the writer finally emitted anchors,
19+ of them invented): repair row logged 10:45:01.347, all 9 dead targets named;
unrepaired components saved 10:45:01.768–.807; all 9 serving 404 from the deployed page.
Second site same day: vonc /about.html, "unlink /how-it-works" logged, href still in the
saved row.

The line above — "Crafting the input was the only route that could ever have closed
this" — deserves its own correction: the crafted route (`content-reviewer`, no
`save_sections` step) was also the only route that could not have EXPOSED this. The
natural induction this file stopped waiting for arrived 24 hours later and refuted the
closure on its first occurrence. Not an argument against crafting inputs — an argument
that a crafted route proves the steps it contains, and the claim must stay that size.

Full mechanism, config citations and fix candidates: `bugs_open/079` REOPENED banner
(moved back from bugs_closed). 016b §9 has the transferable pattern; WRONG_CALLS has the
closure. The fix belongs to whoever picks it up — candidate 1 (repair inside
save_page_sections, where persistence happens) closes the door structurally.

## 2026-07-28 evening — candidate 1 IMPLEMENTED. Committed, submitted, NOT yet live.

Picked up `HANDOFF_2026-07-28_platform_fix_candidate1.md` cold and built it as written.
Session-start checks first: no other session on either action file (`git status` clean on
both, `who-owns.py 079` names this workstream), LLM lane healthy (last `LLM_API_ERROR`
13:13Z, zero in the preceding 90 minutes).

**What shipped into the tree** (commit `5083124e3`, three files, pathspec):
`platform/orchestration/actions/save_sections_link_repair.go` (new) — `repairSectionLinks`
is the pure seam, `repairSectionsBeforePersist` the DB-touching wrapper;
`save_sections_link_repair_test.go` (new, 3 tests); and the wiring +
`saveSectionsLookupPageID` now returning `url` in `save_page_sections_action.go`.
Call site sits between the interactive-tool preservation block and the content-regression
guard, exactly where the handoff put it. `go build ./platform/orchestration/...` clean,
`go test ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/...` all
pass — the shared tree happened to compile, so no `git archive` workaround was needed.

**Blast radius MEASURED before submitting, not asserted** (the 07-28 owner ruling). And the
first measurement was WRONG in the instructive direction:

```sql
-- FIRST GO — 3 rows. jsonb_each over default_config->'workflow'->'steps' only.
--   page-build-handler | page-rerender | tool-recreation-handler
-- I nearly submitted "3 of 6" against a handoff claiming 6.
-- SECOND GO — unfiltered, the string anywhere in the config:
SELECT type, is_active, COALESCE(is_snapshot,false), (deleted_at IS NOT NULL), count(*)
FROM agent_definitions WHERE default_config::text LIKE '%save_page_sections%' GROUP BY 1,2,3,4;
-- 11 rows; 6 live callers + 3 fix-loop rows that merely NAME the action in a footprint map.
```

The narrow query described a small world silently: three agents keep their workflow under a
different container key, so a `->'workflow'->'steps'` walk simply cannot see them. The
handoff's list of six was right and my query was the thing that was wrong. Settled with
`jsonb_path_query_array(default_config,'$.** ? (@.action == "save_page_sections")')`, which
is container-agnostic:

| agent | save steps | validate steps |
|---|---|---|
| page-build-handler | 1 | 1 |
| pageflow-builder | 1 | 0 |
| page-rebuild | 1 | 0 |
| page-rerender | 1 | 0 |
| site-work-orchestrator | 1 | 0 |
| tool-recreation-handler | 1 | 1 |

**That second column is the finding, and it is stronger than the handoff's argument.**
Only 2 of the 6 persistence paths have ANY `validate_page_content` step, so **4 of 6 have
never had dead-link repair by any route at all** — the bug is not merely "the gate's repair
is discarded on the build path", it is "most paths that write body sections were never
within reach of a gate". A candidate-2 fix (repair `sections_metadata` in place inside
validate) is structurally incapable of covering them. Also measured: `repair_internal_links`
appears in **zero** live `agent_definitions` rows, so the step-config key collides with
nothing and the in-code default governs everywhere. "No collision is possible" is a query.

**Council submission** `7c24776e-07f8-4c2e-b1b6-ad3e73c6023c` (default `council-gate`
target, NOT the orchestrator wrapper — the parallelism thread's own handoff says retry that
no earlier than 2026-08-01). Lane LAG 0 at submit; one council already in flight at
`review_editquality`, so mine queues behind it. Budget ~30 minutes.

**Registered as LNK-024** in `docs026_concept_register/register/link-management.md`, status
**built** not deployed, with the landmine written down: *any* transformation applied only to
`clean_html` is discarded by the structured save path — the gate's comment-stripping has the
same defect [INFERRED from the code path, no observed comment damage]. The entry also
carries the open review question (is a persistence-point content transform the right general
pattern, or does each transform belong at its own gate) rather than pretending the scope
question was settled by a bug patch.

**NOT DONE — the fix is inert.** No image built, no roll, nothing verified against a running
pod. `bugs_open/079` stays OPEN and must, because the defect is still reproducible until it
ships. Next, in order: verdict → build/push/deploy with a bumped `IMAGE_TAG` → pod-grep
`"SavePageSectionsAction: repaired dead internal links before persist"` per replica (a
retag is not a rebuild; check `.ID` + `.CreatedAt`) → the zero-LLM live proof the handoff
specifies, a `page_rerender` work item at gamesdesign `bayesian-ranking`, whose stored
sections have carried `href=""` since 07-21 — remembering `handler_agent='page-rerender'`
in the INSERT or the item hard-blocks. Verify the PERSISTED rows, then crawl the SERVED
page. Never the action's return map.

## 2026-07-29 — CLOSED. Proven by a run I did not fire, after the one I did fire proved nothing.

**Council round 2 APPROVED** (corr `7c24776e`, 11 reviewers / 0 unreadable / 4 advisory, none
high). Round 1 was REVISE, gated by `debug_historian` on a HIGH: the plan carried no
deploy-verification step, on a bug whose first closure was inert in production. Fair, and
answered by pre-registering the protocol in round 2.

**The roll happened without me.** By the time I went to deploy, another session had already
built `v1.0.1195`/`1196` from a HEAD containing my commit and rolled the fleet at 22:37:49Z.
Both replicas pod-grepped `marker=1` with the positive control at `1`, against a `v1.0.1194`
baseline of `0`. Round 2 changed only a Go comment and a `_test.go`, so it does not alter the
binary — nothing more to build. **"Committing IS shipping on a shared HEAD" is not a slogan;
it is the delivery mechanism, and it delivered this.**

### The induced repro FAILED, and the reason is the transferable part

The pre-registered zero-LLM test was gamesdesign `bayesian-ranking`, whose stored sections had
carried `href=""` since 07-21. I checked the preconditions properly — `content_data` non-NULL
on all four sections (or the page escalates to the writer and `save_sections` is skipped
entirely), not runtime-fill exempt, not interactive, `rebuild_policy='generic'`, and the step
graph confirmed `reason: section_data_resolved` → `rerender_sections` → **`save_sections`**
(the no-reason branch skips `save_sections` altogether, so the reason is load-bearing).

It ran in 40s, COMPLETED, and the `href=""` count went 2 → 0. **That looked like success and
was not.** Three signals said otherwise:

1. The CTA anchor TEXT was gone too (`Start Ranking Free`, `See How It Works`). Unlinking
   KEEPS inner text — that is the whole design. Text disappearing means something else acted.
2. No `agent_error_log` row with `action='save_page_sections'` existed for that run.
3. The container survived but empty: `brht-cta-row` present, `any_anchor` false.

Cause: `rerender_sections` re-renders each section from `content_data` through the CURRENT
template. That page's `content_data` has `cta_primary_label` and `cta_secondary_label` and
**no url fields at all**, so the template's skip gate (LNK-006) omitted the buttons. The
repair never saw an anchor. **A repro that is regenerated from `content_data` can be destroyed
by the render itself** — check what the template does with the missing field before trusting
one. The 2 unlinks logged against that page at 07:31:46Z were the OUTBOUND seam
(`action='rerender_page'`, LNK-023) acting on the assembled page's chrome — a different call
site, and I would have mis-attributed them if I had not discriminated on `action`.

Side effect I caused and should record: that rerender removed two dead CTA buttons from the
live gamesdesign page. Right outcome per correct-or-absent, but it was a change to a live page
made in service of a test.

### What actually proved it

A **natural** run I did not fire: vetcomparison.uk `index`, 01:58:04.382Z. Five real phantoms
unlinked at the persistence point. It is the exact inverse of this bug's own evidence:

| | the bug (fundamentallyai 07-28) | the fix (vetcomparison 07-29) |
|---|---|---|
| repair logged | 10:45:01.347Z, 10 repairs | 01:58:04.382Z, 5 unlinks |
| components saved | +400ms | +45ms |
| hrefs in saved rows | **all 9 still there** | **all 5 gone** |
| served page | 9 × 404 | all 5 absent |

`/search`, `/about-pricing`, `/about-ownership-disclosure`, `/guides/pet-owner-rights`,
`/claim-listing` — re-probed live, all genuinely 404. Attribution is exact:
`action='save_page_sections'` had been written **0 times before the roll and 1 after**, with
`rerender_page` (20 before / 1 after) as the positive control proving the query works.

### The correction the council forced, which matters more than the fix

`bug_historian` objected that my blast-radius census enumerated `agent_definitions` carrying a
`save_page_sections` STEP NAME — a different question from *who writes
`page_components.rendered_html`*. It was right. Ten Go writers set that column; three persist
LLM prose with no repair at all, including `ApplySectionEditAction`, which
`save_page_sections_action.go:156` itself directs operators to. **The more carefully a session
follows documented practice, the more reliably it bypasses this fix.** Filed as
`bugs_open/136_…section_editor_and_three_siblings…` (note a DIFFERENT `bugs_open/136` exists —
resolve by slug). The round-1 claim "no build path can persist an unrepaired section" was
withdrawn, and the over-broad wording was corrected in the code comment too, because that is
where the next reader would have inherited it.

079 → `bugs_closed/`. 092 and 136 stay open; 092 is the upstream cause of both this bug's
phantoms and gamesdesign's missing url fields.
