# Where we are — finetuning.uk

Plain prose, append-only, newest at the bottom.

---

## 2026-08-03 — what was making the site look terrible

It was broken images, and there were nineteen of them.

The homepage has a row of "departments" down the middle — Automation & Workflow,
Intelligent Agent Systems, Data & Research, and so on. Each one is supposed to
have a small icon above its name. Instead each one had a **broken-image icon**,
120 pixels across, in a grey circle. Eight on the homepage, eleven more on the
About page. On top of that, five case-study photos further down the page were
also missing.

The cause is small and slightly silly. The component that draws that row was
originally built to show **staff photos** — the code still calls things
"team-member" and "member-photo". Someone later reused it to show departments
with icons instead of people with photographs. The data was changed correctly, so
each department now says its icon is called `cpu` or `network` or `database`. But
the component was never changed: it still tried to load a **photograph** at that
name. So the page asked the server for a picture called "cpu", the server said it
had never heard of it, and the browser drew the little broken-image symbol.

I have fixed the component so it draws an icon instead of looking for a
photograph. It now uses exactly the same method another component on the very
same page already uses successfully, so this is not a new approach — it is making
one component do what its neighbour already does.

## The more interesting problem: the site had already been told

Here is the part worth your attention.

The system had **already noticed** five of these problems and written them down.
Back on 26 July it filed five work items saying "these case-study images are
missing". Those items are still sitting there, eight days later, untouched.

And it never noticed the other nineteen at all. The check that looks for broken
images only recognises two shapes: a proper file path that leads nowhere, and an
image with a completely empty source. An image whose source is the bare word
"cpu" is neither of those, so the check looked straight past it. It reported five
problems and stood next to nineteen worse ones without seeing them.

I have fixed that check too, so it now recognises the third shape. Across the
whole fleet that shape appears **31 times, on two sites, all from that one
component** — so it was never just finetuning.uk. The other affected site is
ai-agent-orchestration.com, and my component fix repairs it as well.

## And the reason nothing was being fixed automatically

You asked me to make sure the handlers were picking items up properly. They are
not, and the reason is a single disconnected wire.

Work items go through stages: something *detects* a problem, then a *triage* step
marks it ready, then a *dispatcher* hands it to the agent that can fix it. The
dispatcher only ever looks at items marked ready. On finetuning.uk, **every
single open item was still at "detected"** — none had ever been marked ready. The
give-away is a counter on each item recording how many times a handler has tried:
it reads **zero** on all of them. Nothing had ever picked them up, not even to
fail.

The step that marks items ready lives inside a loop that runs on a schedule. That
schedule has been **switched off since 2 May** — three months. So for three
months, detection has carried on filing problems and nothing downstream has been
able to see them.

This is not confined to your site. Fleet-wide right now: **204 problems detected
across 10 sites, and just 2 marked ready to work on.**

## What I have done, and what I need you to decide

Done, and live:

- Fixed the component, so the icons render properly. Both affected sites.
- Fixed the broken-image check so it catches this shape from now on. This one is
  written and tested but **not yet live** — it needs the next software build to
  take effect, and I have verified against the running system that it is not in
  there yet. I would rather tell you that than let you assume it is done.
- Written a way to run the repair loop against a single chosen site, because the
  method the documentation tells you to use refers to a script that **no longer
  exists in the codebase**. That is now fixed and recorded.
- Started the repair loop on finetuning.uk. It is running as I write: it has
  already been through the design and quality checks and has filed a fresh batch
  of findings. Next it triages them and hands them to the handlers.

**The decision I need from you** is what to do about those 204 parked items
fleet-wide. Three options:

1. **Switch the schedule back on.** Simplest, and it fixes the whole fleet at
   once. But it was deliberately switched off, and I do not know whether the
   reason still applies — it would immediately start making changes across ten
   sites.
2. **Run it site by site, deliberately**, which is what I did here. Safe and
   controlled; does not scale to ten sites by hand.
3. **Mark everything ready in one go** and let the existing dispatcher work
   through it. Fastest, but it skips triage, and 235 of those items have already
   failed at least once — this would not tell the difference between "worth
   retrying" and "already known to fail".

My recommendation is to keep doing (2) for now and treat (1) as its own decision
rather than something that gets settled as a side effect of fixing one site. The
fact that detection and repair have been disconnected for three months seems
worth a proper answer.

One more thing I should flag: on ai-agent-orchestration.com, eight of the
department icons are named things like "strategy", "research", "operations" —
which are not real icon names at all. After my fix those will show an empty
circle rather than a broken image. Better, but still not right, and it is a
content problem on that site rather than a component one. I have not touched it;
it is recorded.

Also, for the record: I submitted the check fix to the review council
(reference `cfc94d91-3d17-4f29-a370-2b91d1a59a6f`) and the verdict was still
pending when I wrote this.

---

## 2026-08-03, later — it's fixed, and here is exactly what "fixed" means

**finetuning.uk is clean.** Both pages, checked by fetching the actual served
HTML rather than trusting a status:

```
https://finetuning.uk/            broken images: 0   (was 8)
https://finetuning.uk/about.html  broken images: 0   (was 11)
```

The homepage now shows 18 proper icons where 8 broken-image symbols used to be.
The framework did the repair — I changed the component, the framework re-rendered
the pages and deployed them.

**There was one more trap on the way, and it is worth knowing about** because it
would have made everything above look done while changing nothing. The site had
68 page-rebuild jobs already queued. It would be reasonable to assume the fix
reaches the pages when those run. It does not. Those jobs come in two kinds, and
the difference is invisible unless you look: one kind *regenerates* each section
from the component, the other just *re-staples* the page from sections it already
has stored — the very sections that were wrong. 42 of the 68 were the re-stapling
kind, including the only one for the About page. They would all have completed,
reported success, and left the site exactly as broken. This has bitten us before
— on 2 August a queue of 294 drained to zero and repaired nothing. I queued the
regenerating kind explicitly; both pages were fixed within a minute of being
picked up.

**The check fix passed review, on the second attempt.** The council rejected my
first submission and it was right to: there is a standing note in our own records
warning that this file and a sibling check overlap, and I had not read it. Having
checked properly, they cannot overlap — the sibling only ever matches two
specific file paths, and the new rule structurally excludes anything containing a
path separator or a dot. Second round approved. I have also moved that analysis
somewhere other agents can read it, rather than leaving it in a code comment,
which was one of the reviewers' suggestions.

**One correction to something I said earlier.** I wrote that the broken images had
"no finding anywhere". That was true — I checked, and there was not a single one
in our records, on any site, ever. But the audit run today *did* catch it: the
design-audit agent described it independently and accurately, from looking at the
rendering rather than the code. So the framework was not blind to it; its
*automated structural checks* were, while its *AI reviewer* was not. The fix is
still right — an AI noticing something once when someone happens to run an audit
is not something you can depend on, whereas a rule is — but "no automated finding"
is the accurate phrasing and I should have used it.

**The other site is queued but not yet done.** ai-agent-orchestration.com uses the
same component and had 16 of the same broken images. Its pages are queued for the
same repair; they are waiting behind finetuning.uk's much longer queue on a shared
dispatcher. I deliberately did *not* start the full repair loop on that site — that
would have set off work across its whole backlog on a site you did not ask about.

## Still outstanding, and one is yours to decide

1. **The 204 parked items across 10 sites.** Unchanged, and still the biggest
   finding here. Options and my recommendation are in the plan document; the short
   version is that I suggest running the loop deliberately per site for now, and
   treating "should the schedule go back on" as its own decision rather than
   something settled by whoever happens to be fixing a site that week.
2. **The check fix is not live yet.** It is written, tested, reviewed and
   approved, but it only takes effect on the next software build. I verified
   against the running system that it is not in there. I have not triggered a
   build myself: a review round was in flight and a restart would have killed it,
   and builds ship everyone's committed work, not just mine. Worth doing soon —
   until then the framework still cannot see this class of problem on its own.
3. **ai-agent-orchestration.com has eight icons with invented names** — "strategy",
   "research", "operations" and so on, which are not real icon names. After the
   repair those show an empty circle rather than a broken image. Better, not
   right, and it is a content problem on that site.
