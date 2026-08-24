# 387 — three pages report `build_status='deployed'` with a timestamp from two hours ago, and 404 on the live site

**Filed** 2026-08-24 by the `bugs_open/364` lane. Found while curling a page before calling a
claims finding "live damage" — the check found the opposite of what was expected, which is why
it is filed rather than assumed.
**Severity** medium-high. Three pages that the platform believes it published are not reachable.
**Class** status-vs-artefact divergence (CLAUDE.md: *"Trust the rendered artefact, not the status"*).
**Status** OPEN, unowned. First-hand evidence below; root cause NOT established — see §4.

## 1. The evidence, both halves

**The DB says deployed**, read 2026-08-24 ~20:1xZ (`site_id=2a8ebf9c-20a2-4c39-b191-840b012371da`):

| page | status | build_status | deployed_at | components |
|---|---|---|---|---|
| adoption-tracker | active | **deployed** | 2026-08-24 16:27:57Z | 3 |
| protocol-tracker | active | **deployed** | 2026-08-24 16:27:49Z | 3 |
| model-directory | active | **deployed** | 2026-08-24 18:43:19Z | 3 |

**The live site says 404**, curled 2026-08-24 ~20:15Z:

```
https://ai-agent-orchestration.com/                            200
https://ai-agent-orchestration.com/adoption-tracker            404
https://ai-agent-orchestration.com/protocol-tracker            404
https://ai-agent-orchestration.com/model-directory             404
https://ai-agent-orchestration.com/definitely-not-a-real-page-xyz 404   <-- CONTROL
```

**The control is load-bearing and is why this is a finding rather than a guess.** A parked or
catch-all domain returns 200 for every path, which would make a 200 meaningless; here an invented
URL returns 404, so the domain discriminates, and the three 404s are real absences. The root
returning 200 proves the site itself is serving. (Both traps are recorded in memory: *"a parked
domain 200s EVERY path"*, and *"curl the target before calling a queue row live damage"*.)

`model-directory` was stamped deployed at **18:43Z and 404s at 20:15Z** — ~1h32m later, so this is
not a propagation window.

## 2. Why it matters

- `deployed_at` is what every downstream reader trusts to mean "this is public". A page that is
  `deployed` and absent is invisible to exactly the sweeps that would otherwise notice it.
- These three pages are **linked from the site's own navigation and from `bugs_open/364`'s census
  as live content** — they carry stored `rendered_html`, so every DB-side check reads them as fine.
- It silently wastes the whole build: three pages' worth of writer, designer and deploy work,
  repeatedly (`agent_error_log` shows these same three pages refused **40 build attempts** between
  them over 60 days for an unrelated claims defect — `bugs_open/364`).

## 3. A second defect, visible only because of the first

`model-directory`'s stored `hero` contains an **unrendered placeholder**:

> "Every model behind our agents, catalogued in one place. This registry tracks **NNN+** AI agents
> across **NNN+** agent types, organised under the eight departments we build for…"

`NNN+` is a template token that was never substituted. **It is not public today only because the
page 404s** — which is the sole reason this is filed here as an observation rather than as a
live-content incident. Fix the deploy and this ships to the public the same hour. Note also that
`checkPlaceholderPatterns` (`validate_page_content.go`) did not convict it: `NNN` is not in
`placeholderPatterns`, whose entries are bracket forms like `[name`, `[company`
(see `bugs_open/218` for that scan's other failure mode).

## 4. What is NOT established — read this before fixing

**The root cause is unknown and three plausible causes fit the same evidence.** Do not pick one
from this file:

1. the deploy wrote to the repo but the static build/routing never picked the pages up (a slug or
   route-map gap — `bugs_closed/015`'s shape, a mistyped `page_type` orphaning a page);
2. the deploy step reported success without publishing (the `complete`-is-not-proof class);
3. the pages are published under a different path than their `name` implies.

The cheap discriminator nobody has run yet: look in the deploy repo for the three files, and read
the deploy action's own record for the 16:27Z / 18:43Z runs. **`build_status` is the instrument
under suspicion here, so do not verify with it.**

## 5. Relations

- `bugs_open/364` (found it; the three pages are that bug's motivating page types).
- `bugs_open/328` (a page that failed to build is still linked from the pages that did),
  `bugs_open/266` (four producers rebuild/redeploy without reading `page_status`) — adjacent
  status-vs-reality defects, neither the same as this.
- `bugs_open/218` (the placeholder scan's coverage), `bugs_closed/015` (page_type orphaning).
