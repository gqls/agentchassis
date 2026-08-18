# Leopardess Consulting — where we are, 2026-08-18

*A milestone read-out, written to be said out loud. The previous one is
`SUMMARY_where_we_are.md`, dated 2026-07-18 — a month ago, and the account below has
moved far enough that it needed a new file rather than an edit. Every figure here was
checked against the live site or the live database on 2026-08-18.*

---

## What we're trying to do

Run a consultancy website that is true. Everything on it should be something we can
show you: a number that comes from a database we operate, a system you can click
through, a claim with a source behind it. And it should read like a person wrote it,
not like a marketing department.

The site is also the shop window for the platform that builds it. So it has a second
job: it has to demonstrate the thing it is selling. A hand-built page would prove
nothing, which is why every change goes through the framework even when that is slower.

## Where we've come from

Two phases, and they were different problems.

**Through July, the problem was truth.** The original site was fluent and largely
fabricated — invented staff, invented clients, capabilities that did not exist. That
was audited out claim by claim, a voice was settled on, and two automatic checkers were
put in place: one that flags a claim with no evidence behind it, one that flags prose
that reads machine-written.

**Through August, the problem became the platform itself.** The site is now inside the
automated improvement loops, and they rewrite it. In one week in August they built five
useful guide pages and, in the same passes, silently broke the services page in five
ways — six images gone, a claim we had removed in July reinstated, a call-to-action
pointing at a calculator, a link to a page that had never been published. None of it
raised an alert. All of it was found by looking at the served page.

That set the working method for everything since: **check the artefact, never the
status.** A job marked complete is not a repaired page. A green log line is not a
shipped fix.

## What we've done

Since the last read-out, in order:

- **Repaired the services page** and proved the repair survived the next platform
  release — which also confirmed that a platform fix protecting hand-authored content
  is genuinely live.
- **Swept the stale figures off the case-studies page** against a live fact register
  that now re-checks itself daily, and redrew its infographic with no numbers in it at
  all, so it cannot go stale.
- **Rebuilt the sitemap**: it listed 27 pages, the site has 36. The four data-trust
  articles, the vendor checklist tool and two guides were being offered to search
  engines not at all. Every one of the 36 was fetched and confirmed to load before it
  went in.
- **Confirmed a tool defect had already been fixed** and proved it by driving the live
  page in a browser rather than trusting the record.
- **Measured the voice backlog** and stopped, because it is not what it appears to be
  (below).
- **Fixed four misdirected buttons on the home page**, and gave the home page two new
  ways of showing what it knows: a band of headline figures, and a chart of the
  Companies House verification work whose numbers are read from the database at build
  time and stamped with the date they were last checked.

Along the way we found and wrote down several traps that would have cost the next
person a day each — including one in our own page-checking tool, which was reporting
pages as perfect that it had never managed to look at.

## Where we are now

The site is in good order and it is being actively improved by the automated loops,
which is both the point and the risk. Concretely:

- **36 pages live, all returning 200**, all now in the sitemap.
- **The home page leads with evidence** rather than burying it — 22 sites, 2,000+
  verified records, 78 agent definitions, 10,000+ news items — and the Companies House
  chart states its figures from the register rather than from typed-in copy.
- **Every home-page button goes where its words say it goes.**
- **The claims layer is working**: a live fact register, re-verified daily, that the
  copy is written against.

Two things are open and one of them needs a decision.

**The voice queue needs your ruling, not more work.** There are 34 outstanding voice
items covering 210 findings, and 138 of the 145 flagged phrases on the whole site come
from a single rule: the one banning the word "trust". Since that rule was written in
July, the platform built this site an entire content pillar about trust — four
articles, a guide, and a tool called the AI Vendor Trust Checklist. The rule is now
flagging our own product name and the titles of research reports we quote. It was right
about what it aimed at, which was us calling ourselves trustworthy instead of showing
it. It just also catches the word used as a subject we write about. My suggestion is to
narrow it to the self-congratulatory forms and leave the plain noun alone — but it is
your rule, so nineteen pages have not been rewritten against it on my judgement.

**The platform's spend cap stopped everything for seven hours yesterday.** The
Anthropic account hit the limit set on it; the build pipeline politely queued
everything rather than failing, so nothing was lost, but nothing moved either, and
twenty-six jobs were waiting. It is running again. The lasting lesson is how invisible
it was: every dashboard said healthy, the scheduler ticked, the dispatcher reported
success every ninety seconds, and the refusal was buried one level down in the record
of each attempt. That is written up so the next person does not spend an hour hunting a
fault in their own work.

## Where we're going

1. **Your call on the "trust" rule**, which unblocks the voice queue. The remaining
   copy fixes that need no ruling are written and ready to run.
2. **Finish the presentation work on the rest of the site.** The home page now shows
   its evidence properly; the same treatment is worth having on the services and
   case-studies pages, which still bury their numbers in paragraphs.
3. **Watch the loops.** The standing question, unchanged since July, is whether a page
   we have fixed stays fixed. Every repair so far has held through a release, and we
   now check rather than hope. That is the discipline worth keeping.
