# README — where we are: model directory pipeline

Owner's running log. Plain prose, append-only, newest at the bottom. Add
dated corrections below rather than editing earlier entries.

---

**2026-07-22 — the ask, and why it grew**

You asked for a big, prominent, frequently-updated section on
ai-agent-orchestration.com listing open and closed AI models — what they do,
what they cost (with a source, not a guess), who runs them, where to find
them and how to use them, plus links out to our own wrapper tools later on
finetuning.uk, and video links if we can manage it. A second section was to
follow: which companies are actually using AI agents, what they claim for
ROI and whether they can back it up, how far they've rolled it out, plus
which agent-communication protocols (like MCP) are catching on.

Partway through, you widened it: rather than a one-off build for this one
site, you want this to be something any of our sites can switch on via its
site spec, and have the fleet build and keep updated automatically. That's
the right call — it turns this from a single content page into a real
product capability, and it happens to be very buildable, because we already
have almost this exact machinery for the news feed each site runs.

I dug into the actual code and database (not just design docs) before
committing to an approach. The news feed's opt-in flag, its scheduled
discovery job, and its "create the page automatically" pipeline are all real
and live, and the model directory can copy that shape closely. Separately,
there's a citation-checking system already built (though not yet switched
on) for exactly the kind of claim we need here — "this model costs $X, and
here's the page that says so, re-checked periodically so a stale number
doesn't sit there quietly wrong." That's the piece that makes "cited cost"
actually mean something rather than being a one-time LLM guess.

I designed it as two small database tables — one for the "things" (a model,
later a company or a protocol), one for individually-cited facts about each
thing (a price, a licence, a claimed ROI figure) — so that when the second
section (company AI-adoption tracking) comes along, it's the same two tables
with a new category, not a rebuild.

You approved the plan. Build order, as agreed: the automated pipeline first
(no hand-typed launch content), model directory before the adoption tracker.
Starting with the database schema now; the research agent and the
page-publishing side come next, each needing a code build and roll before
they're live.

Three things I still need your call on, flagged in the plan doc: whether the
model directory gets its own page or lives as a section on an existing page;
how often we should double-check prices specifically versus more stable
facts; and whether to switch this on for finetuning.uk immediately alongside
ai-agent-orchestration.com, or just the one site for now.

**2026-07-22, later the same day — all the code and schema is written,
tested and committed. Nothing is running yet.**

I built the whole thing in four pieces and committed each one separately as
it became solid, the way we're meant to work in this shared repo:

1. The database tables that hold the directory itself — one for each thing
   (a model, later a company or a protocol), one for every individual cited
   fact about it (a price, a licence type), each fact carrying its source
   URL and the exact quote that proves it. Applied to the live database
   already; I tested that it correctly refuses a duplicate fact for the same
   thing, and rolled that test back so nothing fake is sitting in there.

2. The actual researcher — code that takes a batch of claims a research
   agent has found, re-fetches every cited web page itself, and only keeps
   the ones where the quote genuinely appears on the page. Anything that
   fails goes into a queue for a person to look at rather than being quietly
   dropped or quietly trusted. There's also a separate daily check that goes
   back over everything already published and re-verifies it, so if a
   source changes or a page disappears, we notice.

3. The part that actually turns that data into something a webpage can
   show — both a version baked into the page's HTML (so it works with no
   JavaScript and is visible to search engines) and a live JSON file the
   page quietly refreshes from in the background, so updates don't need a
   full page rebuild.

4. The automatic "does this site need one of these yet" checks, so any site
   can just flip a switch in its configuration and the system will notice,
   build the section, and keep it fed — no separate one-off work for each
   new site.

Along the way I caught two things worth mentioning, because they're exactly
the kind of mistake that looks fine until someone actually tries to use it.
First, I'd initially wired the "run this on a schedule" piece to send its
message to a queue nothing was listening on — it would have looked like it
worked (no error) while doing nothing at all. I found that by checking how
an existing, working feature does the same thing, rather than trusting my
first guess. Second, I noticed the news section's existing on-page script
builds its content in a way that's vulnerable to malicious text if a source
ever contains it — I didn't fix that (not what I was asked to do), but I
made sure I didn't copy the same weakness into the new one, since the model
directory's content comes from a wider, less curated set of sources.

Everything above is written, tested against a real (but disposable, rolled-
back) copy of the database, and committed to the repo. **None of it is live
yet.** The last step — building a new version of our backend, pushing it,
and rolling it out to the cluster — is the one action in all of this that
actually touches the shared, live system everyone else is working on right
now, so I've stopped short of doing that without asking first.

**2026-07-24 — it works. The directory has its first real, verified data.**

Getting from "deployed" to "working" took two days of peeling back an
onion: the research runs kept dying at the web-page-fetching step, and each
fix revealed another, older problem underneath. First, the fetched results
were too big for our internal messaging system, and the component that knew
this just noted it privately in its own logs and told nobody — so the rest
of the system waited twelve minutes for an answer that could never come.
Then, once the answer was small enough to deliver, it turned out to be
written in a format the receiving side couldn't read — a known mistake
that had been fixed in three sister components but had been copied into
this one. And finally, once the answers were arriving and being read, the
proof-checking step was comparing quotes taken from one rendering of a
pricing table against a differently-rendered copy of the same table, so
every genuine fact failed its own proof. The important thing: at that last
stage the system did exactly what we designed it to do — it refused to
publish anything it couldn't prove, and queued the rejects for a person to
look at. That refusal is what made the last problem easy to find.

All three fixes are live (the middle one went through the review council
and passed first time). Tonight the directory holds ten models and
twenty-two priced facts — real current OpenAI model prices, each one
carrying the exact sentence from the vendor's own pricing page that proves
it, and each one due to be automatically re-checked every day so a stale
price can't sit there looking authoritative. I also found and closed a gap
in my own earlier work: the piece that actually delivers the directory
data onto opted-in websites had been planned but never switched on. It's
on now, and it deliberately does nothing until the website page it feeds
exists — which the fleet's own page-building machinery should create on
its next pass, since the pilot site is opted in and the directory now has
something to show.

Worth saying plainly: the model list the first runs found is real but
narrow — it leaned heavily on one vendor's pricing page. The weekly
research query is written to cast wider; whether it does so in practice is
the first thing to watch next week.

**2026-07-24, evening — it's on the site.** The directory is live at
ai-agent-orchestration.com/model-directory.html — a full page listing the
models with their prices, each figure linking to the page that proves it —
with a teaser section on the homepage and the background data file that
lets the page refresh itself between rebuilds. Getting the page created
surfaced two things worth knowing about the wider fleet: site "health
check" scans don't run anywhere automatically (this site hadn't had one
since early May, which is also why its empty tools page sat unnoticed —
that's fixed and live too), and the queue that hands work to builders can
leave a site waiting hours when it's busy with other jobs (evidence filed
with the team that owns that queue). You ruled: keep the health scans
per-site on demand rather than switching them on fleet-wide.
