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
