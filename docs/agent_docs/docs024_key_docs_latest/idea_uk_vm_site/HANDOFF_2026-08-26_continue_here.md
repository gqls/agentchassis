# HANDOFF 2026-08-26 — ⚠ FLEET LLM OUTAGE (credits) is the live headline. The news feed survived its first unattended night. Fresh build v1.0.1341 verified; its new checks and the re-enabled improvement-sweep both touched this site.

**Supersedes `HANDOFF_2026-08-25b_continue_here.md` as the cold-start file** (its §1 news arc is
now proven unattended; its §2 next-steps carry forward below). `HANDOFF_2026-08-16` §4 + §7 still
hold the voice-arc history and head blind spot; `HANDOFF_2026-08-11` §3 for RFC_015.

Cold-start order: **this file → NOTES §X.61–§X.62 → RUNBOOK Phase 6 → `README_where_we_are.md` tail**.

---

## 0. ~~⚠ THE LIVE INCIDENT~~ **RESOLVED 2026-08-26 ~08:58Z — the owner added credit.** Boundary + the burned-row reset: `bugs_open/243` (recurrence section)

> **Update 09:05Z:** last error 08:57:46Z, first success 08:58:29Z, 14 successes / 6 agent
> types by 09:02Z, zero new errors. 33 sub-max-attempts rows self-healed (2 already complete);
> the 20 attempt-exhausted rows were RESET to `triaged` with a backup in
> `bak_credit_burn_20260826` (idea.uk's `ade31076` among them). Nothing in this section is a
> live instruction any more — the paragraphs below stand as the record of the window.

- `[MEASURED 2026-08-26 ~08:55Z]` **1,884** `agent_error_log` rows `%credit balance%`
  (23:47:10Z → 08:50:26Z, still firing); **20** work items burned to terminal `failed`
  (attempt_count 3) across 6 sites — idea.uk's casualty is `ade31076` (`dead_fragment_link`,
  report-example hero). The error is the API's billing message verbatim, sample
  `req_011CeQrxFHTYb4BE9SeTB5Kd`.
- **This is `bugs_open/243`'s class** (08-10 instance resolved same-day when the OWNER added
  credit — not the calendar). **Recurrence appended to 243 today** with evidence + the
  post-recovery sweep. Recovery check: `SELECT max(created_at) FROM llm_call_log WHERE success;`
  — sustained successes, not one probe.
- **OWNER ACTION: add credit** (and if billing looks healthy, check which ORG the fleet key is
  on via the key's "Last used" — it has not been the default org before).
- **Until then:** idea.uk's 34 `triaged` rows (last night's audit wave) dispatch into the outage
  and burn attempts. **After recovery:** re-read every `failed` row from this window as an
  outage casualty before diagnosing anything as a v1.0.1341 regression —
  `SELECT id, item_type FROM site_work_items WHERE status='failed' AND error LIKE '%credit balance%';`

## 1. The deploy — verified at the artefact, and what it changed here

- **v1.0.1341**, binary stamp **`2fb40a960`** (2026-08-25 22:32 BST), read via
  `strings /app/agent-chassis | grep -oE '^[0-9a-f]{40}$'` with both controls passed
  (yesterday-HEAD is an ancestor; stamp is a real commit). The provenance LOG line had already
  rotated — the binary probe is the durable check (NOTES §X.62 §1).
- **New discovery checks fired on idea.uk** (01:26–01:31Z wave), all `detected`:
  `prerequisite_missing` ×3 — **`evidence_base` ABSENT (every claim ungated; `bugs_open/380`
  D1)**, `page_research` never requested, `vertical_landscape` never ran;
  `structure_floor_unmet` (4 of 6 structures / 25 pages); `heading_promise_unmet` (specimen
  page). Owner-triage material; the evidence_base one matters most on this site.
- **`improvement-sweep` RE-ENABLED** 2026-08-25 21:18Z (owner's word, loanzy lane `bf42e9288`),
  900 s cadence. Harmless to the authored news block — no-match still writes nothing. The
  LANDMINES entry and RUNBOOK 6a carry dated corrections.

## 2. News — three unattended trigger passes; the mechanism is now just running

- Passes at 20:45Z, **02:45:53Z** and **08:46:06Z**, all COMPLETED: vm-sites `17597896a`
  (02:46:22Z), served `updated_at 02:46:19Z`, `item_count 6`; 5 sources `error_count 0`;
  0 new feed items (dedup or none — fine). Store: 9 `relevant`, 3 `review`.
- The passes so far needed no LLM (no pending items) so the outage has not touched news yet;
  **a pass WITH new items will leave them `ingested` until credits return** — the renderer keeps
  serving the existing `relevant` set. Degraded, not broken.
- `render_news_section` filed **2 `section_data_resolved` page_rerenders (index, news-index)**,
  `triaged` — server-side news in the pages, not just the JS fetch. They dispatch into the
  outage; the homepage's D-001/D-002 sections are decision-gate-protected (the proven seam).
- Standing items: **retune the 5 keywords after ~a week of items** (RUNBOOK 6e); `api_news` is
  an owner opt-in; goto-redirect links = **`bugs_open/400`** (the ingester's, not ours).

## 3. empty_section — §X.61's prediction happened on schedule

The completeness rotation **re-filed both findings 01:27Z on the LIVE component ids**
(`a1724965`/`1ad768cb`, same item_keys, `triaged`) beside the old `failed` pair still keyed on
the dead ids — `bugs_open/300`'s churn-orphaning, now visible as same-key-twice in one table.
When the new pair dispatches post-outage, **watch whether the handler fills
`{{.section_heading}}` or the next rebuild churns the ids again**. The underlying defect
(templates render `section_heading`/`section_intro`/`eyebrow_label`; `content_data` has no such
keys) is whoever owns tool templates (`bugs_open/357` nearest).

## 4. Queue `[MEASURED 2026-08-26 ~08:55Z]`

| status | count | note |
|---|---|---|
| needs_human_review | 49 | unchanged — the wave filed NO new guard rows (item_key dedup held). 23 of 49 remain the D-001..D-005/lock guard refusals; batch-close recommendation stands (README 08-25) |
| deferred | 37 | +4 `capability_gap` (improvement-loop) |
| detected | 35 | + the new checks (§1) |
| triaged | 34 | the wave: rerender-pages 27 · design-discovery 13 · acceptance 5 · completeness 5 · improvement-loop 5 — dispatching into the outage |
| unresolved | 7 | incl. the 3 stale `tool_crosslink` items (ab-test) |
| failed | 4 | 3 known (§X.61 §4) + `ade31076` (outage casualty) |

## 5. Owner choices still open (unchanged from 08-25b)

1. **ab-test calculator page**: tool-writer rebuild (311's webdesign.co.uk precedent) or retire.
   DB holds 1 of 4 sections; served page works from an old deploy; every rerender fails.
2. **Owner queue batch close** (the 23 guard refusals) — with the owner, not unilaterally.
3. **`api_news` opt-in** for the news feed (LLM-authored items; currently news_search only).
4. Older residuals: first organic Stripe webhook; tools-page card images/heroes; empty-kind →
   SDXL routing; ingress landmines (`ufw allow 80,443` FIRST).

## 6. Traps (new first, then carried)

- **An outage-window `failed` row is a provider casualty, not a code bug** — the error names
  billing; read it before diagnosing the fresh build.
- **The provenance LOG line rotates within hours** — date a build by the binary stamp + ancestry
  controls, never by logs alone.
- **`dispatch_sources` is ASYNC** — one orchestrator run is not a verification; run twice.
- **Identical `length(rendered_html)` across a rebuild = reproduced, not repaired** — re-run the
  predicate on the replacement row (WRONG_CALLS 08-25).
- `news_render_result.item_count` ≠ the snippet's length — read the served file.
- Rolling window on `site_work_items` totals; `attempt_count 0` = never tried; a completion
  count is not an artefact check; the 08-16 §7 set.

## 7. Pointers out of this lane

`bugs_open/243` (the outage class + recovery boundary) · `bugs_open/300` (id churn; our
`empty_section` CONTRIB) · `bugs_open/400` (goto links, ingester's) · `bugs_open/357` +
`bugfix_311_component_keys/` (tool templates/rebuilds) · `bugs_open/380` (evidence_base D1) ·
`dispatch_throughput/` (queue timing) · `news_editorial_features/` (the snippet's headline copy).

---

## ADDENDUM 09:30Z — post-recovery state (NOTES §X.63–§X.64 are now part of the cold-start read)

- **Recovery HOLDING**: 0 credit errors since 08:58Z; the 20 burned rows were reset (backup
  `bak_credit_burn_20260826`). Fleet backlog drains at ~2 items/min against ~1,400 triaged —
  hours, not minutes. The news rerenders (`a10a7110`/`f2fc39d5`) and the empty_section pair
  (`2b52cb30`/`9e6da605`) were still attempt 0 at 09:24Z.
- **`ade31076` diagnosed (NOTES §X.64 §2)**: its rebuild is text-floor-refused (writer 2,062
  visible chars vs 19,918 deployed) and cannot converge; `save_refused_incomplete` `3493b44f`
  is in the owner queue (50th row). Fleet census: the ONLY such case — no bug filed. New owner
  choice: targeted hand-fix of the hero link (`#request-a-report` →
  `#c-report-request-form-request-a-report`, or drop the fragment) vs accept top-of-page landing.
- **Watch also**: `section_edit` `tool_fix` on `6ddcedf4` (ab-test page — bears on choice §5.1);
  design-discovery rotation re-enabled 09:20Z (peer heads-up; `bugs_open/401`) — surprise
  design items within ~2–3 days are the rotation.
