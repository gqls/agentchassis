# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-15 evening — picking up the "other 285"

The fix the 281 lane built for this bug is committed and approved by the council, but it is
not running yet: the chassis that is live was built six hours before that code was committed.
So until the next release rolls, the same accident can happen a third time. Nothing we can do
about that from here except not delay the roll.

Two things in the bug file turned out to be wrong, and one of them mattered. The file said no
page had ever served the bad markup. One had. When the tool-improver "fixed" the asset formatter
on 14 August, it also chose — at random, because it looked up the shared component and took the
first page it found — a learn article called "AI builders: content first", and re-rendered that
page from the poisoned template. Since then that page has been live with an empty article and a
"Related Downloads" box offering three files that do not exist. I put the original article back
(the platform had archived it automatically when it was overwritten, and the archived copy
matches the import fingerprint byte for byte) and queued a plain re-deploy. I still need to
confirm the live page is clean.

The other wrong thing: the earlier decision not to auto-fix ported tools was justified partly by
"no tool has a plan to judge it by". That was counted in the wrong table — 87 tools do have a
current plan, including 14 of the 63 ported webdesign tools. It does not change the decision (there
is still no safe way to fix one ported page without touching the others), but it does mean the
missing piece is exactly one thing: a fixer that writes to the one page instead of the shared
template. That is what I intend to build next, and whether ported findings should then be routed
to it automatically is your call.

## 2026-08-16 morning — the roll landed, and this bug is closed

The new build is running the fence. I did not take that on trust: I sent the platform a deliberate
"rewrite the shared template" request carrying exactly the bytes it already holds, so that if the
fence had somehow not been there nothing bad would have been written — and the request was refused
in under half a second, with the reason recorded ("this component is on 115 pages across 2 sites").
The learn article that had been serving an empty body and fake download links is back to its real
content on the live site; I checked the page itself, not just the job status.

I also added a permanent check that runs on every build: any piece of code that rewrites a
component template must either ask the fence or say, in writing, why its change is meant to reach
every page. I broke it on purpose two ways to make sure it actually fails, then put it back. It has
gone to the council for review (advisory).

One thing for you to decide, and I have written it up rather than built it: today a problem found on
one of the 63 imported webdesign tools goes to a human, because the only automatic fixer would write
to the shared template. The fix for that is a fixer that writes to the one page instead. It is about
a day's work, but it is only worth building if you want those findings fixed automatically — and I'd
suggest only for tools that already have a plan to be judged against (14 of the 63), starting with
one supervised run on the asset formatter. If you'd rather keep them as human-review items, nothing
needs building. The bug itself is closed either way.

Two small corrections to what you were told yesterday, both now recorded: the "recovery" on 8 August
was actually the poisoned version being archived (the recovery came from the regeneration that
followed), and that regeneration itself triggered 154 page re-renders and 73 attempted full rebuilds
of imported pages, all of which were refused by a different safety check. That is the "or worse"
scenario, and it has already happened once — we were saved by the last guard in the chain.

## 2026-08-16 — the D2(b) decision, spelled out (owner asked)

**What it is.** Today a fault found on one of the 63 imported webdesign tools goes to a person,
because the only automatic fixer writes to the shared template (the write this bug was about; now
refused). D2(b) gives the fixer a second way to write: to the ONE page's own stored HTML, so a fix
about one page reaches one page. Three pieces: (1) Go — `update_component_html` takes an optional
page id and, for a shared component, writes that page's row (archived, lock-honouring, same
"is it structurally worse" check), flips nothing else, reports `scope: instance`; without a page
id it behaves as today (refuse). (2) Config for tool-improver, held until that binary rolls — read
the INSTANCE's HTML as the source (today it is fed the shared wrapper, which is how the "fix" became
a fabricated downloads box), pass the page id, deliver by a plain page re-deploy and never the
current `section_edit` step (that re-renders from template + stub data and would discard the
write — it is exactly how the casualty page was made). (3) Routing — which findings go to it.

**The rule.** Don't ship a mechanism nothing exercises (owner 2026-07-29); new authority on a
shared action is an opt-in field, unsafe default off (owner 2026-08-02 §2). Piece 1 satisfies the
second; piece 3 decides whether the first is broken.

**How this case measures.** The mechanism is well-defined and low-risk, and it turns "refuse"
into "write at the right scope". But routing is a policy choice: the 281 lane sent these to
humans on "no per-instance path" (true; this builds it) and "no PLAN to judge by" (counted in the
wrong table — 87 tools have plans, 14 of the 63 imported ones). What that does NOT change: a
page-scoped write fixes the BLAST RADIUS, not the QUALITY. The improver's only two runs on
imported tools produced garbage. Numbers: 63 imported tools; 14 with a PLAN; 49 with none;
`ported_tool_fix` pile today 0 (no post-roll sweep on webdesign yet, so growth is unmeasured).

**Choices.** A — leave it (nothing to build; right if the 63 will be decomposed anyway — 281's
proposal, blocked on bug 204). B — build it, canary ONCE by hand on asset-formatter (has a PLAN,
a real failing criterion, and is the tool that produced the poison), a person diffs the served
page, then route automatically only for the 14 with a PLAN. ~1 day + council. C — build it,
route nothing automatically; a human promotes an item when they want a fix applied.

**Recommendation: B, staged, with an explicit stop at C if the canary's content is not something
you would ship. If you intend to decompose the 63, choose A and I'll close the design note as
"not needed" so it doesn't read as owed.**
