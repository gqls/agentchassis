Done — the copy problem you spotted is fixed, and it turned out to be two layered bugs.

What you saw, and what was actually wrong

The Catalyst page had structure but hollow copy: generic "about the site" text that never named the archetype, plus a garbled stat strip ("24h Clients Served, 1 Daily, 50/50 Awards Won"). Two separate causes:

1. The content-writer wrote boilerplate — it treated each archetype page as a generic about-page. Fixed by authoring canon copy straight into content_data (migration 089), drawn from the spec's actual archetype
   definitions. Catalyst now reads: "The Catalyst doesn't need the best take in the thread — yours is the one the thread grows from… the longest response chains in the Arena trace back to something you filed." CTAs now
   point at the Gauntlet and quiz tools instead of /contact.html.

2. A fleet-wide schema bug the copy pass exposed — after 089, the bodies were right but the stat labels still said "Clients Served". Root cause: on the shared content-block-about component (13 pages across 5 sites),
   the three stat labels and the CTA label were source: "static" with hardcoded business defaults. Every render re-applies them, so no content edit could win — and that's exactly why the LLM had crammed the real label
   into the value on every site ("500+ Models / Clients Served" on robot-hands, "30 yrs / Awards Won" on ai-agent-orchestration). Migration 090 flipped those four fields static→llm (keeping the defaults as fallbacks).

Verified end state

All 8 archetype pages are live and clean — sweep confirms zero boilerplate labels, each stat descriptor is archetype-specific ("Longest / Response chains started", "Widest / Angles opened per Provocation"), and the
CTA reads "Find Your Archetype". The four business sites are untouched: the schema change only bites on re-render, their labels are safely persisted in their own content_data, and I confirmed ai-agent-orchestration's
about page still serves "70+ / Clients Served" with an unchanged DB row.

The transferable lesson (now in the RUNBOOK and running notes): a source: "static" schema field re-applies on every render — no content_data edit or LLM pass can override it. When authored copy "won't stick," check
the field source before assuming the write failed.

Two things I'd still flag for your call, neither blocking:
- The archetype-combinations section on the hub still uses the off-canon 6-name taxonomy (Contrarian/Analyst/Sage…) rather than the canonical 8. Pre-existing; worth aligning in a follow-up.
- The business sites now have a latent improvement available — next time any of them rebuilds its about page, the writer can author proper per-site stat labels via the guidance instead of being locked to "Clients
  Served". Nothing breaks in the meantime.

Worth an eyeball on the refreshed Catalyst page when you have a moment to confirm the voice is right.

--