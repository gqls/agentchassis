> **SUPERSEDED THE SAME EVENING — this summary says "not live", and it went live at 19:09Z and was closed at 19:16Z.** It is kept unedited as the point-in-time record of what we believed at the milestone; the close, its induced-run evidence and the zero-blast-radius target choice are in `bugs_closed/092` and in `NOTES`/`README_where_we_are`. No second summary was written for a seven-minute delta.

# Summary — the page writer and its link constraints (`bugs_open/092`), 2026-07-31

## What we're trying to do

Stop the platform writing internal links to pages that do not exist. When it builds a page,
it is meant to hand the writer the list of pages that actually exist on that site and say
"only link to these". We want that list to be real, to match what the deploy gate will later
accept, and to say something sensible when there is nothing to say.

## Where we've come from

The machinery for this was built a long time ago and has been running on every single build.
It has also been producing an empty list every single time. The step looked for its page
list in four places in the workflow's data; none of those four has ever existed on the
writer's own run. So it produced nothing, and the prompt template — quite correctly — left
out the whole "Internal Linking" section, because there was nothing to put in it. The writer
was then free to guess, and it guessed. Of the fifteen invented link targets the deploy gate
caught in its retained window, all fifteen pointed at pages that exist in no form at all.

This was known. It was written into our own concept register on 12 June as LNK-017, with the
two candidate fixes named. Seven weeks later nothing had changed, and the reason is the
interesting part: **every layer of the failure was silent.** No error, no work item, no
failed build — just a slightly shorter prompt. There was nothing for anyone to notice.

It also got worse in the meantime. A downstream repair used to strip these bad links before
they shipped; on 28 July that repair was shown to have its output discarded at the point of
saving. So for the last few days there has been no backstop at all, and invented links have
been reaching live pages as 404s.

## What we've done

Checked it was still real first — 26 of 26 recent runs, the most recent from this afternoon —
then fixed it at the source.

The step now reads the real page list from the database. There was a genuine choice about
*which* list, because the codebase has three slightly different definitions of "the pages of
a site". We took the one the deploy gate uses, and that is now a single shared definition
used by both, so the two cannot drift apart. The reason matters more than the mechanism: the
gate is what later decides whether a link is broken, and telling the writer one list while
judging it against another produces a writer punished for doing as it was told.

We also fixed the quieter half. An empty list used to produce an empty instruction, which is
to say none at all — the writer that knew least was told least. It now says so explicitly.
And "this site has no pages" no longer looks identical to "I could not find out": those have
opposite fixes, and collapsing them is a large part of why this ran undetected for seven
weeks. The second case now leaves a durable record someone can query tomorrow.

Alongside that: no link address is ever invented from a page name any more, and a 173-line
duplicate implementation that had never been called in its life was deleted — retiring a
standing warning in our notes rather than restating it.

It went through the review council and was approved first time, with six advisory
objections. One of them was right in a way we had talked ourselves out of: we had named the
existing shared code, then copied it anyway with a comment asking the two copies to stay in
step. That is fixed properly now, and pinned by a test.

## Where we are now

Written, tested, reviewed, committed — **and not yet live.** Go changes do nothing until
someone builds and rolls a new chassis image. The bug therefore stays open, which is our own
standing rule: the defect is reproducible until it ships, and a commit is not a deploy.

Nine tests cover it, and we deliberately broke the fix seven different ways to confirm each
test actually catches the thing it claims to. That exercise earned its keep twice: it caught
a claim in a comment that was the exact opposite of the truth, and the fix's own reporting
had a bug where a count could only ever have said "1" no matter how much was left out.

Three of our own mistakes are written up in the fleet's wrong-calls log, because two of them
were the same shape as the bug we were fixing — adding something and not re-running the
checks that read it.

## Where we're going

One thing closes this: the next chassis roll, then a fresh writer run showing a real page
count and a real source, confirmed against the running pod rather than against git or the
image tag. The exact queries and the pod check are in the runbook and in the bug file.

Two things came out of this that are somebody's next job rather than ours. First, this fixes
new pages only — it repairs nothing already deployed, and there are live sites serving broken
links today; those belong to bugs 071 and 097. Second, while auditing the shared site-id
helper the council asked us to check, we found two other callers that fail quietly, one of
which reports success while writing nothing to the link registry — and that registry has
never held a single row. We stopped short of concluding why, because its agent has no runs in
the window we can see and "it's broken" and "it never runs" look identical from here. That
measurement has been handed to the thread that owns the table.
