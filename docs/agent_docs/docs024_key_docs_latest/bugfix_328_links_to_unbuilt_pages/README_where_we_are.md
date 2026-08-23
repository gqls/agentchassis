# Where we are — bug 328, dead links to pages that never got built

Plain prose, append-only, newest at the bottom.

## 2026-08-23, late afternoon — the bug is real, and the reason it survives is not what we thought

The complaint is simple. When one page of a site fails to build, every other page that links to
it still ships with the link. The reader clicks it and gets a "page not found". A site with four
good pages and no link to a missing fifth just looks small; the same site with two dead links
looks broken, and that is a judgement customers make about the whole product rather than about
one page.

I checked whether it is still happening before doing anything else. It is. The loanzy.uk home
page is live and healthy, and it carries three links to two pages that have never existed — two
of them to the same missing page. Same on mortgagecalculator.co.uk, where the missing page has
been advertised from six other pages since late July. Across the whole fleet there are 63 open
records of this, on seven different sites, all pointing at 13 pages that could not be built. The
oldest has been live for a fortnight.

Then the part that changes what we should build. The bug file, updated two days ago by another
lane, says the platform already NOTICES these links — it files a record naming the linking page
and quoting the link — and that the records simply sit in a queue nobody reads. If that were
true, the fix would be plumbing: wire the queue to something.

It is not true, and the rows say so. Those records were picked up. A builder was dispatched at
each one and did run. It failed — 48 times with "no sections ready to build" and 10 times with
"content validation failed" — because the record's only instruction is *build the missing page*,
and the missing page is precisely the thing that cannot be built. So the platform detected the
problem correctly, dispatched correctly, tried the one repair it knows, and that repair is the
one repair that cannot work here. Nobody ever told the pages doing the linking.

The other half of the picture is in the code, and it is a one-line gap with a wide reach. There
is already a shared piece of machinery that cleans dead links out of a page just before it
ships — it runs on the build path, both re-render paths, and at the point where content is
saved. It asks "does this link point at a real page?" and it answers that by looking for a row
in the pages table. A page that was planned and never built HAS a row. So to every one of those
four checks, a link to a page that has never existed on the web looks perfectly fine.

We have already solved exactly this problem once, for the header and footer. That fix asks the
harder question — has this page ever actually been served? — and it is careful about two cases
where the honest answer is "don't touch anything": when the lookup fails, and when the site has
not published anything yet, because on a brand-new site "never served" is true of everything and
means nothing. This bug is the same fix for the body of the page, and the hard part is that
middle case: during a first build, pages go out one at a time, so almost every link points at a
page that has not shipped *yet*. Strip those and you publish a site with no internal links at
all, which is a failure we have shipped before and not noticed. So the design has to tell "not
coming" apart from "not here yet", and it has to do it with something we can actually query.

I have handed that design question to a second model with all the evidence, and I will bring the
answer back through the reviewer council before anything is committed.
