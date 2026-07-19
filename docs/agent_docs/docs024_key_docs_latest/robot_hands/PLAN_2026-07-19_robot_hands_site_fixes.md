# PLAN — robot-hands.com site fixes (R1–R6)

Written 2026-07-19, retroactively for the work done 07-17/19 (the standing-five
directive landed mid-workstream). Decisions **and their reasons**; corrections
to the originating brief are marked as corrections, not edited away.

## Corrections to the originating brief (`HANDOFF_2026-07-17`)

> **CORRECTED 2026-07-18:** the brief attributed the blue chrome to the
> 2026-07-16 `needs_rerender` regeneration. Caught by bisecting the deploy repo:
> the blue entered between gqls/sites `af0ead8da1` (07-08 22:44) and
> `78532b8c63` (07-09 09:13). The regen spread it to 37 pages; it did not
> create it.

> **CORRECTED 2026-07-18:** the brief proposed restoring header/footer from
> `component_versions` snapshots. That table holds **zero rows** for those
> components. Route does not exist; restoration was done by construction
> instead.

> **CORRECTED 2026-07-18:** the brief suspected `hardcoded_section_colors`'
> handler of stripping dark backgrounds. It has never modified anything — every
> run dies `WORKFLOW_INVALID` and is stamped `complete` regardless. Filed as
> `bugs_open/017…unregistered_action…`.

> **CORRECTED 2026-07-18:** "MatchMatrix renders blank" — the page carries 6.3KB
> of visible text. It *read* as blank: white cards (`--color-card-bg: #ffffff`)
> plus `--color-primary` (dark navy) used as a text colour, on a dark page.

## Decisions

**D1 — Fix the theme by construction, not by re-swapping the FK.** The B7
FK-swap was intact; re-running it would have changed nothing. Instead: new
`header-theme-chrome` / `footer-theme-chrome` components carrying **no literal
colours at all**, using the class names the dark layout already styles. Reason:
the failure class is "a regeneration path that isn't theme-aware" — a component
that cannot express a colour cannot regenerate a wrong one. Named to sort after
existing components so no other site's `ORDER BY name LIMIT 1` fallback moved.

**D2 — Rewrite all three palette copies, not just the one that hurt.** Reason:
they are read by different paths (`style_collections` by component rendering,
`palettes` by CSS composition, `css_themes` legacy). Leaving any of them light
guarantees the next regeneration re-diverges.

**D3 — Pin `design_intent.palette` rather than trusting the agent.** Reason: the
design prompt renders only the *structured* palette block; this site had only
free-text `colour_mood`, invisible to the LLM, so every run hit the "no intent →
invent" branch. Pinning is a data fix on one site; the check-side fix (D4) is
the fleet fix.

**D4 — Fix `generic_theme` in code, not per-site.** Reason: it demanded a marker
(`site_specs aspect='webdesign'`) that **no code path has ever written**, so it
fired on every themed site forever. This is a platform defect, so it belongs in
Go and went through the council gate.

**D5 — Learning centre: the hub is canonical; keep the grid page, archive the
residue.** Reason: the hub is the real listing (query-backed, imagery-carded).
The category grid page is linked from body copy across the site — archiving it
would convert every one of those into `phantom_internal_link` churn — so it
stays active but demoted. The plan-residue index page has no inbound links and
its rebuild had failed twice; archived.

**D6 — Retire all six dead article rows rather than build them.** Reason: three
are plan-era scaffolds with placeholder titles, three are duplicate slugs of
guides that already exist and are deployed. Nothing worth building. Their
generated card/hero assets are superseded in the same transaction so they don't
become orphans.

**D7 — Hide the Load More button rather than implement pagination.** Reason: it
has no behaviour anywhere in the fleet, and with three real articles pagination
is not the honest v1. The component-level default was flipped to opt-in, since
a dead control being default-on is the actual defect; explicit opt-ins on other
sites were left for their owners.

**D8 — Split the council submission; don't smuggle the render guard through.**
Reason: the council's `guardian` seat was right that a shared render-merge
change and a single-file check fix have different blast radii and shouldn't be
judged as one unit. The guard was withdrawn and filed as `bugs_open/022` with
its verifications completed, to return as its own submission.

**D9 — Stop the council loop at three rounds.** Reason: CLAUDE.md says one run
per coherent task, not per iteration. At round 3 the standing was 7 of 8 approve
with one dissent whose points were answerable — one by a grep (no other check
shares the shape), one by a small refactor. Spending a fourth round on a
low-severity design point would spend credits against explicit guidance.

## Phasing (as executed)

1. R1 palette + components + pin → re-render CSS → re-render 37 pages.
2. R6 retire rows (unblocks what the listing shows), R3 nav, R5 button — all DB.
3. Second re-render pass to bake the new nav (the first pass predated it).
4. Platform fix (`generic_theme`) through the council; bugs filed for what
   couldn't be fixed in scope.

## Open

- **R4**: two planned tool pages never built; five homepage CTAs point at their
  404s. Belongs with the experience_loop workstream — see the handoff.
- `bugs_open/022` (scheme guard) and `bugs_open/017…unregistered_action…`.
- Neither bug file has been through the 090 diagnosis loop, which became the
  default for durable claims *after* they were written.
