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
