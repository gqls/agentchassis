# SUMMARY — bugs_open/026, the schema-dialect fail-open (2026-07-22)

*Written to be read aloud. First summary of this workstream; the milestone is: the fix is built,
council-approved, and committed, awaiting an image roll.*

## What we're trying to do
Close the structural half of bug 026. The visible bug was a shared news-listing component that,
on a Spanish site, showed English "Loading latest news…" and rendered a blank main heading. The
*structural* question — the one another thread explicitly left for us — was: why did a field the
component marks **required** render empty, save, deploy and serve, with nothing anywhere noticing?

## Where we've come from
The two visible symptoms were already fixed live by another thread's rebuild of how news pages
render (server-rendered items + a translatable loading string). That left the interesting part.
The root cause turned out **not** to be a validator mishandling empty strings — it was
schema-*dialect* blindness. Component input specs come in two formats, an old JSON-Schema one and
a current one; the platform only ever learned to read the current one. The news component was in
the old format, so two separate parts of the system that should have caught the missing heading —
the part that asks the writer to produce content, and the part that refuses to ship an empty
required field — both simply couldn't read the spec and silently concluded "nothing to do." A
reader that understands one format doesn't fail *safely* on another; it fails *open*: "I can't
read this contract" comes out identical to "there is no contract."

## What we've done
Taught the platform to read the old format too, via one shared reader that projects it onto the
current format. Every part of the system whose blindness could silently break or hide content on a
served page now reads through it: generation, the render and re-render gates, the post-deploy
audit, the image-satisfiability check, and call-to-action link derivation. And — this was the
council's sharpest contribution — we added a loud alarm that fires the moment the old format is
ever seen again, at every one of those points, so a future regression (a config reload, a restored
snapshot, a component-builder emitting the old format) can't be silently absorbed the way this bug
was. Parts whose blindness only produces a wrong metric or metadata, not missing content, we left
alone and said so.

## Where we are now
The fix is written, tested, and **council-approved** (three review rounds; the reviewer that kept
pushing was right each time, and the fix is much stronger for it). It's committed but **not yet
live** — it's Go code, so it does nothing until the next time the service image is rebuilt and
rolled out. Honest caveat, unchanged: no component is in the old format today (we re-checked —
zero of ~174), so this repairs nothing currently broken; it closes a trap door and adds the alarm.
Two loose ends (idea.uk and one AI-orchestration news page still showing a stale/blank heading)
turned out to be a *different* known bug — wrong page type — and were handed to that bug's owner,
not fixed here.

## Where we're going
Bug 026 stays open until the fix is actually live, per our own "fixed *and* live" bar. Because
it's defensive, it doesn't warrant forcing a fleet-wide image roll of its own — it'll ride the
next coordinated build, and then we verify it by hand (deliberately feed it an old-format
component with an empty required field and confirm it's now refused and the alarm fires) and close
the case. The visible half (English placeholder, blank heading) is already fixed live and could be
marked done independently.
