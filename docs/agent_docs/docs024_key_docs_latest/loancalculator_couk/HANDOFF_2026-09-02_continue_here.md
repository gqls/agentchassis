# HANDOFF — loancalculator.co.uk · evidence register LIVE (migration 699); two copy corrections await the owner; 385 still open on its build-arm criterion (2026-09-02)

> Supersedes `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/HANDOFF_2026-08-26_continue_here.md`.
> Nothing in that file's 385 arc changed (fix live since 08-26, bug open on the one
> criterion). What is NEW is the owner-directed RFC_060 work, done this session.
> Evidence: NOTES `## 2026-09-02`. Owner prose: `README_where_we_are.md` (same date).

```
site        loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
register    ✅ LIVE — 12 facts (7 CCA 1974 sections, 2 SIs, 3 FCA Handbook rules),
            migration 699 applied 09-02 15:30; council 1f259a95 APPROVED (1 medium
            advisory, VERIFIED not filed: the refresher is map-based and round-trips
            unknown keys losslessly — the typed-writer field-loss hazard stands and
            is with the claims-verification lane). Every quote 12/12 through
            cmd/fcaquotecheck with absent control; every URL title-confirmed.
            PLUS: banned_claims = 8 (migration 707, council 99bd846e) — the
            archetype constraints translated into enforced patterns; all 8
            Go-compiled+probe-fired, 0-match census over all 28 pages, literal-%APR
            excluded on a 2-match census (pedagogy, not promotion). NOTES has the
            supersede-vs-in-place lesson and the 700→707 number collision.
385         OPEN on one criterion: a clean BUILD-ARM rebuild (bug §7b). Fix live
            since 08-26 (LOCK-009).
```

## What the register changes, from today

- The **daily refresher** re-fetches all 12 citation URLs (legislation.gov.uk AND
  handbook.fca.org.uk) and re-checks the verbatim quotes. A `citation_lost` on a
  legislation URL may mean the LAW moved (revised text) — that is the design.
- **`HasScannableRegister` now arms `ScanUnregisteredNumbers`** on future saves for
  this site. Measured-safe: RFC_060 §1c ran the scan over this site's own 474-component
  export with ZERO findings, and `fad209b92`'s regulatory exclusions are live. If a
  claims-floor refusal appears on this site, read it against §1c before treating it as
  damage (`SELECT context FROM agent_error_log WHERE error_code='CONTENT_CLAIMS_FLOOR_DETAIL'`).
- `rule` fields are HUMAN-VERIFIED (dated 2026-09-02 in each fact) until RFC_060
  §3d/Q6's rule-span checker ships. Neither host has rule-level URLs.
- ⚠ `pinned` does NOT survive a re-adoption (`write_site_spec` drops it — LANDMINES).
  After any adoption run, check the identity flags AND
  `SELECT count(*) FROM site_specs WHERE site_id='0162cde4-…' AND aspect='evidence_base' AND is_current;`

## OWNER DECISIONS PENDING (both put in README 2026-09-02)

1. **Copy correction, material:** `tools/overpayment-calculator.html` claims a CCA
   right to overpay "up to 10% of outstanding balance per 12 months" ERC-free. The Act
   has no 10% rule (threshold: £8,000/12mo, s.95A(2)(a); cap 1%/0.5%). A reader could
   rely on this. Register fact `CCA-1974-S95A` carries the correction.
2. **Copy correction, minor:** `tools/settlement-calculator.html` says "ten working
   days"; the prescribed period is 12 (SI 1983/1564 reg 4). Fact `CCA-1974-S97`.
   Per bugs_open/320 §15 the copy is NOT touched on an automated finding — when the
   owner rules, route the edit through the framework (the locked calculators are
   untouched by prose edits to sibling sections; the standing 385 §9 caution about the
   build arm applies until that bug closes).

## Standing state (carried from 08-26, unchanged)

- 385: cause §5c, fix LOCK-009 live both replicas, close = one clean build-arm rebuild
  (rerender waves do NOT qualify). Move to bugs_closed with BOTH paths named on the
  `git mv` commit when it lands.
- Harness: `toolgolden.py --selftest` FIRST, always; golden =
  `acceptance/GOLDEN_2026-08-24_post_385_repair_tool_values.json`.
- The design-rotation and 397 GTM rerender waves may still be passing through —
  churn in `created_at`s/chrome is expected; a mid-wave single sample proves nothing.
- All standing cautions in HANDOFF_2026-08-26 §Standing cautions apply verbatim.
