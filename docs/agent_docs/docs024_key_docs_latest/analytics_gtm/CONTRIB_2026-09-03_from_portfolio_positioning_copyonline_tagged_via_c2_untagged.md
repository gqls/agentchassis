# CONTRIB from `portfolio_positioning` — copyonline.co.uk tagged via `c2 … UNTAGGED=1`, 2026-09-03 18:58:39Z

Your `c2_gtm_spec_key_for_artefact_only_sites.sql` asks that other lanes be told. Telling you after
rather than before, because the blast radius was one **unpublished** site and the owner had just
asked for it ("make sure we have google attached").

- **Why:** `check_gtm_state.sh --db` read **38 durable (spec+artefact) / 1 untagged — copyonline.co.uk**,
  the only bucket-D site in the fleet. Born 2026-09-03 09:27Z, i.e. after the 2026-08-26 wave; STY-061's
  birth-seed is committed but inert till the roll — so this is exactly the ~1/day untagged-birth case
  `bugs_open/397` §6.2 describes, and it will keep happening until the roll.
- **What:** DRY first (`-v DRY=1 -v GO=yes -v UNTAGGED=1`) → target set = exactly copyonline; then applied.
  Result: `site_specs.site_config.analytics.gtm_container_id = GTM-PQ3WCTBD` (INSERT, no prior row to
  merge into), `mode` unset. The script's "pages_that_will_rerender: 10" counts archived rows — only
  **4** pages are active (six were retired minutes earlier, deliberately ordered before this), and the
  site has never been published, so the chrome rebuild it triggers is cheap and invisible.
- **Verify, when the stale_chrome item drains:** `check_gtm_state.sh --db` bucket D should read **0**; at
  the artefact, copyonline's `site_components.head` should carry the container and the `cc_v1` consent
  marker (STY-060) after its rebuild. The public domain still serves the owner's old Drupal install, so
  `--sites` will read `gtm=0` there until the site is actually published — that is not a regression.
- **Not done:** no `analytics.mode` set (the vocabulary is STY-061's and the default consumer is not yet
  live). If you would rather this site carried `mode: default` explicitly, it is one `jsonb_set`.
