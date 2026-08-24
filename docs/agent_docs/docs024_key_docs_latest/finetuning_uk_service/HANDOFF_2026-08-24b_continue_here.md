# HANDOFF 2026-08-24b — the £99 offer page is LIVE and verified; copy-editor run 6 + nav rebuild in flight; owner calls are now the frontier (Stripe deliberately LAST)

**COLD-START for the merged finetuning.uk lane.** Supersedes `HANDOFF_2026-08-24_continue_here.md`
(kept for the seeding/377 detail and the morning's traps). Technical log: NOTES 08-24 a/b/c
sections. Owner prose: README 08-24 entries. Outcome report + addendum for the copy lane:
`copy_quality_two_stage/CONTRIB_2026-08-24_from_the_finetuning_lane_the_exemplar_seed_outcome_and_the_brief_that_taught_the_tell.md`.

## State, verified 2026-08-24 ~19:45 UTC

| thing | state |
|---|---|
| Offer page | **LIVE**: `https://finetuning.uk/your-own-model.html`, deployed 19:19:43Z, served 200 (invented-URL control 404), 6 sections rendered. Hero = ratified proposition + £99; journey; honest concierge section (banned promise ABSENT, "won't scale forever" honesty in); glossary-FAQ; 3× /contact.html (required_links intact). Copy dated `llm_call_log a0355b80` 19:14:42Z |
| 377 (placeholder false positive) | fix LIVE on the 18:32Z roll (NUL-split binary probe, both controls; plain `grep -aq` on `/proc/1/exe` is UNTRUSTWORTHY — BusyBox landmine 08-24). Bug file has a same-day CORRECTION; council APPROVED r1 |
| ⚠ the work item | `gap_plan_new_your-own-model_…` reads `complete` but its `error` column still carries BUILD 1's validation error — STALE, do not re-diagnose from it; build 2's orchestration is clean |
| Tell measurement | build1→2: `X, not Y` 3→0 (its demonstrations removed, incl. this lane's mandated phrase), `rather than` 6→8 (fleet text still demonstrates 7×), owner-tier 0→0, gate hits 9 both builds. Full table in the CONTRIB addendum. The rather-than threshold is copy_quality's D3 / owner |
| copy-editor run 6 | dispatched hand-fired 19:35Z-ish, correlation `a504d92d-745b-45e3-9607-84ed632be386` (queue latency to ~29 min is normal; do NOT re-fire on an absent row). Proposal parks `copy_edit_proposed`/needs_human_review for the OWNER (D2); grade with `copy_quality_two_stage/gate_stage2_edit.py --item <id>` BEFORE acting |
| Nav | page was an ORPHAN (0 inbound links, 0 headers). `nav_drift` item `dc3fe53c` filed 19:44 (`nav_rebuild:<site>` key, triaged, nav-updater). Verify after: `SELECT count(*) FROM pages WHERE site_id='1368e337…' AND rendered_header LIKE '%your-own-model%'` — and the SERVED header of `/index.html` |
| Consultations | copy_quality: outcome + addendum DELIVERED (their reciprocal option: de-demonstrate the 7 fleet `rather than`s and re-render for the second data point — THEIR call). offer-analysis: never replied; our differentiator-[0] call stands. aiao carousel CONTRIB: courtesy, no action |
| Owner calls OPEN | playground booking shape · sample datasets · **Stripe LAST (user 08-24)** · copy-editor proposal review when it parks · (new, small) whether "Your Own Model" staying in the site-wide header is the wanted nav shape — declared in_header=true, nav-updater will ship it |

> ## DELTA ~20:00 UTC, closing the session — BOTH pages are LIVE and verified; what's left is a draining wave and owner decisions
>
> - **`/technical-details.html` LIVE** (deployed 19:38:17, served 200): licences stated exactly
>   as the registered facts, version-pinning honoured in the copy itself ("terms would need
>   checking on their own merits"), both required links present, banned promise absent, tells 6
>   (third data point in the CONTRIB series: zero-demonstration brief → 6, from 9).
> - **copy-editor proposal `8003c51a`** parked for the OWNER (gate FAIL decomposed in NOTES/CONTRIB;
>   the rewrite itself is good). The chrome rerender wave does NOT dangle its component ids
>   (empty-reason rerenders assemble STORED html).
> - **Nav: mechanism proven, artefact pending.** `site_nav_items` has the page; nav-updater
>   spawned a 52-item `page_rerender` wave (empty reason = chrome-only). At 20:00 the wave was
>   still fully `triaged` — site unlocked, zero claimed items, pre-query conditions all pass, and
>   the dispatcher processed this site minutes earlier, so this is FLEET QUEUE PACE, not a fault.
>   **Verify later**: `SELECT count(*) FROM pages WHERE site_id='1368e337…' AND rendered_header
>   LIKE '%your-own-model%'` should climb toward ~54, and the served /index.html header shows
>   "Your Own Model". If the wave is still untouched after several hours, THEN investigate the
>   dispatcher's site rotation — not before.
> - Licence facts registered (`ft-licence-llama33/mistral7b/phi35mini`), terms question list in
>   README 08-24d. **Remaining work is owner-gated** (proposal review, booking shape, sample
>   datasets, terms answers) **except Stripe, which the user ordered LAST (2026-08-24).**

## Next work, in order

1. **When copy-editor run 6 lands**: read the proposal, run `gate_stage2_edit.py --item <id>`,
   summarise for the owner in README (do NOT apply without him — D2). If the run produced
   nothing by ~20:15Z, check `orchestration_states` by the correlation before suspecting a drop.
2. **Verify the nav rebuild shipped** (query above + served /index.html header). If `dc3fe53c`
   parks or fails, 016b + `bugfix_149_nav_membership/` is the prior art lane.
3. **Body cross-links (optional, cheap)**: nav gives chrome links only. Inbound BODY links from
   index/services/pricing would use `scripts/fire-internal-linker.sh` per page — decide whether
   the owner wants that before firing three re-renders.
4. **The technical page** ("for your technical person", ratified shape principle 3): next build
   through the same path (page row + needs_content_page; the WHOLE recipe is proven now). ⚠ Its
   copy needs LICENCE facts — verify each base model's licence IN WRITING first (PLAN Phase 0
   step 2; never from memory) and register what's stated into `evidence_base.facts` BEFORE the
   build, as with £99.
5. **Legal/terms extension** (customer training data: retention, deletion, licence, playground
   hours): the commitments are OWNER OPERATIONAL DECISIONS (same class as the retracted
   "person checks every run") — draft the QUESTION LIST for him, do not invent terms.
6. **Stripe payment link: LAST, per the user 2026-08-24.**

## Standing traps for this lane (new ones only; older sets in the 08-24 file + RUNBOOK)

- The dispatcher sweeps `pipeline='build'` + status `triaged` (a column default covered us once).
- `scripts/fire-copy-editor.sh` exists and self-checks (migrated by the 327 lane 19:51) — use it,
  not a hand-rolled kcat envelope.
- A `complete` item can carry a PREVIOUS attempt's error text for ever.
- Binary probes: NUL-split (`tr "\0" "\n" | grep -Fc`), both controls through the SAME pipeline.
- 2× component-creator/`component_selector` FAILED at `store_component` (18:59, no error
  recorded) on this site — unowned observation, not on our path; mention if it recurs.
