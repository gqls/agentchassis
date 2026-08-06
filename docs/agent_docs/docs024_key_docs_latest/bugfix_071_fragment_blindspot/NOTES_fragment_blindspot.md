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
