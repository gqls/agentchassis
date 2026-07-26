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

---

## 2026-07-26, later — it is live, and the bug is closed

A fresh chassis build went out (v1.0.1171) with the code fix in it. I checked it was
genuinely in the running container rather than just in git — grepped the binary for a
phrase that only exists because of this change, plus a phrase that existed before it and a
phrase that exists nowhere, so a false pass would have shown up. All three behaved.

Then I re-rendered the two broken homepages and measured all four sites. Both broken sites
now render styled news cards. The two that already worked are unchanged — I checked that
properly rather than assuming it: the styling now sitting inside their pages is a
byte-for-byte match with what their stylesheet already served. So we added the missing
paint; we did not repaint anyone.

That was the whole test, and it passed, so **the bug is closed** and the file has moved to
the closed folder.

Two things I want to be straight about.

**The first is that the half of the fix that actually did the work was the simpler half.**
The two news components now carry their own styling, and that is what showed up on the
pages. The cleverer half — the page assembler collecting styles as it builds — deliberately
did nothing, because it checks whether a component already carries its own styles and
stands aside if so. That is exactly what it was built to do, but it does mean it is
installed rather than proven in the wild. So I proved it a different way: I ran its
database query by hand, twice — once against the world as it is (it correctly found nothing
to add) and once against a simulated page from before today's change (it correctly produced
the 3,355 characters that were missing for eighty days). Both directions behave. It is a
safety net, and the net is real, but nothing has fallen into it yet.

**The second is a mistake I made and got away with.** I intended to fire the light kind of
page refresh, the one that just restitches what is already stored. The script's
documentation says you get that by passing no reason. I passed an empty reason, which is
not the same thing — in shell, an empty argument counts as absent, so it quietly selected
the *heavy* refresh instead, which re-renders every section on the page. On two live
customer sites.

Nothing was harmed: that script has a guard that refuses to run when a page has missing
content, precisely because the heavy path would otherwise regenerate the text, and both
pages were clean. But that was luck, not care. Worse, the heavier path is what carried the
fix onto the page — so the mistake was rewarded, which is the sort that repeats. I have
written it up in the standing wrong-calls log and put the warning in the runbook next to
the command.

One loose end: the change was sent to the reviewer council hours ago and no verdict ever
came back. There is no trace of it in the system at all, which looks like the request was
lost rather than merely slow. I have not resubmitted, because "no rows" is also what a
queued job looks like and our own guidance says not to spend the credits twice on that
evidence. The practical effect is that both commits are honestly marked as un-reviewed
rather than carrying a review stamp they did not earn.
