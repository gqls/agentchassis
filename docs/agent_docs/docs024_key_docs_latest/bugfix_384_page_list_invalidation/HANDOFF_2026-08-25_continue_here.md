# HANDOFF — bugs_open/384 page-list invalidation · continue here
**Written 2026-08-25 ~09:50Z. Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/`.**
Cold-start read order: this file → `RUNBOOK_page_list_invalidation.md` (every command) → `NOTES_page_list_invalidation.md` tail (what went wrong and the check for each) → `bugs_open/384_HANDOFF_2026-08-24_….md` (the bug, with its own filer's correction).

---

## 1. The bug, in one paragraph

A page's listing card (`assets.purpose='card'`, entity-linked) lands, and the listing that renders it keeps showing a text-only card. The listing's items are a **stored snapshot** — `page_components.content_data`, filled from a `query.*` source at the last section resolve — and every routine re-render is *assemble mode*, which re-ships that array verbatim. Only a re-render carrying `spec.reason='section_data_resolved'` re-runs the query, and nothing in the card-landing chain asked for one. **The tell was never "the page wasn't re-rendered"**: dartsonline's listing was re-rendered three times after its cards landed and stayed stale every time. `spec->>'reason'` on those completed rows is the variable that discriminates.

## 2. What is now LIVE (proven at the artefact, not at git)

**The roll happened.** `[MEASURED 2026-08-25 09:37–09:38Z]` `service_binary_capabilities` (written by each pod at startup from the registry):

| name | kind | pods | commit |
|---|---|---|---|
| `page_list_stale` | discovery_check | **201** | `635f2d32f5bb` |
| `page_list_stale` | discovery_check | 25 | `4c996e1b5cb9` |
| `orphan_pages` (positive control) | discovery_check | 201 / 25 | same two |
| `no_such_check_xyz` (negative control) | — | **absent** | — |

`635f2d32f` is this lane's Phase-2 round-3 commit and is an ancestor of `4c996e1b5cb9`, so **both** running images carry the whole fix. That is the proof — the `build provenance` log line had already scrolled out of `--tail=400`, which is normal and is why the capability table is the check.

**Live behaviour as of this roll:** every `derive_card_asset` (card landed) and every `flag_page_image_rebuild` where no card derive was raised (page image landed) now files one `page_rerender` / `spec.reason='section_data_resolved'` per consumer page, keyed `page_rerender_<page>_<site>_section_data_resolved`, with `spec.cause` naming the event.

**Still NOT live, by design:** the `page_list_stale` sweep. It is *registered* in the binary (above) and *not enabled* — `completeness-discovery-agent.run_checks.config.checks` has 44 names and does not include it. Enabling it is migration `docs/agent_docs/sql_for_agents/603_enable_page_list_stale_HOLD.sql`, applied by hand. See §5, decision 1.

## 3. Commits (all council-reviewed, both correlations APPROVED)

| commit | what |
|---|---|
| `9a00a1ee9` | Phase 1 — the seam + both event callers + register PBP-048 |
| `4a268a0a7` | PBP-048 index row |
| `49a67a002` | Phase 2 — `page_list_stale`, held migration 603, per-event bound, RFC_052 |
| `50f9b13c4` | round-2 revisions (⚠ its `Council-Submitted:` trailer is EMPTY — see `95eccaece`) |
| `95eccaece` | the trail note naming `50f9b13c4`'s correlation + WRONG_CALLS |
| `635f2d32f` | round-3 revisions (603 baseline at apply time, exposure acknowledgement) |
| `a3e37cef2`, `5fa9a9d46` | close-out: register + index status |

Council: `c2873f56` (Phase 1) APPROVED first round; `2005a846` (Phase 2) APPROVED at round 3 — r1 and r2 each found a real defect, listed in NOTES.

## 4. What changed TODAY (2026-08-25), and it changes what the sweep is for

**⚠ CORRECTION to this lane's own claim, recorded in `WRONG_CALLS.md`.** The bug file, PBP-048 and the NOTES all said the sweep's first pass would repair *"the 4 sites holding the 14 stale `tool-cta` entries"*. **That is false, and my own council revision made it false.** The round-2 bound (only count consumers whose template actually renders `.image`) narrows the shared lookup, so it narrows the SWEEP as well as the event seam — and `tool-cta` (**59** live instances as of 2026-08-25) renders no image.

**Simulated against the live fleet under the shipped predicate `[MEASURED 2026-08-25 09:42Z]`: the sweep would file ZERO items today, on every site.** Every listing that actually renders images is current.

That is the correct behaviour (a stale-but-invisible array is a re-render for no visible change, and if such a template is ever changed to render the image, `template_changed` re-resolves the array anyway — it is in `check_rerender_mode`'s reason list). But it means **enabling the sweep buys insurance, not a repair.** Decide it on those terms, not on the 14 entries.

## 5. THE DECISIONS THAT ARE YOURS

### Decision 1 — enable the sweep (migration 603)? *Recommendation: yes, but it is a spend with no measurable return today.*
- **For:** it is the only thing that catches a producer nobody wired, or the event seam failing silently. A backstop switched on after the next incident is not a backstop. It is also the only way to exercise the check in production at all — today it has never fired anywhere.
- **Against:** zero findings today, so it costs a `queryresolve.Resolve` per distinct source per site per sweep and returns nothing measurable for now. Its items ride the shared `page_rerender` path, which can escalate a page to the LLM content writer when a section lacks a required `source:"llm"` field — baseline `[MEASURED 2026-08-25]` **1 escalation in 40** `section_data_resolved` runs over 14 days.
- **How to tell it is working rather than blind, once on:** the per-run summary finding carries `stale` / `current` / `unknown`. `stale = 0` **with `current > 0`** means it looked and found things current. `stale = 0` with `current = 0` means it saw nothing — that is the blind case, and it is what the counter exists to distinguish.
- Command + controls: `RUNBOOK` §"Phase 2 — enabling the sweep". The migration reads its own baseline at apply time, refuses a second application, and carries the escalation-exposure note and the roll-back rule in its header.

### Decision 2 — RFC_052: generalise the source→consumer lookup, or leave it image-specific?
`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_052_source_to_consumer_derivation_as_shared_infrastructure.md`. The architecture seat approved the fix but asked for the seam to be *named* as shared infrastructure rather than discovered later. Today `render_news_section` (795 rows live+archive) and `render_directory` (63) each **hard-code their one consumer page**; the new lookup could serve both if its declaration went from a boolean ("reads page images") to a per-source dependency set.
- **Leave it:** costs nothing now; the third hard-coded consumer name arrives eventually, and the day a site grows a *second* news or directory page it goes stale exactly the way dartsonline's listing did.
- **Generalise it:** one architecture round plus a migration of two producers. Nothing is broken today, so this is a scheduling call, not an urgency.

### Decision 3 — `rebuild_blog_listing` writes `"image": ""` for every listed post
`rebuild_blog_listing_action.go:212-220` bypasses the image projection entirely and blanks the image on every article it writes. **Latent today** (`[MEASURED 2026-08-24]` 0 of 3 `blog-index` pages list a post that has a card), and *now covered by the sweep* — `blog-listing_pre_037` does render `.image`, so a blanked listing would be filed as stale. Choose: fix the action to use the shared projection (small, and closes the door), or leave it latent and let the sweep repair it after the fact.

### Decision 4 — the 14 stale `tool-cta` entries: leave them. *No decision needed unless you disagree.*
They are invisible (the template renders no image), and if that template is ever changed the `template_changed` re-render re-resolves them. Recorded so nobody re-discovers them as a defect.

## 6. What I would do next, in order

1. **Finish the acceptance test** (§7 — it was in flight when this was written; the RUNBOOK has the protocol and the script).
2. Apply **603** if decision 1 is yes, then read the first summary finding for `current > 0`.
3. Watch the escalation rate for a week against the 1-in-40 baseline (query in 603's header); roll back and bring the number to the owner if the new items escalate materially above it.
4. Leave RFC_052 for the architecture track.

## 7. ACCEPTANCE TEST: PASSED at the item level `[MEASURED 2026-08-25 09:49:32Z]`

An induced card landing on `barrel-shapes` (one of the four originally-broken cards — the motivating case) produced **exactly N=2** consumer items, and nothing else:

| consumer page | policy | status | reason | created_by | handler | item_key |
|---|---|---|---|---|---|---|
| `guides-index` | generic | triaged | `section_data_resolved` | `derive_card_asset` | `page-rerender` | `page_rerender_guides-index_<site>_section_data_resolved` |
| `index` | generic | triaged | `section_data_resolved` | `derive_card_asset` | `page-rerender` | `page_rerender_index_<site>_section_data_resolved` |

Spec shape as designed: `{cause: "card_landed:barrel-shapes", domain, reason, page_id, page_name, consumes: ["query.blog_posts"]}`. The acceptance item itself completed cleanly (`attempt_count=0`, no error). **N=2 equals the site's consumer count under the shipped predicate**, neither row sits on an `owned` page, and both keys are the shared `PageRerenderItemKey` spelling — so the sweep would collapse onto them.

Pre-state captured for the causation leg: `index` `deployed_at` 2026-08-24 23:24:30, `guides-index` 23:27:07. The remaining leg — the two re-renders running COMPLETED with `escalated=false` and advancing `deployed_at` — is ordinary platform behaviour on an existing path; it was in flight when this was written, and the poll command is in the RUNBOOK.

## 7a. Protocol, and the trap I hit running it

**Script:** `scripts/induce_card_landing.sh <domain> <page_name>` in this lane. It prints N (the site's consumer count under the *shipped* predicate), pre-counts, publishes with an asserted receipt, and prints the assertion.

**The assertion:** exactly **N** `page_rerender` rows carrying `spec.cause='card_landed:<page>'`, none on a `rebuild_policy='owned'` page, then N `page-rerender` runs COMPLETED with `escalated=false`. For dartsonline N=**2** (`index`, `guides-index`).

⚠ **The served page is NOT the measurement here.** dartsonline serves 12/12 cards with images today (re-verified 09:42Z) because the filer hand-repaired it on 08-24. A listing that already looks right proves nothing about the seam. **The rows are the measurement.**

⚠ **The trap, cost me one run:** dispatching `asset-deployer` straight to `system.agent.generic.requests` by kcat lands on a chassis pod with **no S3 client**, and `derive_card_asset` fails at `derive_card_asset_step` with *"storage client not available"* — correlation `30a6f05b-b7d6-41fd-94e1-7a1405a1696c`, a real FAILED row in `orchestration_states`. That is my dispatch, not the fix. **The production route is the work item**: file `needs_content_image` → handler `asset-deployer`, status `triaged`, key `content_image:<page>`, spec = `discovery_checks.ContentImageSpecJSON`'s shape, and let `build-dispatch-loop` claim it (235 claims in the 6 hours to 09:40Z, so the loop is live). Copy the shape from a **completed row of the same item_type**, per the LANDMINE. The acceptance item filed that way is `efa8918b-d888-4661-af9f-bc47f9219b10` (`created_by='bugfix_384_acceptance'` — that is how you find it).

**Cleanup owed:** the induced card re-derivation is idempotent (same hero, same crop, same `asset_key`), so nothing needs undoing at the artefact. If the acceptance item is still open when you read this, let it run or cancel it — it is not load-bearing.

## 8. Where the knowledge lives

- **Bug:** `bugs_open/384_HANDOFF_2026-08-24_a_landed_card_image_never_invalidates_the_listing_that_renders_it.md` — read its own CORRECTED block; the filer's first mechanism was wrong and says so.
- **Register:** **PBP-048** in `docs/agent_docs/docs026_concept_register/register/page-build-pipeline.md` (+ index row). Names every `section_data_resolved` producer with the census dated, and the shared `item_key` shape — the owner-ruling condition for converging producers on one key.
- **LANDMINE:** *"A `query.*`-fed array in `page_components.content_data` is a SNAPSHOT — an assemble-mode re-render re-affirms it"* — footprints `queryresolve.go`, `page_components.content_data`, `derive_card_asset_action.go`. Swept into `babac6c9b` by another lane's commit (declared); verifier armed under correlation `17d2a5fe`.
- **WRONG_CALLS:** three entries from this lane — the runs-vs-items unit slip, the empty `Council-Submitted:` trailer, and (2026-08-25) the bound that silenced the sweep's motivating case.
- **Code:** `queryresolve/consumers.go` + `queryresolve.go` (`pageImageSources`), `actions/page_list_reresolve.go`, `discovery_checks/check_page_list_stale.go`, `discovery_checks/content_image_helpers.go` (`PageRerenderItemKey`).

## 9. Peer lanes that touched this (all answered, nothing outstanding)
`bugs_open/326` (anti-churn posture — agreed, not an `insertWorkItem` site), `bugs_open/357` (key delegation; filed the `workItemTerminalStatuses` reading-trap landmine), `bugs_open/352` (key namespace, retraction census — **two of their figures were wrong and retracted; I re-ran all three**), `bugs_open/333` (owned-page exclusion — **their correction was itself wrong twice; the cause-based number is 13 of 18 in 14 days**). The rule this lane earned: **a peer's number is another doc — re-run it before it goes into anything with a date on it.**
