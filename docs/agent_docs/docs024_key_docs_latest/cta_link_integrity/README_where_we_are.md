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