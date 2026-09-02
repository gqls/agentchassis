# NOTES — bugfix 443 fallback-tier subjects (append-only, newest at the bottom)

## 2026-09-02 — session "bugs_open/443" resumes the handoff

- **Ownership check:** `who-owns.py 443` → filing lane (finetuning_uk_service) active on the
  SITE, but the bug file is a handoff ("Filed as bugs_open/443 with fix candidates ordered" in
  their 2026-09-02 handoff). No dirty or recent-fix commits on
  `load_page_sections_from_spec_action.go` beyond apis.uk's tier-1 mechanism (`35905c547`).
  Messaged the live `finetuning` session; they confirmed handoff and sent five findings (below).
- **Bug still valid:** finetuning.uk `site_plans` rows = 0 `[MEASURED 2026-09-02]`; deciding
  arm unchanged at :508–515 (both `section_facts` and `section_subjects` gated on
  `specSource == "site_plan_tables"`).
- **The mechanism this rides is LIVE as of today:** 638 applied 2026-08-26; **639 applied
  2026-09-02 by apis.uk, pod-verified** (`plan_sections` config carries both
  `section_facts` and `section_subjects` wirings — read from the live `agent_definitions`
  row, not the seed); 640 applied 2026-09-02 (second attempt; first correctly refused by the
  anchor guard); **641 NOT applied** (owner-read gated) → `sectionPlanItem.Subject` is stamped
  and writer-inert fleet-wide until it lands. Evidence: seed file APPLIED lines + live config
  query in RUNBOOK.
- **Census (mine, independent):** 6 plan-less real sites; 203 deployed pages
  (57/47/41/30/15/13 — reconciles exactly with the bug file's corrected figure); **11 pages
  repeat a component type** (finetuning 4, gaswholesalers 4, ai-agent-orchestration 3; other
  three sites 0). My query used `count(*) FILTER (WHERE EXISTS …)` so the 37 empty-sections
  pages were counted, not dropped (the finetuning session hit the CROSS JOIN LATERAL
  drop-trap and warned me; my shape dodged it by construction — noted, not virtue).
- **Finetuning session's five findings** (their message, folded into the bug file by them):
  (1) use `your-own-model` as the verification canary (verbatim "How it works" ×3, built
  2026-08-27); (2) do NOT link the owner's "very AI sounding" verdict (2026-08-25) to that
  copy (postdates it); (3) their census matches mine 11 = 4+4+3; (4) `section_facts` shares
  the gate — fix must carry both arrays; (5) brief-splitting is an option to TEST, not a
  recommendation (count-match unmeasured fleet-wide; misalignment is the failure the gate
  exists to prevent) — I am not pursuing it: same-guess objection as cross-tier alignment.
- **Their second message:** `our-position-on-ai` repeats with two NON-ADJACENT blocks →
  trigger is per-type count page-wide, not adjacency (my detector design already counts
  per-type, page-wide). copy_quality_two_stage lane independently concluded the fix belongs
  at the subject-publishing tier, not in any prompt. Parked stage-2 copy proposal `8003c51a`
  sits on `your-own-model` — whoever grades it needs the 443 context (their lane carries it).
- **Serve check, 11/11:** curled all 7 previously-unread damaged pages on gaswholesalers.com
  and ai-agent-orchestration.com with per-domain invented-URL controls (both 404 → 200s are
  real). All 7 repeat; verbatim-identical h2 pairs on 5 of 7 (case-study-kafka,
  containment-first, service-areas, supply-terms, wholesale-pricing); subject-level
  (near-verbatim) repetition on enterprise-reference-deployment and pricing-transparency.
  With finetuning's 4/4: **11 of 11 census-flagged pages serve real repetition; ≥8 verbatim.**
  The layout census is a confirmed damage proxy in every case tested.
- **Tier resolution of the 11:** all TIER 3 (aspect exists for all 6 sites but does not name
  these pages with sections; per-page membership query in RUNBOOK — the site-level
  "aspect present" query MISLEADS, first draft of mine did exactly that before the per-page
  join corrected it).
- **Detector bound:** 25 repeat-layout pages fleet-wide (14 plan-carrying + 11 plan-less);
  0 repeat-pages have plan subjects today.
- **Storage decision inputs:** `pages.content_direction` is free-form writer steering fed to
  the prompt (bug 025) — wrong home for a machine array; `page_spec` is admin/purpose
  metadata; object-form entries in `pages.sections` rejected (downstream expects strings —
  validate_plan's own normalise comment; `jsonb_array_elements_text` errors on objects).
  → dedicated sibling columns. 19 candidate writer files of `pages.sections` as of
  2026-09-02 → no CHECK constraint (would error them), no destructive trigger; read-guard +
  visible degrade + 5b repair.
- **No 090 run, substitution declared** (2026-07-31 ruling): the deciding arm is a five-line
  condition read directly, both censuses were derived twice independently (filing session and
  this one, identical numbers), and the predicted symptom was measured at the served artefact
  11/11. A third re-read is what 090 would buy; the council gate reviews the FIX. Queue
  checked: 0 `awaiting_diagnosis`, no open work items on the mechanism.
