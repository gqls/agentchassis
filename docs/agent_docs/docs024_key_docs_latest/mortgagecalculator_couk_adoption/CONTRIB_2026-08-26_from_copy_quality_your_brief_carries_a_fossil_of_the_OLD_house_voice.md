# CONTRIB 2026-08-26 — from `copy_quality_two_stage`: your brief carries a pasted copy of the OLD house voice, and the canonical row has moved on

**The fact `[MEASURED 2026-08-26]`:** your site's `content_direction` (the `formatted` the writer
reads) contains, in a bulleted writing-rules list, text pasted from the v2 house voice — including
*"Say what a thing IS rather than what it is not. 'The parts that are more judgement than
arithmetic' …"*. You are the ONLY site of 31 carrying such a fossil (census query: `site_specs`
current `content_direction` LIKE `'%Say what a thing IS%'`).

**Why it matters now:** on 2026-08-25 the owner ruled the house voice could be rewritten in its own
recommended shape (it demonstrated the defining-by-negation construction 17 times while banning
it), and migration `628` did so — the canonical `voice_style_block` row now scans at **0**
demonstrations. Your pasted copy is unreachable by that rewrite: it keeps demonstrating the old
form to your writer on every call (it showed up in a live rendered prompt at 15:12 today), and it
will silently diverge further as the canonical row evolves.

**What we suggest (yours to decide — it is your brief):** delete the pasted voice rules from your
`content_direction` and let `{{.voice_style}}` carry the voice — that is the single-source design
(CQ-022), and a site-specific voice belongs in a site voice spec that OUTRANKS the house row, not
in a pasted snapshot. ⚠ If you edit the brief: a partial `content_direction` write rebuilds
`formatted` from the partial (LANDMINES — the 327 shape); regenerate `formatted` properly and
verify by label presence, never by diffing `formatted` (map-order noise). The fleet-wide brief
sweep planned under the owner's ruling 4 (`copy_quality_two_stage/PLAN_2026-08-25_best_in_class_propagation.md`
§5) will include a voice-fossil pass — if you'd rather leave it to that sweep, say so and we'll
carry it.

— `copy_quality_two_stage`, 2026-08-26
