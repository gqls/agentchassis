Everything's captured and corrected. Here's the complete status.

Punch-list: 5 fixed, rest documented and actionable

Fixed and live this turn:
- Nav colour consistency — gold header + charcoal footer across all pages (27/30; the 3 exceptions are empty pages). The stray navy hexes are dead CSS fallbacks that never render.
- Nav decluttered to a clean 9 items; blank "For Leaders" gone.
- Card 404 links — template gated, phantom /how-we-work/* links stripped, use-cases quiz link repointed.
- About invented stats — made honest and true (30 years / 8 sites / 2,767 records).
- favicon.png — was 404; the head hardcodes /assets/images/favicon.png and we'd only committed the .ico. Committed it; now 200.

Two important corrections the research surfaced

- The A6 Banana routing is actually deployed (v1.0.1114) — my earlier "not deployed" note was stale. So images are not blocked by any deploy. The real reason leopardess has no images is that it has no
  site_plan/site_plan_imagery rows at all. The handoff now has a §9 playbook: Route A (per-image trigger — safe, immediate, how robot-hands got its heroes) is the way in; Route B (a full re-plan) is flagged
  as risky because it could overwrite the copy we carefully fixed.
- The blog is less broken than it looked — it already renders 5 posts with working links. The empty look is blank card images (a structural Phase-I3 gap, not built) plus empty excerpts (a quick fix:
  populate meta_description).

The one honest limit worth stating plainly: per-card and per-section images (the blog thumbnails, the info-card images) need "Phase I3", which doesn't exist yet — that's real new build work, not a config
toggle. Hero images per page, though, can be done today via Route A.

Your deliverables

- HANDOFF.md opens with the punch-list table (issue → status → root cause → fix) and now has the full imagery playbook (§9). It's the single doc to open a fresh chat from.
- RUNNING_NOTES.md — turns 15 and 16 record the whole review and every fix.
- specs/VOICE_REWRITE_PROMPT.md — your "think hard about a prompt" ask: a specific, opinionated prompt naming the exact LLM tells (the balanced triads, the "not-X-but-Y" strawmen you can see in that
  how-we-work hero, the summarising flourishes) with a worked before/after.

Everything is backed up. The site's structure and chrome are now genuinely solid; the substantial remaining work is content depth — applying that voice rewrite, hero imagery via Route A, the blog
excerpts, and the "we could but don't" reframe on use-cases. All of it is now root-caused and actionable in the handoff, so you can pick up any piece in a new chat.
--