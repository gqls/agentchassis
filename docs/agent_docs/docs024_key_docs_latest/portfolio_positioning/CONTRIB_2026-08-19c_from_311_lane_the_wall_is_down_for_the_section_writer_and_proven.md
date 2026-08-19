# CONTRIB from the bugfix_311 lane — your handoff's "311 is NOT fixed" is stale; the section half is live AND proven, the tool half awaits a roll (2026-08-19 20:20 UTC)

For `portfolio_positioning/HANDOFF_2026-08-19_continue_here.md` §3 ("THE WALL") and §1 (the
halt). Evidence: `docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/NOTES_311_fix.md`;
status line of `bugs_open/311_HANDOFF_2026-08-18_…md` updated the same minute.

- **Section writer (`store_generated_component`)** — commit `17d883333`, council `fc3ac5f4`
  APPROVED, LIVE on v1.0.1315, and **exercised on a real collision 2026-08-19 16:23Z**: loanzy's
  `loans-car-finance-calculator` diverted to a site-scoped base row instead of deadlocking; all
  eight loanandmortgagecalculator.co.uk incumbents byte-identical afterwards; the page rebuilt
  20:16Z and serves 4 `<input>` where it served 0. Your frozen pilot was not needed and was not
  touched.
- **Tool writer (`create_tool_component`, RFC_036 §9.3)** — commit `e24bc9c0f`, council
  `ceae30f2` APPROVED round 1, **NOT ROLLED** (chassis v1.0.1315 at 20:17Z). Per the owner's
  ruling the pair is the precondition for wave 1, so wave 1 still waits on a roll — nothing else.
- **A correction that affects your wave-1 expectations:** a page parked on the old failure does
  NOT heal by itself. `needs_rebuild` has no consumer; each hollow tool page needs a `needs_page`
  re-render filed (recipe in the 311 RUNBOOK). For a GREENFIELD build this is moot — the
  one-shot path files its own items and the store now succeeds first time.
- `adversecreditmortgage.co.uk` build #1's 41 held `needs_page` items: when the halt lifts they
  will run through the fixed writer; expect diversions (finding rows) rather than
  `needs_new_component` failures, and expect the `loans-credit-health-check`-class max_tokens
  failure to be unaffected (upstream of the store).

> **CORRECTED 2026-08-19 20:40Z (same lane):** "NOT ROLLED — chassis v1.0.1315 at 20:17Z" was a
> stale carry-forward; v1.0.1316 rolled at 17:13Z and carries BOTH halves (stamp `07eeba4a1`,
> ancestry TRUE for `e24bc9c0f` and `17d883333`, both literals present, controls clean). The
> pair is LIVE. What wave 1 now waits on is only the tool half's first real exercise
> (webdesign_tool_rebuilds Phase D, `tool-ab-test-calculator`) — and the owner's halt.
