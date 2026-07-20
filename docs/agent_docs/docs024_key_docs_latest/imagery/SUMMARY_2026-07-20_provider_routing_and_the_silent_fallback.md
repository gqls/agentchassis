# Image provider routing, and making a silent fallback speak — 2026-07-20

## What we're trying to do

Every image the platform generates goes to one of two providers. One of them
(Gemini, "Banana") renders legible text and can anchor to a site's brand references;
the other (SDXL) cannot do either. Which provider serves a given image is decided by
its *kind* — hero, icon, logo, infographic and so on. The goal is simple to state:
each kind should reach the provider that can actually serve it, sites should be able
to override that choice without a code change, and — the part this phase is really
about — **when the system falls back to the weaker provider because nobody told it
about a kind, that fact must be impossible to miss.**

## Where we've come from

The original bug looked like "hero images use the wrong model". It wasn't. The real
defect was the *mechanism*: provider selection was a hand-written switch whose
catch-all branch quietly chose the weaker provider. A kind nobody had routed was
indistinguishable, at runtime, from one deliberately placed there. That is how two
separate defects shipped — `content_hero`, then `hero` — each found months later by a
human looking at a bad picture, because generation reported success either way. A
gibberish flowchart went out as a client's homepage.

The fix replaced the switch with an enumerable table plus a pure function, put `hero`
(the fleet's largest kind, 84 of 155 planned images) on the capable provider, and made
the choice overridable per site in configuration rather than code.

## What we've done

**Proved the routing fix actually works in production.** Until this morning it was
deployed but never exercised — no image had been generated since the roll. Then
dartsonline generated a hero and seven icons, and every one went to the right
provider, with the adapter's own log showing it making that choice. That closes R1.

**Built the missing half: the fallback now leaves a record, not a log line.** The
reviewers' standing objection was that detection living only in pod logs depends on
someone tailing the right pod at the right moment — and this repo's history says that
is exactly how defects survive for months. It is now a durable row in the errors table
the dashboards already read. The design constraint that made this hard is worth
keeping in view: the image service has no database access (deliberately), and the
component that *does* have one must not second-guess a routing table compiled into a
different binary. So the image service detects against the table it actually ships and
says so in its reply; the chassis, which owns the database, writes it down.

**Costed the change.** Roughly fourteen times more per hero image, but about £4 a
month in absolute terms, with the fleet's entire image bill under £12. Approved on
that basis. The batch option halves it if volume climbs.

## Where we are now

The mechanism is live on both services and verified running on all four containers —
checked by looking inside the binaries for strings the change itself creates, not by
trusting a version tag.

Two things are honestly unfinished. **The council verdict is REVISE, not approved** —
an eighth review round is queued, and no review trailer has been claimed. And the code
reached production ahead of that verdict, because it was sitting uncommitted in the
shared working folder when a sweep commit picked it up. Nothing was lost, but the
sequence was wrong, and the lesson is counter-intuitive enough to be worth stating:
holding work back to avoid shipping it unreviewed is what exposed it. Commit narrowly
and let the commit message carry the review status.

The reviewers earned their cost twice in this phase, and both catches are recorded
because both were things confidence would not have found. One vetoed the first version
outright: it wired the new reporting into plumbing every job in the system crosses,
with no restriction — a foundational change described as a small fix. The second
noticed that a partly-garbled report would keep its readable entries and silently drop
the rest, making a half-broken message look perfectly healthy. That is the same
disease the whole change exists to cure, one level down.

## Where we're going

The next item is the one that now looks most valuable: **nothing in the pipeline reads
the text inside a picture we generate.** The capable model still misspells
occasionally — a map that rendered "REPRETITIVE" is the standing example — and
generation reports success regardless. The shape is a vision pass after generation
that flags misspellings and any number not in the request, routed to human review, so
an image whose text failed can never publish itself. After that: infographic figures
drawn from the audited evidence base, so a picture structurally cannot state an
unverified number.

Resume from `bugs_open/011` §7 and
`HANDOFF_2026-07-20_provider_routing_011.md`.
