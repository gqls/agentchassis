
<!-- SOURCE: U14_docs019_runbooks.md -->
### Site Quality Programme — the three-way split and seven legs
- **category:** NEW:site-quality
- **status-signal:** partial
- **status-evidence:** site_quality(1) "MEASURED BASELINE (2026-07-06, the four rendered pages)" table (zero nav/img/svg/script everywhere) and the A/B/C split with legs 1–7; handed off from builder §B6 2026-07-06.
- **what:** The programme closing the gap between "deploys" and "best in class" for relay-built sites, evidence-first: split failures into A dispatched-but-stuck (LEG 1 site chrome, LEG 2 design/stylesheet delivery, LEG 3 imagery items), B delivered-but-poor (LEG 4 content depth, LEG 7 link integrity), C never-in-scope (LEG 5 feeds/graphics/games as planning criteria, LEG 6 the disabled improvement loop) — and fix in that order (dispatch before content before scope). Pre-stated decision rules; the diagnosis loop named as the deeper instrument when a direct read is ambiguous.
- **sources:** docs019/RUNBOOK_site_quality(1).md#the-task; docs019/RUNBOOK_site_quality(1).md#three-way-split; docs019/RUNBOOK_builder_route(21).md#B6
- **relations:** work-item relay; build pump (improvement-sweep disabled); content-regression guard; coverage baseline
- **verify-later:** §B6 query set results; /assets/css/styles.css existence in the sites repo

<!-- SOURCE: U14_docs019_runbooks.md -->
### Site-chrome gap hypothesis (relay path lacks chrome rendering)
- **category:** NEW:site-quality
- **status-signal:** unknown
- **status-evidence:** site_quality(1) "Zero <nav> on every page ⇒ hypothesis: the RELAY build path lacks the site-chrome rendering step (pageflow-builder has render_site_components …; build-site-planner … has populate_nav_tables — nav DATA — but no chrome-render step was observed)."
- **what:** Open hypothesis from the measured baseline: relay-built pages ship without header/footer/nav because the relay path never runs an equivalent of pageflow-builder's render_site_components; nav DATA exists (populate_nav) but chrome is never rendered/injected at assembly. Was briefly the F0 pilot before being reassigned; remains LEG 1's core question.
- **sources:** docs019/RUNBOOK_site_quality(1).md#measured-baseline; docs019/RUNBOOK_diagnosis_fix_loop(9).md#f0-pilot-original
- **relations:** site quality programme; work-item relay spine; F0 pilot history
- **verify-later:** site_components rows for dartsonline; assembler chrome injection; render_site_components reachability from the relay

<!-- SOURCE: U14_docs019_runbooks.md -->
### Site Quality Programme — the three-way split and seven legs
- **category:** NEW:site-quality
- **status-signal:** partial
- **status-evidence:** site_quality(1) "MEASURED BASELINE (2026-07-06, the four rendered pages)" table (zero nav/img/svg/script everywhere) and the A/B/C split with legs 1–7; handed off from builder §B6 2026-07-06.
- **what:** The programme closing the gap between "deploys" and "best in class" for relay-built sites, evidence-first: split failures into A dispatched-but-stuck (LEG 1 site chrome, LEG 2 design/stylesheet delivery, LEG 3 imagery items), B delivered-but-poor (LEG 4 content depth, LEG 7 link integrity), C never-in-scope (LEG 5 feeds/graphics/games as planning criteria, LEG 6 the disabled improvement loop) — and fix in that order (dispatch before content before scope). Pre-stated decision rules; the diagnosis loop named as the deeper instrument when a direct read is ambiguous.
- **sources:** docs019/RUNBOOK_site_quality(1).md#the-task; docs019/RUNBOOK_site_quality(1).md#three-way-split; docs019/RUNBOOK_builder_route(21).md#B6
- **relations:** work-item relay; build pump (improvement-sweep disabled); content-regression guard; coverage baseline
- **verify-later:** §B6 query set results; /assets/css/styles.css existence in the sites repo

<!-- SOURCE: U14_docs019_runbooks.md -->
### Site-chrome gap hypothesis (relay path lacks chrome rendering)
- **category:** NEW:site-quality
- **status-signal:** unknown
- **status-evidence:** site_quality(1) "Zero <nav> on every page ⇒ hypothesis: the RELAY build path lacks the site-chrome rendering step (pageflow-builder has render_site_components …; build-site-planner … has populate_nav_tables — nav DATA — but no chrome-render step was observed)."
- **what:** Open hypothesis from the measured baseline: relay-built pages ship without header/footer/nav because the relay path never runs an equivalent of pageflow-builder's render_site_components; nav DATA exists (populate_nav) but chrome is never rendered/injected at assembly. Was briefly the F0 pilot before being reassigned; remains LEG 1's core question.
- **sources:** docs019/RUNBOOK_site_quality(1).md#measured-baseline; docs019/RUNBOOK_diagnosis_fix_loop(9).md#f0-pilot-original
- **relations:** site quality programme; work-item relay spine; F0 pilot history
- **verify-later:** site_components rows for dartsonline; assembler chrome injection; render_site_components reachability from the relay
