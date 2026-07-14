# Where we are — the self-healing platform, in plain language

*2026-07-14. Written to be read calmly, start to finish, with no code needed.
This follows on from `SUMMARY_where_we_are_2026-07-13.md` (the day the repair
pipeline opened its first pull request). Companion docs, if you want the
detail: `RUNBOOK_diagnosis_fix_loop(10).md` (how it works + every gotcha),
`DESIGN_triage_and_escalation.md` (today's architecture),
`NOTES_running_fixloop(10).md` (the turn-by-turn story),
`HANDOFF_diagnosis_fixloop_2.md` (the fresh-chat entry point).*

---

## The short version

Last week we proved the **repair shop** works: given a diagnosed bug, the
platform can plan a fix, argue it through a reviewer council, write the code
in a cage, prove it compiles, and open a pull request that only you can merge.
What it could *not* yet do was notice its own problems — every case still had
to be found and written up by a human.

**Today that changed. The platform now notices.** Two new watchers went live:

- **Triage** — a sorter that reads every recorded failure across the fleet and
  routes only the genuine code bugs into the repair queue.
- **Silent-check** — an inspector that looks at the websites themselves for
  promises not kept, so that even problems *nobody recorded anywhere* get
  found.

By this afternoon the pipeline had, entirely by itself, found three real
problem patterns — including the very bug we'd been using as our hand-picked
benchmark — and filed each one as a tidy, deduplicated case in the diagnosis
queue. The cases sit there, inert, until a human says "go". Nothing runs on a
schedule; nothing merges itself; nothing has become more autonomous — it has
become more *observant*.

## The idea, in one picture

Think of the platform as three tiers.

1. **The workers** build the websites — pages, images, components.
2. **The immune system** watches the work: checkers that spot problems and
   handlers that apply known remedies. This already existed.
3. **The repair shop** is for problems whose cause is in the *code*: diagnose
   with evidence, plan a fix, review it, implement it in a cage, open a pull
   request. This is what we proved last week.

The missing piece was the corridor between tiers 2 and 3: when a remedy keeps
failing, or a problem exists that no checker ever sees, how does that reach
the repair shop without a human noticing it and writing it up? Today the
corridor opened — with a strict doorkeeper.

## Triage — the doorkeeper

Once fired (by hand, for now), triage reads every failure the platform has
recorded and sorts them into four bins:

- **Genuine code bugs** — a handler is failing with a real error in its own
  logic. These, and only these, are escalated to the diagnosis queue.
- **Operational blips** — timeouts, a pod that died. No code fix exists;
  they're surfaced for the operations layer, never escalated.
- **No evidence** — something failed but recorded no error text. There is
  nothing for a diagnostician to work from, so these are held for a human.
- **Missing capabilities** — "no handler exists for this yet" is a roadmap
  decision, not a bug, so those go to the roadmap list, never the repair shop.

Two guardrails matter as much as the sorting. **Deduplication**: fifty pages
failing the same way become *one* case, always — the repair shop can never be
buried. **A hard cap**: at most three escalations per sweep while trust is
built. And its first live run vindicated the design: roughly half of all
recorded "failures" turned out to be operational noise the filter correctly
kept away from the repair queue, leaving just two genuine code-bug patterns to
escalate.

## Silent-check — the inspector for what nobody reported

The bug that started this whole workstream taught us the uncomfortable lesson:
the worst failures are silent. The darts site had a navigation menu proudly
linking to a guides page that had never been built — and no record anywhere
said so. No failure, no alert, nothing to triage. The page was simply blank,
and only a human clicking the link would ever know.

Silent-check closes that gap. It inspects the sites themselves and asks: is
the platform keeping its promises? Its first check is exactly the darts
signature — *a page linked in a site's navigation that was never built, with
no record anywhere that anyone is on it*. The "no record anywhere" clause is
the point: silent-check deliberately reports only what the immune system
cannot see. If any work item already covers a page — open, failed, or waiting
on a human — silent-check stays out of it.

On its first live run it found the darts problem on **two** sites (the second
one we didn't know about), filed them as one platform-level case — because one
cause fixed once beats the same fix applied site by site — and triage routed
it into the diagnosis queue. Our hand-picked benchmark bug has now been found
by the machine, through the front door.

A second check — pages that are live but serving nothing at all — currently
runs in report-only mode. Some of those blanks may be deliberate (content we
removed on purpose during the marketing audit), so it writes its findings to a
report for your eyes and escalates nothing until you decide it should.

## A neighbouring team made this smaller, in a good way

Another workstream (the empty-sections thread) recently built a completion
gate: a handler can no longer mark certain jobs "done" unless the fix is
verifiably real. We reconciled the two efforts before building. Their gate
prevents one whole flavour of silent failure *at the source*; the platform
already had a "two strikes and it's flagged" rule for recurring problems; so
silent-check was built deliberately narrow — it covers only the class neither
of those can ever see: problems where no work record exists at all. No
duplicated machinery, clean boundaries, both sets of notes cross-reference.

## The afternoon's best moment

Between two of silent-check's sweeps, a different team's work created proper
work items for one affected site's missing pages. On its next sweep,
silent-check noticed those pages were no longer *silent* — someone was
officially on them — and closed its own finding, on its own, with an honest
note saying why. The other site's finding stayed open. That is the bookkeeping
we designed — findings that open only when nobody is looking at a problem and
close themselves when somebody is — demonstrated in production, by accident,
within twenty minutes of going live.

## Why you're still in control

- Both watchers are **deterministic** — ordinary code, no AI judgement — so
  the routing can't hallucinate.
- Both shipped in **preview mode** first; each was flipped live only after a
  human read its preview and agreed with every line.
- Escalated cases are **parked, inert**. Diagnosing them costs money and
  starts nothing by itself; it waits for your go.
- Everything is **fired by hand**. No schedules, no cadence, until you choose
  otherwise.
- And as before: the repair shop cites evidence or abstains, builds before it
  opens a pull request, and **never merges its own work**.

## Honest caveats

- Three cases are found and filed — **none is fixed yet**. The next spend
  decision is yours: dispatch the diagnosis loop on them, or let them wait.
- The darts case is now machine-found, but its *fix* is still the one the
  council has always (rightly) said is an architecture decision for a human.
- The blank-pages check stays report-only until you've reviewed its findings —
  some may be deliberate removals.
- Watchers only catch what they check for. The invariant list will grow one
  careful check at a time; each new one ships in preview mode first.

## Where this leaves the roadmap

The immune system and the repair shop are now one connected loop: **detect →
sort → diagnose → plan → review → implement → gate → your merge.** Still to
build, in order: **close the loop honestly** (after a fix ships, re-verify the
original problem and close or reopen the case on evidence); **one page for
everything** (fold escalations, capability gaps, and verifications into your
daily digest); and later, the wider reviewer council. The standing rule
remains: more awareness before more autonomy.
