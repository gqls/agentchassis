# The website fleet that repairs itself

**agentchassis** is a platform where AI agents plan, build, and operate a
fleet of content websites — interactive tools, guides, games, news feeds.
Humans make the taste decisions (approve a logo, set a brand mood, reject a
layout). Agents do everything between: plan the pages, write the content,
generate the imagery, deploy to git-backed hosting — and continuously
**inspect their own output and file the fixes themselves**.

## One week, in numbers (8–10 July 2026)

- **9** deployed sites, **144** cooperating agent definitions
- **1,187** work items processed on the test site alone (1,031 complete)
- **33-page site rebuilt from scratch** — live news feed (9 sources),
  5 interactive engineering calculators — largely overnight, unattended
- **14 bespoke hero images** generated overnight: ~90 seconds each,
  prompt → image model → optimisation → git commit
- **14 corrupted components** discovered fleet-wide; **10 healed within a
  day** — the first one detected, queued, regenerated, and verified with
  **zero human involvement**; the rest fix themselves on schedule
- **~1 hour** to restyle an entire site (wrong layout → dark engineering
  portal, CSS re-rendered and deployed)
- **2** new permanent self-healing checks now policing the whole fleet

## Why it matters

1. **Maintenance is a loop, not a backlog.** The corrupted components were
   the platform's own historical damage. It found them, fixed them, and
   installed a guard so the class of bug can never silently return.
2. **Humans only at the taste layer.** Every mechanical step between two
   human judgements is agent work — including diagnosing why something
   looks wrong.
3. **The system tells the truth about itself.** One fix this week made
   failed jobs *report as failed* instead of quietly claiming success —
   because a fleet you can trust starts with honest telemetry.

*Details, evidence, and the full engineering story:
`SHOWCASE_imagery_workstream.md` and `SHOWCASE_technical_architecture.md`
in the same folder. Numbers verified against production, 2026-07-10.*
