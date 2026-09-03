# NOTES — bugs_open/384 page-list consumer invalidation (append-only, newest at the bottom)

## 2026-08-24 ~19:15Z — pick-up and ownership
- Bug filed 18:12Z and CORRECTED 18:20Z by the `dartsonline_traffic` lane (transcript `fe285621…`), which hand-repaired the dartsonline instance by dispatching `page-rerender` with `spec.reason='section_data_resolved'` for `index` and `guides-index` at ~18:17Z. Served `/` re-verified here 19:17Z: 12 article-cards, 12 with `<img>`.
- The filing session is ALIVE (owner asked it a Cloudflare question at 19:14Z) but has moved on; its transcript's only fix-site symbol hits are from its own 18:19Z reads. `who-owns` → "(none identified)". Nobody is building the platform fix → taken up here.
- Their NOTES were last written 08-22; today's 384 work lives only in the bug file.

## 2026-08-24 ~19:20Z — measurements (all against the live DB; the disconfirming result was possible for each)
- 43 `query.*` fields across 25 components. `content-listing`'s `articles` field: `{"type":"array","source":"query.blog_posts","required":true,"on_missing":"skip_section"}`.
- Pair census (card asset ↔ stored entry): `content-listing` 32 pairs / 0 stale (post-repair); `tool-list` 37 / 0; `tool-cta` 62 / **14 stale**, 5 of them written AFTER the card landed. `tool-cta`'s template does not render `image`, so no served defect there.
- 41 card landings in 14 days over 8 sites. `page-rerender` runs 1,323 COMPLETED / 31 FAILED in 14d; `section_data_resolved` escalated 1 of 25.
- FAILED classes: 15 `failed to load page info: FATAL` (DB blip), **12 `OWNED_PAGE_GUARD` on save_sections** — so a reasoned re-render of an `owned` page fails; the emitter must exclude them.
- Existing `section_data_resolved` producers: `render_news_section` (182 rows, `page_rerender`), `render_directory` (63, `needs_page`), `rerender-pages` (4), manual (6). None fires on a card landing.
- `blog-index` pages whose listed posts have a card: **0 of 3** → `rebuild_blog_listing`'s `"image": ""` write is latent, not live.

### Missteps so far (each caught before it became a claim)
1. Assumed `pages.slug` — no such column (`\d pages`: it's `url`/`name`). Check: `\d` before SQL, as CLAUDE.md says.
2. Assumed `input_schema->'fields'` is an array — it is an object; `jsonb_array_elements` errored. Check: look at one row's `jsonb_typeof` first.
3. First census counted `entry_image=''` over every source and printed news/directory arrays as "20/20 empty" — those entries carry no `image` key at all. The number was true and meant nothing. Check: measure the JOIN to the card, not the absence of a key.
4. `cd platform/...` in one Bash call left the cwd drifted for later calls (`No such file or directory` on relative paths). Check: absolute paths.
None of these reached a document as an assertion; recorded here, not in WRONG_CALLS (the bar there is a claim written down that turned out false).

## 2026-08-24 ~19:35Z — coordination sent
- SendMessage to `bugs_open/326` (new page_rerender emitter + edit to `flag_page_image_rebuild_action.go` and its test), `bugs_open/357` (one-line delegate in `create_rerender_items_action.go`), `bugs_open/352` (Phase-2 check is their population), `bugs_open/333 [cb419e]` (owned-page exclusion at the lookup).
- `rebuild-cascade.md` is dirty with another lane's REB-003 correction (the 182 lane) — REB-008 would carry a same-file passenger, so the register entry goes into the CLEAN `page-build-pipeline.md` instead.

## 2026-08-24 ~19:45–20:50Z — code, tests, mutations, peer corrections
- Built: `queryresolve.pageImageSources` + `SourceReadsPageImages` + `PageImageSources`; `queryresolve/consumers.go` (`PageListConsumerPages`, owned pages excluded in the SQL); `discovery_checks.PageRerenderItemKey` (exported; `actions.pageRerenderItemKey` delegates); `actions/page_list_reresolve.go` (`requestPageListReresolve`, `reresolvePageListsAfterCard`, `reresolvePageListsAfterPageImage`); callers in `derive_card_asset_action.go` (after the provenance decision) and `flag_page_image_rebuild_action.go` (after `emitContentCardDerive`, inside the tx).
- Tests: lockstep (drives every `queryHandlers` entry under a recording sqlmock; declared set must equal the set whose SQL contains `ca.purpose = 'card'`, both directions); consumers filter/dialect/owned-clause-on-the-statement; seam branches (queued/deduped/partial/lookup_failed/no_consumers; provenance-not-recorded skip; derive-raised defer); key one-spelling + not `page_rerender:`-prefixed; call-site ratchet (comment-stripped, anchored on `provenanceRecorded := true` and `cardEmit := emitContentCardDerive(`); the two `FlagPageImageRebuild_*_StillEmits` tests now script the lookup + INSERT and assert `page_list_reresolve == "queued"`.
- **Mutations, all red then restored:** M1 declare `section_index_for` → lockstep FAIL ("stale declaration"); M2 undeclare `blog_posts` → FAIL ("reads the card join but is NOT declared"); M3 drop the owned clause from the SQL → consumers test FAIL; M4 delete the derive_card_asset call → ratchet FAIL; M5 replace the flag_page_image_rebuild call with a constant → ratchet FAIL AND `_StillEmits` unmet-expectation FAIL. `go test ./platform/orchestration/actions/...` green (3 packages).
- `argContains` name collided with `write_render_audit_findings_test.go` (352 lane) → renamed `specArgWith`.
- **CORRECTION (from the 333 lane, recorded in WRONG_CALLS):** my "12 FAILED = OWNED_PAGE_GUARD" was RUNS (`orchestration_states`); the ITEM count is **4**; the bigger owned-page failure population is `cta_links_stale` at `rerender_sections` — not this bug. The exclusion stays; it prevents 4, not 12.
- From the 352 lane: `idx_swi_dedup` is `(site_id, item_key)` across ALL item types; my key (`page_rerender_<page>_<site>_<reason>`) is disjoint from the `needs_page` producers' `page_rerender:<page>`; 0 pairs ever carried two types. Noted in the helper's doc comment with the date. Also: the Phase-2 Resolved arm must fire only on a positive non-empty match.
- From 357: name the producer set + key shape in the register entry (owner ruling 2026-08-02 §1) → done in PBP-048. From 326: the `summary:` anchor line in `flag_page_image_rebuild` is a ratchet anchor — untouched.
- `idx_swi_dedup` read from `pg_indexes` also excludes `'cancelled'`, which CLAUDE.md's quote of the Go list omits — noted, not touched (DEV-026 territory).
- Register entry went into the CLEAN `page-build-pipeline.md` (PBP-048), not `rebuild-cascade.md` (dirty with the 182 lane's REB-003 edit — would have been a same-file passenger).

## 2026-08-24 ~21:00Z — three peer corrections to figures I had already written (none committed yet)
> **CORRECTED (357 lane):** the line above saying "`idx_swi_dedup` … also excludes `'cancelled'`, which CLAUDE.md's quote of the Go list omits" is WRONG on both halves. CLAUDE.md does not quote `workItemTerminalStatuses` at all (the lockstep note is in the auto-memory index), and `cancelled` IS in the Go list — `work_items_common.go:56`, sitting AFTER a seven-line comment about itself, which a truncated read shows as six entries plus a comment. I asserted an omission from a partial read. The canonical side-by-side lives in the comment above `workItemClosedStatuses` (terminal: 7; closed: 5 — `unresolved` and `failed` are OPEN per RFC_010).
> **CORRECTED (333 lane retracting its own correction):** the "4 items" was classified by the `OWNED_PAGE_GUARD` marker, which was born 2026-08-19; by CAUSE (`error LIKE '%rebuild_policy=owned%' OR '%OWNED_PAGE_GUARD%'`) my own re-run: **13 of 18** owned-page page-rerender failures in the last 14d (live table) are ownership refusals, all `cta_links_stale`. The exclusion keeps MY items off owned pages; the `cta_links_stale` population is another producer's — I have removed the "different defect, not this bug" sentence from the bug file. WRONG_CALLS entry rewritten to carry all three steps.
> **CORRECTED (352 lane retracting):** "0 cross-type (site_id,item_key) pairs ever" was a live-table count; live+archive it is **20** (all colon-shaped; one is `needs_page`+`page_rerender` on `page_rerender:llm-cost-calculator`). `needs_page` keys: 491 colon / 0 underscore across live+archive; `page_rerender` keys: 16,097 underscore / 46 colon (hand dispatches). My underscore key cannot meet theirs; the hazard is colon-shaped hand dispatches. Producers of `section_data_resolved` page_rerender rows live+archive: 1,289 rows / 53 `created_by`; standing = `render_news_section` 795, `rerender-pages` 203, `completeness-discovery-agent` 2; the rest hand dispatches. Doc comment, PBP-048 and the council rationale corrected before dispatch.
- Lesson for this lane, in one line: **a peer's number is another doc** — re-run it before it goes into a comment with a date on it. Three of three peer figures needed a re-run; two were wrong.

## 2026-08-24 ~19:55Z — committed, submitted, verifier armed
- Commit **`9a00a1ee9`** (17 files, explicit pathspec): the seam, both callers, tests, PBP-048, the bug-file block, this lane's four docs. Trailer `Council-Submitted: c2873f56-f45f-42ca-9664-13cceda1f199` (dispatched ~19:52Z; budget ~30 min — find the run by payload: `SELECT current_step, status FROM orchestration_states WHERE collected_data->'input_data'->>'fix_correlation_id'='c2873f56-f45f-42ca-9664-13cceda1f199';`).
- My LANDMINES entry was swept into `babac6c9b` (352 lane, declared) and my WRONG_CALLS entry into `f38c0a019` (333 lane) by their pathspec commits between my append and my commit — the CLAUDE.md "your uncommitted work is not safe" case; HEAD carries the CORRECTED WRONG_CALLS text (checked with `git show HEAD:… | grep -c "13 of 18"` → 1). Nothing lost.
- `landmines-verify-dispatch.sh` reported "already in sync" because another lane's sync had consumed my entry's new-status → armed by hand: `trigger-landmine-verifier.sh 'LANDMINES.md#a-query-fed-array-in-pagecomponents-contentdata-is-a-snapshot-an-assemble-mode-r'` → correlation `17d2a5fe-3cb2-4a58-8bb8-a19449479cf7`.
- Pre-commit advisories: (1) `register-entry-without-row` — PBP-048 needs a row in `000_concept_index.md` (follow-up commit, forward-only); (2) optional-key parity NOT CHECKED because the WORKING TREE does not build (other lanes' WIP) — `scripts/verify-head-builds.sh` run on HEAD instead; I added no `ActionInputSpec` optional keys, so RFC_022's budget is untouched.
- Council submission: 8 edits, 50,527 bytes, real diffs; DRY_RUN passed first; dispatched once with the corrected figures (a stale-phrase assert tripped on the sentence QUOTING 352's retraction, not on a stale claim — checked before trusting the dispatch).

## 2026-08-24 ~20:20Z — Phase 2 built: the `page_list_stale` sweep
- `discovery_checks/check_page_list_stale.go` + `_test.go`; held migration `603_enable_page_list_stale_HOLD.sql` (+ ROLLBACK) appending the name to `completeness-discovery-agent.run_checks.config.checks` (step key `run_checks`, pipeline `content`, 44 checks measured tonight).
- **Design decision, reversing what I told the 352 lane:** NO `Resolved` arm. A `page_rerender` is an action request that completes on its own, and the key is SHARED with other `section_data_resolved` producers (PBP-048 names them) — a retraction keyed on it could close `render_news_section`'s request on a page that also carries a listing. A positive observation about one field is not an observation about the whole key. 352 told.
- **Test gap I caught in my own test:** unknown and current both file nothing, so the first cut could not tell them apart — added a per-run summary finding (`{"summary":true, stale, current, unknown}`) and asserted on it. That is the only place the split is observable from outside.
- Mutations: M6 (empty fresh resolve counted as current) → `Unresolvable…` test FAIL ✓; M7 (membership counted as stale) → `FilesOne…` test FAIL on `/gone.html` ✓. Restored, package green; full `actions` suite re-run in the background.
- Comparison is image-only and both-sides-only on purpose; fresh resolve at the hard cap (24) so a schema-limited stored array is always a subset. First live sweep will hit the 4 sites holding the 14 stale `tool-cta` entries (a no-LLM re-render each).

## 2026-08-24 ~20:35Z — Phase 1 verdict, the bound, RFC_052, Phase 2 submitted
- **Phase 1 council `c2873f56`: APPROVED 20:00Z** — 13 in body, 13 voted, 0 unreadable. Five advisory objections: editquality (diff packaging — the 8th edit bundled three test files), bug_historian (invariant enforced only at two call sites; lookup failure leaves only a log — the disposition also lands in `orchestration_states.collected_data`, and Phase 2 is the backstop), reuse_agent (two hard-coded producers not migrated onto the generic lookup), guardian (bound the per-event consumer count), architecture (`needs_rfc`: the first general source→consumer derivation as shared infra). Approved items 1–3 recorded; 4 and 5 ACTED on:
- **Bound, measured first:** consumer pages by schema vs by rendering `.image` — loancalculator **26 → 1** (25 are `tool-cta`, which never renders the image), fundamentallyai 6→1, gamesdesign 6→2, robot-hands 3→3. Filter added to the shared lookup SQL (`cc.html_template ~ '\.image\y'`, asserted on the statement), plus a per-event cap of 24 in `requestPageListReresolve` with a Warn naming the un-filed pages and `Capped` in the result map (test: 27 → 24 + 3).
- **RFC_052** (`architecture_review/RFC_052_source_to_consumer_derivation_as_shared_infrastructure.md`) — the seam named as shared infra; the open question is whether to generalise the per-base declaration to a dependency set and migrate `render_news_section`/`render_directory` onto it.
- **Retraction census re-run by me** (352 lane's question): `page_rerender` 18,360 rows / 122 producers / **0** ever retracted; every self-retracted type has exactly one retraction authority. That is the measured form of "no Resolved arm on a shared key".
- Phase 2 submitted: council `2005a846-af0e-447f-97c6-3d9702f94979` (6 edits: check, test, 603 + rollback, the filter, the cap). Committed with `Council-Submitted`.

## 2026-08-24 ~20:50Z — Phase 2 round 1 REVISE → round 2 submitted (``, trail 2005a846)
- REVISE at 20:19Z: 14 seats, 0 unreadable, 4 advisory objections (editquality: Phase-1 type not shown + regex case; bug_historian: bare `continue` on unparseable content_data; guardian: show the two Phase-1 edits regress nothing; debug_historian: the 603 header's binary-verification recipe was a GUESSED query). All four acted on.
- **Misstep, recorded:** my first mutation for the unparseable path (M8) PASSED — the single-component fixture was already counted unknown by `compared == 0`, so removing the new `unknown = true` changed nothing observable: a guard in series (the memory entry "a mutation that PASSES usually hit a guard in series", exactly). The discriminating fixture is two components — one unreadable, one current — and with it M8 goes red. Check: when a mutation passes, ask which OTHER guard produced the same outcome before calling the code dead.
- The 603 header now uses `service_binary_capabilities` (kind='discovery_check') with a positive control (`orphan_pages`: 594+61 pods on two commits tonight) and a negative one (`no_such_check_xyz`). The earlier recipe named a `capabilities` column that does not exist — an `[UNVERIFIED]` shape wearing settled clothes.
- Predicate made case-insensitive (`~*`); measured 0 of 7 consumer templates spell `.Image`. M9 (drop the predicate) and M3 (drop owned clause) both red on the statement assertion. Phase-1 named tests: 20 queryresolve + 19 actions PASS; full suite green.
- 352's positive control for the no-Resolved rule recorded: `needs_rerender` 635 rows / 21 filers / 17 retracted / ONE authority — single authority, not few producers, is the condition.

## 2026-08-24 ~20:30Z — round 2 dispatched on the SAME correlation; one empty trailer, recorded
- `RESUBMIT_CORR` keeps the trail under ONE correlation: round 2 is `2005a846-af0e-447f-97c6-3d9702f94979` again (dispatched ~20:28Z after trimming the plan from 71,389 to 49,102 bytes — the 64 KB cap bit on full-file sketches).
- **Misstep:** commit `50f9b13c4` carries `Council-Submitted:` with NOTHING after it. The first round-2 dispatch had failed the size cap, the validator's non-zero exit was masked by `| tail -2`, and `set -e` did not stop at the empty-corr test either (the heredoc-composed message was built regardless). An empty trailer asserts nothing, so it is not a false claim, but it is malformed and `098` cannot join it. Forward-only: this docs commit carries the correct trailer and names `50f9b13c4` as the commit it vouches for. Check: never pipe the validator; guard with `[ -n "$CORR" ] || exit 1` BEFORE composing a message, and read the printed corr — on a resubmit it is the OLD one by design.
- The full `actions` suite showed one FAIL — `TestUpdateWorkItemStatus_ArmedAndDefectPersists_RefusesTheCompletion` — in `update_work_item_status_verification_test.go`, an UNTRACKED file of the 345 lane; it passes alone and references nothing of 384's. Not this change; not chased.

## 2026-08-24 ~20:55Z — round 2 REVISE → round 3 submitted (same trail `2005a846-af0e-447f-97c6-3d9702f94979`)
- r2 objections: editquality (cap test sketched under the production file name), bug_historian (the shared re-render path can escalate to the writer; acknowledge or mitigate), debug_historian (`pages.status` literal set and the "44 checks" baseline were remembered, not queried). Guardian now approves; 11 approvals.
- Acted: test repackaged as its own edit; exposure ACKNOWLEDGED AND DEFERRED to the human applying 603 (baseline 1/25 escalations, the re-read query, and the roll-back rule are in the migration header; a pre-check duplicating the escalation rule was considered and rejected for drift); 603 reads its baseline at apply time (temp table, after == before + 1, every pre-apply name survives).
- **Measured, and it corrected my own literal's story:** `pages.status` holds only `active` 805 and `archived` 66 — `deployed` does not occur. The `IN ('active','deployed')` set stays for parity with the resolvers (a consumer page is chosen by the rule its items are chosen by) and the comment now says so instead of implying both values are live.
- The first r3 dry run was REFUSED by the gate (an edit with an empty sketch after I split edit 6) — and this time the guard stopped the chain before any commit. The check from the earlier slip, applied.

## 2026-08-24 20:48Z — Phase 2 council APPROVED at round 3
- `2005a846` r3: approved 20:48Z (dispatched 20:38Z). Trail: r1 REVISE 20:19Z (4 objections) → r2 REVISE 20:33Z (3) → r3 APPROVED. Every objection was a real defect or a real gap in evidence; none was a matter of taste. The two corrections that would otherwise have shipped: a remembered "44" as a migration verify baseline, and a `pages.status` literal I had described as the live vocabulary when `deployed` does not occur.
- Both correlations approved; the commits carry `Council-Submitted:` and 098 credits them at report time (no amend, forward-only). The one malformed trailer (`50f9b13c4`, empty) is covered by `95eccaece`'s note and the WRONG_CALLS entry.
- **State at hand-off:** everything is committed and HEAD builds (`verify-head-builds.sh` OK at `49a67a002`; later commits are docs + one comment + the migration text). NOTHING IS LIVE until the next chassis roll. Owed after the roll, in the RUNBOOK: (1) prove the roll at the binary; (2) the induced-landing protocol (exactly N consumer items with `spec.cause`); (3) apply `603_HOLD` by hand after the capability probe with its controls; (4) re-read the escalation rate a week on.

## 2026-08-25 ~09:35–09:50Z — post-roll: the roll is proven, and my own bound had silenced the sweep's motivating case
- **Roll proven at the artefact, not at git.** `service_binary_capabilities` (kind='discovery_check', written by each pod at startup from `checks.Names()`): `page_list_stale` on **201** pods @`635f2d32f5bb` (this lane's r3 commit) and 25 @`4c996e1b5cb9`; `orphan_pages` present on both (positive control); `no_such_check_xyz` absent (negative). `635f2d32f` is an ancestor of `4c996e1b5cb9`, so both images carry the whole fix. The `build provenance` log line had already scrolled past `--tail=400` — expected, and exactly why the capability table is the check.
- `603` correctly NOT applied: the checks array is 44 names and does not include `page_list_stale`.
- **CORRECTION (WRONG_CALLS 2026-08-25):** three of this lane's docs said the sweep's first pass would repair the 14 stale `tool-cta` entries. The round-2 renders-image bound — added at the guardian's request, measured and right on its own terms — narrows the SHARED lookup, so it narrows the sweep too. `[MEASURED 2026-08-25 09:42Z]` the sweep would file **zero** items fleet-wide. Bug file and PBP-048 corrected; the check that would have caught it: when you narrow a shared predicate, re-run the census of the population you cited as motivating, under the NEW predicate.
- Served dartsonline re-verified 09:42Z: 12 cards, 12 with `<img>` — and stated in the handoff as NOT the measurement, since the filer hand-repaired it on 08-24.
- **Acceptance-test trap, cost one run:** dispatching `asset-deployer` straight to `system.agent.generic.requests` by kcat lands on a chassis pod with no S3 client → `derive_card_asset` FAILS at `derive_card_asset_step` with "storage client not available" (correlation `30a6f05b`, a real FAILED row). The production route is the **work item** (`needs_content_image` → `asset-deployer`, `triaged`, key `content_image:<page>`, spec = `ContentImageSpecJSON`'s shape), claimed by `build-dispatch-loop` — which handles this type (13 completed, last 08-23) and is live (235 claims in the 6h to 09:40Z). Item `efa8918b-d888-4661-af9f-bc47f9219b10`, `created_by='bugfix_384_acceptance'`.
- Escalation baseline refreshed for the 603 header's rule: `section_data_resolved` runs over 14 days = 39 not-escalated / **1** escalated.

## 2026-08-25 ~09:50–10:05Z — the chain closed, and it disproved one line of my own protocol
- `index`: item `complete`; `page-rerender` COMPLETED, **escalated=false**, `section_count=4 / rerendered=4 / carried=0`, steps visited through `deploy_page` + `update_status`; array rewritten 09:50:58 (was 08-24 23:15:40), 12 entries / 0 empty.
- **CORRECTION to my own RUNBOOK:** I had written "require `pages.deployed_at` to advance". It did not, and that is CORRECT — the array was already current, so the HTML is byte-identical and the deploy is a no-op. A protocol that demands a downstream artefact move when nothing changed manufactures a false negative. The discriminating signals are `spec.cause` (on the item AND the run) and `page_components.updated_at`. Fixed in the RUNBOOK and the handoff.
- `guides-index`: NOT this change — `spawn_agent … failed to create responses topic … dial tcp 10.20.161.251:9092: i/o timeout`, a Kafka broker dial timeout; item back to `triaged`, `attempt_count=0`, retries next loop turn. **I nearly attributed it to the remembered spawn→call handshake race; the full error text says broker dial.** Read the error, don't pattern-match it. Same shape on 2 sites in 7d.
- `pages.build_status='needs_rebuild'` on `index` (1 of 40 dartsonline pages) with `deployed_at` predating my run — pre-existing page state; this seam files work items and never touches `pages.build_status`.

## 2026-08-25 ~11:30–11:55Z — the OWNER RULED ON ALL FOUR DECISIONS, and one of them was re-scoped by a second question

Owner's answers to the handoff's §5: **(1) enable the sweep. (2) generalise it now. (3) fix the
action. (4) fix the tool-cta entries — by changing the template.** Decision 4 then drew a second
ruling once the visual outcome was measured (below).

### Decision 1 — 603 APPLIED, and the verify I first wrote could not have failed
- Pre-apply controls re-run immediately before applying, not quoted from the handoff: `page_list_stale`
  on **301** pods @`4c996e1b5cb9` (was 201+25 across two commits yesterday — the fleet has converged),
  positive control `orphan_pages` on the same 301, negative control `no_such_check_xyz` absent.
  `git merge-base --is-ancestor 635f2d32f 4c996e1b5cb9` → yes.
- Applied by hand (`psql -v ON_ERROR_STOP=1 -f -`), exit 0, snapshot captured. Live row: 44 → **45** checks.
- **Misstep 1, and it is the classic one:** my first independent verify asserted
  `checks @> '["orphan_pages","content_image_missing"]'` and came back FALSE — because
  `content_image_missing` was never in that array. I had invented a control naming a value I had
  myself measured absent 20 minutes earlier. A control that asserts something untrue reports a
  failure that is not there; the twin failure (a control that cannot fail) came next.
- **Misstep 2, worse:** I then compared the live row against `agent_definitions` id
  `b05773e0-…` believing it was the pre-image, because `snapshot_agent()` RETURNED that id. It is the
  **source** id — the live row itself. So the query compared a row to itself and returned "every
  pre-apply name survives: true" with `names_added` empty, which reads as a clean verification and is
  literally nothing. **Check: a diff whose two sides can be the same row is not a diff.** The real
  pre-image is in `agent_definitions_backup` (`type`, `snapshot_reason LIKE '603_%'`,
  `snapshot_taken_at DESC`), which has its own columns — `agent_type` and `snapshot_reason` do NOT
  exist on `agent_definitions`, which is how I noticed.
- Verified properly against that pre-image: before 44, after 45, every pre-apply name survives,
  `names_added = ["page_list_stale"]`, **names_lost = none**. That version can fail; the first two could not.
- Owed, per 603's own header: the first sweep summary must show `current > 0` (a `stale=0, current=0`
  is the BLIND case, not a pass), and the escalation rate re-read in a week against the baseline
  refreshed today — **1 of 36** `section_data_resolved` runs escalated in 14 days.

### Decision 3 — the action was not "latent", it was a live competing writer
- `[MEASURED 2026-08-25]` 3 live blog-index listings, arrays written 2026-08-24 14:01–14:02,
  **47 articles, 47 blank images**: ai-agent-orchestration 16/16, finetuning 20/20, leopardess 11/11.
- `rebuild_blog_listing` is dispatched by ONE live agent (`rerender-pages`) as an **unconditional**
  step; 42 runs, all COMPLETED, in 14 days. And leopardess's `blog-listing_pre_037` declares
  `articles ← query.blog_posts` AND renders `.image` — so it **is** a 384 consumer. The seam fills the
  image on a card landing; the next rerender-pages run blanks it. Two writers, one field, last one wins.
- **But the fix changes no stored byte today:** 0 of those 47 articles has a card asset or a plan hero,
  so the shared projection yields `""` for all 47 as well. Door-closing, not repair — said that way in
  the code, the tests, the commit and here, because "we fixed the blank images" would be false.
- Shipped `7720dc76c`, `Council-Submitted: 170147b4-947d-45c3-8f31-d4b2d1bd5336`.
- **The test trap I avoided by checking:** the cheap guard would be a source scan for the literal
  `"image": ""`. It passes VACUOUSLY — the comment I wrote on `blogPostsQuery` quotes that literal
  while explaining the defect, and first occurrence wins. Tests drive `scanBlogArticles` with mock rows
  instead. Mutation-verified: reverting to `""` turns 2 red; dropping the 3 Scan destinations turns 3 red.
- **And one test overclaimed until I mutated it.** `TestBlogListingScanContractMatchesTheProjection`
  asserts a short row is skipped — it does NOT go red when the Scan drops the image destinations
  (a 9-column row errors either way, so the skip happens for the wrong reason). Its comment now says
  so and names the drift it DOES catch. A test's blind spot is only visible by mutating it.

### Decision 4 — re-scoped by measurement, then by the owner
- `tool-cta` declares only url/title/nav_label/meta_description for its items; `image` is stored purely
  because `plan_sections` (`plan_sections_action.go:2402`) stores the resolver's full item map verbatim.
  Nothing declared it, nothing renders it.
- **The framing that settled it:** a stored-but-unrendered key RE-STALES after any repair, because the
  resolver always returns it and the seam deliberately skips non-rendering consumers. Only two states
  are stable — the key is RENDERED, or it is NOT STORED.
- **"Not stored" is UNSAFE today, measured:** of 28 live (component, query-array-field) declarations,
  **17 render an item key their schema omits** — every directory listing renders `.url` without
  declaring it, and all three blog listings declare no item keys at all. A projection to declared keys
  would blank live content. Recorded so nobody retries it as the tidy fix.
- Owner chose RENDER. Then the visual measurement drew a second ruling: of 228 tool-cta entries,
  62 would show a card crop, **144 a full-bleed plan hero — every one on loancalculator.co.uk**, 42
  nothing. Cause: `WebPath()` is card-first/**hero-fallback**, and loancalculator has **0 of 10** tool
  pages with a card (loanandmortgagecalculator 0 of 19). `content_image_missing` is NOT enabled, so
  nothing was deriving them. **Owner: derive the missing cards first.**
- D1 fired 11:40Z: **29** `needs_content_image` items (19 + 10), `created_by='bugfix_384_toolcta_cards'`,
  production work-item route (NOT kcat — that trap cost a run yesterday). Queue checked first: no open
  item and no `content_image:%` key on either site. `loanandmortgagecalculator`'s dispatch loop is
  active (22 runs/3h); `loancalculator`'s had not run in 3h, so expect it to lag.

### Not mine, but HEAD is carrying it: the optional-key parity test cannot compile
`go test ./cmd/config-key-audit/` fails to BUILD — `livedeclarations_test.go:151,153` reference
`livespec.DeferredDeclarations`, which `platform/livespec/livespec.go:77` records as **renamed on
2026-08-23**. Committed at HEAD by the 363 lane (`18661b3c7`), nothing dirty in the tree. So the
WFA-013 parity test that CLAUDE.md instructs every author to run **cannot be run by anyone**, and the
pre-commit hook reports it as "NOT CHECKED (the tree does not build)" — which reads as a local problem
and is not one. `go build ./...` is clean; it is the TEST that does not compile. Flagged to the owner;
not fixed here (another lane's file, and `who-owns` should route it).

## 2026-08-25 ~12:00–13:15Z — decisions 2, 3 and 4 shipped; and a REVISE that was right

### Decision 4 — LIVE and PROVEN at the artefact
- Migration **614** (gated `{{if .image}}` thumbnail + CSS, and `image` DECLARED in the items
  block) and **615** (the fan-out) applied by hand. Simulated 614's two `replace()` calls
  read-only against the live row first: anchors 1 and 1, 6,102 → 6,421 bytes, result matches
  `~* '\.image\y'`. Applied output identical.
- `tool-cta` now appears in the consumer predicate on 10 domains summing to exactly **40**
  pages — the number 615 then filed, with **0 on archived and 0 on owned pages**.
- **Proof at the artefact, not the status** `[MEASURED 2026-08-25 13:06Z]`: 6 pages re-rendered
  so far, every one carrying `tool-cta-card-thumb` (7 occurrences = 1 CSS rule + 6 item images
  on loancalculator, 6 = 1 + 5 on finetuning). A real src:
  `<img class="tool-cta-card-thumb" src="/assets/images/card-tool-car-finance-calculator.jpg" alt="" loading="lazy">`.
  Control: **zero** `src=""` anywhere in tool-cta output — the gate works.
- Escalations so far: **0 of 5** completed runs, against the 1-in-36 baseline.
- **⚠ 1 of 40 FAILED, and it is NOT this change.** `tool-automation-savings-estimator`
  (ai-agent-orchestration) was refused by the section component floor: `77→37 class attributes
  (48% kept, floor 50%)`. Nothing was written. The refusing slot is the page's own bespoke tool
  section, not `tool-cta` — and **that page already failed 3 times on 2026-08-24**, before any
  of this lane's work; those were the fleet's only other floor refusals in 14 days. Pre-existing
  divergence between its stored HTML and what its template+content_data regenerate; my fan-out
  triggered a 4th attempt at it. The guard did its job.

### Decision 4's real reasoning, for the next reader
`derive_card_asset` **CROPS AN EXISTING HERO** — it does not generate imagery. That is why the
29 D1 items all completed while only 10 cards landed: loancalculator's 10 tool pages had heroes
to crop, loanandmortgagecalculator's 19 had none ("no hero asset to derive from: no active
page, content, or site hero" — the action's own words in the completed rows). So
loanandmortgagecalculator's 12 tool-cta entries stay blank, correctly and permanently, via the
gate. **A `complete` work item is not a repaired artefact**, again.

### Decision 2 / RFC_052 — APPROVED first round (`e1d32ca2`), committed `72469c556`
- `sourceDependencies` replaces the boolean: per base, dependency class → the item keys it
  feeds. The KEY LIST is what generalises it — named keys ⇒ the renders-key template filter;
  NO keys ⇒ whole item set, filter OMITTED. Backwards, that filter returns nothing for news and
  directories and both producers silently stop notifying. Asserted and mutation-verified.
- `business_directory` reads **`business_intel`**, not `directory_entities` — its own class.
  Merging them on the word "directory" would notify the wrong consumers on every publish.
- **The lockstep test caught my own wrong declaration on its first run**, then turned out to be
  blind rather than right: `resolveBusinessDirectory` returns early with no exporter config, so
  under an all-empty mock it never issues its `business_intel` query and reported a CORRECT
  declaration as stale. Fixed by FEEDING THE GATE, not by an exclusion list — an exclusion
  would hide exactly the read the test exists to catch.
- Measured no-ops before migrating: news 16/0/0, directory 5/0/0 and per kind 1/1/1/2, has-shipped
  floor 62→62. **The news figure first read "2 schema-only" — my own AND/OR precedence bug**
  (WRONG_CALLS). Parenthesised: 0.
- RFC_052 CLOSED, and its own premise CORRECTED: the two producers never hard-coded a consumer
  PAGE, they hard-coded the component FUNCTION set. Slower failure, still real.

### Decision 3 — round 1 REVISE, and the objection was right
- `170147b4` REVISE on a **gating** objection from bug_historian: my own comment described a
  silent-degradation path — scan error → skip → empty listing → caller keeps the stale listing
  → step reports success → nobody told — and I had documented the exposure without closing it.
  **Closed** by splitting the cases: some rows scanned = skip one bad post; NO rows scanned
  though rows were offered = projection/Scan divergence, return an ERROR; genuinely empty result
  set is never an error. Three tests, mutation-verified (disabling the branch turns 2 red).
- The guardian's uncapped-query objection turned out to be worse than cost: `query.blog_posts`
  caps at 24, this action did not, and **webdesign.co.uk has 40 eligible posts** — so the two
  writers of one listing would produce 24 and 40. Fixed by sharing ONE cap
  (`queryresolve.PageListingHardCap`); no-op on the three listings written today (16/20/11).
- **Two reviewer checks contradicted my numbers and I re-ran both rather than defending them:**
  (a) "47 blank images" vs their **55** — they are right fleet-wide (94 listed / 55 blank / 9
  pages); 47/47 is the population THIS action writes, and my phrasing read as fleet-wide. The
  other 8 sit on `section-index` pages it never touches and are **correctly** blank (0 cards, 0
  heroes). (b) "leopardess is a live consumer" returning 0 — re-run against the exact
  `PageListConsumerPages` predicate it returns the row; their query was shaped differently. My
  claim stands, and now has the row behind it.
- The audit bug_historian asked for, **with its limit stated**: grep finds zero other actions
  hand-writing a blank into a listing item map. A DYNAMIC audit is unavailable —
  `page_component_history` stamps **1.4%** of rows (309 of 21,491) with a `component_id` and its
  writers read as `(none)` or a raw socket. So "who wrote this component" is not answerable
  fleet-wide today; the blocker is writer-stamp adoption (`bugs_open/355` A1), not this change.
- The rename left **4 stale `pageImageJoins` references in comments** in three other files.
  Corrected; the grep is clean including comments.

### 603 — applied, but the verification I owe is NOT yet satisfied
The sweep has run twice since (lampenkap 11:58, cv1 12:59) and filed **0 items**, as predicted.
But the one summary finding is `consumer_pages: 1, stale: 0, current: 0, **unknown: 1**` —
lampenkap has ONE page and ZERO tool pages, so its `tool-list` array is legitimately empty and
an empty resolve counts as UNKNOWN by design. **`current > 0` is still outstanding**: the
rotation has not yet reached a site with a non-empty listing. ⚠ And note the reporting hazard —
`stale=0, current=0, unknown=1` is indistinguishable at a glance from the BLIND case 603's
header warns about; `consumer_pages` is what proves the lookup ran.

## 2026-08-26 ~10:00Z — the 603 proof arrived, and the fan-out finished

**The `current > 0` proof I owed is OBTAINED.** Overnight the rotation reached sites with
non-empty listings: `loancalculator.co.uk` **`consumer_pages: 25, stale: 0, current: 25,
unknown: 0`** (08:25Z), plus robot-hands 3/3, finetuning 3/3, loanandmortgagecalculator 2/2,
vonc 1/1, webdesign 1/1, garden-tools 1/1. **Items filed all-time: 0** — predicted, and correct.

So the sweep is proven LOOKING, not blind, and the distinction the counter exists to draw did
its job: yesterday's only reading was `current:0, unknown:1` on lampenkap, which has ONE page and
ZERO `tool` pages — a legitimately empty listing, classified UNKNOWN by design. agritec is the
mixed case (`current:1, unknown:1`). The hazard stands: `stale=0, current=0, unknown=N` reads
identically to the blind case at a glance, and `consumer_pages` is the field that discriminates.

**tool-cta fan-out finished: 39 of 40 complete, 39 pages showing thumbnails, 0 escalations in 42
runs** against the 1-in-36 baseline. The 1 failure is the pre-existing shrink-guard page
(`tool-automation-savings-estimator`), which had already failed 3× on 08-24 before this lane
touched anything.

**Replied to the filing session** (`agentchassis-51`), which wrote to make sure the corrected
mechanism was the one built on — it was, and their correction is the foundation of the whole
fix. Sent back the four things they could not have known: the second writer
(`rebuild_blog_listing`), `bugs_open/404` (a THIRD place the assemble-vs-resolve mode gate is
stale, which is their defect one seam along), the sweep proof above, and the `deployed_at`
correction to the acceptance protocol they wrote.

## 2026-08-26 ~20:45Z — fresh build verified, FOUR natural demonstrations, residual pinned to owned pages

- **Fresh chassis build `b34c24f4c65b` (95 pods) rolling alongside `e7f1045fddec` (700).** All four
  of this lane's Go commits are ancestors of BOTH, and the new build is a strict descendant of the
  old, so behaviour holds whichever pod serves. ⚠ **My first ancestry check used yesterday's sha
  and returned "NOT in the running build" for all four** — a false negative made entirely of a
  hardcoded value a roll had superseded. Always read the stamp, never remember it.
- **The seam has now proven itself FOUR times on NATURAL triggers**, which is the evidence the
  induced acceptance test could never give: leopardess 14:42:45 → items filed 14:42:46 → array
  rewritten 15:30:34, 11/11 entries with an image; finetuning 17:25:45 → 3 items → all complete by
  18:44, arrays rewritten 19:13–19:15, **0 blank on all three generic listings**; vonc 19:59:30 →
  fired, re-render pending. Plus three landings that correctly DEDUPED (`deduped: 3, queued: 0`) —
  four landings in fourteen minutes produced three items, not twelve.
- **Zero escalations** across every seam-driven run this lane has produced (baseline 1 in 36).
- **The residual is now pinned and measured:** fleet-wide blank-where-a-card-exists splits
  **14 on `owned` pages (3 pages) / 1 generic (vonc, seam in flight)**. So every generic page
  repairs itself and owned pages never do — the seam, the sweep and the `template_changed` fan-out
  all exclude them by design (`save_sections` refuses an owned page). Migration `614` made that
  visible by giving `tool-cta` an image. Recorded in the bug file's CLOSE-OUT and in the new
  handoff as the one genuinely unfinished piece; the remedy shape is `486`'s
  `section_edit` → `section-editor` route, which is a new seam and belongs in its own round.
- **New handoff written:** `HANDOFF_2026-08-26_continue_here.md`, superseding the 08-25 one.
  384 is ready to close; the handoff's §6 carries the both-paths `git mv` trap and says the
  residual must not close with the bug.

## 2026-08-31 ~16:10Z — DO NOT CLOSE 384: a live recurrence on a generic page, and today's clean census is flattered

Picked the lane up after five idle days (last lane commit 2026-08-26). Everything below is
measured today; the 08-26 handoff's figures were re-derived, not carried.

- **The lane's code is still live.** Fleet is on a SINGLE build `ef06af0e0afc` (342 pods,
  `last_seen_at` current). All four lane commits — `7720dc76c`, `bafd4411c`, `72469c556`,
  `efc0db7bc` — are ancestors of it. Read the stamp, did not remember it (§3's own trap).

- **⚠ THE HEADLINE, AND IT REVERSES THE HANDOFF: the defect is BACK on a generic page.**
  `leopardessconsulting.co.uk/blog` carries **2 blank entries where a card exists**, and they are
  the **first two cards in the grid**. Verified at the SERVED ARTEFACT, not the store:
  `curl https://leopardessconsulting.co.uk/blog.html` → 11 `src=".../card-*.jpg"` for 13 array
  entries, and both guides render as `<article class="article-card hover-lift">` straight into
  `<div class="article-card__content">` with no image node. This is the original 384 symptom.

- **The seam is NOT what failed. It fired, correctly, nine times.** Cards landed 2026-08-27
  22:37:25 and 22:37:49; items were filed within 40 ms. Every spec is right —
  `reason=section_data_resolved`, correct `page_id`, `consumes=["query.blog_posts"]`, and the
  component's own field source IS `query.blog_posts` (checked `content_components.input_schema`,
  component `blog-listing_pre_037`), so the dependency scoping matches exactly.

- **What failed is the CONSUMPTION, and in two different ways:**
  1. **Two items completed green and repaired nothing.** `e1f2dd23` (created 22:37:25.98,
     complete 22:58:18) and `5f78c1e4` (09:39:12 → 09:40:36 on 08-28) both deployed
     `blog.html` + `tools/assets/blog-listing.js` with real commit shas. But
     `page_components.updated_at` for the listing row is still **2026-08-27 21:34:20** — an hour
     BEFORE the cards landed. Per RUNBOOK line 120 (measured 08-25) that column advances whenever
     the array is rewritten, so **the array was never rewritten**. A `complete` work item is not a
     repaired artefact — the memory-index lesson, paid again.
  2. **Seven more items sit `unresolved`, `attempt_count=0`, never picked up** — oldest
     2026-08-28 09:30, newest today 10:37. Same `item_key`, so the dedup key did not collapse them
     either.

- **I do NOT have the mechanism for (1) and am not asserting one.** The 08-27/28 runs are
  unrecoverable: `orchestration_states` retains ~1 day (oldest row 2026-08-30 15:07). The live
  gate is NOT the suspect — I read the live `page-rerender` row and its `check_rerender_mode`
  condition does include `section_data_resolved`. Candidates read but not tested: the
  `plan.Status != "ready"` carry branch (`rerender_page_sections_action.go:509`), and the
  `listedOnly` floor in the `blog_posts` resolver. This is a durable cross-cutting claim with a
  non-obvious cause — i.e. exactly the CLAUDE.md trigger for a `090` diagnosis run, which has NOT
  been fired yet.

- **⚠ TODAY'S CENSUS IS FLATTERED — do not read `generic ≈ 0` as health.** Two demand controls,
  both measured today, and neither was in the 08-26 read-out:
  - **Card production has been ZERO for two days.** Cards landed per day: 08-26 **89**,
    08-27 **109**, 08-28 **46**, 08-29 **18**, 08-30 **0**, 08-31 **0**. The seam has had no
    demand since 08-29, so a low blank count measures an idle producer as much as a working fix.
  - **Fleet work-item completion collapsed.** `page_rerender` created vs now-complete:
    08-24 1400/1390 (99%), 08-27 2138/1947 (91%), **08-28 338/146 (43%), 08-29 179/7 (4%),
    08-30 210/3 (1.4%)**, 08-31 300/144 (48%). 1,076 `page_rerender` rows sit `unresolved` in 7
    days, alongside 1,395 `undeployed_asset` and 466 `required_fields_missing`. Fleet-wide, not
    this lane — the shape matches `bugs_open/413` (selector/loader ordering starvation, owned by
    the `dispatch_throughput` lane). **This lane must not close on a queue that is not running.**

- **The owned residual is UNCHANGED and still correct as described:** 14 blank entries across the
  same 3 `rebuild_policy='owned'` pages. Structural, expected, and still needs its own seam.

**Conclusion: §1 and §6 of `HANDOFF_2026-08-26_continue_here.md` are superseded — 384 is NOT
ready to close.** Corrected in place at the head of that file.

### 2026-08-31 ~16:25Z — CORRECTION to the entry above: the queue collapse is a DIAGNOSED, OWNED, now-ENDED credit outage — and that SHARPENS the residual finding

I wrote the collapse up as an unexplained fleet symptom "matching the `bugs_open/413` shape". That
was under-read: I had not checked the `dispatch_throughput` lane's own record before characterising
their subsystem. Their NOTES (line ~1102, `[MEASURED 2026-08-31 09:45Z]`) already carries it:

- **It is an Anthropic CREDIT-BALANCE outage, not a dispatch defect.** Window
  **~2026-08-28 → 2026-08-31 ~08:40–09:00Z (~2.5 days)** — "D4 case 4". The night ran 100%
  failure; the 09:00 hour today ran 0/35 failed. **Recovery ~09:00Z today.**
- Their post-recovery sanity read is CLEAN: trigger 53 fires/h, arrivals ramping, floor clean,
  **zero stuck claims, zero pins**. Nothing owed dispatch-side.
- So `bugs_open/413` (selector/loader starvation) is NOT what I was looking at. Naming it was a
  guess from shape alone. The check I skipped and should not have: **read the owning lane's record
  before describing their subsystem's failure** — `scripts/who-owns.py` and their NOTES tail cost
  under a minute, and I had already been told the lane existed.

**What this does to my own finding — it makes it sharper, not weaker:**

- The **seven `unresolved` items** (08-28 09:30 → 08-31 10:37) are the outage. Expected. Not a
  defect, and they should now drain post-recovery — **that is a thing to re-read, not to assert.**
- The **zero card production on 08-30/08-31** is the outage too (image generation is API-bearing).
  My "demand control" reading stands as a caveat on the census, but its CAUSE is known and ended.
- **The 2026-08-27 22:58:18 completion is NOT covered by any outage.** The credit outages this
  month ran 08-25 23:47→08-26 08:55, 08-27 11:30→13:35, and 08-28→08-31 09:00. **22:58 on 08-27
  sits in a healthy window between the second and third.** It completed green, deployed
  `blog.html` with a real sha, and left `page_components.updated_at` at 21:34 — an hour before its
  own trigger's cards landed. **That single run is the whole open question**, and it is now
  isolated from every confound I had stacked around it.

So the bug file's UPDATE item 5 should be read as: the census IS flattered, the cause is the
outage, and the outage does NOT explain item 3. Corrected there too.

## 2026-09-02 ~16:05Z — re-read on a recovered queue: page unchanged, and the sweep has NEVER RUN (12/12 born terminal)

Owner asked to re-read the page before firing a `090`. Right call — the re-read killed the last
confound and turned up a second defect that had been hiding behind the first.

**The demand controls PASS now, which is what makes today's reading admissible.** 08-31's reading
was flattered by a dead queue; today's is not. `page_rerender` created/complete: **09-01 1164/953
(82%), 09-02 1179/1084 (92%)**; cards landing again (10, then 6). So "still broken" today is a
real observation, not an idle system.

**The page is byte-identical to 08-31** — 38,319 bytes, same 11 card `src`s for 13 entries, both
guides still text-only, still first in the grid. `page_components.updated_at` still 08-27 21:34.

**MISSTEP, mine, corrected: I mis-attributed the nine items on 08-31.** I wrote "the seam fired
nine times". It did not, and one `GROUP BY created_by` would have told me on the day:
- `derive_card_asset` (the seam): **6 items, ALL complete** — both of ours among them.
- `completeness-discovery-agent` (`spec.check=page_list_stale`, the sweep): **9 items, all stuck.**
I had the rows in hand on 08-31 and grouped by status but not by creator. **The check: when a set
of rows is going to carry a claim about WHO acted, group by the actor before counting them.** It
changed the finding materially — the seam is exonerated, the sweep is the victim.

**DEFECT #2 — the sweep this lane shipped has never run once.** `page_list_stale`, live+archive,
all time: **12 items, 12 `unresolved`, `attempt_count=0`, zero claimed, zero completed.** Three
pages on three sites. Mechanism, read at source and confirmed by fingerprint at the artefact:
- two-strike arm counts `status IN ('complete','failed')` on the shared `item_key`
  (`load_work_item_actions.go:1985-1993`); at `>=2` it brands the row `unresolved` (`:2033`);
- `unresolved` ∈ `workItemTerminalStatuses` (`work_items_common.go:48`) and claim takes only
  `('triaged','approved')` (`claim_work_item_action.go:135`) — **born terminal, unclaimable**;
- every stuck row carries `[unresolved after 6 attempts] …`, and the strike composition on that key
  is **6 `complete`, 0 `failed`**. All six "attempts" were defect #1's false successes.

**The two defects are coupled: our own first line disables our own second line.** The shared
`PageRerenderItemKey` — chosen deliberately for dedup, and celebrated in the 08-26 handoff §4 as
the thing that collapsed four landings into three items — is what carries the strikes across from
one producer to the other.

**Prior art check paid off: this is `bugs_open/389`'s class, already filed 2026-08-25**, whose
class 1 literally reads "completes, strikes, and manufactures `unresolved` stock". `who-owns.py 389`
→ OWNED by the `bugfix_308` lane. So the evidence went in as a **CONTRIB into 389**, not a new bug
and not a fix. The one thing I could not find stated there, and added: the strikes and the parked
item came from **two different producers sharing a key on purpose**, and all six strikes were
successes rather than failures.

**A watch item of ours turns out to be vacuous.** §8.1's "zero escalations against a 1-in-36
baseline" is zero over an empty denominator — the sweep never ran. Eight days of that reading as
health. Same family as `a post-fix ZERO needs a DEMAND control` and `enabled + a fresh tick ≠ ever
RUN`; the check that would have caught it is one query — **count the RUNS before quoting a RATE.**

**Still open:** defect #1's mechanism, now sharply targeted for the `090` — a correctly-specced
`section_data_resolved` item, consumed in a healthy window, completing green without rewriting its
array. Not fired; owner's call.

## 2026-09-02 ~17:15Z — the 090 (UNVERIFIABLE) and what its missing-evidence list exposed

Run: intake `d4f745e6-3f79-42a8-8f71-bb611736912c`, **run corr
`149ec925-ffb7-41eb-806a-1595b8ff2226`**, 5 iterations, 3 orchestrations COMPLETED, ~15 min
end-to-end. Verdict **`UNVERIFIABLE` / "NOT CONFIRMED (stopped: iteration-cap)"**.

**It was worth the run, and not because it agreed with me.** It reproduced my state evidence
independently, then refused to conclude and listed what it lacked — and one item on that list was a
hole in MY reasoning that I had not seen: *"the row was never touched" vs "touched but the value
was unchanged"*. I had been treating a frozen `page_components.updated_at` as proof of no write.

- **`page_components.updated_at` is NOT trigger-maintained.** No trigger sets it; only the app's
  own UPDATE does. So my inference was unsound even though it turned out correct. The RUNBOOK's
  08-25 line ("that column advances when the array is rewritten") is an observation about the
  writer's SQL, not a database guarantee — I had promoted it to one.
- **`page_component_history` is the sound instrument.** Two archive triggers fire on real change
  (rendered_html changed; or content_data changed with html static). No row at 22:58:18 or
  09:40:36 ⇒ those runs changed neither. Alternative closed properly this time.

**Then the history table answered a question I had not thought to ask: WHO writes these arrays.**
`application_name` is on every archive row.
- This listing's last four writes are all `action:rebuild_blog_listing` (08-26 15:30:33,
  08-26 22:02:54, 08-27 06:55:03, 08-27 21:34:20).
- Fleet census, every component with a `query.*` array field, 14 days `[MEASURED 2026-09-02]`:
  `action:rebuild_blog_listing` 5 writes / 1 page; `psql` 1 write / 1 page (another lane's hand
  write). **`rerender_page_sections`: ZERO.**
- And the live `page-rerender` workflow has **no `rebuild_blog_listing` step** —
  `check_rerender_mode → rerender_sections → check_escalated → save_sections → render_page →
  check_skipped → deploy_page → update_status → complete`.

**So the seam files an item whose workflow renders and deploys but has not, in 14 days, rewritten a
listing array.** That is why it completes green with a real commit sha and repairs nothing.

**⚠ This casts doubt on §4 of the 08-26 handoff — the "four natural demonstrations".** The
08-26 15:30:33 archive row IS the celebrated leopardess repair (pre-image 11 blank of 11), and its
writer is `action:rebuild_blog_listing`. The seam files `page_rerender`. **[INFERRED]** the other
three are the same shape; nobody has checked their `application_name` and that is the next
session's first job. If it holds, this lane's headline evidence was attributing another action's
repairs to its own seam — the whole "proven four times on natural triggers" claim rests on it.

**Two seeding mistakes, both mine, both now in the RUNBOOK:**
1. `SEED_SCOPE` with `file.go:EntrySymbol` gave the loop the entry function and omitted
   `rerenderFlatSections`, the callee holding the deciding branch. **Seed whole files** unless you
   already know the branch is inside the symbol you name.
2. The bundle's agent auto-gather does not fetch workflow steps, so the loop listed
   `check_rerender_mode`'s config as missing — **a fact I already had and failed to put in the
   symptom text.** If a live config value is load-bearing, quote it in the symptom.

**Where this leaves the diagnosis.** Sharper than when I started: not "why does the sections branch
skip the section" but **"is `rerender_page_sections` supposed to rewrite a `query.*` array at all,
or was the seam pointed at the wrong vehicle from the outset?"** A re-run with whole-file seeds and
the workflow steps quoted is cheap and is the obvious next move — owner's call.

### 2026-09-02 ~17:35Z — RETRACTION of my own fleet census, 20 minutes after writing it

Checking the other three demonstrations returned **0 rows for finetuning.uk and vonc.com**, which
was too clean. Tested the filter instead of believing it:
**44,781 of 45,540 `page_component_history` rows (98.3%) match no live `page_components` row**,
because `save_page_sections` DELETES and RE-INSERTS the row (`op=delete`,
`source=save_page_sections_overwrite`) rather than updating in place. My census joined on
`pc.id = h.component_id`, so it ran over 1.7% of the table — and the 1.7% it kept was precisely the
in-place writers. **`rebuild_blog_listing` looked like "the only writer" because it is the only one
whose history rows stay joinable.**

Corrected, keyed on `page_id` (14d, listing-hosting pages): `save_page_sections_overwrite`
**4,969 rows / 161 pages**, newest today 17:26; `action:section_editor.update` 93/49; `psql` 12/9.
So the rerender path's writer is prolific. **RETRACTED in `bugs_open/384` and logged in
`WRONG_CALLS.md`.**

**What survives:** the leopardess/blog finding, because it was keyed on `page_id` from the start —
four writes since 08-25, all `rebuild_blog_listing`, **none at 22:58:18 or 09:40:36, not even a
`save_page_sections_overwrite` row.** So on this page, two runs reached `deploy_page` with a real
commit sha and the save step wrote nothing.

**Next question, replacing the retracted one:** why does a run reach deploy with `save_sections`
writing nothing? Untested candidates: every section carried so the save is a no-op; `check_escalated`
routes around it; or `render_page`/`rerender_single_page` deploys from stored `content_data`
independently of the save. **Anyone re-checking the §4 demonstrations must key on `page_id`** — a
`save_page_sections` repair is invisible to a component-id join.

### 2026-09-02 ~17:55Z — §4 doubt WITHDRAWN (3 of 4 genuine), and defect #1 localised to the blog-listing path

Checked the other demonstrations keyed on `page_id`. **My doubt was wrong and I withdraw it.**
finetuning's three (19:13:31 `tool-ai-readiness-checker`, 19:14:21 `tools` incl. slot `tool-list`,
19:15:17 `model-approach-selector`) are all genuine `save_page_sections_overwrite` writes, each
paired with slot deletes. vonc's `archetypes` repaired at 08-27 08:12:56, same way — later than the
handoff's "pending", so that entry was optimistic in timing, not false. **Only the leopardess
15:30:33 attribution was wrong** (that was `action:rebuild_blog_listing`).

So the seam is NOT aimed at the wrong vehicle in general — I was one step too quick to generalise
from the retracted census. It works wherever the listing write goes through `save_page_sections`.

**The defect is BLOG-LISTING-specific.** `leopardessconsulting.co.uk/blog`, complete history, all
time: `rebuild_blog_listing` 5 writes (08-24 → 08-27 21:34:20), two app writes on 08-11, and
`save_page_sections` **exactly once, on 2026-07-12** — never since. The `page-rerender` workflow has
no `rebuild_blog_listing` step, so on this page the item renders from stored `content_data`,
deploys a real sha, and writes nothing. On a tool-listing page the same item repairs correctly.
**That is why exactly one generic page fleet-wide is stuck and every aggregate reads healthy.**

**Last open question:** `rebuild_blog_listing` last ran 08-27 21:34:20, **63 minutes before** the
cards landed at 22:37, and has not run in six days. What triggers it, and why has it not fired?
Answer that and defect #1 closes. Not investigated.

**Method note for whoever picks this up.** Three of my four measurement errors this session were the
same shape — an instrument that could only return the answer it returned (`updated_at` with no
trigger behind it; a history join to a live table whose rows are recreated; a component-id key where
the writer deletes and reinserts). The one that worked was keyed on the identifier the system does
NOT recreate. **Ask what your key survives before you ask what your query returns.**

---

## 2026-09-02 — CONTRIB from the 440 lane (not this lane's session): your `listing_stale` reason is out-of-vocabulary and silently assembles

FYI from `bugs_open/440` (the spec.reason routing/annotation split, RFC_062): one live
`page_rerender` item carries `spec.reason='listing_stale'` (2026-08-24) — routing-shaped, not
in the five-value gate vocabulary, so it took `else_step: render_page` and re-shipped stored
HTML. If a listing re-render was intended there, that item did nothing, green. If assemble was
intended, no action needed — but when the 440 split lands you'll want `listing_stale` either
added to the vocabulary (a migration pasting `CheckRerenderModeConditionClause()`) or kept as a
pure annotation deliberately. Your call, your lane; flagged so the choice is made rather than
inherited.

## 2026-09-03 ~09:30Z — RESOLVED at the artefact: 13/13, chain confirmed by the brand, my date wrong by 21h

- **Repaired.** 13 of 13 card images (42,483 bytes). Repairing write `2026-09-02 23:20:15`,
  `action:rebuild_blog_listing`, pre-image 13 articles / 2 blank.
- **Chain CONFIRMED by the state discriminator, not the timing one.** Three items filed 23:12–23:22
  (`deactivated_head`, `stale_chrome`, `improvement_rerender_…`) were **unbranded**, dispatched and
  completed, after six days in which every one was born `unresolved`. Growth door parked nothing.
- **MISSTEP: my predicted date was 21h late because I aged out the WRONG KEY.** I used the
  blog-listing key (second-newest strike 08-27 22:37 ⇒ 09-03 22:37). The keys that gate
  `rerender-pages` service are `deactivated_head` (08-26 01:57:58, 21:21:48 ⇒ lifts 09-02 01:57:58)
  and `improvement_rerender_…` (08-26 02:02:26, 22:01:22 ⇒ 09-02 02:02:26). **The key carrying the
  symptom and the key gating the repair were different objects.** Worse, I had written "still broken
  on 09-04 ⇒ chain REFUTED" into the handoff — that would have discarded a correct mechanism. The
  brand test saved it. **Enumerate the keys before dating anything; prefer a STATE discriminator to
  a TIMING one, and say which decides.** In `WRONG_CALLS.md`.
- **The roll is not the cause** — the brand lifted ~02:00 on 09-02, 19h before the 21:00Z roll. But
  the filing happened post-roll and I cannot prove the latency was rotation rather than a restart.
  `[UNRESOLVED, immaterial]`: a pod restart does not un-brand a row, so the brand evidence decides.
- **Census now:** generic 5 blanks / 2 pages, **all in-flight** (4.3h and 7.1h old); owned 14 / 3.
- **What remains** is in the new handoff: one owner decision (close on rotation vs close the
  blog-listing gap), the owned residual, and the never-run sweep pending 389.

## 2026-09-03 ~10:1xZ — OWNER RULING (stays open) and a LIVE reproduction that falsifies two of my claims

- **OWNER RULING:** *"keep it open until those are checked and fixed."* Option B. My Option-A
  recommendation (close on rotation, with latency measured) is superseded. Recorded in the bug file
  and the handoff so it is not re-litigated.
- **First check found the defect LIVE.** `designblog.co.uk/index`: component `content-listing`,
  source `query.blog_posts`, **`save_page_sections`-maintained**. Seam fired correctly (5 items,
  `section_data_resolved`, complete, two with matching `consumes`). All four target pages active,
  `page_type='blog-post'`, four active card assets with `asset_key` and matching `site_id`. Array
  rewritten 05:06:21 and 05:25:28, **both after all cards existed**, and holds 4 entries / 4 blank.
- **CARRY HYPOTHESIS REFUTED at the run**, which is still inside `orchestration_states` retention:
  `section_count=4, rerendered=4, carried=0, escalated=false, skipped=false`. Nothing carried. So
  `plan.Status != "ready"` (line 509) is dead for this case — **I carried that candidate through two
  handoffs.** Note how it died: not by reading more code, but because the reproduction was fresh
  enough that the run record still existed. **Diagnose the youngest instance, not the one you
  happen to be attached to.**
- **RETRACTED: "the defect is BLOG-LISTING-specific"** (written yesterday, wrong today).
- **WEAKENED: "four of five demonstrations are genuine."** I verified the finetuning WRITES
  happened; I did not verify they produced non-blank images. The 0-blank figure is the 08-26
  census, a different measurement. **A write is not a repair — the same distinction this lane has
  now paid for three times** (`complete` is not repaired; a write is not a correct write).
- **090 fired on the live case** (~10:1xZ), seeding corrected per the 09-02 lessons: whole files,
  and the live workflow routing quoted in the symptom rather than assumed to be fetched.
  Slug `query_blog_posts_resolves_empty_image_despite_active_cards`.
- **Untested hypothesis, recorded so it is not mistaken for a finding:** `query.blog_posts` resolves
  without populating `articles`, so `plan.ResolvedData` lacks the key and `mergedContent` retains
  the stored blank array, while the section still counts as `rerendered` because HTML was produced.
