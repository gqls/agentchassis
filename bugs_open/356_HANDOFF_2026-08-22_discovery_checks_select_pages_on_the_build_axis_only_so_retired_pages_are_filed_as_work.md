# 356 — discovery checks select pages on the BUILD axis only, so **retired pages are filed as work for handlers that all correctly refuse them**

> ## ⏳ FIXED IN THE TREE, NOT YET LIVE — stays OPEN, 2026-08-22
>
> **The instance fix and the class guard are committed** (`24d0bc251`, council **APPROVED**
> round 1 corr `4cf291a2`, follow-up `fa283256d`). This file stays in `bugs_open/` because the
> bar is **fixed AND live** and half of this is not:
>
> | half | state |
> |---|---|
> | `check_orphan_pages`' lifecycle arm | **Go — INERT until the next `agent-chassis` build and roll.** The defect is still reproducible in production today. |
> | the posture registry + guard | **test-only — live in CI on merge**, protecting the tree now |
> | the other **17** routing gaps | **NOT FIXED.** Declared, counted and named in the registry; each needs its own judgement (§6-B). |
>
> **Do not close this on the roll alone.** §7 STEP 0 is a pod/ancestry check, and the census
> arm needs its disconfirming PAIR. The remaining 17 gaps are the larger half of the ticket.
>
> **The 090 loop did NOT ratify this file** — one run died on infrastructure, one returned
> UNVERIFIABLE. See the banner below before citing anything here as loop-confirmed.

**Filed 2026-08-22** by the session named `bugs_open/298`, taking that file's explicitly
unclaimed adjacent finding. **The damage is LIVE and RECURRING** — the same
archived pages have been re-detected and re-dispatched on every discovery rotation for four
months.

**Lane docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_356_orphan_check_lifecycle_axis/`

> **On the 090 loop, stated plainly because the 2026-07-31 owner ruling requires it.** This
> asserts a cross-cutting structural cause that lives in shared infra, so it went to the
> diagnosis loop **before** the root cause was asserted here. Two runs, and **neither
> confirmed this file. Read this block before citing anything below as loop-confirmed.**
>
> - **Run 1** (intake `4480a3a7`, run `7bac4520`): **FAILED, not refuted.** `call_diagnoser` died
>   on `kafka … topic partition has no leader` — the documented spawn→call handshake race. Its
>   intake row is still parked at `diagnosing` and is deliberately NOT cancelled.
> - **Run 2** (run `cd3ad443`, `FORCE=1`, justified because run 1 left an explicit FAILED row
>   with an infrastructure error — evidence for a retry, unlike a MISSING row, which is only
>   latency): **UNVERIFIABLE, `stopped_by: scope-not-narrowing`.** Its own words: *"Diagnosis NOT
>   confirmed … no fix proposed. Hand to a human with the full trail; do NOT auto-conclude."*
>
> **UNVERIFIABLE is not CONFIRMED and it is not REFUTED — it is "I could not get there".** It
> reached one of this file's central citations independently (quoting
> `AND ` + datahelpers.PageHasShippedPredicateFor("p") + ` from `findOrphanPagesSQL`, and
> `HandlerAgent: "internal-linker"` from `Run`), and named **three gaps** it could not close.
> All three are closed here by first-hand reading, which is the substitution the owner ruling
> requires be stated rather than left silent:
>
> 1. *"the full body of `findOrphanPagesSQL` is not in the bundle"* — read in full, §3. ⚠ Its
>    content search for `PageWantedLivePredicateFor` did not return this file, which it correctly
>    treated as UNKNOWN rather than ABSENT. Note the likely cause and the irony: **the fix had
>    already been committed when run 2 executed**, so the file now DOES contain the symbol and a
>    current index would have returned it — the code index is pinned and lags (`MEMORY`:
>    code-index-says-fresh-while-stale).
> 2. *"the internal-linker remedy path's page predicate is unverified — we have only the step
>    NAME"* — read from the LIVE `agent_definitions` row, quoted in §4 and in the RUNBOOK.
> 3. *"`unresolved` is a status the dispatcher does not claim — no citation in this bundle"* —
>    verified first-hand 2026-08-22: `workItemTerminalStatuses` includes `"unresolved"`
>    (`work_items_common.go:42-48`) and `workItemDispatchableStatuses = {"triaged","approved"}`
>    (`work_items_common.go:174-177`).
>
> **So: the loop did not ratify this file. The evidence below is first-hand and each claim carries
> its own citation; treat it as such, and if you can get a CONFIRMED verdict out of a third run
> with a narrower symptom, record it here — including if it refutes me.**

---

## 1. The one-paragraph version

`pages` carries two independent axes, and the platform's own predicate family says so:
the **BUILD** axis (`PageHasShippedPredicateFor` — "has this ever been served") and the
**LIFECYCLE** axis (`PageWantedLivePredicateFor` — `status = 'active'`, "does the platform
still want it"). `platform/orchestration/datahelpers/links.go` states the contract in terms:
*"Pair this with whichever build-axis arm YOUR question needs; do not expect one combined
helper."* **`OrphanPagesCheck` pairs nothing.** It takes the build arm alone, so an
**archived** page that was deployed once is enumerated as an orphan and filed as a work item
telling a handler to make it *more reachable*. Every one of the three handlers it routes to
already applies a lifecycle arm and therefore refuses. The disagreement is entirely internal
to the platform, and it repeats for ever.

## 2. How it presented — 298's adjacent finding, answered

`bugs_closed/298` recorded, without chasing it: *"15 of 38 completed `internal-linker` items
found NO target page"*, noting that such a run is indistinguishable by `status` from one that
linked successfully. Re-measured 2026-08-22 `[MEASURED]`:

| completed `internal-linker` items | count |
|---|---|
| total | 34 |
| **found NO target page** | **17** |
| found a target page | 9 |
| unreadable (`result` is the spawn record — `bugs_closed/287`) | 8 |

**All 17 of the no-target items name a page whose `pages.status` is `archived`.** Not a
majority — all of them. `site_match` is true on every row, so this is not a site-id mixup.

⚠ The read is `result->'response'->'target_page'`, **not** `result->'target_page'`. The
shallow path returns NULL on all 34 rows and reads as "everything is unreadable". That was
this lane's misstep 1 — see `WRONG_CALLS.md`.

## 3. Root cause

`findOrphanPagesSQL`, `platform/orchestration/actions/discovery_checks/check_orphan_pages.go:204-245`:

```go
FROM pages p
WHERE p.site_id = $1
  AND ` + datahelpers.PageHasShippedPredicateFor("p") + `      // BUILD axis — the only page-row arm
  AND p.url IS NOT NULL ...
```

`PageHasShippedPredicateFor` is `NOT (deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed')`
(`platform/orchestration/datahelpers/links.go:293-295`). It says nothing about `status`. An
archived page that shipped before it was retired satisfies it.

⚠ **The file contains `status = 'active'` twice and neither is on the page row** — both are
on `site_nav_items` (`check_orphan_pages.go:220`, `:226`). A grep-count scores this file 2 and
calls it armed. That is why the audit in §5 was done by reading SQL, not by grepping.

## 4. Why this is ONE producer defect, not three handler defects

**Every one of the three remedy paths already carries a lifecycle arm. The producer is the
sole outlier.** Each verified first-hand 2026-08-22:

| branch | item type → handler | the remedy path's OWN page predicate | what an archived page does there |
|---|---|---|---|
| unflagged | `needs_internal_links` → `internal-linker` | live `agent_definitions` step `load_target_page`: `… AND p.status = 'active'` | target resolves empty → `check_target_found` takes `else_step` → completes having done nothing |
| nav-flagged | `nav_drift` → `nav-updater` | `navPageScopeSQL`, `platform/orchestration/actions/nav_prune_floor.go:128`: `status IN ('active','deployed','pending')` | the rebuild **cannot** add the page → the finding is **unsatisfiable** and recurs for ever |
| blog | `orphan_blog_posts` → `rerender-pages` | `blogPostsQuery`, `platform/orchestration/actions/rebuild_blog_listing_action.go:110-111`: `p.status IN ('active','deployed')` | the listing **cannot** include the page → same unsatisfiable shape |

Read the live `agent_definitions` row for the first one, never
`sql_for_agents/101_internal_linker.sql` — the seed records what the agent *was*.

### The second-order cost: it retires the queue for the pages that ARE affected

`writeWorkItem`'s anti-churn ladder (`platform/orchestration/actions/load_work_item_actions.go:1468-1504`,
read first-hand) counts prior `complete`/`failed` rows on the same `item_key` and at
`terminalCount >= 2` rewrites `item.status = "unresolved"`. **`unresolved` is TERMINAL and
undispatchable** (`work_items_common.go:40-46`, `:172-175`). So the no-op completions are
counted as failed attempts, and `bugs_closed/313` already established that this is what parked
the linker's 20 queued items. Of those 20, **15 name a page that is `active`** — legitimate work,
retired as collateral.

## 5. Blast radius — this is a CLASS, and the orphan check is one of eighteen

`[MEASURED 2026-08-22]` by reading the SQL of all 71 non-test files in
`platform/orchestration/actions/discovery_checks/`. **18 checks can select an archived page
AND route it to a handler.** A representative sample was re-verified by hand against the
source, in both directions (unarmed and clean) — the spot-checks agreed.

Highest blast radius first; these are the fix targets:

| # | check | the page-row arm it actually has | routes to |
|---|---|---|---|
| 1 | `check_sectionless_pages.go:116` | `COALESCE(p.status,'') <> 'deleted'` | `needs_content_page` → `page-build-handler`, **priority 90** |
| 2 | `check_orphan_pages.go:207` | build axis only | `rerender-pages`, `nav-updater`, `internal-linker` |
| 3 | `check_component_standards.go:445` | `p.build_status IN ('deployed','active')` | `page-build-handler` |
| 4 | `check_empty_sections.go:172` | none | `page-build-handler` |
| 5 | `check_literal_markdown.go:392` | none | `page-rerender` / `section-editor` |
| 6–18 | `check_contact_form_undeliverable`, `check_required_fields_missing`, `check_placeholder_contact`, `check_content_duplication`, `check_integrity` (×2), `check_phantom_internal_links` (container scan only — its target set IS armed), `check_hardcoded_section_colors`, `check_component_template_corrupted`, `check_page_component_status_drift`, `check_revenue_shape` (×2), `check_empty_blog`, `check_news_feed` (×3), `check_placeholder_image_in_use` | none, or a `build_status`-only arm | various |

> ⚠ **`check_sectionless_pages`' arm excludes NOTHING.** `pages.status` has exactly two live
> values — `active` 759 / `archived` 65 `[MEASURED 2026-08-22]`. There is no `'deleted'` row and
> there never has been in this vocabulary, so `<> 'deleted'` is a no-op that **reads** like a
> lifecycle filter. Same trap in reverse: `p.status IN ('active','deployed')`
> (`check_content_image_missing`, `check_voice_tells`) works only by accident — `'deployed'` is
> not a `pages.status` value, it is a `build_status` value. Both spellings are the same
> confusion between the two columns that `archived_page_guard.go:67-71` warns about.

Also selecting archived pages but NOT routing (flag-only, `handler_agent = ""`) and therefore
lower priority: `check_decision_guards`, `check_forced_text_colors`, `check_image_source_unsatisfiable`,
`check_image_url_404`, `check_page_canonical_collision`, `check_section_source_drift`,
`check_site_unreachable`, `check_truncated_component`.

> **⚠ These eight split 6/2 in the posture registry, and the split is deliberate — do not read it
> as a discrepancy.** Six are `PostureObserves`: they file at `handler_agent: ""` and seeing a
> retired page is **correct** for them, so there is nothing to fix. Two are `PostureKnownGap`
> despite also being flag-only today — `check_page_canonical_collision` (no status filter at all;
> its safety comes from a Go-side `activeCount >= 2` gate, not from the query) and
> `check_section_source_drift` (`COALESCE(status,'') <> 'deleted'`, the inert spelling). Those two
> are **debt, not decisions**: nobody chose to see archived pages there, and each is one routing
> change away from becoming a live instance of this bug. The registry records the difference
> because "deliberately unfiltered" and "accidentally unfiltered" look identical in the SQL.

**Two side-effect cases, no work item, worth their own look:**
`check_unlinked_components.go:57` runs `UPDATE page_components … JOIN pages` with no lifecycle
arm — it silently re-links components on archived pages. `check_undeployed_assets.go:296` lets
an archived page's HTML *suppress* a finding (fail-quiet, opposite direction).

### The adoption gap that explains the class

Only **3 files** in `discovery_checks/` call `datahelpers.PageWantedLivePredicateFor`
`[MEASURED 2026-08-22]`. The rest hand-spell the intent in at least four different ways —
`p.status = 'active'`, `status NOT IN ('deleted','archived')`, `status IN ('active','deployed')`,
`COALESCE(status,'') <> 'deleted'` — two of which do not do what they appear to. This is the
same hand-copied-predicate drift that `bugs_closed/185` fixed for the build axis by deriving
every spelling from one helper; the lifecycle axis never got the same treatment.

## 6. Fix candidates, ordered by what closes the door

**A — the individual case (necessary, not sufficient).** Add the lifecycle arm to
`findOrphanPagesSQL` so the producer agrees with its own three remedies:

```go
AND ` + datahelpers.PageHasShippedPredicateFor("p") + `
AND ` + datahelpers.PageWantedLivePredicateFor("p") + `
```

Closes today's 34 archived rows. Alone it fixes one check of eighteen and teaches the next
author nothing.

**B — the framework fix: make the omission visible and unrepeatable.** A **declared posture
registry** in the `discovery_checks` package: every check that selects from `pages` names the
lifecycle posture it takes and why, asserted by a coverage test that fails when a new check
queries `pages` with no entry. The estate's existing idiom for exactly this shape —
`InboundLinkSurfaces`' lockstep test, `COMPONENT_WRITE_ALLOWED`'s reason-carrying allow-list.

**The posture is NOT binary, and a two-valued registry would be wrong.** Three legitimate
postures exist in the tree today:

1. **Armed** — `status = 'active'`. For a check whose remedy mutates or re-links the page.
2. **Deliberately unarmed** — the check must see retired pages. `check_unverified_claims`
   documents this at `check_unverified_claims.go:347`.
3. **Nuanced** — `NOT (p.status = 'archived' AND <never deployed>)`, i.e. *keep* archived
   pages that are still serving, drop the ones that never shipped
   (`check_unverified_claims.go:458`, mutation-tested at
   `check_unverified_claims_archivedskip_test.go`). This is the sophisticated posture and any
   design that cannot express it will push authors back to hand-spelling.

⚠ **The allow-list landmine applies and must be designed against** (`LANDMINES.md`: *a declared
key silences your own detector*; `COMPONENT_WRITE_ALLOWED`'s own note that adding the two known
gaps "converts a live debt into a false all-clear"). Mitigations, both required:
   - The registry distinguishes **reviewed-and-deliberate** from **known gap**, and the audit
     **reports the known-gap count** rather than passing silently. A backlog of 17 must read as
     a backlog of 17.
   - For an `Armed` declaration the test asserts the **SQL actually contains a lifecycle arm**,
     so the entry records a decision the code must then honour. Mutation-prove it: a check
     declaring `Armed` whose arm is deleted must fail the test (`LANDMINES.md`: *a mock's own
     bookkeeping cannot assert a negative — MUTATE to prove a guard*).

**C — NOT proposed: a blanket guard at the work-item filing seam.** The seam is real and
singular — `platform/orchestration/actions/discovery_checks.go:238-264` maps every
`WorkItemSpec` onto `insertWorkItem` (verified first-hand) — and `bugs_open/266` fixed its
four-producer problem exactly that way. **Rejected here** because §6-B shows the right answer
differs per check: a seam-level archived filter would break `check_unverified_claims`, which is
right to look at archived pages. 266's state (`archived` = "nothing may deploy this") admits no
exceptions; this one does. Two further reasons it would not even work: the seam sees only
`item.pageID`, and `check_orphan_pages` **does not set `PageID`** — its three `WorkItemSpec`
literals omit the field, so the id lives only in `spec->>'page_id'` as TEXT. And the seam's
established posture is *demote, never refuse* (`load_work_item_actions.go:1516-1521`), with
`blocked` being non-terminal and therefore still holding the dedup slot.

**D — NOT proposed: sweep the existing rows.** They resolve themselves once A lands — they stop
being re-raised, and the terminals age out of `idx_swi_dedup` (the self-healing note in
`bugs_closed/313`). `RFC_006`'s ruling is that a one-off deletion is not a class fix.

**E — worth its own ticket, not folded in here:** `check_orphan_pages` should set
`WorkItemSpec.PageID`. ~25 checks do; these three do not, so their rows land with
`page_id IS NULL` and are invisible to any page-keyed reaper or guard. Small, and independent
of everything above.

## 7. How to verify a fix

**⚠ STEP 0, AND NOTHING BELOW COUNTS WITHOUT IT: prove the fix is in the RUNNING BINARY before
you read the census.** Raised by the council's `debug_historian` seat (corr `4cf291a2`, medium),
and it is right — this was missing from the first version of this section. The fix is Go
embedded in `agent-chassis`, inert until a roll, and **a census run against the old binary
returns exactly the numbers a working fix would eventually produce is not what happens — it
returns the OLD numbers, which reads as "the fix did not work" — but the dangerous direction is
the other one**: if the archived pages have meanwhile aged out or been re-detected under a
different key, a stale binary can produce a falling count that looks like confirmation. Ask the
artefact, per service, never git and never the tag:

```bash
# Which commit is the running chassis built from? (startup line — it SCROLLS; empty
# means "not in range", NOT "unstamped")
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <the 356 commit> <the stamp> && echo "fix is aboard"

# If the line has scrolled, probe the binary — WITH BOTH CONTROLS in the same breath.
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec ${POD#pod/} -- sh -c '
  grep -aq "PageWantedLivePredicateFor" /proc/1/exe && echo "PRESENT (expected)"
  grep -aq "PageHasShippedPredicateFor" /proc/1/exe && echo "control PRESENT (expected)"
  grep -aq "ZZnotarealsymbol356"        /proc/1/exe && echo "control PRESENT (MUST NOT APPEAR)"'
```
⚠ Never `strings` (absent from the debian-slim images) and never a discovery grep for "some
40-hex string" (it matches Go's internal digit table). ⚠ **Per SERVICE, not per fleet** — one
release tag can straddle several revisions (`bugs_open/249`). ⚠ Note the probe above is weaker
than usual: `PageWantedLivePredicateFor` was ALREADY in the binary before this fix (3 other
checks call it), so its presence proves the roll happened, not that THIS change is aboard. The
commit-ancestry check is the discriminating one; use the probe only to confirm the pod is
running the build you think it is.

Then, and only then:

- **Census arm.** The §5 fleet census (RUNBOOK, "The fleet census") returns **0 archived rows**
  in the `needs_internal_links` / `nav_drift` / `blog` branches, where today it returns 15 / 3 / 16.
  It is grouped by `status`, so it can come out either way.
- **Disconfirming pair, and it must be a PAIR.** A named archived page (e.g.
  `case-study-news-pipeline`, filed 3× since 2026-04-24) is absent from the check's output after
  the fix; **and a named `active` orphan on the same site is still present.** The second arm is
  what discriminates — a filter that returned nothing at all would also satisfy the first.
- **Mutation arm for the registry test.** Delete the lifecycle arm from a check declared
  `Armed` and confirm the test FAILS — **with the decoys left in place**, i.e. against the real
  file, not a minimal reproduction. A clean-room mutation passes on a guard that is blind to the
  nav join, which is how two versions of this test shipped green while asserting nothing.
- **Not a valid check:** work-item counts alone. These are Go changes, inert until a chassis
  image is rebuilt and rolled, and the rotation is hourly/one-site with a 7-day per-site stamp —
  so a fall in new rows lags the roll by days and a flat count proves nothing either way.

## 7b. Council round 1 — APPROVED, and what its advisory objections changed

**Council `4cf291a2-9a45-427f-96ad-8b164250b3c7` — APPROVED round 1**, 15 reviewers, 3 advisory
objections, none high-severity. Recorded here because two of the three named real gaps and the
fixes are in the same commit as this section.

| seat | objection | what happened |
|---|---|---|
| `debug_historian` (medium) | "the plan is silent on confirming the fix is live in the running pod versus just merged" | **Right, and acted on** — §7 now opens with a STEP 0 pod/ancestry check and refuses to let the census be read without it. |
| `reuse_agent` (medium) | "did you check whether `COMPONENT_WRITE_ALLOWED` exposes a reusable posture TYPE this could extend?" | **Checked, and no** — `COMPONENT_WRITE_ALLOWED` is a **Python dict in `scripts/pattern-check.py`**, not a Go type. There is no Go posture vocabulary to extend, so this is not a second parallel implementation. Answered by measurement. |
| `reuse_agent` (low) | "no evidence an existing Go comment-stripping helper was searched for" | **Not searched — and searching found one worth NOT using.** See below; the finding is the useful part. |
| `editquality` (low) | "51 registry entries vs 71 files in the package — 20 unaccounted for" | **Right** — the arithmetic is now stated in the registry's own header (51 query `pages`, 20 do not; and 23 gaps = 17 still-routing + 6 non-routing + this file's fix = §5's 18). |
| `editquality` (medium) | edit 5 listed the bug file as `add` when it was already committed | A submission bookkeeping error, not a code one. Correct: the file landed in `f546f6e3c` ahead of the change, deliberately, so the evidence stands whether or not the fix is approved. |
| `guardian` / `architecture` (low) | the sensor is a build-gate on all future contributors to the package, and should be flagged as a new contribution requirement rather than absorbed as "test-only" | **Flagged explicitly**, here and in WII-025: writing a new discovery check that queries `pages` now requires declaring a posture. Also confirmed: nothing outside `discovery_checks` mirrors or depends on this vocabulary, so it is not a cross-package expectation. |
| `prior_art_librarian` | asked for the "3 files call the helper" and "18 checks" counts to be audited before shipping a wrong denominator | **Re-measured 2026-08-22.** 3 files before this fix, **4 after** (this fix is the fourth); 14 non-test call sites repo-wide. 71 files / 51 query pages. Both load-bearing counts hold. |

### The sibling comment-stripper: found, measured, deliberately not reused

`actions/diagnose_load_runtime_schema_test.go:252` already declares a `stripGoComments` —
a byte-scanner for `//` and `/* */`. **It carries the exact defect this lane's own v2 had**
[MEASURED 2026-08-22, both inputs executed against a verbatim copy]:

```
input : req.Header.Set("Accept", "*/*")\nrows := db.Query(`FROM real_table`)
output: req.Header.Set("Accept", "*          <- rest of file DISCARDED
```

So reusing it would have imported the bug the new one exists to fix. **Not filed as a bug and
not edited from here:** it scans exactly one file, `diagnose_load_runtime_action.go`, which
contains **0** occurrences of `*/*` — so it is **latent there, not blind today**. Recorded for
its owning lane rather than fixed uninvited. This is the objection earning its keep: the answer
to "did you look?" was no, and looking produced a finding.

### `bug_historian`'s note on `bugs_closed/077`, answered rather than deflected

The seat approved but flagged that this bug's shape is `bugs_closed/077`'s — *"a detector must
PARTITION its population by the handler's remit, and file the residue as a capability gap"* —
and that **this fix does the partitioning but not the residue-filing**, which 077 treats as
load-bearing rather than optional.

That is a fair reading and the omission is deliberate, so it is stated rather than left implicit.
Filing the excluded archived pages as a `capability_gap` would assert *"the platform has no
handler for this finding"*. For a retired page that is **not true**: there is no finding. The
page is unreachable because we retired it, which is the intended end state, and 098's answer to a
retired page that is still serving is **retraction**, which already exists and is a different
pipeline's job. Filing a gap per archived orphan would manufacture ~34 undispatchable rows
asserting a capability we do not want. **What 077's residue rule genuinely points at here is the
real hole in §8 — no detector for "archived AND still serving" — and that is recorded there as
unowned rather than smuggled into this fix.**

## 8. What this bug does NOT claim

- It does **not** claim archived pages should be invisible to discovery. `bugs_open/266`'s note
  from the `bugfix_168` lane (2026-08-14) warns specifically against that, and it is right: an
  archived page **can still be serving 200 to the public**, which is 266's whole damage. The
  claim here is narrower — a check whose remedy makes a page *more reachable* must not fire on a
  page the platform has retired.
- ~~**There is no detector for "archived and still serving" anywhere in `discovery_checks/`**~~
  — **FILED 2026-08-22 at the owner's direction as `bugs_open/359`**, with the absence re-scoped
  repo-wide and cluster-wide, and the damage measured at the artefact: **3 archived pages serving
  200 with same-domain 404 controls**, one of them (`robot-hands.com/gripper-catalog.html`)
  byte-identical to what `bugs_open/266` recorded 8 days earlier and still unnoticed. This bug's
  fix neither creates nor closes that gap; 359 owns it now.

## Related

`bugs_closed/298` (this file takes its unclaimed adjacent finding; its own cap defect is fixed,
live and proven) · `bugs_closed/313` (the dead branch, and the source of the `unresolved`
parking mechanism) · `bugs_open/266` (the same axis confusion at DEPLOY time, and the shared-seam
fix precedent; **read its 2026-08-14 note before adding any page-status filter**) ·
`bugs_closed/098` (archiving must retract, not re-link) · `bugs_closed/185` (the build axis got
the single-helper treatment the lifecycle axis never did) · `bugs_open/349` (the other end of the
same family: `PageWantedLivePredicateFor` is lifecycle-only, so a never-built page passes it) ·
`bugs_closed/287` (why 8 of the 34 results are unreadable) · `bugs_open/149` (the discovery
checker-layer defect queue).
