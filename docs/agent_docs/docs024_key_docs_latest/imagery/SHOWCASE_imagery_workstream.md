# A fleet of websites that builds and repairs itself — one week, in numbers

*A shareable summary of the agentchassis platform, illustrated with real
figures from the imagery workstream of 8–10 July 2026. Written for readers
who have never seen the project. Refresh the numbers before reusing —
they date quickly by design: the system keeps improving them.*

---

## What this is

agentchassis is a platform where AI agents plan, build, and operate a fleet
of content websites — interactive tools, buyer's guides, games, news feeds,
articles. Humans set direction and approve taste decisions; the agents do
the rest: they plan each site's pages, write the content, generate the
imagery, deploy everything to git-backed hosting, and — the interesting part
— **continuously inspect their own output and fix what's wrong**.

Today: **9 deployed sites**, run by **144 agent definitions** cooperating
over a message bus. Everything an agent does becomes a "work item" in a
queue; discovery checks scan the sites and file new work items when they
find problems; a dispatcher hands them to specialist agents. One test site,
robot-hands.com (an industrial-robotics reference platform), has been through
**1,187 work items, 1,031 completed** — plans, pages, images, fixes — nearly
all without a human touching anything.

## The week's story: making the sites *look* the part

The goal for this workstream: best-in-class visuals — every page with its own
relevant imagery, one consistent brand voice per site, a permanent logo, and
an audit loop that *enforces* the standard instead of hoping for it.

**A full site rebuild, largely overnight.** robot-hands.com was re-planned
from scratch by the planning agent: a 33-page site including a news section
(9 live news sources, zero erroring) and 5 interactive engineering
calculators. The content agents then built the pages and the imagery agents
generated **14 bespoke hero images overnight** (about 90 seconds each,
end-to-end: prompt → image model → storage → optimisation → git commit),
unattended.

**Finding out why the images didn't show — three bugs, one evening.**
The images generated perfectly but pages showed the same placeholder
everywhere, or nothing. The investigation found three distinct causes
(a naming mismatch between what components ask for and what the pipeline
produces; expiring storage URLs used where permanent paths belonged; and —
the deep one — components whose HTML templates had been saved as *rendered
output* years of iterations ago, with no slots left to fill). Result after
the fixes: **16 distinct hero images, each on its own page, zero expiring
URLs, and one remaining empty image slot site-wide** (a feature that isn't
built yet, not a bug).

**The self-healing punchline.** That third bug — corrupted templates —
affected **14 components across 4 different sites**. The platform already
*detected* the problem (its quality scanner had flagged every one) and
already had a *repair* path (an agent that regenerates a component from its
schema). What was missing was a 200-line bridge between the two. Once built:

- The first corrupted component was **detected, queued, regenerated, and
  verified clean with zero human involvement** — the full loop, autonomous.
- **10 of the 14 were healed within a day** of discovering the problem class.
- The remaining 4 fix themselves whenever their sites' next inspection cycle
  runs. Nobody needs to remember they exist.
- The guard is now permanent: any future component saved in this broken state
  gets caught and repaired automatically, fleet-wide.

**A whole-site restyle in about an hour.** The rebuild had left robot-hands
on a "formal brochure" layout — wrong for an engineering tools platform. The
diagnosis: the site's classification lacked the tags the layout-matcher
scores against, and the layout library *already contained* the right answer
(a dark tool-portal layout, itself grown from a previous instance of the same
gap — the library learns). One data fix and one CSS re-render later, the
entire site restyled from light brochure to dark engineering portal, with
its own electric-blue palette preserved.

**Brand consistency as data, not vibes.** Each site now carries a
machine-readable style guide (palette, medium, mood, things to avoid,
approved reference images). Every image generation reads it — with
hard-earned nuance: photographic subjects get the full brand voice, icons
get palette only (a photographic direction on an icon prompt makes the model
paint a photo *around* the icon — learned the hard way), and logos get
nothing at all, because an approved logo is generated once, then **locked**:
the storage layer now refuses to overwrite it, ever.

## Numbers to quote

| Stat | Value |
|---|---|
| Deployed sites in the fleet | 9 |
| Cooperating agent definitions | 144 |
| Work items on the test site alone (all-time / completed) | 1,187 / 1,031 |
| Pages in the rebuilt test site | 33 (incl. 5 interactive calculators, live news) |
| Hero images generated unattended overnight | 14 (~90s each, prompt→git) |
| Distinct per-page heroes after the render fix | 16, zero expiring URLs |
| Corrupted components found fleet-wide / healed within a day | 14 / 10 (rest on autopilot) |
| Human involvement in the first autonomous template repair | none |
| New permanent self-healing checks added this week | 2 |
| Time to restyle an entire site (layout + CSS + redeploy) | ~1 hour |

## Why it's interesting

1. **The system repairs itself — including its own history.** The corrupted
   templates were old self-inflicted damage from an earlier, less careful
   version of the pipeline. The current system found them, fixed them, and
   installed a guard against recurrence. Maintenance is a loop, not a
   backlog.
2. **Humans only at the taste layer.** Approving a logo, choosing "no, not
   the brochure look", setting the brand mood — human. Everything mechanical
   between those decisions — agents.
3. **Institutional memory is documents, not people.** Every diagnosis,
   decision, and fix this week is written into plan/notes/runbook/handoff
   files that any future session (human or agent) picks up cold. The layout
   library remembering its previous gap — and having the answer ready — is
   the same principle in data.
4. **Failures became visible before they became invisible successes.** One
   fix this week was purely epistemic: work items that failed used to report
   as "complete". Now they report as failed. A fleet you can trust starts
   with a fleet that tells the truth about itself.

---

*Sources: RUNNING_NOTES_imagery_best_in_class.md (Turns 1–22),
PLAN_imagery_best_in_class.md, and live production queries, 2026-07-10.*
