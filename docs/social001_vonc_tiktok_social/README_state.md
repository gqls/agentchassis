Done — the copy problem you spotted is fixed, and it turned out to be two layered bugs.

What you saw, and what was actually wrong

The Catalyst page had structure but hollow copy: generic "about the site" text that never named the archetype, plus a garbled stat strip ("24h Clients Served, 1 Daily, 50/50 Awards Won"). Two separate causes:

1. The content-writer wrote boilerplate — it treated each archetype page as a generic about-page. Fixed by authoring canon copy straight into content_data (migration 089), drawn from the spec's actual archetype
   definitions. Catalyst now reads: "The Catalyst doesn't need the best take in the thread — yours is the one the thread grows from… the longest response chains in the Arena trace back to something you filed." CTAs now
   point at the Gauntlet and quiz tools instead of /contact.html.

2. A fleet-wide schema bug the copy pass exposed — after 089, the bodies were right but the stat labels still said "Clients Served". Root cause: on the shared content-block-about component (13 pages across 5 sites),
   the three stat labels and the CTA label were source: "static" with hardcoded business defaults. Every render re-applies them, so no content edit could win — and that's exactly why the LLM had crammed the real label
   into the value on every site ("500+ Models / Clients Served" on robot-hands, "30 yrs / Awards Won" on ai-agent-orchestration). Migration 090 flipped those four fields static→llm (keeping the defaults as fallbacks).

Verified end state

All 8 archetype pages are live and clean — sweep confirms zero boilerplate labels, each stat descriptor is archetype-specific ("Longest / Response chains started", "Widest / Angles opened per Provocation"), and the
CTA reads "Find Your Archetype". The four business sites are untouched: the schema change only bites on re-render, their labels are safely persisted in their own content_data, and I confirmed ai-agent-orchestration's
about page still serves "70+ / Clients Served" with an unchanged DB row.

The transferable lesson (now in the RUNBOOK and running notes): a source: "static" schema field re-applies on every render — no content_data edit or LLM pass can override it. When authored copy "won't stick," check
the field source before assuming the write failed.

Two things I'd still flag for your call, neither blocking:
- The archetype-combinations section on the hub still uses the off-canon 6-name taxonomy (Contrarian/Analyst/Sage…) rather than the canonical 8. Pre-existing; worth aligning in a follow-up.
- The business sites now have a latent improvement available — next time any of them rebuilds its about page, the writer can author proper per-site stat labels via the guidance instead of being locked to "Clients
  Served". Nothing breaks in the meantime.

Worth an eyeball on the refreshed Catalyst page when you have a moment to confirm the voice is right.

-----

Done. Everything in the approved plan that can be produced in an authoring session is implemented, tested, and committed as `0f301cc75` on `085_debug_and_feature_loops` (18 files, +1894/−41). One caveat: **this repo has no git remote configured and no `gh` CLI, so I could not push or open a PR** — the commit is local, which matches this project's documented practice ("images are built from the local filesystem via the Makefile; commits are decoupled").

**What shipped in Go** (one chassis image, all builds green, new unit tests passing):

- **Root-cause repair path** — `chooseCTATargets` v2 in `resolve_internal_links_action.go` now ranks tool/game pages ahead of content hubs and returns full targets, so `setCTAField` writes `cta_target_title` alongside every URL (copy can be authored *for* the real destination). A new `ctaExcludedDestination` helper fixes the blind spot where `/contact.html` never matched the excluded-area set.
- **`rerender_page_sections`** gained a CTA recompute gated strictly on `reason == "cta_links_stale"` — plain `image_landed` rerenders are byte-identical to before. Its exception rule keeps authored links that are real, non-excluded, and non-circular; everything else (phantom, `/contact.html`, self-links, empty) gets the recomputed target. This means vonc's 19 misdirected CTAs repair through the **generic** path — no hand SQL.
- **Two new discovery checks**: `misdirected_cta` (anchor text names page X, href goes to page Y → one `cta_links_stale` rerender item per page; distinctive copy naming *no* page with a bad href → `cta_names_unknown_destination` for human review — the Arena case) and `incomplete_page_group` (plan-promised siblings half-deployed; tool/game gaps go to review per TP-004). Plus the definite `orphan_pages` fix: nav-flagged tool pages are no longer excluded, so vonc's missing Gauntlet/Quiz nav entries will finally surface as `nav_drift`.
- **Header CTA validation** in `render_site_components` — falls back to the resolver's ranking instead of an absent contact page.

**SQL migrations** (091–095 in `minilobby_task/`, all with backups + hard-failing verification blocks): 091 flips the four `pages.*` CTA schema sources to `renderer` fleet-wide; 092 enables the three checks and wires `cta_links_stale` into page-rerender plus the writer guidance; 093 fixes only the two `/how-it-works*` prose phantoms; 094 plans `tool-arena`, queues the tool-generator item with the full v1 spec (daily provocation, localStorage takes, the five Arena Reactions, remix-chain visual), and records the PLAN doc; 095 retargets the Arena CTAs with a pre-flight that refuses to run before the page is deployed.

**`RUNBOOK_link_integrity_task.md`** sequences the live rollout — image → symbol verification in the pod → 091/092 → fleet-safety gate → pre-repair discovery (which must reproduce all four finding classes on vonc, proving detection) → repair → Arena → the closing discovery pass that must come back clean. Running notes are updated with the session log, including the standing TL-001 note that the Arena page must never receive a generic full rebuild.
Done. Everything in the approved plan that can be produced in an authoring session is implemented, tested, and committed as `0f301cc75` on `085_debug_and_feature_loops` (18 files, +1894/−41). One caveat: **this repo has no git remote configured and no `gh` CLI, so I could not push or open a PR** — the commit is local, which matches this project's documented practice ("images are built from the local filesystem via the Makefile; commits are decoupled").

**What shipped in Go** (one chassis image, all builds green, new unit tests passing):

- **Root-cause repair path** — `chooseCTATargets` v2 in `resolve_internal_links_action.go` now ranks tool/game pages ahead of content hubs and returns full targets, so `setCTAField` writes `cta_target_title` alongside every URL (copy can be authored *for* the real destination). A new `ctaExcludedDestination` helper fixes the blind spot where `/contact.html` never matched the excluded-area set.
- **`rerender_page_sections`** gained a CTA recompute gated strictly on `reason == "cta_links_stale"` — plain `image_landed` rerenders are byte-identical to before. Its exception rule keeps authored links that are real, non-excluded, and non-circular; everything else (phantom, `/contact.html`, self-links, empty) gets the recomputed target. This means vonc's 19 misdirected CTAs repair through the **generic** path — no hand SQL.
- **Two new discovery checks**: `misdirected_cta` (anchor text names page X, href goes to page Y → one `cta_links_stale` rerender item per page; distinctive copy naming *no* page with a bad href → `cta_names_unknown_destination` for human review — the Arena case) and `incomplete_page_group` (plan-promised siblings half-deployed; tool/game gaps go to review per TP-004). Plus the definite `orphan_pages` fix: nav-flagged tool pages are no longer excluded, so vonc's missing Gauntlet/Quiz nav entries will finally surface as `nav_drift`.
- **Header CTA validation** in `render_site_components` — falls back to the resolver's ranking instead of an absent contact page.

**SQL migrations** (091–095 in `minilobby_task/`, all with backups + hard-failing verification blocks): 091 flips the four `pages.*` CTA schema sources to `renderer` fleet-wide; 092 enables the three checks and wires `cta_links_stale` into page-rerender plus the writer guidance; 093 fixes only the two `/how-it-works*` prose phantoms; 094 plans `tool-arena`, queues the tool-generator item with the full v1 spec (daily provocation, localStorage takes, the five Arena Reactions, remix-chain visual), and records the PLAN doc; 095 retargets the Arena CTAs with a pre-flight that refuses to run before the page is deployed.

**`RUNBOOK_link_integrity_task.md`** sequences the live rollout — image → symbol verification in the pod → 091/092 → fleet-safety gate → pre-repair discovery (which must reproduce all four finding classes on vonc, proving detection) → repair → Arena → the closing discovery pass that must come back clean. Running notes are updated with the session log, including the standing TL-001 note that the Arena page must never receive a generic full rebuild.

----

That's a wrap on WS1–WS3. Here's the full, verified state:

WS1 (Go) — live and confirmed across two deploys. Your two chassis builds (v1.0.1116, then v1.0.1117) both carry the code; I re-verified symbols in-pod after the second one flipped over mid-task.

WS2 (migrations 091+092) — applied, verified. 4 CTA URL fields flipped off the pages.contact/pages.services root cause; 3 checks enabled; cta_links_stale wired into page-rerender. Fleet safety gate
passed: an untouched business site's CTA is byte-identical (DB row and live HTML), and a plain image_landed rerender proved genuinely no-op on CTAs (confirmed by hash, not by trusting the work-item's
misleading stale-error status).

WS3 (vonc repair) — done, verified by artifact and by curl on the live site.
- Discovery run before repair proved the generic loop actually works: it independently found the misdirected CTAs, the 2 true 404s, and both Arena CTAs — without me telling it where to look.
- 093 fixed the two /about.html 404s.
- All 9 archetype pages + archetypes.html: primary CTA → Gauntlet, secondary → Quiz, confirmed both in the database and in the live served HTML (the leftover /contact.html links I see now are legitimate —
  header button and footer nav, not the bug).
- index.html is still mid-queue (not stuck, just slow — the dispatcher stalled repeatedly on zombie claims during this run; I used the documented reset-and-retry recovery each time). Worth a curl check on
  your end in a few minutes.

Three things the generic loop surfaced that weren't in the original plan — none touched, all deliberately parked:
1. check_phantom_internal_links.go has no data-runtime-fill guard (unlike the other checks), so it misfires on vonc's deliberate empty-shell provocation-card/lobby-grid components. Real gap, low urgency
   (items sit inert at detected) — worth a follow-up patch.
2. The quiz's result card has a static href="" "Get Your Full Report" CTA with no runtime-fill marker — genuinely undetermined whether that's a planned feature or dead promise. Routed to human review,
   untouched.
3. about.html's gauntlet-cta component uses its own field names (cta_primary_url/cta_secondary_url), so it's invisible to our fix — and it has both a misdirected secondary CTA and the same generic
   stat-label bug (090's pattern: "Happy Customers"/"Avg. Rating") on a component 090 never touched. Needs its own small migration.
4. nav_drift correctly never fired — turns out Gauntlet and Quiz are already linked from the site footer, just not in the formal site_nav_items table. Less broken than originally assessed; no action
   needed.

Not started: WS4 (Arena). Want me to move on to that next, or knock out the three residual findings above first?


---


All done — the handoff's top job is complete, verified live end to end, and an unrelated fleet-wide breakage found during verification is fixed too. Here's the full picture.

The top job: detection and repair now match

Go (shipped in v1.0.1125, verified in-pod by binary string table): ctaFieldNames — the one shared set used by both the build-time link resolver and the repair-time applyCTARecompute — now covers archetype-grid
(cta_url), archetype-combinations and gauntlet-cta (cta_primary_url/cta_secondary_url), and content-block-about (cta_url), alongside hero/call-to-action. Single-URL components use an empty secondary slot, guarded
in both writers so nothing writes under an empty key or files a bogus HITL item. A new TestCTAFieldNamesContract unit test pins the map to the live schema field names — mechanising the field-name lesson from the
095 bug. The checker's stale "hero/CTA only" comments were updated.

Schema (migration 098, applied, verify block passed): the five site_specs.*-sourced URL fields flipped to renderer, with a backup table. Deliberately untouched: content-block-about.cta_url (already llm),
provocations-archive-list (097b's static-Arena pin is correct — adding it to the map would let a rebuild misdirect it), the required flags (verified in code that a renderer source short-circuits planSection before
required-field handling), and the still-static label fields (that's the step-3 content pass, anti-fabrication).

Proof on vonc: the archetypes page's "Explore All Archetypes" button — rendered /contact.html purely from the schema lock, with empty stored content_data — now points at the Gauntlet, confirmed in both the database
row and live curl. The other four flagged anchors are authored links to real pages, which the recompute correctly keeps; after two completed repair attempts each, discovery escalated them to unresolved for human
review — the loop's designed terminal state: repairable links self-heal, copy decisions get parked loudly. Fleet-wide, four other sites' content-block-about CTAs pointing at contact/about pages will now self-repair
on their next flagged rerender.

Guideline cross-check (as requested, against 000/001/002/003)

- 001 Step Zero / reuse: no new agents, actions, or registry entries — the whole fix extends one existing shared map plus two three-line guards; both consumers stay consistent by construction.
- 001 verification discipline: planSection's renderer handling checked at point of use before flipping required:true fields (dated claim in 098's header); deployed binaries verified against the pod, never git.
- 003 Input Schema v2: renderer is the documented source for render-time-assembled fields, which is exactly what these are now; no fallbacks present, avoiding 097's renderer-with-fallback revert trap.
- 003 source-of-truth: the recompute writes into resolved_data and the rerender persists content_data = stored ⊕ resolved — confirmed in the repaired row.
- 002/004 loop architecture: unchanged — the recompute still gates on reason=cta_links_stale only, so image_landed/section_data_resolved rerenders stay byte-identical.

Found and fixed en route: fleet-wide work-item insert breakage

My verification re-runs failed, and the trail led to a real incident: migration 157 (the claims session, yesterday) added 'cancelled' to idx_swi_dedup's excluded statuses, but Go's workItemTerminalStatuses —
interpolated into insertWorkItem's ON CONFLICT ... WHERE — stayed at six. Postgres could no longer infer the arbiter index, so every keyed work-item insert failed (SQLSTATE 42P10) and every discovery run's output
was lost fleet-wide — while looking "clean" (zero items). I added cancelled to the Go list (commit 5e2711997, shipped as v1.0.1127 via the new committed-ref build target), then swept four more stale hardcoded
copies of the list (21e74808e) — including one ON CONFLICT in create_tool_cross_link_items that had already been silently failing. Post-fix discovery runs complete cleanly with working inserts, verified this
morning. Full write-up added to 016b ("split-contract-drift"), and a memory saved so any future idx_swi_dedup migration moves in lockstep with the Go list.

Housekeeping and what's next

Six stale vonc items were closed by artifact (the 093 /how-it-works phantoms and arena nav phantoms — live curl shows zero occurrences; the two runtime-fill shell re-raises → wont_fix per the standing ruling). Docs
are updated and committed: handoff addendum, RUNBOOK close-the-loop marked executed, RUNNING_NOTES entries for both days including a correction of my own first-day claim. One observation left for the fixloop
workstream: unresolved rows sit outside the dedup index, so each qualifying discovery pass adds a duplicate triplet — harmless but noisy.

Remaining from the handoff, in order: investigate the gauntlet/quiz needs_rebuild flags (#2), the content pass on unlocked components (#3 — needs a 096-style label unlock for archetype-grid/archetype-combinations
first), and the two minor loose ends (#4), plus reviewing the three unresolved copy-decision triplets.


----


