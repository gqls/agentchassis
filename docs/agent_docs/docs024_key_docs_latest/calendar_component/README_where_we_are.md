This is the plain-English log for whoever wants to know what's going on with "calendar"
work on the platform, without reading code or SQL.

**2026-09-02.** The owner asked me to find everything we've said about calendars across
the whole project and, if nobody already owned it, to make this the thread that does.

I looked everywhere — old planning documents, closed and open bug reports, the
component library, the database. Here's the honest picture: there was no dedicated
"calendar team" before today. What exists is one specific thing — a building block
called `period-calendar` that lets a page show content organised into named periods
(months, seasons, quarters) as a real ordered list, rather than a wall of prose. It was
built ten days ago as part of a fix for a different, broader problem: a gardening site
had a page headed "what your shed needs, month by month" that only ever delivered four
paragraphs, because none of the page's building blocks could hold a list. That fix built
three new blocks at once — a checklist, a comparison table, and this calendar block —
and the calendar is the only one of the three that's actually been picked up and used
since.

As of today, it's been placed on two sites: the gardening site that motivated it, and —
new, nobody had noticed this yet — a farm insurance blog post about seasonal risks. The
other two blocks (checklist, comparison table) have been offered on every site built
since and chosen zero times. That's a real, open question, though not one this thread
needs to solve today: why is the calendar the one that gets picked?

There's also a more interesting loose thread. On that first gardening site, the system
was asked (informally, in the brief) to cover the gardening year month by month, and
instead of using the one-page calendar block, it built an entire separate page for every
month — seventeen extra pages, wired into the site's navigation. Nobody had planned for
that, and nobody has decided whether it's a good outcome (a fuller, more explorable
calendar) or a worse one (seventeen pages that depend on infrastructure elsewhere in the
pipeline that isn't fully reliable yet, versus one self-contained block that can't break
the same way). That's the first real decision waiting in this thread.

One more thing worth knowing so it doesn't cause confusion later: the word "calendar"
also shows up all over the place for something completely unrelated — a news article on
the darts site about how tournament scheduling has changed over the years, called
"darts-calendar-density". That's a chart, not this component, and has nothing to do with
this thread's work. If "calendar" turns up somewhere in the project and it doesn't ring
a bell against what's written here, it's probably that.
