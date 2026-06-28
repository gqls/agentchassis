Here's the open backlog, pulled from the journal and grouped roughly in the order it makes sense to tackle. (For orientation: the coordinator result-extraction fix is done and field-validated — everything below is still open.)
idea.uk page content — the page builds now, but with holes

Wire the section-data reconciler. reconcile_section_data_action.go was built on 2026-06-02 but never wired to a host (registry entries pending). This is the likely root of both the empty differentiators and the unresolved pricing — so it's probably the structural fix behind two of the defects.
Empty differentiators (7 blank cards) — confirm it's the reconciler vs a writer content gap, then fix.
Unresolved CTAs (hero + call-to-action buttons render empty) — point them at the real intake (/request or the email) or at real destination pages. Tied to the thin plan having no hub pages.
Dead contact form (posts to #contact) — wire to the real flow or drop it.
Pricing section (needs_section_data) — capture the £29 into site_specs.pricing, or drop the section.
Thin nav/footer (only Home; footer "Our Services" lists only Refund Policy) and empty meta description — minor, downstream of the four-page plan.

Design leadership

The design_intent (parchment/rust, avoid blue/gradient/rounded) doesn't reach the install — the theme block came through empty and the header/footer hardcode blue. Sub-parts: classifier should emit a structured palette (or the cascade parses hex from the prose); curate distinctive layouts/typography/palettes; make the header/footer and other components read palette variables instead of hardcoded colours; add Fraunces/IBM Plex or accept the fallback; decide whether tool-portal-dark reads as leading.

Content ambition / mission

Standing-ambition default in domain-submitter — a Go action carrying a default in the mission aspect, merged with any owner mission (reuse persist_mission_brief/write_site_spec, no schema change). You still need to finalise the principle wording (draft is in the notes).

Framework follow-ups (lower priority)

Optional final confirmation: the resolveResultSpec … mode=flatten log line on the 09:51–10:08 writer run.
Broader validation of the fix: requeue the gamesdesign confirmed-stub page (the unambiguous proof), and/or a full rebuild that exercises the writer on every page.
Optional deprecated-key rename migration (output_field→result_from, output_fields→multiple_output_fields, output→result_mapping) — old names keep resolving, so no rush.

Housekeeping

Tidy duplicate idea.uk spec rows and the stale duplicate index work item (bce22606).

VM deploy — Phase 2 (the big one)

Static build to /var/www/idea.uk; Go tool stays on 127.0.0.1:8080; nginx serves / and the framework pages, proxies the reserved tool paths (/request, /audience-check, /confirm, /approve, /decline, /op, /stripe/webhook, /order/, /internal/, /health, /capacity, /terms, /refund-policy, /privacy); never static-serve the webhook or operator paths; full pay→webhook→deliver test; cutover and rollback are each one nginx line; DNS unchanged.

Eventual

Full fresh rebuild of idea.uk as the end-to-end confirmation and a clean, improved site — after the content, design, and mission work is in.

One genuinely separate item, parked from the live-service work (not the chassis site): rewriting the £29 report and audience-check language/format (less dense, nicer layout) — deferred a while back and never picked up.
The immediate fork is still the one from last turn: the empty sections (starting with wiring/identifying the reconciler) or the design palette. If it's useful I can drop this list into the runbook as an open-items checklist so it's in one place for next time.