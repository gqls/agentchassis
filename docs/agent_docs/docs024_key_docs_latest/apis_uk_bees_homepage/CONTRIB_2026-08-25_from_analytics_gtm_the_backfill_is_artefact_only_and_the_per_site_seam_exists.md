# CONTRIB 2026-08-25 — from `analytics_gtm` (session "google"): your 08-24 GTM backfill is artefact-only, and the per-site id seam your §3 designs already exists

Written into your directory because two things in `HANDOFF_2026-08-25_continue_here.md` are now
wrong in ways a reader cannot see, and one of them is a build you told the next session to do.
Nothing here is a criticism of the roll-out — it served the tag on 24 sites within the hour and
every check you ran was the right check for what it measured.

## 1. "All 27 heads carry it" — true at 13:13 on 08-24, false by 19:20 the same day

Your backfill inserted the snippet into `site_components.rendered_html`. The head templates do not
contain the snippet; they contain `{{if .gtm_container_id}}…{{end}}`, and that value is read from
`site_specs` (`aspect='site_config'`, `data->'analytics'->>'gtm_container_id'`) — the seam the
07-31 rollout used for its 14 sites (`analytics_gtm/sql/c1_gtm_fleet_rollout.sql:168-176`). A chrome
render regenerates the head from template + inputs, and with no key the gate is false.

`[MEASURED 2026-08-25]` **agritec.uk lost the tag at 2026-08-24 19:20:53** — its lane added a tool
page (`add_tool` → `nav_drift` → `nav_rebuild`), the head regenerated, and the platform filed
`chrome_divergence_overwritten:site_component:head:920676eab287` at `needs_human_review` the same
second (the only such head row in the estate). The other **12** of your 13 are at
`updated_at = 2026-08-24 13:13:52` with no key: `apis.uk cookly.uk garden-tools.uk lendzy.co.uk
loanandmortgagecalculator.co.uk loancalculator.co.uk loancash.co.uk loanzy.uk
mortgagecalculator.co.uk noted.co.uk remortgagecalculator.uk webdesign.uk` — each reverts on its
first nav change. **`bugs_open/397`.** Your own §1 "apis.uk … no footer … GTM present" is in that
list; so is your "1 × `overwrite: REFUSED` … our own lock working" — the lock protects
`page_components`, not the head slot.

**Fix:** `analytics_gtm/sql/c2_gtm_spec_key_for_artefact_only_sites.sql` — writes the key, merges
into existing `site_config` rows, dry-run verified. ⚠ It is a **rebuild wave** (241 pages: the
`stale_site_components` check fingerprints `site_config` and `rerender-pages` force-rerenders every
slot and page), so it is held for the owner's timing, exactly as your 08-24 fan-out was.

**Cheap check for next time:** when a value lives inside `{{if .x}}` in the template, census `x` in
the spec, not the value in the artefact — `analytics_gtm/scripts/check_gtm_state.sh --db`.

## 2. §3 "Per-site analytics id … read `sites.settings->>'analytics_container_id'` in `RenderFallbackHead`" — please do not build this

- The per-site seam **already exists and is live**: `site_config.analytics.gtm_container_id`
  (STY-050, 07-31, 14 sites keyed, `input_schema` source `config.analytics.gtm_container_id`). A
  second key in `sites.settings` is two switches for one light, and the concept-register bar
  ("another workstream could call this and would not know it exists") is the thing that happened
  here.
- `RenderFallbackHead` (`component_library.go:2243`) runs only when the head *component fails to
  render*. A tag there hides a broken head and never reaches a working site.
- Your intent is right and the seam already satisfies it: **empty key ⇒ no tag; nothing hardcoded;
  a third-party site never gets ours unless someone writes its spec.** What is missing is the other
  direction — *our* new sites get no key at birth (no Go writer touches `site_config`; the four
  sites built 08-24/25 have no row at all). That is 397 §6.2, council scope, and it belongs with
  the handover question (`web_admin_console`), not in the fallback renderer.

## 3. Two smaller things

- Your §1 "GA4 NOT published — nothing is being recorded" is still true at 16:06 BST 08-25
  (container version 2, `"tags":[]`). Your §4a walkthrough is the pending owner action; it is
  cross-referenced from `analytics_gtm/HANDOFF_2026-08-25_continue_here.md`, which is now the
  cold-start for fleet tracking.
- `webdesign.uk` 302s to `webdesign.co.uk`; a curl without `-L` reads `gtm=0` there and is not a
  regression.
