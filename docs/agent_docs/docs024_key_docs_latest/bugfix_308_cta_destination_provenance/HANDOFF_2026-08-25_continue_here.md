# HANDOFF — `bugs_open/308`, CTA destination provenance. Continue here.

**Written 2026-08-25 ~11:00Z. Supersedes `HANDOFF_2026-08-23_continue_here.md`.** Lane dir:
`docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`
Read NOTES §§8-14 (2026-08-24/25) for everything below in evidence-grade detail.
Newest milestone read-out: `SUMMARY_2026-08-24b_…`. Owner log: `README_where_we_are.md`.

## 0. State in one paragraph

Phase A (provenance) and Phase B (shared candidate universe + tie/self-link refusal) are LIVE
(v1.0.1337) and proven **at estate scale**: fleet misdirected findings **301 → 135** and
machine-fixable **171 → 11** in the 24 h to 2026-08-25 morning, via four hand-released site
sweeps (all verified in SERVED bytes: gaswholesalers 25→0, leopardess 23/23, ai-agent-orch
20/22, lendzy 20/20) plus the hourly rolling sweep, unattended. The 215-item "stuck backlog"
question that headed the previous handoff is RESOLVED: they were two-strike labels, not stuck
jobs (78% obsolete when re-measured; Fable-audited). **Migration
`555_requeue_misdirected_cta_stock.sql` must NOT be built on the old premise — retire it.**

## 1. What remains — DECISIONS TAKEN 2026-08-25 ~11:40Z, see NOTES §15

> Items 1 and 5 are DONE: both bulk closes executed (109 wont_fix + 74 cancelled, 65 remain),
> `bugs_open/308` is CLOSED to `bugs_closed/`, and Phase C is re-filed as **`bugs_open/389`**
> — which now owns items 2-4 below. This handoff stays as the lane's map; a fresh session
> continuing the work should START from `bugs_open/389_HANDOFF_2026-08-25_…`.

## 1. What remains, in order (historical numbering)

1. **Two owner GO/NO-GOs on the human-review queue (248 open `cta_names_unknown_destination`
   items)** — NOTES §14 table: (a) close 155 STALE items (buttons gone; ids in
   `scratchpad …/fleet2/stale_unknown_ids.txt`, but RE-DERIVE before acting — HTML moves daily;
   the method is the probe in §14); (b) close the demoted excluded-area category (109, ~0
   measured precision, arm no longer files). Genuine residue after both: **~65**.
2. **Phase C — the only unbuilt half of what the bug file asks.** `suggested_target` still has
   no consumer; no completion verifier, so "complete and unchanged" still reports success.
   THREE proven classes now demand it: uncovered component (124 findings today), owned page
   (now parks `deferred` under 333's door), data-less legacy component (aao /blog, HTML frozen
   since April). `VerifyMisdirectedCTAResolved` = re-run `ctaClassifyAnchor` on the page before
   a `cta_links_stale` rerender may complete; and the detector should not file a `page_rerender`
   for a page with ZERO covered findings (that is the two-strike stock manufacturer, NOTES §8).
   > **CORRECTED 2026-08-25 (333 lane): the "(now parks `deferred` under 333's door)" clause above is
   > FALSE** — 0 owned-page `cta_links_stale` rows have ever parked (135 complete / 108 unresolved /
   > 96 failed / 22 cancelled / 1 triaged, live+archive); the door keys on the target handler's
   > declaration and `page-rerender` declares none, by design. See
   > `CONTRIB_2026-08-25_owned_page_cta_rows_do_not_park_under_333s_door.md` in this directory.
3. **Optional widening**: `tool-cta` (12 findings), `tool-list` (5), `case-studies-grid` (1)
   carry `cta_url`-shaped schema fields — adding them to `ctaFieldNames` is council-gated work
   that converts ~18 human findings to machine ones. `article-body`/`ported-*`/`generic-text-block`
   (~84) are prose: human by design.
4. **RFC_047 §10** — the undecidable-tie route to `offer-analyser` needs a page-level output and
   a route back from a refused match. Unstarted.
5. `bugs_open/308` itself: bar #1 is met at population scale; what keeps it open is Phase C
   (its §3 "what is NOT proven"). Owner may close-and-refile Phase C or keep 308 open for it.

## 2. Operating facts a fresh session must not re-derive wrongly

- **Two-strike deferral arithmetic**: the brand fires on ≥2 terminal siblings IN the 7-day
  window — compute ageout from the SECOND-NEWEST strike (my 08-29 dartsonline deferral was
  wrong; the estate corrected it 16 minutes after ageout). <3h repeats now DEFER, not drop
  (`f16c87beb`, 326-D). Owned-page findings park at `deferred` with `builder_needed` (333).
  > **CORRECTED 2026-08-25 (333 lane): FALSE for this lane's findings** — they target `page-rerender`,
  > which declares no refusal, so the door never parks them; 0 ever have. Same CONTRIB as above.
- **The hourly rolling sweep is live and Phase B repairs stick** — do not hand-fire sweeps
  except to jump the queue; pre-flight per NOTES §9 if you do
  (`scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh <domain> completeness`).
- **Verify at pages.url, never a name-derived URL** (leopardess /guides/ lesson, §11); a
  repair may rewrite LABEL and URL together, so old-label tick-lists under-count (§12).
- **Extraction traps**: `kubectl exec | psql` truncates large results silently (assert row
  counts); `COPY TO STDOUT` text-escapes — use the base64 recipe (§9 method, RUNBOOK).
- The offline probe (imports real datahelpers, reproduces the live detector exactly —
  control-verified twice) lives at `scratchpad …/probe/main.go`; rebuild it from NOTES §9's
  description if the scratchpad is gone.
- Kubeconfig token = 3-day expiry; observers die, the fleet does not (§9).

## 3. Key artefacts

| what | where |
|---|---|
| council approvals | Phase A `e4336931…`, Phase B `00732119…`, self-link `49addc8d…` |
| release corr ids | gasw `9917776c…`, leo `9b52142b…`, aao `ba594f7b…`, lendzy `5484e5df…` |
| backlog re-measurement + audit | NOTES §§8, commits `6971b5be8`, `bc105aa3f` |
| census figures (dated) | 2026-08-24: 301/171 · 2026-08-25: 135/11 — re-derive, never quote |
