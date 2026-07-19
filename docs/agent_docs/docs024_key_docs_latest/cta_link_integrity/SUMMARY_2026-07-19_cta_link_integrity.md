# SUMMARY — the broken buttons, in plain terms

**For the owner. 2026-07-19.**

---

## What you found

You clicked four buttons on the Leopardess site and couldn't tell what they were for:
*Start Ranking Free*, *See How It Works*, *Start the Guide*, *Visit the Tool*. You were
right on both counts. They are broken, and they are also mislabelled — which is why they
made no sense.

They all sit on the two tool pages: the LLM cost calculator and the AI agent ROI estimator.

## What each one actually does

**Start Ranking Free** takes you to the contact page. The label belongs to a completely
different tool — a "Bayesian ranker" that lives on another site. The page has borrowed that
tool's front panel and inherited its wording, which is frozen and can't be edited from the
page's own content. That's also where "Calculate Rankings" and "Try the Bayesian Ranker"
come from on a page about LLM costs.

**See How It Works** goes nowhere at all. It has no web address behind it — the link is
literally empty, so clicking it just reloads the page.

**Start the Guide** is meant to jump you down the page to a section called "guide start".
That section doesn't exist. Nothing happens when you click it.

**Visit the Tool** sends you off the site entirely. On one page it goes to
`leopardess.contactforsales.com`, which doesn't exist — you've confirmed you never set that
subdomain up. On the other it goes to your `.com` domain, which is yours but currently
serves a blank page. Either way, a visitor who clicks it leaves your working site for
nothing. The system invented that first address by gluing together two domains you really
own, which is exactly why no automatic check spotted it.

## Why it happened

A button is made of two separate things: the words on it, and the address it points to. The
system treats those as unrelated. **Nothing anywhere checks that a button with words on it
also has somewhere to go.**

Worse, the words usually come from a built-in default that gets re-applied every single
time the page is rebuilt — so they can't be corrected from the page. The address, meanwhile,
might be missing, empty, or (in one case here) *demanded from the AI without giving it
anywhere to look it up*. When you require an answer from something that has no way of
knowing, it makes one up. That's where the fake web address came from. That's not the AI
misbehaving; it's the form it was handed.

## The part that should concern you most

**The system already found one of these, correctly, two days before you did.**

On 17 July it wrote down: *the "Start Ranking Free" button on the ROI estimator points at
the contact page*. Right button, right page, right explanation. It filed that note in a
queue marked "needs a human" — and nothing in the platform ever reads that queue. There are
**34 notes sitting in it for this site**, the oldest from 13 July.

So this isn't only a case of the system failing to notice. It noticed, wrote it down
clearly, and put it somewhere nobody looks. Those are two different problems and they need
two different fixes. Adding more checking without fixing the second one just makes a bigger
pile that nobody reads.

## How widespread it is

This is not a Leopardess problem.

- **51 dead or dud buttons across 7 of your 11 sites.**
- **Around 84% of the buttons in the shared component library** are built in a way that
  produces a dead link when the address is missing — instead of simply not showing a button.
  There is a written rule in the platform saying it should do the latter. It's almost
  universally not followed.
- The dead "Start the Guide" button appears on **four pages across three sites**, including
  robot-hands and finetuning.

## What I propose

Four stages, in the plan. In plain terms:

1. **Teach the system that a button needs a destination**, and check it before the page
   ships. Right now that check is impossible because the two halves of a button aren't
   linked to each other in the data — so the first job is linking them. This is the single
   most valuable change, and it also fixes a related weakness where only six components
   could ever be repaired automatically. After it, all of them can.
2. **Change the components so a missing address means no button**, rather than a dead one.
   Mechanical, and it makes the whole class of problem impossible rather than merely detected.
3. **Build something that actually acts on the findings** — repair where there's a real
   destination, and remove the button where there honestly isn't one. Sending a button
   labelled "Start Ranking Free" to the contact page is what created this mess; the
   honest answer is to not show the button.
4. **Fix the four buttons on Leopardess** — but at the component level, because anything
   fixed directly on the page can be wiped out by the separate re-planning bug (001).

## Two judgement calls I'd flag

**I don't recommend running the experience loop on this yet**, even though you asked whether
it was needed. It's a *detection* loop, and detection isn't what's missing — these buttons
are already correctly described in 34 unread notes. Running it now would add to a pile
nothing drains. Build the handler first, then it becomes useful.

**I don't recommend switching the new check to "block the build" straight away.** There are
30 empty links live right now across 7 sites; a strict check would fail the next rebuild of
most of your fleet. The pattern that works here is to warn first, clear the backlog, then
tighten. A gate that fires on everything gets switched off, and then nothing is checked at all.

## Where things stand

Research and diagnosis are **done and evidenced**. The plan is written. **Nothing is fixed
yet** — that's deliberate, per your instruction that the fix needn't happen in this thread.

- Bug report: `bugs_open/023_HANDOFF_2026-07-19_cta_label_url_pairing_unchecked.md`
- Plan, evidence and commands: `docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/`

One open question for you: **what do you want `leopardessconsulting.com` to do?** It's
yours, it currently serves a blank page, and at least one button on the live site points at
it. Redirecting it to the `.co.uk` would turn one of these four broken buttons into a
working one for free.
