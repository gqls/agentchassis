# HANDOFF — brochure component library / fundamentallyai.com — 2026-07-28b (afternoon)

**Cold-start document, supersedes `HANDOFF_2026-07-28_continue_here.md`** (read that one
next anyway — its §3 cap story, §4b owner rule, §4c–4e design directions and landmines
L11–L17 all still stand; only its §4 action list is superseded here).

## What changed this afternoon (evidence: NOTES 07-28 afternoon entry)

- **§4.1–4.4 of the morning list are DONE.** Canary complete; capabilities chart correct
  (2 charts, 085 held); link crawl run; **both `needs_imagery` items complete with stored
  assets**. The llm-cost-calculator hero is live and serving
  (`/assets/images/content-hero-llm-cost-calculator.jpg`, 200, no new broken refs).
- **`bugs_open/079` is REOPENED — the link repair's output is DISCARDED at save.**
  The gate repaired all 9 invented links on the 10:05Z build; `save_page_sections`
  persisted the unrepaired sections 400ms later (structured `sections_metadata` path
  wins; `clean_html` is fallback-only). Full mechanism + fix candidates in the 079
  REOPENED banner. Consequence: **there is NO downstream mitigation for invented links
  or srcs — 092 (writer never gets constraints) is now the live front line.**
- **`/assets/illustrations/` is pure invention** (1 component fleet-wide — the regressed
  carousel). The §4b regression is one defect with the links, filed into 071/092.
- **Anthropic spend cap EXHAUSTED until 08-01 00:00 UTC** — councils + diagnosis loop
  die on `provider=anthropic`; failures show `COMPLETED` with the refusal in
  `__step_error` (099 family). Gemini text/image (this site's lanes) work fine.

## Constants

Unchanged — see the morning handoff §1. DB access, site/plan/page ids, `.html`-only hrefs.

## Next actions, in order

1. **Verify the selector tool's hero landed.** Work item `56fbcc9a` ("Re-render
   tool-model-approach-selector after its image asset landed") failed at 12:20Z
   because the image didn't exist yet, parked `needs_human_review`, retried ~14:55Z
   after asset `d76f0282` landed. Check:
   `curl -s https://fundamentallyai.com/tools/model-approach-selector/index.html | grep -o 'content-hero[^"]*'`
   and the item's status. If it failed again, read the error before retrying — the
   "missing data" premise is gone, so a second failure means something new.
2. **The capabilities page still serves 13 broken refs** (9 links + 4 carousel svgs).
   Repair now has a decision to make that the morning handoff couldn't see: a hand
   repair still won't survive a rebuild (071's proven point), and the automated repair
   is proven vacuous (079 REOPENED). Options: (a) fix 079 candidate 1 (repair inside
   `save_page_sections`) then re-render; (b) scoped section re-render with corrected
   content_data for the carousel (LLM-free, safe, but same invention next rebuild);
   (c) wait for 092. The owner's "replace before deleting" rule applies to the images.
3. **`bugs_open/128`** (`image_url_404` never makes an HTTP request) — untouched, still
   worth a thread.
4. **After 08-01:** fire the owed 090 diagnosis verification on the 079 reopen claim,
   and any council submissions queued during the cap.
5. **Owner design directions** (§4c carousels/cliffhanger, §4d experience-register join
   check + four wrong names, §4e shapes registry) — all untouched, all still the
   biggest open design surface. Start with the §4d join check: it is a few lines.

## Landmine added this session

- **L18 — a repair that logs durably still may not persist.** `CONTENT_LINK_REPAIR_DETAIL`
  rows are records of what the gate COMPUTED, not of what was saved. Two
  representations of the page travel the build (`clean_html` + `sections_metadata`);
  writes to the one that loses are silently vacuous. Verify content fixes against the
  saved `page_components` row, then the served page — never the action's return map.
  (016b §9 has the general pattern.)
