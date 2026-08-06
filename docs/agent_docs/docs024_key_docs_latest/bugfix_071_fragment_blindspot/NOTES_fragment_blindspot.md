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
