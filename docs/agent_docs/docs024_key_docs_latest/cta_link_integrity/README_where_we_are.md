Done. Research, plan, and bug report are filed and committed.

Your four buttons are all genuinely broken — in four different ways

┌─────────────┬─────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│   Button    │    What it does     │                                                         Why                                                         │
├─────────────┼─────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Start       │ goes to             │ Label belongs to a different tool — a Bayesian ranker. It's a frozen source:static fallback that re-applies every   │
│ Ranking     │ /contact.html       │ render, so page content can't override it. Same origin as "Calculate Rankings" and "Try the Bayesian Ranker"        │
│ Free        │                     │ appearing on an LLM cost page.                                                                                      │
├─────────────┼─────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ See How It  │ href="" — nothing   │ Its URL resolves from an unset site_specs key → empty → the template is ungated, so it emits an empty href instead  │
│ Works       │                     │ of no button.                                                                                                       │
├─────────────┼─────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Start the   │ href="#guide-start" │ Hardcoded fragment in the template; no such id exists on any page using it. Dead on 4 pages / 3 sites.              │
│ Guide       │                     │                                                                                                                     │
├─────────────┼─────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Visit the   │ NXDOMAIN / blank    │ Field is source:llm + required:true with nothing to look up. It invented two different hostnames on adjacent pages. │
│ Tool        │ page                │  Thanks for the domain clarification — that it's built from two domains you do own is what makes it undetectable by │
│             │                     │  heuristic.                                                                                                         │
└─────────────┴─────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Root cause: a button's label and its URL are unrelated schema fields, and nothing anywhere expresses "a label implies a destination". Each of the four then
fell in a different check's blind spot, and the blind spots tile the space.

Two findings that outrank the buttons

It's fleet-wide, and measured: 51 dead/suspect controls across 7 of your 11 sites, and 75 of 89 URL-bound CTA anchors in the component library are ungated
(~84%) — violating the platform's own written LNK-005 invariant. That invariant is agreed, documented, and almost entirely unenforced.

The sharper one is a delivery gap, not a detection gap. The platform caught "Start Ranking Free" correctly — right component, right page — on 17 July, two days
before you clicked it. It filed the finding at needs_human_review, which triage never promotes and no handler consumes. Grepping the whole platform for those
item types returns emission sites only, zero consumers. There are 34 such notes open on this site, oldest 13 July.

So adding more checks without building a handler would make the pile bigger, not smaller.

Two places I'd push back on the brief

Don't invoke the experience loop yet. You asked whether it was needed — it's a detection loop, and detection isn't what's missing. Running it now re-discovers
what's already correctly written in 34 unread items. Build the handler first.

Don't ship the build-time check as a blocker straight away. 30 empty hrefs are live right now; a strict gate fails most of the fleet's next rebuild and gets
switched off. Stage it — the repo already learned this with phantom_internal_links.

The highest-leverage fix also isn't the obvious one: replace the hardcoded 6-component ctaFieldNames map with pairs derived from input_schema. Four migrations
have hand-patched that map with the same lesson.

Filed: bugs_open/023, plan/notes/runbook/summary in docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/, pattern in 016b §9, leopardess HANDOFF
updated. Commits db9a4259b and 47a86c61b. Nothing fixed, per your scope.

One question for you: leopardessconsulting.com is yours but serves a blank 114-byte page, and a live button points at it. Redirecting it to the .co.uk would
turn one of these four into a working button for free.

-----
----

**Correction to the summary above (2026-07-19, later the same session).** Two claims in
the table above were superseded within the hour — kept here rather than edited away,
because the wrong turns are the part that can't be rederived:

1. **"built from two domains you do own" is wrong.** The model has no knowledge of the
   owner's domain estate. `leopardessconsulting.com` is just the obvious `.com` variant of
   the site's own name (the owner happening to own it is coincidence — he said so, and he
   was right). `leopardess.contactforsales.com` is a **transform of the real contact email
   in the site's own identity spec**, `leopardess@contactforsales.com`, `@` swapped for a
   dot. Parts true and in-context; only the recombination invented. That makes it
   deterministically checkable (plan step P1.5), and it generalises — 6 sites carry
   `<label>@contactforsales.com` in their current identity spec.

2. **The redirect option is dropped.** `leopardessconsulting.com` → `.co.uk` was offered as
   a free fix for one button; the owner declined it in favour of the root cause. Correct
   call — it would only have made a fabricated URL resolve.

**And the finding that dropping it exposed:** all four broken buttons live in
`page_components` positions 1 and 2 of the two tool pages — `bayesian-ranking-hero-tool`
(wrong component) and `tool-guide-intro` (wrong page type: a guide intro on a tool page).
The genuine tool sits at position 3 and is fine. So "Visit the Tool" is a self-link *by
construction*, which is why `required:true` had nothing to resolve against. The fabricated
external URL then **concealed** the self-link — `LinkScopeExternal` is skipped by every
check, whereas an internal self-link would likely have tripped `misdirected_cta`.

Owner scope decision: **full platform fix** (schema-derived CTA pairing, anchor gating,
work-item handler, site fix, planner cause). In progress.

----

**2026-07-19, later — grounding the fix, and a docs convention change.**

You chose the full platform fix over the redirect, so I stopped writing plans and went and
read the actual code — the earlier map of how CTA checking works came from a research agent,
not from me, and I didn't want to start editing on someone else's reading. Three things came
back different enough to matter.

**The hardcoded list of "components that have buttons" covers about 15% of them.** There is a
six-entry list in the code that decides which components the platform is allowed to repair. I
derived the real answer from the component library instead: **33 component types, 119 button
destination fields.** The list covers five of the thirty-three. So "the platform can detect
this but not fix it" isn't an edge case — it's nearly every button on every site.

**My proposed rule was wrong and testing caught it.** I'd written that a button destination is
"any field ending in `_url` that has a matching label field". Run against the live library,
plain `_url` matching also swallows **32 image fields** — logos and photographs. Had I written
that as specified, the link resolver would have started rewriting image sources. Requiring the
matching label field fixes it cleanly. Cheap to catch now, expensive to catch after a fleet
rebuild.

**The invented-web-address problem is 21 fields, not one.** The field that produced
`leopardess.contactforsales.com` is not unusual: **21 of the 119 destination fields ask an AI
to write a web address**, across 7 components. Every one of them is the same instruction that
produced your broken button — write a URL you have no way of looking up.

**One thing I want to flag rather than just do.** Fixing the derivation alone won't hold.
Fields sourced from site settings get re-resolved on every page render and overwrite whatever
the link resolver wrote — only 6 of the 119 are currently in a state where the resolver wins.
So there's a choice: another migration flipping fields over (the same migration has now been
written four times), or change the render step to stop overwriting fields the resolver owns.
The second is more invasive but retires the recurring migration. I've flagged it in the plan
rather than picking unilaterally.

**And a change to how I keep these docs, at your direction.** CLAUDE.md now describes a
standing **five**, not four, with a cadence attached to each — this file is now an official
part of it rather than something you maintain by hand alone. I'll append to it at every
natural break like this one, so you stop copy-pasting. It's recorded as *your* document:
append-only, never rewritten, corrections go underneath rather than editing your words. That
rule exists because I overwrote it earlier today after wrongly deciding it was a stray file.
The running notes now also explicitly require every misstep, and SUMMARY moves to a
milestone cadence — once or twice a day — so you can talk about progress while it's
happening rather than reading a report at the end.

Nothing is fixed yet. Next step is the shared derivation helper with tests, since three
separate places in the code need to agree on what a button destination is.

----

**2026-07-19, later still — the diagnosis loop came back, and it caught me out.**

You asked me to put the "same migration written four times" question to the diagnosis loop
rather than take my own preference. Good call, because it corrected me.

**The verdict was CONFIRMED** — the mechanism I described is real. When a page is rebuilt,
certain button-destination fields get looked up fresh from the database every time, and that
lookup overwrites whatever the link resolver had worked out. So the resolver can compute the
right answer and have it thrown away on the next render. That is genuinely what happens.

**But I had one of the details wrong, and it matters.** I told you that only fields marked
`renderer` survive, and that `static` fields overwrite the resolver. The loop showed me the
code says otherwise: `static` and `renderer` are handled by the *same* line and both leave the
resolver alone. The fields that actually get overwritten are the ones that pull from site
settings or the page list.

Where I went wrong is worth knowing, because it's the kind of mistake that repeats. I'd
worked out how `static` behaves for button *labels* — where it does re-apply every time, which
is exactly why "Start Ranking Free" is frozen onto your page — and then assumed it behaved the
same way for button *destinations*. It doesn't. The same setting does opposite things
depending on whether a default value is declared alongside it. My original finding about the
frozen label stands; my generalisation from it didn't.

**The practical effect is that the job got smaller and sharper.** I'd told you the problem
covered around 113 fields across 33 components. Measured properly against the correction:
**83 fields across 18 components**. The rest were never at risk.

**One more thing the loop found on its own** — a component in your header still carries
`source: pages.contact`, with a hard-coded fallback to the contact page. That is the actual
fossil of the old "every button on every site points at contact" bug, still sitting in the
schema.

**What it didn't do, by design:** it diagnosed the cause but didn't pick between the two
remedies. That's the correct division — the diagnosis loop tells you what's wrong, the council
gate reviews a proposed fix. So the direction question is still open, but now it's a
better-informed question about a smaller target.

**A trap I nearly walked into.** The diagnosis trigger defaults to running against `main`.
I checked first: `main` is 345 commits behind and carries an early version of the very list
I was asking about — two entries where the working branch has six. Had I taken the default,
the loop would have confidently diagnosed a version of the code where the problem barely
exists. That's now written into the runbook as a standing check.

Next: I'd like your steer on the two remedies, or I can put a proposal through the council
gate and let it arbitrate.

----

**2026-07-19, evening — the council answered, and it told me off. Correctly.**

You asked me to let the council arbitrate rather than take my own preference. It did, and the
answer was better than either of the two options I gave it.

**The verdict was REJECTED** — a hard veto from the guardian seat. That means my plan does not
ship as written. But the useful part isn't the veto, it's the direction underneath it, which
almost every seat agreed on:

**Do both, in order.** Run the migration now as a safe stopgap, *and* build the proper
consolidation — but with its riskiest change staged carefully first, which mine wasn't. Then
retire the migration pattern for good. It was never A or B; it was A now, B properly, then A's
pattern deleted. I framed it as a choice and that framing was the weakest thing about the
submission.

**The error it caught in my plan is one I should have caught myself.** I had two code changes:
a low-risk one and a high-risk one. I gave the *low*-risk one a careful "log what would happen,
change nothing, read the results first" rollout — and let the high-risk one, which I had
personally labelled "the contested edit", go live immediately. Exactly backwards. Two separate
reviewers flagged it, one at high severity, and one of them made the point that stings: the
plan clearly knows how to de-risk a change, and simply didn't do it where it mattered.

**Three other real defects**, none cosmetic:
- My code would have left a button's destination **empty** in a case where the schema marks it
  required — the very failure we're trying to eliminate, applied to up to 83 fields.
- My "smarter rule" for finding buttons doesn't remove the old fragility, it **moves** it: a
  component that names its fields unconventionally still silently disappears from both
  detection and repair. I'd admitted this in the risks section and then proposed nothing to
  address it. The fix is to make it complain loudly instead of failing quietly.
- I specified no way to **verify the change is actually live** in the running system after
  deploy — which is a standing rule here precisely because code changes do nothing until an
  image is rebuilt.

**The one genuine disagreement among reviewers** is worth your attention, because it's your
original question. Five seats said a migration written four times means the *repetition is the
defect* and a fifth just postpones a sixth. The guardian disagreed: each migration only touches
configuration rows, never live code on the hot path, so four cheap migrations beat one
expensive outage. Both are defensible. The compromise everyone landed on — migration now,
proper fix staged behind it — is what I'd now recommend.

**Also worth knowing:** the first attempt at this submission never reached a verdict at all. It
was destroyed by a known open bug where a single over-long reviewer response discards the
entire round, including the other five seats' work. That's now recorded as the third occurrence,
with new evidence about why it's hard to avoid.

Nothing has been built. Next step is a revised plan along the council's lines, which I'd
resubmit against the same trail.
