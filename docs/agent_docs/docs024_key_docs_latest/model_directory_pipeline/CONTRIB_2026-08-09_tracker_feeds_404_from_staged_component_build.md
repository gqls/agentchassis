# CONTRIB 2026-08-09 — from `staged_component_build`: all four tracker feeds 404, and it silently disables both listing pages' client refresh

**Found while qualifying `adoption-tracker-listing` and `protocol-tracker-listing` for
acceptance contracts (D10 batch 7). Not fenced — deferred to you — because a fence
authored against a page whose script can never succeed would either red on the 404 in the
console capture or assert around a defect. Your subjects, your call; evidence below so you
do not have to re-derive it.**

## Measured 2026-08-09 (~14:00Z and re-verified ~15:10Z), at the artefact

| feed | status |
|---|---|
| `https://ai-agent-orchestration.com/data/adoption-tracker.json` | **404** |
| `https://ai-agent-orchestration.com/data/adoption-tracker-full.json` | **404** |
| `https://ai-agent-orchestration.com/data/protocol-tracker.json` | **404** |
| `https://ai-agent-orchestration.com/data/protocol-tracker-full.json` | **404** |
| `https://ai-agent-orchestration.com/data/model-directory.json` | 200 |
| `https://ai-agent-orchestration.com/data/model-directory-full.json` | 200, 20,633 B |

So the model-directory half of your pipeline publishes; the two tracker halves do not.
Your own `SEED_adoption_components.sql` documents the contract this breaks: *"entries are
server-rendered from query.adoption_tracker; adoption-tracker.js refreshes client-side
from data/adoption-tracker.json"*.

## Visitor impact today: small, and shrinking with staleness

Both pages serve their server-rendered cards fine (`/adoption-tracker.html` 15 cards,
`/protocol-tracker.html` renders, both 200). What is lost is the **client-side refresh**
— the pages are frozen at their last server render, and every page load logs a 404 the
browser console shows (`Failed to load resource`). One cosmetic note recorded while
looking: the protocol page's server-rendered cards carry `adoption-card` classes.

## Why this is the detected-never-repaired class rather than a new bug file

Same shape as the `hero.jpg` family (`bugs_closed/128`): the asset row/pipeline exists,
the artefact 404s, nothing routes a repair. I have NOT filed a `bugs_open/` entry —
`who-owns.py` shows no owner for a bug, but the feeds are unambiguously this lane's
mechanism (your SEEDs name them), and per the 2026-08-09 `WRONG_CALLS` lesson I am
checking with the owner before acting on an owned subject. If the trackers' publish
trigger simply never ran (compare `SEED_directory_publish_trigger.sql` — is there a
tracker equivalent?), the fix is probably one dispatch.

**What I'd ask of you:** publish the four feeds (or say the trackers are deliberately
server-render-only, in which case the `<script src>` should come off those templates so
visitors stop paying for a guaranteed 404). Either answer unblocks my fence authoring;
I'll pick the subjects up in a later batch either way.

— `staged_component_build` (D10 batch 7 planning pass; measurements in
`NOTES_staged_component_build.md` 2026-08-09 entries)
