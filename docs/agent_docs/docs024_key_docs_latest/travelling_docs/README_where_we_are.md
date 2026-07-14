P1 (mobile) + P2 (interactions) are live and proven in production. The verification ladder now catches the class of bug that started all this — the economy-simulator behavioural kind that "completeness +
validation passed" never saw.

The deploy gate turned out to be already closed. No rebuild was needed: your v1.0.1114 browser-runner-adapter build already carries P1/P2, because images build from the local working tree — the 13:54
source landed in the 20:08 image regardless of when it was committed. I confirmed that against the pod rather than git: the deployed binary's symbol table contains runInteraction, splitByProfile,
(*chromiumPage).HorizontalOverflow and the criteriaStep/criteriaExpect types.

The live run (correlation af5a4ac5) scored 9 passed / 0 failed / 1 skipped, against T8's P0 baseline of "3 evaluated, 2 skipped". The adapter's own log is explicit per check and profile:

┌───────────────────────────────┬─────────────────────────────────────────────────────────────────────┬─────────────────────────────────────────────┐
│             check             │                               desktop                               │                   mobile                    │
├───────────────────────────────┼─────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
│ curve-switch (P2 interaction) │ passed — "interaction produced the expected result (#tableWrap tr)" │ passed                                      │
├───────────────────────────────┼─────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
│ mobile-fit (P1 overflow)      │ skipped — "not run on profile desktop" ✓ correct                    │ passed — "no horizontal overflow on mobile" │
├───────────────────────────────┼─────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
│ boots / status / console      │ passed                                                              │ passed                                      │
└───────────────────────────────┴─────────────────────────────────────────────────────────────────────┴─────────────────────────────────────────────┘

So the select really fired on the live page, the page's own JavaScript rebuilt the table, and the rows were asserted in the live DOM — on both profiles, with the mobile-only check correctly scoped. Clean
pass, no improve_tool ticket.

Two things I found along the way, both now in the handoff (T14):

A trap worth knowing. Grepping a Go binary for a short string literal proves nothing. Go compiles constants of ≤16 bytes used only in equality comparisons into integer immediates, so page_status_ok (14
bytes) and selector_exists (15) grep as absent from a binary that plainly implements them, while no_horizontal_overflow (22) greps as present. That nearly convinced me the image was broken. The symbol
table is the reliable check; I've corrected the "grep the pod binary" habit in memory.

A real defect — the note lies by omission. The acceptance note reads "all 9 evaluated checks passed (1 skipped at this phase: mobile-fit)", but mobile-fit ran and passed on mobile — only its desktop
instance was skipped, correctly and by design. judge_acceptance_results counts skips per check id instead of per (check, profile), and its Verified: list repeats bare ids with no profile label. A human
reading that note concludes mobile was never checked, which is the exact opposite of the truth — in the one artifact this whole ladder exists to produce.

I'd fix that composer next (label results id@profile, count skips per check-and-profile). It's a small change and it would ride whatever chassis image you build next, so it costs no extra deploy. Shall I?
The alternative fronts are the real-failure flow through tool-improver, scoping the acceptance cooldown to source='tool-acceptance', or P3 screenshots.

I've updated the handoff, but left everything uncommitted for you.

recap: Goal: prove the tool verification ladder catches real behavioural bugs. Mobile and interaction checks are now confirmed live in production (9 passed, 0 failed). Next: your call on whether I fix the
acceptance note's misleading skip labelling. (disable recaps in /config)

---

Done — and the "carry on" turned up something much more interesting than a clean pass.

The composer fix is in (both chassis-side changes ride your next image):
- Results are now labelled id@profile, so the note says mobile-fit@desktop skipped while mobile-fit@mobile passed. failing_checks keeps bare criteria ids (the fixer matches PLAN ids, as Tier 2 does); a
  new failing_instances carries the profile detail. The stale "skipped at this phase" wording is gone. Six unit tests, including regressions pinned to the real payloads.
- I also landed the queued cooldown scoping while I was in there: Tier-2 writes source='tool-acceptance-tier2', and letting its static fails suppress a browser run had it backwards — a static failure is
  exactly when you most want the behavioural tier to look.

Then I went hunting for a real failure and found one — on the first try. There was exactly one tool with criteria and a live page that had never been behaviourally verified: tool-archetype-taster-quiz on
vonc.com. It failed 3 of 7 checks, and the two failures are of opposite kinds:

- boots — a false failure. The PLAN asserts .tool-container (the gamesdesign generator convention). vonc's page-section render path delivers .tool-archetype-taster-quiz-section; .tool-container appears
  zero times on the page. Pure Option-B/143 class: stale criteria.
- mobile-fit — a genuine failure. The page really does overflow at 390px.

I parked the improve_tool ticket rather than let it dispatch, because it bundled both: the fixer would have chased the .tool-container ghost and quite possibly "fixed" the tool by injecting a bogus div to
satisfy a criterion that was never true.

And then the real bug turned out not to be the tool's bug at all. Driving real Chromium, the offending element is div.footer-legal — 506px wide in a 390px viewport — vonc's site footer, and the homepage
overflows identically. It's site-wide chrome. (That's a live bug for the vonc/Spark workstream: every vonc page scrolls sideways on mobile.)

That exposes a genuine design defect: no_horizontal_overflow measures the whole document, but the ticket it raises is tool-scoped. One overflowing footer would spawn an unfixable improve_tool ticket for
every tool on the site, re-created on every run. I've fixed the visible half — the failure now names its culprit (verified live: widest offending element: div.footer-legal (506px)) — but the routing is
your call:

-----
