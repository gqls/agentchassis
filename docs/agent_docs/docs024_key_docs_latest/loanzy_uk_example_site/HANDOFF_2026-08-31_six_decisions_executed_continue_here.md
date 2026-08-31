# HANDOFF 2026-08-31 — the owner ruled on all six open decisions; five executed, one designed. COLD-START HERE.

Supersedes `HANDOFF_2026-08-26_council_live_farmer_dispatched_continue_here.md` as the lane
cold-start (that file stays the council-mechanics reference — its §3 proofs are now DONE, see
its appended correction). Site: farmerinsurance.uk `99cae989-2413-430d-b026-59dfeeb638c0`.
Canaries: loanzy.uk, lendzy.uk. RFC: `architecture_review/RFC_056_…`. Bug of the day:
`bugs_open/417` (planner logo exemplar) — Go-free, migrations 669/670 shipped against it.

## 1. The six owner decisions (2026-08-31, verbatim-in-substance) and their state

| # | decision | state |
|---|---|---|
| 1 | logos carry **NO WORDS** | **DONE** — migration `669` applied+verified (planner exemplar → text-free form, snapshot `669_pre_logo_exemplar_no_words`); council corr `3b666f0f-0055…` (Council-Submitted on the commit; resolve at report time) |
| 2 | **rewrite** the 19 propagated plan prompts | **DONE** — migration `670` applied+verified: 10 verbatim rows replaced, 9 varied rows carry the dated override; backup `bak_670_plan_imagery_wordmark` (19 rows, kept); post-census licence=0. ⚠ verify by counting the **LICENCE** (`no text outside the wordmark`), never the prohibition (417 §2) |
| 3 | carousel **default ON fleet-wide**; card ORDER = offer/benefit thread decides, or randomise per refresh | **RELAYED** with the ruling quoted: offer-analysis lane (SendMessage, acked earlier rounds) + `staged_component_build/CONTRIB_2026-08-31_from_loanzy_lane…` (their lane has no live session — the file is the channel). Implementation is THEIRS; they hold the measurements (flag on 1/42; semantic_tags-derived vocabulary; grid-mis-tagging grep trap) |
| 4 | **delete** the seven farmer tools | **IN FLIGHT** — see §2. 21 pages archived; 8 files deleted; 13 refused pending link-drops; 6 moot `owned_page_review` rows cancelled |
| 5 | build the **"this site is ready"** growth posture | **DESIGNED, not built** — full design + producer census in `PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md` "ADDITION 2026-08-31". Key: `sites.settings->maintenance_profile->>'growth_posture'='hold'`; on hold, the three tool-chain producers file growth in the RECORD SHAPE (deferred + handler ''); `source='owner-request'` exempt; audit-seat growth is ALREADY held by record mode. Next session implements (Go helper + 3 call sites), council gate, register on ship |
| 6 | **cancel** the 59 pre-door opinion rows | **DONE** — census re-run first (reproduced 59 exactly, all `needs_human_review`), cancelled with the owner-reason stamped in `result.cancelled_reason`. The 44 `capability_gap` deferred rows are a DIFFERENT class, untouched by design |

## 2. Decision 4's remaining half — the retraction loop (TAKE THIS OVER FIRST)

State at handoff: `pages.status='archived'` on all 21 (ids in the 12:2x commit/NOTES; the
cluster = 8 `page_type='tool'` + 13 tool-guide blog/guide pages). First retraction run
(orch `fbae5325-…`, COMPLETED): **21 considered, 8 dispatched for deletion, 13 REFUSED** —
"still linked from live content", referrers in the `RETRACTION_AUDIT` row's
`context->'editorial_inbound'` in `agent_error_log` (16 live pages, hero + CTA slots).
**16 `page_rerender` items filed** (item_key suffix `_tool_retirement`, promotable —
handler/pipeline set; see the RUNBOOK's hand-filing recipe, learned this day) so the links
drop at render. A Monitor was watching them at session end.

Then, in order:
1. When all 16 are terminal: **re-fire the retraction for the 13 refused ids** (same
   `216_TRIGGER_page_retraction.sh` recipe, RUNBOOK). Expect refusals=0; any residue names
   its referrer — rerender that page and go again.
2. **Acceptance is TWO-PART** (216's header): every one of the 21 urls 404s now, AND still
   404s after the next ~08:0x/20:0x refresh with zero fresh `page_rerender` rows for them.
   The 8 already-deleted urls: 1 confirmed 404 at 12:30Z, the rest unchecked — curl all 21.
3. Nav check on the survivors: no remaining page should link a retired url
   (`editorial_inbound` empty on a dry-run, or grep the served html).

## 3. Watches + follow-ups a new session inherits

- **Council verdict** for 669+670: corr `3b666f0f-0055-4e3a-813f-29458007af3f` — watch
  doc_notes (RUNBOOK "verdict queue read"), not orchestration_states. On APPROVED nothing to
  do (Council-Submitted resolves at report time); on REVISE, answer it.
- **Farmer's logo has a WORDED wordmark ("farmerinsurance") by the owner's earlier same-day
  instruction; decision 1 (“no words”) postdates it.** Flagged to the owner in the six-decisions
  reply — do NOT regenerate unless he says so. 670 already rewrote farmer's stored prompt, so
  any future regeneration comes out text-free automatically.
- **Record-mode council: all proofs DONE** (retraction behavioural: 57 self-retracted /
  228 streaked / 1,819 held; origin stamping total post-roll; §11e from durable stores).
  Remaining RFC_056 FOLLOW-UPs: 5 (design-audit child fail-open), 6 (verdict-release
  surface — 1,819 held rows and growing makes this warmer), 1 (growth-refusal — now
  decision 5's build). FOLLOW-UP 7 CLOSED by decision 6.
- **Owner review routing (08-31 morning)**: all six findings landed — copy lane has the
  tone-vacuum variant; news ask sits in `bugfix_316_…/CONTRIB_2026-08-31_…` (capability
  absent fleet-wide, seam web_search_action.go, TLD default at seed time, population count
  owed); Drewberry link-outs in `bugfix_206_…/CONTRIB_2026-08-31_…`; logo chain closed
  (417 + 669/670 + regenerated assets, favicon/og re-derived 11:15Z, eye-verified).
- **417 disposition**: fix candidate 1 SHIPPED (669+670). Candidate 2 (pixels-vs-identity
  check) deliberately unbuilt — architecture-scope, classifier-inherits-gaps caution on
  file. 417 can move to bugs_closed once the council verdict lands and the exemplar census
  stays clean one more build (fixed AND live bar: 669/670 are DB config = live now).
- **Residual quality note, not churn**: farmer's favicon shrinks the whole worded logo →
  wordmark illegible at 16px (NOTES 08-31). A mark-only favicon crop is imagery-family
  work; do nothing on farmer.

## 4. The lane's standing state (unchanged from the 08-26 handoff except as above)
Record-mode council LIVE fleet-wide (mig 624 + 623/625/629); promoter has 5 doors incl.
origin; silence-retraction proven in production; farmer serving 18 active pages post-cull
(39 − 21). Working docs: NOTES (newest at bottom), RUNBOOK (hand-filing recipe + verdict
queue + re-fire recipe), OWNER_REVIEW_2026-08-31 (findings + routing outcomes), PLAN
(ADDITION 08-31 = decision 5 design), WRONG_CALLS + LANDMINES entries per their files.

---
> **CORRECTED 2026-08-31 (same session, later):** §2's "16 page_rerender items filed … so the
> links drop at render" was WRONG twice over, and round 2 of the retraction refused all 13 to
> prove it: (a) the guard reads **stored content_data** precisely BECAUSE a rerender rebuilds
> from it (`retract_page_graph.go:145` — the estate had this documented); (b) the CTA
> recompute inside page-rerender is **REASON-GATED — it runs only for
> `reason=cta_links_stale`** (`rerender_page_sections_action.go:539`), and my items said
> `tool_retirement`, so the recompute never ran. The 16 plain rerenders were a wasted cycle.
> The working repair (in flight at correction time): **15 `cta_links_stale` rerenders**
> (item_key `misdirected_cta:<page>:<site>` — the check's own shape) so `applyCTARecompute`
> re-mints the CTA/hero targets from the LIVE page set; then re-fire the retraction for the
> 13. The CTA values were safe to recompute because their `__cta_minted` stamps MATCH the
> stored values — minted, not authored (cta_provenance.go). Watch the guide-list component on
> guides-index separately — a LIST is outside the recompute's field set.
