# Where we are — the news cards with no styling

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-26

Someone filed a bug yesterday saying that on two of our sites the news cards render as
bare, unstyled text, while on three other sites the same cards look right. They measured
it carefully and said honestly that they had no idea why. That last part turned out to be
the useful bit, because the reason is not where anyone would look.

Here is what is actually going on. Every site has one stylesheet, `styles.css`. That file
is written once, by the design agent, when the site's design is generated. At that moment
the system looks at which components the site uses and pastes in the matching styles for
them. And then it never looks again. Nothing in the platform regenerates that file when
the site changes. So if a site gains a news section six weeks after its design run — which
is exactly what happened — the news markup appears on the page and the styles for it were
never written anywhere. The page is asking for a paint colour that was never mixed.

The dates make it concrete. ai-agent-orchestration.com's stylesheet was written on 2 May.
Its homepage first carried news cards on 21 July. Eighty days apart. relojistas.com's
stylesheet was written on 16 July and its homepage gained news *today*. In both cases the
stylesheet is older than the thing it is supposed to be styling.

I nearly got this wrong, and it is worth saying how. My first explanation was simply "the
stylesheet is older than the markup". Neat, fits both broken sites. Then I checked one of
the *working* sites as a control and it broke the theory: robot-hands.com's stylesheet was
also written before its markup, thirteen hours before, and it is styled perfectly. So the
simple version was wrong. The real answer is a bit more careful — the system also counts
components that are *planned* but not yet built, and robot-hands had news planned before
its design run. The point is that checking a site that *works* is what caught it. If I had
only looked at the two broken ones I would have written a confident, wrong explanation
into a handoff, with dates attached to make it look checked.

The obvious repair — just regenerate those two stylesheets — is a trap. Regenerating a
stylesheet re-runs the whole design pass, and that re-rolls the site's colours. We would
have fixed a news section by redesigning two live customer sites. Not worth it.

So the fix is different in kind. Instead of hoping the site-wide stylesheet happens to
contain what a component needs, the page now carries its component's styles with it: at
the moment a page is assembled, the system looks up the styles for the components actually
on that page and puts them in the page itself. Whatever builds the page also styles it, so
the two can't drift apart. That is the real repair, because it fixes the whole class
rather than the news cards specifically.

Alongside that, you asked for both approaches, so the two news components have also been
given their own built-in styles — which, it turns out, is what 86 of our 94 components
already do. These two were simply stragglers. I made the two mechanisms aware of each
other so they never both apply the same styles twice: if a component already carries its
own, the page-level injection leaves it alone.

One thing I want to flag because it nearly bit me. There is a rule buried in the code
about *where* a style block has to sit inside a component — it has to come before the
component's script, or the system silently throws it away when it saves the component. The
obvious placement, tacking it on at the end, would have lost 3,355 characters of styling
with no error message anywhere and a component row that looked perfectly healthy. I tested
it both ways rather than trusting the reading: right placement keeps 4,566 bytes, wrong
placement keeps 1,190.

**Where this leaves us.** The code fix is written, tested and committed, but it is Go code,
so it does nothing until the next time the images are rebuilt and rolled out. The database
half is applied and live, but it can't show up on any page until those pages are
re-rendered. Per your decision I have not touched either live site and have not re-rendered
anything — so as of right now, those two sites still look exactly as they did this morning.
Nothing has visibly changed and nothing was risked.

I have left the bug **open** rather than closing it, because our rule is that a bug closes
when it is fixed *and live*, and this is fixed but not live. The bug file now carries the
full diagnosis and the exact check to run after the next roll: prove the code is in the
running pod, re-render the two homepages, then measure all four sites — and the two sites
that already worked must come back *unchanged*, which is what proves we added the missing
styling rather than quietly restyling three customers' sites.

The change has also gone to the reviewer council. It was queued behind eleven other jobs
when I submitted, so the verdict will take a while; that is normal and does not mean
anything went wrong.
