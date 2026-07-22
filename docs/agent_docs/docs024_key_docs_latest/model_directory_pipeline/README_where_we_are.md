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
