# SUMMARY 2026-08-08 — directory-build-handler (`bugs_open/206`)

## What we're trying to do

Two pages on vetcomparison.uk had never worked: `/directory/index.html`,
which is supposed to list the site's vet practices, and `/guides/index.html`,
which is supposed to list its advice guides. Both 404'd. The homepage's own
"Search the directory" link pointed at the dead page — that was the original
complaint (`features_open/021`). The job was to find out why, and build
whatever was actually missing so both pages serve for real.

## Where we've come from

The diagnosis (2026-08-06) found two separate gaps stacked on top of each
other. First, `entity-directory` pages were explicitly listed in the
platform's own builder-dispatch map under a comment naming them a builder
that didn't exist yet — the map *knew* it couldn't handle this page type.
Second, the `directory-listing` component — the piece of a page's plan that's
supposed to say "put the business list here" — pointed at a query name
(`query.directory_entries`) that had never been registered anywhere in the
platform. Zero pages fleet-wide used it. So even a page with a plan couldn't
have rendered a real list.

The fix had two Go pieces: a new query resolver that reads a site's own
directory-export configuration and queries the same business data the
site's JSON export already uses (so the two can never disagree), and a new
action that fills in a page's layout plan only when it has none at all —
built that way specifically so it can never silently overwrite a plan a
human or another process already set. A new agent, `directory-build-handler`,
chains those two pieces together and then hands off to the platform's
existing generic page builder to do the actual write-and-deploy. No new
content-writing logic was needed.

That design went through three rounds of the council review gate before it
was approved. Round 1 wanted the delegation step hardened so it couldn't be
routed to the wrong workflow. Round 2 wanted a real audit of whether other
query resolvers in the same package shared the same "missing config looks
like zero results" ambiguity (they don't — this one is the only resolver
whose data source is itself optional per-site config) — and, while
answering that, this session's account of it caught a real second bug: a
verification check in the migration script that used `!=` against a
possibly-missing JSON key, which in Postgres evaluates to NULL and silently
never fires. That was fixed to use `IS DISTINCT FROM` instead, which does
what was intended. Round 3, with all of that answered and evidenced rather
than just asserted, approved.

## What we've done

Round 3 approved this morning. From there: the two database migrations
that register the query resolver's binding and seed the new agent were
applied by hand and verified through their own built-in checks. The two
work items that had been sitting parked and unbuildable — one for each
page — were re-pointed at the new handler and put back in the ordinary
dispatch queue, the same queue every other page build goes through. No
special manual dispatch was used; that was the entire point of building a
real handler instead of working around the gap by hand.

Two more problems turned up only once real dispatches actually tried to
run the new handler — the kind of defect that never shows up by reading
the code, only by watching it fail. The delegation step's field mapping
had the child agent's *own* field names prefixed with the parent's dot-path
(`input_data.site_id` instead of `site_id`), so the child rejected the call
outright. Fixed by a follow-up migration. Then, past that, the build got
further — through planning, content-writing, and saving — before dying at
the status-update step, because that step needs to know which page it's
updating and the delegation hadn't forwarded the two fields it reads for
that. A second follow-up migration mirrored in the fields the platform's
own normal dispatcher already sends to every other build, so the same gap
can't recur silently for the next consumer of this pattern.

Both pages have now been rebuilt and deployed through the ordinary
pipeline: `directory/index.html` lists 60 real vet practices with genuine
UK postcodes, alphabetically; `guides/index.html` lists exactly the three
real guide pages that exist, plus the real advice-tool call to action — no
placeholder or invented entries in either. The bug file has its closure
evidence written in (it stays in `bugs_open/`, not moved to `bugs_closed/`,
per a standing owner instruction from earlier this month), and the
concept register's entry for this capability has been updated from "built,
not yet live" to "deployed, proven live."

Along the way, a side question got answered for free. The owner asked
whether the platform's own self-healing "improvement loop" would have
caught this bug on its own. The answer turned out to be: it would have
*detected* the symptom (it already had, on other pages, going back to
early August), but its fix dispatch has its own separate bug — it rebuilds
the page that merely *contains* a broken link, not the page the broken
link is pointing *at* — so it would have marked itself successful without
ever fixing anything, and kept re-detecting the same problem forever. That
is now its own filed case (`bugs_open/220`), deliberately kept as a
separate piece of work rather than folded into this one.

## Where we are now

Both pages are live and correct, verified by fetching them directly rather
than trusting a status field. The homepage's directory link now goes
somewhere real. The council-reviewed design is fully deployed with no
outstanding config gaps that are known about. This closes out the original
ask.

Two loose ends are named but deliberately not pursued here: `entity-page`
(individual practice pages) stays unbuilt on purpose, waiting on a separate,
much larger crawl that is only a small fraction done; and `bugs_open/220`,
the wrong-page-rebuild defect in the improvement loop, is real, evidenced,
and reproducible, but is its own lane.

## Where we're going

Nothing further is owed on this specific bug. The pattern it leaves behind
— a delegating agent must satisfy the *whole* input contract of the thing
it calls, every step that reads a field, not just the first one — is worth
watching for elsewhere in the platform, since it is exactly the kind of gap
that only shows up when something actually gets dispatched for real, not
when the code is read. `bugs_open/220` is the next piece of related work,
whenever it's picked up, and is unrelated to this handler beyond having been
found by the same investigation.
