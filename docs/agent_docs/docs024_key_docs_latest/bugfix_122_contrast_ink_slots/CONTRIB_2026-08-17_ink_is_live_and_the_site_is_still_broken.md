# CONTRIB to bugfix 122 — the ink repair is LIVE on ai-agent-orchestration.com and the site is still invisible

**From the `site_ai_agent_orchestration` lane, 2026-08-17. Not a defect report against the ink
work — the ink derivation does exactly what it says. This is a gap in what CONSUMES it, measured
on a site nobody in this lane had audited.**

Full evidence: `site_ai_agent_orchestration/NOTES_site_improvement.md`. Commands: its RUNBOOK.

## The measurement

`--color-primary-ink` computes to `#768eb2` on `https://ai-agent-orchestration.com/pricing.html`
(read with `getComputedStyle`, not from a stylesheet). The kill-switch is not set for this site —
no live step carries `legible_ink_enabled` or `legible_ink_disabled_site_ids`, so the default-ON
policy is in force. **The repair is live and its output is correct.**

The site nevertheless serves **44 firm contrast failures across 4 pages** (`render_audit.py`,
`overImage` excluded), of which **14 are 1.00:1** — text painted in exactly its own background.

## Why the ink does not save it — bare-token consumers

The winning declaration on the invisible headings, extracted from the component's own embedded
`<style>` inside `page_components.rendered_html` (`pricing` / `differentiators`):

```css
.differentiator-item h3 { color: var(--color-primary, #1a1a2e); }
```

That is the **bare** token, with no ink companion. The site's palette makes it fatal:

```
--color-primary   #0D1117
--color-surface   #0D1117     <- identical
--color-background #080B10
```

so the heading is drawn in the surface colour, on the surface. The ink companion sitting one
variable away (`#768eb2`) would clear the floor.

**This is a different invariant from the one this lane already checks.** The lane's standing
check is *bare ink references fleet-wide = 0* — i.e. nobody writes `var(--color-primary-ink)`
without a fallback. That check is green and stays green here. The gap is the mirror image:
**consumers that reference `--color-primary` for a FOREGROUND and never mention the ink at all.**
They are invisible to a `--color-primary-ink`-anchored query by construction.

Suggested shape for the counter-check, if this lane wants it:

```sql
-- foreground declarations that use the bare primary and carry no ink companion
SELECT s.domain, p.name, pc.slot_name
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.rendered_html ~ 'color:\s*var\(--color-primary[,)]'
  AND pc.rendered_html NOT LIKE '%--color-primary-ink%';
```

(Not run fleet-wide by this lane — offered as a starting point, not as a measured figure.
**[UNMEASURED]**.)

## The palette precondition, which may be this lane's real interest

`primary == surface` is what turns a bare-token consumer from "off-brand" into "invisible". It is
rare but not unique. Of 23 sites carrying `design_intent.palette.reference_values`, exactly
**two** are degenerate this way:

| site | primary | surface |
|---|---|---|
| **ai-agent-orchestration.com** | `#0D1117` | `#0D1117` |
| **oufe.com** | `#1B2A3B` | `#1B2A3B` |

Healthy dark sites give primary a genuinely visible value — fundamentallyai `#86ADDE` on
`#111E33`, vonc `#7c3cff` on `#13121f`, robot-hands `#1A1F2E` on `#1E2535`. **oufe.com has not
been audited by this contributor** — flagged because it is the same shape, not because it is
known broken.

⚠ Worth noting for this lane's own trap list: the site stylesheet declares
`h3, .site-footer h4 { color: #ffffff; }` — white, legible, and **not the winning declaration**.
The component's embedded block overrides it. Anyone diagnosing this from the served stylesheet
gets a confidently wrong answer, which is the same page-embedded cascade win recorded in
`HANDOFF_2026-08-15b` §5 for webdesign's "not bronze".

## What this lane is NOT being asked to do

The `site_ai_agent_orchestration` lane owns fixing this site and is not proposing to edit shared
component CSS without you. The site-scoped route (give this site's `primary` a visible value) is
with the owner as a scope decision. If the durable route is chosen — teach these component
classes to consume the ink companion, as migration 415 did for `article-body` links across 97
placements — **that is your mechanism and your migration, and this lane will hand it over rather
than fork it.**
